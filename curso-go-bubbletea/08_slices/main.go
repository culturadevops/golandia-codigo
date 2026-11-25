// Lección 8 · El modelo es un valor... sus slices no.
//
// Update recibe el modelo por valor y devuelve otro: parece que cada estado es una foto
// independiente. Por eso se guarda «el estado anterior» para deshacer, o para comparar en una
// prueba. La trampa: copiar un struct copia la CABECERA del slice, no sus elementos. Los dos
// modelos apuntan al mismo array. Marcas un servidor en el nuevo y aparece marcado en el viejo.
// No hay error, y el «deshacer» no deshace.
package main

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"curso-tui/comun"
)

type servidor struct {
	nombre  string
	marcado bool
}

type lista struct {
	servidores []servidor
	cursor     int
	copiar     bool // true = copiar el slice antes de modificarlo
}

func (m lista) Init() tea.Cmd { return nil }

func (m lista) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "down":
			m.cursor++
		case "space":
			if m.copiar {
				m.servidores = slices.Clone(m.servidores) // ✅ array propio para el estado nuevo
			}
			m.servidores[m.cursor].marcado = !m.servidores[m.cursor].marcado // ⛔ sin Clone: el array es compartido
		}
	}
	return m, nil
}

func (m lista) View() tea.View {
	var b strings.Builder
	for i, s := range m.servidores {
		c, x := " ", " "
		if i == m.cursor {
			c = ">"
		}
		if s.marcado {
			x = "x"
		}
		fmt.Fprintf(&b, "%s [%s] %s\n", c, x, s.nombre)
	}
	return tea.NewView(b.String())
}

func probar(titulo string, copiar bool) {
	comun.Titulo(titulo)
	var antes tea.Model = lista{servidores: []servidor{{nombre: "web-01"}, {nombre: "web-02"}, {nombre: "db-01"}}, copiar: copiar}
	despues, _ := antes.Update(comun.Tecla("space")) // marca web-01
	comun.Pantalla("despues (marcado web-01)", despues.View().Content)
	comun.Pantalla("antes (para deshacer)", antes.View().Content)
	fmt.Printf("  ¿el estado anterior sigue sin marcar? %v\n", !antes.(lista).servidores[0].marcado)
}

func main() {
	probar("8.1 · La trampa: modificar el slice en sitio", false)
	probar("8.2 · Lo correcto: slices.Clone antes de modificar", true)
}
