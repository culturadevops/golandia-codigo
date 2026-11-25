// revisar: comprueba /salud en una lista de servidores, todos a la vez, con un tiempo máximo
// por servidor. Un solo binario: se copia al servidor y funciona, sin instalar nada.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type resultado struct {
	host string
	ok   bool
	dur  time.Duration
	err  string
}

func revisar(ctx context.Context, cli *http.Client, host string) resultado {
	ini := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://"+host+"/salud", nil)
	resp, err := cli.Do(req)
	if err != nil {
		return resultado{host, false, time.Since(ini), "sin respuesta"}
	}
	resp.Body.Close()
	return resultado{host, resp.StatusCode == 200, time.Since(ini), resp.Status}
}

func main() {
	lista := flag.String("hosts", "hosts.txt", "un servidor por línea")
	maxim := flag.Duration("timeout", 2*time.Second, "tiempo máximo por servidor")
	limite := flag.Int("max", 0, "cuántos a la vez como mucho (0 = todos)")
	enJSON := flag.Bool("json", false, "salida en JSON, para otra herramienta")
	flag.Parse()

	f, err := os.Open(*lista)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var hosts []string
	for s := bufio.NewScanner(f); s.Scan(); {
		hosts = append(hosts, s.Text())
	}

	cli := &http.Client{Timeout: *maxim}
	ini := time.Now()
	res := make([]resultado, len(hosts))
	n := *limite
	if n <= 0 {
		n = len(hosts)
	}
	turno := make(chan struct{}, n) // semáforo: como mucho n peticiones a la vez
	var wg sync.WaitGroup
	for i, h := range hosts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			turno <- struct{}{}
			defer func() { <-turno }()
			res[i] = revisar(context.Background(), cli, h)
		}()
	}
	wg.Wait()

	sort.Slice(res, func(a, b int) bool { return res[a].host < res[b].host })
	if *enJSON {
		type fila struct {
			Host string  `json:"host"`
			OK   bool    `json:"ok"`
			Ms   float64 `json:"ms"`
		}
		var filas []fila
		for _, r := range res {
			filas = append(filas, fila{r.host, r.ok, float64(r.dur.Microseconds()) / 1000})
		}
		json.NewEncoder(os.Stdout).Encode(filas)
	}
	caidos := 0
	for _, r := range res {
		if !r.ok {
			caidos++
			fmt.Fprintf(os.Stderr, "CAÍDO  %s  (%s)\n", r.host, r.err)
		}
	}
	fmt.Fprintf(os.Stderr, "%d servidores · %d bien · %d caídos · %.1f s en total\n",
		len(res), len(res)-caidos, caidos, time.Since(ini).Seconds())
	if caidos > 0 {
		os.Exit(1) // para cron, systemd o el monitoreo: distinto de cero = algo va mal
	}
}
