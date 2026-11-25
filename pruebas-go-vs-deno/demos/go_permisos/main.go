// Lo mismo en Go: un binario de Go puede hacer todo lo que pueda el usuario que lo ejecuta.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	b, err := os.ReadFile("/etc/passwd")
	if err == nil {
		fmt.Println("leer /etc/passwd: PERMITIDO,", strings.Count(string(b), "\n"), "líneas")
	}
	fmt.Println("leer la variable HOME: PERMITIDO,", os.Getenv("HOME"))
	out, err := exec.Command("sh", "-c", "id -u").Output()
	if err == nil {
		fmt.Println("ejecutar sh: PERMITIDO, uid", strings.TrimSpace(string(out)))
	}
}
