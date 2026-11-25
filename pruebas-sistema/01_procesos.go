//go:build ignore

// 1 · Lanzar y CONTROLAR un proceso del sistema desde Go.
// Lo que se demuestra: no es solo "ejecutar un comando" (eso lo hace cualquiera).
// Es que el proceso hijo muere solo cuando vence el plazo, sin hilos ni señales a mano.
package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// --- a) ejecutar y capturar la salida -----------------------------------
	salida, err := exec.Command("uname", "-sr").Output()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("a) el sistema dice: %s\n", strings.TrimSpace(string(salida)))

	// --- b) un proceso que se cuelga, con plazo -----------------------------
	// "sleep 30" no va a terminar nunca dentro de nuestro presupuesto de 1 segundo.
	ctx, cancelar := context.WithTimeout(context.Background(), time.Second)
	defer cancelar()

	empezo := time.Now()
	err = exec.CommandContext(ctx, "sleep", "30").Run()
	tardo := time.Since(empezo)

	fmt.Printf("b) pedi 30 segundos de sleep y volvi en %.1fs\n", tardo.Seconds())
	fmt.Printf("   el error que devuelve: %v\n", err)
	fmt.Printf("   motivo del contexto:   %v\n", ctx.Err())
}
