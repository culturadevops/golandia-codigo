// Lección 9 · Probar el modelo sin terminal.
//
// Update es una función: entra un modelo y un mensaje, sale un modelo y un comando. No hace
// falta terminal para probarla: se le mandan mensajes y se mira lo que devuelve.
//
// La trampa: comprobar que «q sale» mirando solo que el comando no es nil. Las funciones no se
// pueden comparar con ==, así que `cmd == tea.Quit` no compila y se acaba escribiendo
// `cmd != nil`. Pasa en verde... aunque el comando sea otro. Este visor tiene justo ese fallo.
package visor

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

type refrescoMsg struct{}

func refrescar() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return refrescoMsg{} })
}

// Visor muestra logs y se refresca cada segundo.
type Visor struct {
	Lineas int
	Pausa  bool
}

func (m Visor) Init() tea.Cmd { return refrescar() }

func (m Visor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case refrescoMsg:
		if !m.Pausa {
			m.Lineas++
		}
		return m, refrescar()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "p":
			m.Pausa = !m.Pausa
		case "q":
			return m, refrescar() // ⛔ EL FALLO: copiado del caso de arriba; debería ser tea.Quit
		}
	}
	return m, nil
}

func (m Visor) View() tea.View { return tea.NewView("logs") }
