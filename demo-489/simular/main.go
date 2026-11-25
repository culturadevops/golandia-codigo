// Simula a Discord: firma cada interacción con Ed25519 (con una clave de PRUEBA generada aquí) y la envía al bot.
// En producción la clave privada la tiene Discord y tú solo copias la pública del portal de desarrolladores.
package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func enviar(privada ed25519.PrivateKey, cuerpo string, falsificar bool) {
	marca := strconv.FormatInt(time.Now().Unix(), 10)
	firma := ed25519.Sign(privada, []byte(marca+cuerpo))
	if falsificar {
		firma[0] ^= 0xFF // un bit cambiado: firma inválida
	}
	req, _ := http.NewRequest("POST", "http://127.0.0.1:8080/interacciones", bytes.NewBufferString(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature-Ed25519", hex.EncodeToString(firma))
	req.Header.Set("X-Signature-Timestamp", marca)
	t0 := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Printf("→ %d en %v · %s", resp.StatusCode, time.Since(t0).Round(time.Microsecond), bytes.TrimSpace(b))
	fmt.Println()
}

func main() {
	semilla, _ := hex.DecodeString(os.Getenv("SEMILLA_PRUEBA"))
	privada := ed25519.NewKeyFromSeed(semilla)
	casos := []struct {
		titulo, cuerpo string
		falsa          bool
	}{
		{"PING de Discord (type 1)", `{"type":1}`, false},
		{"misma petición con la firma alterada", `{"type":1}`, true},
		{"/hola", `{"type":2,"data":{"name":"hola"},"member":{"user":{"username":"golandia"}}}`, false},
		{"/dado lados:20", `{"type":2,"data":{"name":"dado","options":[{"name":"lados","type":4,"value":20}]},"member":{"user":{"username":"golandia"}}}`, false},
		{"/dado lados:1", `{"type":2,"data":{"name":"dado","options":[{"name":"lados","type":4,"value":1}]},"member":{"user":{"username":"golandia"}}}`, false},
	}
	for _, c := range casos {
		fmt.Println("#", c.titulo)
		enviar(privada, c.cuerpo, c.falsa)
	}
}
