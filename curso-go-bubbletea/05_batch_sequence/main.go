// Lección 5 · tea.Batch no tiene orden.
//
// La trampa: al arrancar hay que (1) leer la configuración, (2) conectar con el servidor que dice
// esa configuración y (3) pedir la lista. Se meten los tres en tea.Batch porque «así arrancan
// todos». Batch los lanza A LA VEZ, cada uno en su gorrutina: los mensajes vuelven en el orden en
// que terminan, no en el que los escribiste. No hay error: solo un panel que a veces conecta
// antes de saber adónde.
//
// tea.Sequence los lanza de uno en uno: el segundo no empieza hasta que el primero ha vuelto.
package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"curso-tui/comun"
)

type pasoMsg string

// paso simula un trabajo que tarda `d` y vuelve con su nombre.
func paso(nombre string, d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return pasoMsg(nombre)
	}
}

type arranque struct {
	ordenado bool
	llegadas []string
	inicio   time.Time
}

func (m arranque) Init() tea.Cmd {
	cmds := []tea.Cmd{
		paso("1-leer-config", 300*time.Millisecond),
		paso("2-conectar", 100*time.Millisecond),
		paso("3-pedir-lista", 200*time.Millisecond),
	}
	if m.ordenado {
		return tea.Sequence(cmds...)
	}
	return tea.Batch(cmds...)
}

func (m arranque) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if p, ok := msg.(pasoMsg); ok {
		m.llegadas = append(m.llegadas, fmt.Sprintf("%s (t=%d ms)", p, time.Since(m.inicio).Milliseconds()))
		if len(m.llegadas) == 3 {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m arranque) View() tea.View { return tea.NewView(strings.Join(m.llegadas, "\n")) }

func probar(titulo string, ordenado bool) {
	comun.Titulo(titulo)
	final, _ := comun.Correr(arranque{ordenado: ordenado, inicio: time.Now()}, nil, 0, 5*time.Second)
	for i, l := range final.(arranque).llegadas {
		fmt.Printf("  llegó %d.º: %s\n", i+1, l)
	}
}

func main() {
	fmt.Println("  escritos en este orden: 1-leer-config (300 ms) · 2-conectar (100 ms) · 3-pedir-lista (200 ms)")
	probar("5.1 · La trampa: tea.Batch", false)
	probar("5.2 · Lo correcto cuando importa el orden: tea.Sequence", true)
	fmt.Println("\n  Batch: más rápido, sin orden. Sequence: en orden, la suma de los tiempos.")
}
