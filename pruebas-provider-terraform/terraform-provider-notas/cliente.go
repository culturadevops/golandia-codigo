package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// El cliente de notas-api: lo único del provider que sabe de HTTP.
type cliente struct {
	base string
	http *http.Client
}

type nota struct {
	ID        string   `json:"id,omitempty"`
	Titulo    string   `json:"titulo"`
	Texto     string   `json:"texto"`
	Etiquetas []string `json:"etiquetas"`
	Clave     string   `json:"clave,omitempty"`
	Creada    string   `json:"creada,omitempty"`
	Version   int64    `json:"version,omitempty"`
}

var errNoExiste = errors.New("la nota no existe (404)")

func (c *cliente) hacer(ctx context.Context, metodo, ruta string, in, out any) error {
	var cuerpo bytes.Buffer
	if in != nil {
		json.NewEncoder(&cuerpo).Encode(in)
	}
	req, err := http.NewRequestWithContext(ctx, metodo, c.base+ruta, &cuerpo)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return errNoExiste
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s: %s", metodo, ruta, resp.Status)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *cliente) crear(ctx context.Context, n nota) (nota, error) {
	var r nota
	return r, c.hacer(ctx, "POST", "/notas", n, &r)
}

func (c *cliente) leer(ctx context.Context, id string) (nota, error) {
	var r nota
	return r, c.hacer(ctx, "GET", "/notas/"+id, nil, &r)
}

func (c *cliente) cambiar(ctx context.Context, id string, n nota) (nota, error) {
	var r nota
	return r, c.hacer(ctx, "PUT", "/notas/"+id, n, &r)
}

func (c *cliente) borrar(ctx context.Context, id string) error {
	return c.hacer(ctx, "DELETE", "/notas/"+id, nil, nil)
}
