// Package comun: lo que comparten las lecciones para correr un programa de Bubble Tea SIN
// terminal y enseñar lo que de verdad pasó. Nada de lo que sale en el vídeo se escribe a mano.
package comun

import (
	"fmt"
	"io"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Titulo imprime la cabecera de un apartado.
func Titulo(t string) { fmt.Printf("\n== %s ==\n", t) }

// Pantalla imprime lo que el usuario vería, sin los códigos de color, dentro de un marco.
func Pantalla(nombre, vista string) {
	lineas := strings.Split(strings.TrimRight(ansi.Strip(vista), "\n"), "\n")
	ancho := len([]rune(nombre)) + 3
	for _, l := range lineas {
		if w := ansi.StringWidth(l); w > ancho {
			ancho = w
		}
	}
	fmt.Printf("  ┌─ %s %s┐\n", nombre, strings.Repeat("─", ancho-len([]rune(nombre))-2))
	for _, l := range lineas {
		fmt.Printf("  │%s%s│\n", l, strings.Repeat(" ", ancho+1-ansi.StringWidth(l)))
	}
	fmt.Printf("  └%s┘\n", strings.Repeat("─", ancho+1))
}

// Tecla construye la pulsación que manda el terminal: "q", "space", "enter", "ctrl+c", "down"...
func Tecla(s string) tea.KeyPressMsg {
	switch s {
	case "space":
		return tea.KeyPressMsg{Code: ' ', Text: " "}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	}
	r := []rune(s)
	return tea.KeyPressMsg{Code: r[0], Text: s}
}

// Teclas convierte una lista de nombres de tecla en mensajes.
func Teclas(nombres ...string) []tea.Msg {
	ms := make([]tea.Msg, len(nombres))
	for i, n := range nombres {
		ms[i] = Tecla(n)
	}
	return ms
}

// Correr arranca un tea.Program de verdad (bucle de eventos, renderizador y todo), pero sin
// teclado ni pantalla: los mensajes se los mandamos nosotros con p.Send, con una pausa entre
// uno y otro como haría una persona. Si el programa no termina solo en `limite`, se mata y se
// dice. Devuelve el modelo final y si terminó por su cuenta.
func Correr(m tea.Model, msgs []tea.Msg, pausa, limite time.Duration) (tea.Model, bool) {
	p := tea.NewProgram(m,
		tea.WithInput(nil),
		tea.WithOutput(io.Discard),
		tea.WithWindowSize(80, 24),
		tea.WithoutSignals(),
	)
	go func() {
		for _, msg := range msgs {
			time.Sleep(pausa)
			p.Send(msg)
		}
	}()
	hecho := make(chan tea.Model, 1)
	go func() {
		final, _ := p.Run()
		hecho <- final
	}()
	select {
	case final := <-hecho:
		return final, true
	case <-time.After(limite):
		p.Kill()
		return <-hecho, false
	}
}
