#!/usr/bin/env bash
# Demo del vídeo 463 «html/template vs text/template» — se EJECUTA y su salida va a SALIDA.md.
# La MISMA plantilla y los MISMOS datos (un comentario y un enlace que escribió un usuario), renderizados con
# text/template y con html/template. Se ve qué escapa cada uno y en qué contexto. Directorio temporal.
set -u
T=$(mktemp -d)
cd "$T"
go mod init demo/plantillas >/dev/null 2>&1

cat > main.go <<'EOF'
package main

import (
	htmltpl "html/template"
	"os"
	texttpl "text/template"
)

const pagina = `<p>{{.Comentario}}</p>
<a href="{{.Web}}">mi web</a>
<script>var nombre = "{{.Nombre}}";</script>
`

type Datos struct{ Comentario, Web, Nombre string }

func main() {
	d := Datos{
		Comentario: `<script>robar(document.cookie)</script>`,
		Web:        `javascript:robar(document.cookie)`,
		Nombre:     `Ana"; robar(document.cookie); var x="`,
	}
	switch os.Args[1] {
	case "text":
		texttpl.Must(texttpl.New("p").Parse(pagina)).Execute(os.Stdout, d)
	case "html":
		htmltpl.Must(htmltpl.New("p").Parse(pagina)).Execute(os.Stdout, d)
	}
}
EOF

echo '$ cat datos   # lo que escribió el usuario'
echo 'Comentario: <script>robar(document.cookie)</script>'
echo 'Web:        javascript:robar(document.cookie)'
echo 'Nombre:     Ana"; robar(document.cookie); var x="'
echo
echo '$ go run . text   # text/template'
go run . text
echo
echo '$ go run . html   # html/template'
go run . html
echo
echo '$ diff <(go run . text) <(go run . html) | grep -c "^>"   # líneas que cambian'
diff <(go run . text) <(go run . html) | grep -c "^>"
cd / && rm -rf "$T"
