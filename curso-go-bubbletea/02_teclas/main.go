// Lección 2 · Las teclas.
//
// Dos trampas que no dan error:
//   - En la v2, la barra espaciadora ya no se llama " ": se llama "space". Un `case " "` copiado
//     de un tutorial de la v1 compila y no salta nunca.
//   - El terminal está en modo crudo: Ctrl+C ya no es una señal, es una tecla más. Si tu Update
//     no la atiende, tu programa no se puede cerrar.
package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"curso-tui/comun"
)

// marcador marca o desmarca servidores con la barra espaciadora.
type marcador struct {
	marcados int
	tecla    string // la última tecla, tal y como la nombra Bubble Tea
	v1       bool   // true = escrito como en la v1
}

func (m marcador) Init() tea.Cmd { return nil }

func (m marcador) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		m.tecla = k.String()
		if m.v1 {
			switch k.String() {
			case " ": // ⛔ así era en la v1
				m.marcados++
			}
		} else {
			switch k.String() {
			case "space": // ✅ así es en la v2
				m.marcados++
			}
		}
	}
	return m, nil
}

func (m marcador) View() tea.View { return tea.NewView(fmt.Sprintf("marcados: %d", m.marcados)) }

// visor sale con "q", y nada más.
type visor struct{ ctrlc bool }

func (m visor) Init() tea.Cmd { return nil }
func (m visor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "q":
			return m, tea.Quit
		case "ctrl+c":
			if m.ctrlc {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}
func (m visor) View() tea.View { return tea.NewView("logs en vivo... (q para salir)") }

func main() {
	comun.Titulo("2.1 · La barra espaciadora, escrita como en la v1")
	var m tea.Model = marcador{v1: true}
	for i := 0; i < 3; i++ {
		m, _ = m.Update(comun.Tecla("space"))
	}
	fmt.Printf("  3 pulsaciones de espacio -> %s   (Bubble Tea la llama %q)\n",
		m.View().Content, m.(marcador).tecla)

	comun.Titulo("2.2 · La misma, escrita como en la v2")
	m = marcador{}
	for i := 0; i < 3; i++ {
		m, _ = m.Update(comun.Tecla("space"))
	}
	fmt.Printf("  3 pulsaciones de espacio -> %s\n", m.View().Content)

	comun.Titulo("2.3 · Cómo se llama cada tecla")
	for _, t := range []string{"q", "space", "enter", "esc", "down", "ctrl+c"} {
		fmt.Printf("  %-8s -> String() = %q\n", t, comun.Tecla(t).String())
	}

	comun.Titulo("2.4 · La trampa: un programa que no atiende Ctrl+C")
	_, salio := comun.Correr(visor{}, comun.Teclas("ctrl+c", "ctrl+c", "ctrl+c"), 50*time.Millisecond, 2*time.Second)
	fmt.Printf("  3 veces Ctrl+C -> ¿terminó? %v   (hubo que matarlo a los 2 s)\n", salio)

	comun.Titulo("2.5 · Lo correcto: Ctrl+C también sale")
	_, salio = comun.Correr(visor{ctrlc: true}, comun.Teclas("ctrl+c"), 50*time.Millisecond, 2*time.Second)
	fmt.Printf("  1 vez Ctrl+C   -> ¿terminó? %v\n", salio)
}
