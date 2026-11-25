// Lección 6 · Componer modelos: un modelo dentro de otro.
//
// Los componentes de bubbles (spinner, textinput, list...) son modelos con su propio Update.
// Tu modelo los guarda dentro y es TU Update el que tiene que pasarles los mensajes.
//
// Dos trampas, ninguna da error:
//   - No reenviar los mensajes: el spinner se queda quieto y la caja de texto no escribe.
//   - Reenviarlos a una caja de texto sin Focus(): recibe las teclas y las ignora.
package main

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"curso-tui/comun"
)

type finMsg struct{}

type buscador struct {
	reenviar bool
	enfocar  bool
	spin     spinner.Model
	caja     textinput.Model
	frames   map[string]bool // fotogramas distintos que llegó a pintar el spinner
}

func nuevo(reenviar, enfocar bool) buscador {
	b := buscador{reenviar: reenviar, enfocar: enfocar, spin: spinner.New(spinner.WithSpinner(spinner.Line)),
		caja: textinput.New(), frames: map[string]bool{}}
	b.caja.Placeholder = "filtrar servidores"
	return b
}

func (m buscador) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spin.Tick, tea.Tick(time.Second, func(time.Time) tea.Msg { return finMsg{} })}
	return tea.Batch(cmds...)
}

func (m buscador) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(finMsg); ok {
		return m, tea.Quit
	}
	if !m.reenviar {
		return m, nil // ⛔ nadie le pasa nada a los hijos
	}
	if m.enfocar && !m.caja.Focused() {
		m.caja.Focus()
	}
	var c1, c2 tea.Cmd
	m.spin, c1 = m.spin.Update(msg) // ✅ cada hijo recibe el mensaje
	m.caja, c2 = m.caja.Update(msg) //    y devuelve su comando, que también hay que devolver
	return m, tea.Batch(c1, c2)
}

func (m buscador) View() tea.View {
	m.frames[m.spin.View()] = true
	return tea.NewView(m.spin.View() + " buscando · filtro: [" + m.caja.Value() + "]")
}

func probar(titulo string, reenviar, enfocar bool) {
	comun.Titulo(titulo)
	final, _ := comun.Correr(nuevo(reenviar, enfocar), comun.Teclas("w", "e", "b"), 100*time.Millisecond, 5*time.Second)
	b := final.(buscador)
	fmt.Printf("  tecleado «web» en 1 s -> filtro: %q · fotogramas distintos del spinner: %d\n",
		b.caja.Value(), len(b.frames))
	comun.Pantalla("al final", b.View().Content)
}

func main() {
	probar("6.1 · La trampa: no reenviar los mensajes a los hijos", false, false)
	probar("6.2 · Reenviados, pero la caja sin Focus()", true, false)
	probar("6.3 · Lo correcto: reenviar, devolver sus comandos y enfocar", true, true)
}
