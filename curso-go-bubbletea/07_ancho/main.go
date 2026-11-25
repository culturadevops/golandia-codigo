// Lección 7 · El ancho de lo que se ve no es len().
//
// La trampa: una tabla de servidores con nombres con tilde. Se alinea con fmt.Printf("%-15s")
// o calculando con len(), como siempre. No hay error: la tabla sale torcida.
//   - len() cuenta BYTES: «é» son 2, «✅» son 3.
//   - %-15s cuenta RUNAS. Y una tilde puede ser UNA runa («é», U+00E9) o DOS («e» + U+0301, el
//     acento suelto que se pega a la letra anterior). Se ven idénticas. La segunda forma (NFD) es
//     la que usa macOS en los nombres de fichero, y es la que llega cuando el inventario sale de
//     un Mac o de un fichero que pasó por uno.
//   - Los iconos (✅ ❌) ocupan 2 columnas siendo 1 runa.
//
// Lo correcto: medir con lipgloss.Width, que cuenta columnas de terminal.
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"curso-tui/comun"
)

// nfd escribe la tilde como letra + acento suelto, como lo guarda macOS.
func nfd(s string) string {
	r := strings.NewReplacer("á", "á", "é", "é", "í", "í", "ó", "ó", "ú", "ú")
	return r.Replace(s)
}

type fila struct{ nombre, estado, latencia string }

var filas = []fila{
	{"web-01", "ok", "12 ms"},
	{nfd("almacén-norte"), "lento", "870 ms"},
	{nfd("facturación"), nfd("caído"), "—"},
	{"caché", "ok", "3 ms"},
}

// rellenar completa hasta n COLUMNAS de pantalla.
func rellenar(s string, n int) string {
	if w := lipgloss.Width(s); w < n {
		return s + strings.Repeat(" ", n-w)
	}
	return s
}

func main() {
	comun.Titulo("7.1 · Tres formas de medir lo mismo")
	fmt.Printf("  %s %6s %6s %9s\n", rellenar("texto", 22), "len()", "runas", "columnas")
	for _, c := range []struct{ nombre, s string }{
		{"web-01", "web-01"},
		{"caché (é de una runa)", "caché"},
		{"caché (e + acento)", nfd("caché")},
		{"✅ (icono)", "✅"},
	} {
		fmt.Printf("  %s %6d %6d %9d\n", rellenar(c.nombre, 22), len(c.s), utf8.RuneCountInString(c.s), lipgloss.Width(c.s))
	}
	fmt.Printf("  ¿«caché» == «caché»? %v   (se ven iguales; no lo son)\n", "caché" == nfd("caché"))

	comun.Titulo("7.2 · La trampa: alinear con %-15s")
	for _, f := range filas {
		fmt.Printf("  |%-15s|%-8s|%8s|\n", f.nombre, f.estado, f.latencia)
	}

	comun.Titulo("7.3 · Lo correcto: rellenar por columnas con lipgloss.Width")
	for _, f := range filas {
		fmt.Printf("  |%s|%s|%8s|\n", rellenar(f.nombre, 15), rellenar(f.estado, 8), f.latencia)
	}

	comun.Titulo("7.4 · Lo mismo, con un estilo de lipgloss")
	nombre := lipgloss.NewStyle().Width(15)
	estado := lipgloss.NewStyle().Width(8)
	for _, f := range filas {
		fmt.Printf("  |%s|%s|%8s|\n", nombre.Render(f.nombre), estado.Render(f.estado), f.latencia)
	}
}
