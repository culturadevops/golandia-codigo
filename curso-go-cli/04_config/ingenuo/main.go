// Lección 4, versión ingenua: dos formas habituales de mezclar flags, entorno y archivo.
// Las dos compilan, ninguna da error, y las dos se equivocan en algún caso.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

const defecto = 10 * time.Second

func main() {
	timeout := flag.Duration("timeout", defecto, "tiempo máximo por endpoint")
	flag.Parse()

	// Lo que dice el archivo y lo que dice el entorno, si dicen algo.
	var archivo struct{ Timeout string }
	if datos, err := os.ReadFile("vigia.json"); err == nil {
		json.Unmarshal(datos, &archivo)
	}
	deArchivo, _ := time.ParseDuration(archivo.Timeout)
	deEntorno, _ := time.ParseDuration(os.Getenv("VIGIA_TIMEOUT"))

	// A) archivo, luego entorno, y al final el flag... que siempre tiene valor.
	a := defecto
	if deArchivo > 0 {
		a = deArchivo
	}
	if deEntorno > 0 {
		a = deEntorno
	}
	a = *timeout

	// B) el «arreglo» de siempre: usar el flag solo si no vale lo de por defecto.
	b := defecto
	if deArchivo > 0 {
		b = deArchivo
	}
	if deEntorno > 0 {
		b = deEntorno
	}
	if *timeout != defecto {
		b = *timeout
	}

	fmt.Printf("A (el flag al final)      timeout=%s\n", a)
	fmt.Printf("B (comparar con defecto)  timeout=%s\n", b)
}
