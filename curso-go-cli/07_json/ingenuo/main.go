// Lección 7, versión ingenua: JSON «directo de la struct». Compila y jq lo lee... casi siempre.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"curso-cli/comun"
)

type item struct {
	URL      string        `json:"url"`
	OK       bool          `json:"ok"`
	Latencia time.Duration `json:"latencia"` // ⛔ time.Duration sale en nanosegundos, sin unidad
}

func main() {
	soloCaidos := flag.Bool("solo-caidos", false, "muestra solo los caídos")
	avisar := flag.Bool("avisar", false, "avisa si un endpoint va lento")
	flag.Parse()

	var out []item // ⛔ si no se añade nada, json.Marshal da null, no []
	for _, u := range flag.Args() {
		r := comun.Revisar(context.Background(), u, 10*time.Second)
		if *avisar && r.Latencia > 200*time.Millisecond {
			fmt.Println("aviso: va lento", u) // ⛔ texto suelto en medio del JSON
		}
		if !*soloCaidos || !r.OK() {
			out = append(out, item{URL: r.URL, OK: r.OK(), Latencia: r.Latencia})
		}
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
