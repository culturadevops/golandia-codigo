package main

import "fmt"

func main() {
	edades := map[string]int{"ana": 34, "recien": 0}

	// Sin comma-ok no puedes distinguir estas dos situaciones:
	fmt.Println("edades[recien] =", edades["recien"])
	fmt.Println("edades[nadie]  =", edades["nadie"])

	for _, k := range []string{"recien", "nadie"} {
		v, ok := edades[k]
		fmt.Printf("%-7s valor=%d existe=%v\n", k, v, ok)
	}

	// El idiom que esto habilita: contar sin inicializar nada.
	cuenta := map[rune]int{}
	for _, r := range "abracadabra" {
		cuenta[r]++ // el valor cero hace que esto funcione a la primera
	}
	fmt.Println("a aparece", cuenta['a'], "veces")
}
