// Leccion 2 — Tipos estaticos y el valor cero: en Go no existe None.
package main

import "fmt"

func main() {
	fmt.Println("== 2.1 · Toda variable nace con un valor ==")
	var puerto int
	var nombre string
	var activo bool
	var etiquetas []string
	fmt.Printf("  int     -> %d\n", puerto)
	fmt.Printf("  string  -> %q\n", nombre)
	fmt.Printf("  bool    -> %v\n", activo)
	fmt.Printf("  slice   -> %v  (es nil: %v)\n", etiquetas, etiquetas == nil)
	fmt.Println("  No hay None para un int. Si no pones valor, vale su CERO.")

	fmt.Println()
	fmt.Println("== 2.2 · LA TRAMPA: el mapa no da KeyError, da el valor cero ==")
	puertos := map[string]int{"web": 443, "db": 5432}
	fmt.Printf("  puertos[\"cache\"]  -> %d\n", puertos["cache"])
	fmt.Println("  En Python esto era un KeyError. Aqui devuelve 0, y 0 parece")
	fmt.Println("  un puerto valido. No hay error: hay un dato falso.")

	fmt.Println()
	fmt.Println("== 2.3 · La forma de preguntar si ESTA: la coma ok ==")
	if p, ok := puertos["cache"]; ok {
		fmt.Printf("  cache -> %d\n", p)
	} else {
		fmt.Println("  cache no esta en el mapa   <- ok == false")
	}
	p, ok := puertos["web"]
	fmt.Printf("  web -> %d, ok=%v\n", p, ok)
}
