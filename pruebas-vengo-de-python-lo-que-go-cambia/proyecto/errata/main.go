// El mismo error de tipeo, en Go.
package main

import "fmt"

func enviarAlerta(m string) { fmt.Println("ALERTA:", m) }

func avisar(discoLleno bool) string {
	if discoLleno {
		enviarAlrta("disco lleno") // ← la errata
	}
	return "todo bien"
}

func main() {
	fmt.Println(avisar(false))
	fmt.Println(avisar(true))
}
