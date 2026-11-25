package main

import (
	"os"
	"text/template"
)

func main() {
	t := template.Must(template.New("x").Parse("hola, {{.Nombre}}: {{len .Items}} items\n"))
	t.Execute(os.Stdout, map[string]any{"Nombre": "Ana", "Items": []int{1, 2, 3}})
}
