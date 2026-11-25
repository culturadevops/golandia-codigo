// Lección 1 · El bucle: mensaje → Update → View.
//
// La trampa: View parece «la función que pinta», así que se le mete trabajo (leer un fichero,
// preguntar a una API, recorrer una lista grande). No da error. Pero View no se llama cuando
// tú quieres: se llama cada vez que el modelo cambia. Medimos cuántas veces.
package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"curso-tui/comun"
)

// contador es el modelo más pequeño posible: un número y nada más.
type contador struct{ n int }

func (m contador) Init() tea.Cmd { return nil }

func (m contador) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "q":
			return m, tea.Quit
		case "up":
			m.n++ // m es una COPIA: el cambio solo existe porque devolvemos m
		}
	}
	return m, nil
}

func (m contador) View() tea.View {
	return tea.NewView(fmt.Sprintf("Contador: %d\n\n↑ suma · q sale\n", m.n))
}

// ── la trampa ─────────────────────────────────────────────────────────────────────────────

var updates, views, lecturas int

// leerServidores hace como que lee un inventario de 200 líneas (un fichero, una API...).
func leerServidores() []string {
	lecturas++
	s := make([]string, 200)
	for i := range s {
		s[i] = fmt.Sprintf("web-%03d", i)
	}
	return s
}

type panelMal struct{ cursor int }

func (m panelMal) Init() tea.Cmd { return nil }
func (m panelMal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updates++
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "q":
			return m, tea.Quit
		case "down":
			m.cursor++
		}
	}
	return m, nil
}
func (m panelMal) View() tea.View {
	views++
	s := leerServidores() // ⛔ trabajo dentro de View
	return tea.NewView(fmt.Sprintf("%d servidores · cursor en %s", len(s), s[m.cursor]))
}

// panelBien lee UNA vez, al crearse, y View solo pinta lo que ya tiene.
type panelBien struct {
	servidores []string
	cursor     int
}

func (m panelBien) Init() tea.Cmd { return nil }
func (m panelBien) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updates++
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "q":
			return m, tea.Quit
		case "down":
			m.cursor++
		}
	}
	return m, nil
}
func (m panelBien) View() tea.View {
	views++
	return tea.NewView(fmt.Sprintf("%d servidores · cursor en %s", len(m.servidores), m.servidores[m.cursor]))
}

func main() {
	comun.Titulo("1.1 · El modelo es un valor: Update devuelve el siguiente")
	var m tea.Model = contador{}
	for _, t := range []string{"up", "up", "up"} {
		m, _ = m.Update(comun.Tecla(t))
	}
	comun.Pantalla("tras pulsar ↑ tres veces", m.View().Content)

	comun.Titulo("1.2 · La trampa: leer datos dentro de View")
	teclas := append(comun.Teclas(strings.Split(strings.Repeat("down ", 40), " ")[:40]...), comun.Tecla("q"))
	comun.Correr(panelMal{}, teclas, 5*time.Millisecond, 10*time.Second)
	fmt.Printf("  41 pulsaciones -> Update: %d veces · View: %d veces · inventario leído: %d veces\n",
		updates, views, lecturas)

	comun.Titulo("1.3 · Lo correcto: leer una vez y que View solo pinte")
	updates, views, lecturas = 0, 0, 0
	comun.Correr(panelBien{servidores: leerServidores()}, teclas, 5*time.Millisecond, 10*time.Second)
	fmt.Printf("  41 pulsaciones -> Update: %d veces · View: %d veces · inventario leído: %d vez\n",
		updates, views, lecturas)
	fmt.Println("  View puede llamarse muchas veces y en cualquier momento: tiene que ser barato.")
}
