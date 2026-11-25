// concgo: un programa mínimo que usa cgo (llama a una función de C). Sirve para ver qué pasa al compilarlo
// en cruzado para ARM: con CGO_ENABLED=1 hace falta un compilador de C para ARM, y el binario ya no es estático.
package main

/*
int doble(int x) { return 2 * x; }
*/
import "C"

import "fmt"

func main() {
	fmt.Println("doble(21) desde C =", C.doble(21))
}
