package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, entrada := range []string{"42", "cuarenta"} {
		if n, err := strconv.Atoi(entrada); err != nil {
			fmt.Println("no es un número:", entrada)
		} else if n > 18 {
			fmt.Println(n, "es mayor de edad")
		} else {
			fmt.Println(n, "es menor de edad")
		}
	}
}
