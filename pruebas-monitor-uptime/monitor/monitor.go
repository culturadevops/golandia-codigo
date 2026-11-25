// El monitor de verdad: lo que se enseña en el vídeo. Solo librería estándar.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"
)

type Sitio struct {
	Nombre string
	URL    string
}

type Monitor struct {
	Cliente     *http.Client
	Intervalo   time.Duration
	Umbral      int  // fallos seguidos para declarar un sitio caído
	SoloCambios bool // false = el monitor ingenuo: avisa en cada fallo
	Alertar     func(msg string)
	AlRevisar   func(s Sitio, dur time.Duration, err error) // solo para medir en la demo
}

// Revisar hace UNA petición. Cuenta como fallo: no conectar, pasarse del tiempo o un código >= 400.
func (m *Monitor) Revisar(ctx context.Context, s Sitio) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return err
	}
	resp, err := m.Cliente.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// Vigilar es el bucle de UN sitio, en su propia goroutine: un sitio lento no retrasa a los demás.
func (m *Monitor) Vigilar(ctx context.Context, s Sitio) {
	caido, fallos := false, 0
	t := time.NewTicker(m.Intervalo)
	defer t.Stop()
	for {
		ini := time.Now()
		err := m.Revisar(ctx, s)
		if ctx.Err() != nil {
			return
		}
		if m.AlRevisar != nil {
			m.AlRevisar(s, time.Since(ini), err)
		}
		switch {
		case err != nil && !m.SoloCambios:
			m.Alertar(fmt.Sprintf("🔴 %s falló: %s", s.Nombre, motivo(err)))
		case err != nil:
			fallos++
			if !caido && fallos >= m.Umbral {
				caido = true
				m.Alertar(fmt.Sprintf("🔴 %s CAÍDO tras %d fallos seguidos: %s", s.Nombre, fallos, motivo(err)))
			}
		default:
			if caido {
				m.Alertar(fmt.Sprintf("🟢 %s volvió", s.Nombre))
			}
			caido, fallos = false, 0
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// motivo traduce el error a algo legible en el aviso.
func motivo(err error) string {
	var ne interface{ Timeout() bool }
	switch {
	case errors.Is(err, syscall.ECONNREFUSED):
		return "conexión rechazada"
	case errors.As(err, &ne) && ne.Timeout():
		return "no respondió a tiempo"
	default:
		return err.Error()
	}
}

// Discord devuelve un Alertar que publica en un webhook de Discord: un POST con {"content": "..."}.
// La URL sale de una variable de entorno: un webhook es una credencial, no va en el código.
func Discord(url string) func(string) {
	cliente := &http.Client{Timeout: 5 * time.Second}
	return func(msg string) {
		cuerpo, _ := json.Marshal(map[string]string{"content": msg})
		resp, err := cliente.Post(url, "application/json", bytes.NewReader(cuerpo))
		if err != nil {
			log.Printf("no pude avisar: %v", err)
			return
		}
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("Discord respondió %d", resp.StatusCode)
		}
	}
}

func webhookDeEntorno() string {
	u := os.Getenv("DISCORD_WEBHOOK_URL")
	if u == "" {
		log.Fatal("falta DISCORD_WEBHOOK_URL")
	}
	return u
}
