// Un bot de Discord con comandos de barra (/dado, /hola) por HTTP: Discord envía cada interacción a tu URL,
// firmada con Ed25519. Solo librería estándar de Go: net/http, crypto/ed25519, encoding/json.
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
)

// Tipos que define Discord (docs: «Receiving and Responding»).
const (
	tipoPing    = 1 // PING
	tipoComando = 2 // APPLICATION_COMMAND
	respPong    = 1 // PONG
	respMensaje = 4 // CHANNEL_MESSAGE_WITH_SOURCE
)

type interaccion struct {
	Type int `json:"type"`
	Data struct {
		Name    string `json:"name"`
		Options []struct {
			Name  string  `json:"name"`
			Value float64 `json:"value"`
		} `json:"options"`
	} `json:"data"`
	Member struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
	} `json:"member"`
}

func responder(w http.ResponseWriter, tipo int, texto string) {
	w.Header().Set("Content-Type", "application/json")
	cuerpo := map[string]any{"type": tipo}
	if texto != "" {
		cuerpo["data"] = map[string]string{"content": texto}
	}
	json.NewEncoder(w).Encode(cuerpo)
}

func main() {
	clave, err := hex.DecodeString(os.Getenv("DISCORD_PUBLIC_KEY"))
	if err != nil || len(clave) != ed25519.PublicKeySize {
		log.Fatal("falta DISCORD_PUBLIC_KEY (la clave pública de tu aplicación, en hexadecimal)")
	}
	publica := ed25519.PublicKey(clave)

	http.HandleFunc("POST /interacciones", func(w http.ResponseWriter, r *http.Request) {
		cuerpo, _ := io.ReadAll(r.Body)
		firma, _ := hex.DecodeString(r.Header.Get("X-Signature-Ed25519"))
		marca := r.Header.Get("X-Signature-Timestamp")
		// Lo firmado es la marca de tiempo seguida del cuerpo, tal cual llegó.
		if !ed25519.Verify(publica, append([]byte(marca), cuerpo...), firma) {
			http.Error(w, "firma inválida", http.StatusUnauthorized)
			return
		}
		var in interaccion
		if err := json.Unmarshal(cuerpo, &in); err != nil {
			http.Error(w, "json inválido", http.StatusBadRequest)
			return
		}
		switch {
		case in.Type == tipoPing:
			responder(w, respPong, "")
		case in.Type == tipoComando && in.Data.Name == "hola":
			responder(w, respMensaje, fmt.Sprintf("¡Hola, %s! Te respondo desde Go.", in.Member.User.Username))
		case in.Type == tipoComando && in.Data.Name == "dado":
			lados := 6
			for _, o := range in.Data.Options {
				if o.Name == "lados" {
					lados = int(o.Value)
				}
			}
			if lados < 2 {
				responder(w, respMensaje, "Un dado necesita al menos 2 lados.")
				return
			}
			responder(w, respMensaje, fmt.Sprintf("🎲 Salió %d (dado de %d)", rand.IntN(lados)+1, lados))
		default:
			responder(w, respMensaje, "No conozco ese comando.")
		}
	})
	log.Println("bot escuchando en :8080/interacciones")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
