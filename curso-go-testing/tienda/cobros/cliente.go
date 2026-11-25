// Package cobros es el cliente HTTP de la pasarela de pagos externa.
//
//	POST {Base}/cobros  {"cliente":"ana","centavos":1250}  ->  201 {"id":"cob-123"}
//
// Un 5xx se reintenta con espera exponencial (Espera, 2×Espera, 4×Espera...); un 4xx no.
package cobros

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"curso-testing/tienda"
)

type Cliente struct {
	Base     string
	HTTP     *http.Client  // nunca http.DefaultClient: no tiene timeout (lección 4)
	Intentos int           // 4 por defecto
	Espera   time.Duration // 1 s por defecto
}

type ErrHTTP struct {
	Codigo int
	Cuerpo string
}

func (e *ErrHTTP) Error() string { return fmt.Sprintf("pasarela: HTTP %d: %s", e.Codigo, e.Cuerpo) }

func (c *Cliente) Cobrar(ctx context.Context, cliente string, m tienda.Monto) (string, error) {
	intentos, espera := c.Intentos, c.Espera
	if intentos == 0 {
		intentos = 4
	}
	if espera == 0 {
		espera = time.Second
	}
	for i := 0; ; i++ {
		id, err := c.una(ctx, cliente, m)
		if err == nil {
			return id, nil
		}
		e, ok := err.(*ErrHTTP)
		reintentable := !ok || e.Codigo >= 500
		if !reintentable || i+1 >= intentos || ctx.Err() != nil {
			return "", err
		}
		// select, no time.Sleep: si el ctx se cancela durante la espera, se sale YA (lección 8).
		t := time.NewTimer(espera << i)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return "", ctx.Err()
		}
	}
}

func (c *Cliente) una(ctx context.Context, cliente string, m tienda.Monto) (string, error) {
	cuerpo, _ := json.Marshal(map[string]any{"cliente": cliente, "centavos": int64(m)})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Base+"/cobros",
		bytes.NewReader(cuerpo))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	// Leer y cerrar SIEMPRE: si no, la conexión no vuelve al pool (lección 4).
	defer resp.Body.Close()
	datos, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", &ErrHTTP{Codigo: resp.StatusCode, Cuerpo: string(bytes.TrimSpace(datos))}
	}
	var r struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(datos, &r); err != nil || r.ID == "" {
		return "", fmt.Errorf("pasarela: respuesta inválida %q", datos)
	}
	return r.ID, nil
}
