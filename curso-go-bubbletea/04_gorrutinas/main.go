// Lección 4 · No se toca el estado desde otra gorrutina.
//
// El caso de siempre: una gorrutina vigila un log y, cada vez que llega una línea de error,
// suma uno a un contador que el modelo guarda por puntero. Compila y no da error. Dos fallos:
//   1. La pantalla no se entera: View solo se llama tras un mensaje, y aquí no hay mensaje.
//   2. Es una carrera de datos: dos gorrutinas leen y escriben el mismo int sin cerrojo.
//      `go run -race ./04_gorrutinas -carrera` lo caza (hace falta cgo; ver ejecutar_todo.py).
//
// Lo correcto: la gorrutina no toca nada. Manda un mensaje con p.Send y el que suma es Update.
package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	tea "charm.land/bubbletea/v2"

	"curso-tui/comun"
)

const errores = 10

type estado struct{ errores int }
type errorMsg struct{}
type finMsg struct{}

var (
	pintado string    // lo último que View puso en pantalla
	inicio  time.Time // para saber CUÁNDO se pintó cada cosa
	pintas  []string  // cada vez que View pintó un valor distinto: «t=… ms -> valor»
)

type monitor struct{ e *estado }

func (m monitor) Init() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(time.Time) tea.Msg { return finMsg{} })
}

func (m monitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case errorMsg:
		m.e.errores++ // ✅ solo Update toca el estado
	case finMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m monitor) View() tea.View {
	v := fmt.Sprintf("errores en el log: %d", m.e.errores)
	if v != pintado {
		pintas = append(pintas, fmt.Sprintf("t=%4d ms -> %s", time.Since(inicio).Milliseconds(), v))
	}
	pintado = v
	return tea.NewView(v)
}

// vigilarMal: la gorrutina escribe directamente en el estado.
func vigilarMal(e *estado) {
	for i := 0; i < errores; i++ {
		time.Sleep(100 * time.Millisecond)
		e.errores++ // ⛔ otra gorrutina, el mismo int, sin mensaje
	}
}

// vigilarBien: la gorrutina solo avisa.
func vigilarBien(p *tea.Program) {
	for i := 0; i < errores; i++ {
		time.Sleep(100 * time.Millisecond)
		p.Send(errorMsg{}) // ✅
	}
}

func correr(bien bool) (*estado, string) {
	e := &estado{}
	pintado, pintas, inicio = "", nil, time.Now()
	p := tea.NewProgram(monitor{e}, tea.WithInput(nil), tea.WithOutput(io.Discard),
		tea.WithWindowSize(80, 24), tea.WithoutSignals())
	if bien {
		go vigilarBien(p)
	} else {
		go vigilarMal(e)
	}
	p.Run()
	return e, pintado
}

func main() {
	carrera := flag.Bool("carrera", false, "solo la versión mala, para -race")
	flag.Parse()
	if *carrera {
		correr(false)
		return
	}

	comun.Titulo("4.1 · La trampa: la gorrutina suma directamente")
	e, p := correr(false)
	fmt.Printf("  errores de verdad: %d\n  en pantalla al salir: %q\n", e.errores, p)
	contar()

	comun.Titulo("4.2 · Lo correcto: la gorrutina manda un mensaje")
	e, p = correr(true)
	fmt.Printf("  errores de verdad: %d\n  en pantalla al salir: %q\n", e.errores, p)
	contar()
}

func contar() {
	fmt.Println("  lo que se vio, y cuándo (los errores llegan de t=100 a t=1000 ms; sale en t=1500):")
	for _, x := range pintas {
		fmt.Println("    " + x)
	}
}
