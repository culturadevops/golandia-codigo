// Lección 3 · Los comandos: el trabajo lento va FUERA de Update.
//
// La trampa: pulsas «r» para refrescar y, dentro de Update, se llama a la API. Funciona: los
// datos llegan. Pero mientras la llamada dura, el bucle está parado: ni el reloj avanza, ni se
// atienden teclas, ni se pinta nada. La pantalla se congela y nadie ve un error.
//
// Lo correcto: Update devuelve un tea.Cmd, una función que Bubble Tea ejecuta en otra gorrutina.
// Cuando acaba, su resultado vuelve al bucle como un mensaje más.
package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"curso-tui/comun"
)

const latencia = time.Second // lo que tarda «la API»

type tickMsg time.Time
type datosMsg struct{ servidores int }

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// llamarAPI es el trabajo lento. Como Cmd, devuelve un mensaje en vez de tocar el modelo.
func llamarAPI() tea.Msg {
	time.Sleep(latencia)
	return datosMsg{servidores: 42}
}

type panel struct {
	bloqueante   bool
	cargando     bool
	servidores   int
	frames       int // ticks atendidos en total
	framesEspera int // ticks atendidos MIENTRAS se esperaba a la API
	teclasEspera int // teclas atendidas mientras se esperaba
	esperaMs     int64
	pedido       time.Time
}

func (m panel) Init() tea.Cmd { return tick() }

func (m panel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.frames++
		if m.cargando {
			m.framesEspera++
		}
		return m, tick()
	case datosMsg:
		m.servidores, m.cargando = msg.servidores, false
		m.esperaMs = time.Since(m.pedido).Milliseconds()
		return m, tea.Quit
	case tea.KeyPressMsg:
		if m.cargando {
			m.teclasEspera++
		}
		if msg.String() == "r" && !m.cargando {
			m.pedido = time.Now()
			if m.bloqueante {
				d := llamarAPI() // ⛔ el bucle entero espera aquí
				m.servidores = d.(datosMsg).servidores
				m.esperaMs = time.Since(m.pedido).Milliseconds()
				return m, tea.Quit
			}
			m.cargando = true
			return m, llamarAPI // ✅ se pasa la función, no su resultado
		}
	}
	return m, nil
}

func (m panel) View() tea.View { return tea.NewView(fmt.Sprintf("servidores: %d", m.servidores)) }

func probar(titulo string, bloqueante bool) {
	comun.Titulo(titulo)
	// «r», y después 5 teclas más mientras se espera, como haría alguien impaciente.
	final, _ := comun.Correr(panel{bloqueante: bloqueante},
		comun.Teclas("r", "down", "down", "down", "down", "down"), 150*time.Millisecond, 10*time.Second)
	p := final.(panel)
	fmt.Printf("  la API tardó %d ms · servidores: %d\n", p.esperaMs, p.servidores)
	fmt.Printf("  durante la espera: %d fotogramas del reloj · %d teclas atendidas\n",
		p.framesEspera, p.teclasEspera)
}

func main() {
	probar("3.1 · La trampa: llamar a la API dentro de Update", true)
	probar("3.2 · Lo correcto: devolver un tea.Cmd", false)
	fmt.Println("\n  Los dos traen los mismos datos. Uno congela la pantalla un segundo entero.")
}
