// El mismo tipo equivocado, en Go.
package main

import "fmt"

func siguientePuerto(puerto int) int { return puerto + 1 }

func main() {
	config := map[string]string{"puerto": "8080"} // leído de un archivo: es texto
	fmt.Println("arranca")
	fmt.Println(siguientePuerto(config["puerto"]))
}
