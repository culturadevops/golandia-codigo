// vigia: reinicia los contenedores que Docker marca como «unhealthy».
//
// Habla con la API de Docker por el socket Unix, con net/http y nada más.
// Solo toca contenedores con la etiqueta que se le pasa (-etiqueta), nunca otros.
//
//	vigia -modo eventos  -etiqueta golandia.autoheal=demo
//	vigia -modo sondeo   -etiqueta golandia.autoheal=demo -cada 10s
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const socket = "/var/run/docker.sock"

// La API de Docker es HTTP normal; solo cambia por dónde viaja: un socket Unix en vez de TCP.
var cli = &http.Client{Transport: &http.Transport{
	DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "unix", socket)
	},
}}

// El host da igual (se ignora); la versión de la API va fija para que una actualización de Docker no cambie las respuestas.
const api = "http://docker/v1.44"

var inicio = time.Now()

func logf(format string, a ...any) {
	fmt.Printf("%7.1fs  %s\n", time.Since(inicio).Seconds(), fmt.Sprintf(format, a...))
}

type contenedor struct {
	ID     string            `json:"Id"`
	Names  []string          `json:"Names"`
	Labels map[string]string `json:"Labels"`
}

func filtros(f map[string][]string) string {
	b, _ := json.Marshal(f)
	return url.QueryEscape(string(b))
}

// enfermos pide a Docker los contenedores con la etiqueta y health=unhealthy.
func enfermos(ctx context.Context, etiqueta string) ([]contenedor, error) {
	u := api + "/containers/json?filters=" + filtros(map[string][]string{"label": {etiqueta}, "health": {"unhealthy"}})
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("docker respondió %d: %s", resp.StatusCode, b)
	}
	var cs []contenedor
	return cs, json.NewDecoder(resp.Body).Decode(&cs)
}

// reiniciar = POST /containers/{id}/restart?t=N : SIGTERM, N segundos de gracia, SIGKILL, y arranca otra vez.
func reiniciar(ctx context.Context, id string, gracia int) error {
	u := fmt.Sprintf("%s/containers/%s/restart?t=%d", api, id, gracia)
	req, _ := http.NewRequestWithContext(ctx, "POST", u, nil)
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("docker respondió %d: %s", resp.StatusCode, b)
	}
	return nil
}

// freno: como mucho `max` reinicios por contenedor dentro de `ventana`. Si un contenedor se enferma
// cada vez que arranca, reiniciarlo en bucle no arregla nada: se deja y se avisa.
type freno struct {
	max     int
	ventana time.Duration
	hist    map[string][]time.Time
}

func (f *freno) permite(id string, ahora time.Time) bool {
	var recientes []time.Time
	for _, t := range f.hist[id] {
		if ahora.Sub(t) < f.ventana {
			recientes = append(recientes, t)
		}
	}
	f.hist[id] = recientes
	if len(recientes) >= f.max {
		return false
	}
	f.hist[id] = append(recientes, ahora)
	return true
}

func curar(ctx context.Context, id, nombre string, fr *freno, gracia int, avisados map[string]bool) {
	if !fr.permite(id, time.Now()) {
		if !avisados[id] {
			logf("⛔ %s: %d reinicios en %s y sigue enfermo · lo dejo y aviso", nombre, fr.max, fr.ventana)
			avisados[id] = true
		}
		return
	}
	logf("reinicio %s", nombre)
	t := time.Now()
	if err := reiniciar(ctx, id, gracia); err != nil {
		logf("error al reiniciar %s: %v", nombre, err)
		return
	}
	logf("%s reiniciado (la llamada tardó %.1fs)", nombre, time.Since(t).Seconds())
}

// sondeo: cada `cada`, pregunta quién está enfermo.
func sondeo(ctx context.Context, etiqueta string, cada time.Duration, fr *freno, gracia int) error {
	avisados := map[string]bool{}
	tick := time.NewTicker(cada)
	defer tick.Stop()
	for {
		cs, err := enfermos(ctx, etiqueta)
		if err != nil {
			return err
		}
		for _, c := range cs {
			curar(ctx, c.ID, strings.TrimPrefix(c.Names[0], "/"), fr, gracia, avisados)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
}

type evento struct {
	Type   string `json:"Type"`
	Action string `json:"Action"`
	Actor  struct {
		ID         string            `json:"ID"`
		Attributes map[string]string `json:"Attributes"`
	} `json:"Actor"`
}

// eventos: Docker empuja cada cambio de salud en cuanto ocurre (GET /events es un stream de JSON, uno por línea).
func eventos(ctx context.Context, etiqueta string, fr *freno, gracia int) error {
	avisados := map[string]bool{}
	// Los que ya estaban enfermos antes de empezar a escuchar no generan evento nuevo: se miran una vez al arrancar.
	cs, err := enfermos(ctx, etiqueta)
	if err != nil {
		return err
	}
	for _, c := range cs {
		curar(ctx, c.ID, strings.TrimPrefix(c.Names[0], "/"), fr, gracia, avisados)
	}
	u := api + "/events?filters=" + filtros(map[string][]string{"type": {"container"}, "event": {"health_status"}, "label": {etiqueta}})
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		var e evento
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			continue
		}
		nombre := e.Actor.Attributes["name"]
		logf("evento: %s · %s", nombre, e.Action)
		if e.Action == "health_status: unhealthy" {
			curar(ctx, e.Actor.ID, nombre, fr, gracia, avisados)
		}
	}
	if ctx.Err() != nil {
		return nil
	}
	return sc.Err()
}

func main() {
	modo := flag.String("modo", "eventos", "eventos | sondeo")
	etiqueta := flag.String("etiqueta", "", "solo contenedores con esta etiqueta (clave=valor)")
	cada := flag.Duration("cada", 10*time.Second, "intervalo del sondeo")
	gracia := flag.Int("gracia", 2, "segundos entre SIGTERM y SIGKILL al reiniciar")
	max := flag.Int("max", 3, "reinicios máximos por contenedor en la ventana")
	ventana := flag.Duration("ventana", 2*time.Minute, "ventana del freno")
	flag.Parse()
	if *etiqueta == "" {
		fmt.Fprintln(os.Stderr, "falta -etiqueta: el vigía nunca toca contenedores sin etiqueta")
		os.Exit(2)
	}
	// Apagado limpio: Ctrl+C o el SIGTERM de docker stop cancelan el contexto y cortan el stream de eventos.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	fr := &freno{max: *max, ventana: *ventana, hist: map[string][]time.Time{}}
	logf("vigía en modo %s · etiqueta %s", *modo, *etiqueta)
	var err error
	if *modo == "sondeo" {
		err = sondeo(ctx, *etiqueta, *cada, fr, *gracia)
	} else {
		err = eventos(ctx, *etiqueta, fr, *gracia)
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	logf("vigía apagado")
}
