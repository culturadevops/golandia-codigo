package panel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var pestanas = []string{"Buckets", "Colas", "Tablas"}
var unidades = [][2]string{{"objeto", "objetos"}, {"mensaje", "mensajes"}, {"elemento", "elementos"}}

func unidad(p, n int) string {
	if n == 1 {
		return unidades[p][0]
	}
	return unidades[p][1]
}

// Cada cuánto se refresca solo.
const Intervalo = 2 * time.Second

type teclas struct {
	Pestana, Arriba, Abajo, Filtrar, Refrescar, Purgar, Salir key.Binding
}

var kb = teclas{
	Pestana:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "pestaña")),
	Arriba:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑↓", "moverse")),
	Abajo:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓", "bajar")),
	Filtrar:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filtrar")),
	Refrescar: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refrescar")),
	Purgar:    key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "purgar")),
	Salir:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "salir")), // lección 2
}

// ── mensajes ──────────────────────────────────────────────────────────────────────────────
type datosMsg struct {
	gen   int
	datos [3][]Recurso
	err   error
}
type tickMsg struct{}
type purgadaMsg struct {
	cola string
	err  error
}

// Modelo es el panel entero.
type Modelo struct {
	fuente    Fuente
	pestana   int
	cursor    int
	datos     [3][]Recurso
	err       error
	cargando  bool
	gen       int // lección 3 + 5: cada carga lleva número; la respuesta vieja se descarta
	refrescos int
	spin      spinner.Model
	filtro    textinput.Model
	filtrando bool
	confirmar *Recurso // cola pendiente de confirmar la purga
	aviso     string
	ayuda     help.Model
	ancho     int
}

// Nuevo crea el panel sobre una fuente.
func Nuevo(f Fuente) Modelo {
	ti := textinput.New()
	ti.Placeholder = "nombre"
	ti.Prompt = "/"
	return Modelo{fuente: f, spin: spinner.New(spinner.WithSpinner(spinner.Line)), filtro: ti,
		ayuda: help.New(), ancho: 80, gen: 1, cargando: true}
}

// cargar es el trabajo lento: va en un tea.Cmd, nunca dentro de Update (lección 3).
func (m Modelo) cargar() tea.Cmd {
	gen, f := m.gen, m.fuente
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var d [3][]Recurso
		var err error
		if d[0], err = f.Buckets(ctx); err != nil {
			return datosMsg{gen: gen, err: err}
		}
		if d[1], err = f.Colas(ctx); err != nil {
			return datosMsg{gen: gen, err: err}
		}
		if d[2], err = f.Tablas(ctx); err != nil {
			return datosMsg{gen: gen, err: err}
		}
		return datosMsg{gen: gen, datos: d}
	}
}

func esperar() tea.Cmd { return tea.Tick(Intervalo, func(time.Time) tea.Msg { return tickMsg{} }) }

func (m Modelo) refrescar() (Modelo, tea.Cmd) {
	m.gen++
	m.cargando = true
	return m, tea.Batch(m.cargar(), m.spin.Tick)
}

// Init NO devuelve modelo: un m.gen++ aquí se perdería y la primera carga se descartaría
// siempre. Por eso Nuevo ya arranca en gen 1 y cargando.
func (m Modelo) Init() tea.Cmd {
	return tea.Batch(m.cargar(), m.spin.Tick, esperar())
}

// visibles aplica el filtro a la pestaña actual.
func (m Modelo) visibles() []Recurso {
	f := strings.ToLower(m.filtro.Value())
	var out []Recurso
	for _, r := range m.datos[m.pestana] {
		if strings.Contains(strings.ToLower(r.Nombre), f) {
			out = append(out, r)
		}
	}
	return out
}

func (m Modelo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ancho = msg.Width
		m.ayuda.SetWidth(msg.Width)
		return m, nil

	case datosMsg:
		if msg.gen != m.gen {
			return m, nil // ⛔ respuesta de una carga vieja: llegaría después y pisaría la nueva
		}
		m.cargando = false
		m.refrescos++
		m.err = msg.err
		if msg.err == nil {
			m.datos = msg.datos
		}
		if n := len(m.visibles()); m.cursor >= n {
			m.cursor = max(0, n-1)
		}
		return m, nil

	case tickMsg:
		if m.cargando {
			return m, esperar()
		}
		m, c := m.refrescar()
		return m, tea.Batch(c, esperar())

	case purgadaMsg:
		if msg.err != nil {
			m.aviso = "no se pudo purgar " + msg.cola + ": " + msg.err.Error()
		} else {
			m.aviso = "purgada: " + msg.cola
		}
		return m.refrescar()

	case spinner.TickMsg:
		if !m.cargando {
			return m, nil // se deja de pedir fotogramas: el spinner se apaga
		}
		var c tea.Cmd
		m.spin, c = m.spin.Update(msg) // lección 6: al hijo se le pasa su mensaje
		return m, c

	case tea.KeyPressMsg:
		if m.confirmar != nil {
			cola := *m.confirmar
			m.confirmar = nil
			if msg.String() == "s" {
				f := m.fuente
				return m, func() tea.Msg {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					return purgadaMsg{cola: cola.Nombre, err: f.Purgar(ctx, cola)}
				}
			}
			m.aviso = "purga cancelada"
			return m, nil
		}
		if m.filtrando {
			switch msg.String() {
			case "enter", "esc":
				m.filtrando = false
				m.filtro.Blur()
				if msg.String() == "esc" {
					m.filtro.SetValue("")
				}
				m.cursor = 0
				return m, nil
			}
			var c tea.Cmd
			m.filtro, c = m.filtro.Update(msg)
			m.cursor = 0
			return m, c
		}
		m.aviso = ""
		switch {
		case key.Matches(msg, kb.Salir):
			return m, tea.Quit
		case key.Matches(msg, kb.Pestana):
			m.pestana = (m.pestana + 1) % len(pestanas)
			m.cursor = 0
		case key.Matches(msg, kb.Abajo):
			if m.cursor < len(m.visibles())-1 {
				m.cursor++
			}
		case key.Matches(msg, kb.Arriba):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, kb.Filtrar):
			m.filtrando = true
			return m, m.filtro.Focus() // lección 6: sin Focus no escribe
		case key.Matches(msg, kb.Refrescar):
			return m.refrescar()
		case key.Matches(msg, kb.Purgar):
			v := m.visibles()
			if m.pestana == 1 && len(v) > 0 {
				r := v[m.cursor]
				m.confirmar = &r
			} else {
				m.aviso = "solo se purgan colas"
			}
		}
	}
	return m, nil
}

var (
	estTitulo  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ADD8"))
	estActiva  = lipgloss.NewStyle().Bold(true).Reverse(true).Padding(0, 1)
	estPestana = lipgloss.NewStyle().Padding(0, 1)
	estCursor  = lipgloss.NewStyle().Bold(true)
	estTenue   = lipgloss.NewStyle().Faint(true)
	estError   = lipgloss.NewStyle().Foreground(lipgloss.Color("#e5534b"))
	colNombre  = lipgloss.NewStyle().Width(28) // lección 7: ancho en columnas, no en bytes
	colCuenta  = lipgloss.NewStyle().Width(8).Align(lipgloss.Right)
)

func (m Modelo) View() tea.View {
	var b strings.Builder
	estado := fmt.Sprintf("refrescos: %d", m.refrescos)
	if m.cargando {
		estado = m.spin.View() + " cargando"
	}
	b.WriteString(estTitulo.Render("nubepanel") + "  " + estTenue.Render(estado) + "\n\n")

	var ps []string
	for i, p := range pestanas {
		etiqueta := fmt.Sprintf("%s (%d)", p, len(m.datos[i]))
		if i == m.pestana {
			ps = append(ps, estActiva.Render(etiqueta))
		} else {
			ps = append(ps, estPestana.Render(etiqueta))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, ps...) + "\n\n")

	if m.err != nil {
		b.WriteString(estError.Render("sin conexión: "+recortar(causa(m.err), m.ancho-16)) + "\n")
		b.WriteString(estTenue.Render("se reintenta solo cada "+Intervalo.String()) + "\n")
	} else {
		v := m.visibles()
		if len(v) == 0 {
			b.WriteString(estTenue.Render("  (nada que enseñar)") + "\n")
		}
		for i, r := range v {
			linea := "  " + colNombre.Render(r.Nombre) + colCuenta.Render(fmt.Sprint(r.Cuenta)) + " " + unidad(m.pestana, r.Cuenta)
			if i == m.cursor {
				linea = estCursor.Render("> " + colNombre.Render(r.Nombre) + colCuenta.Render(fmt.Sprint(r.Cuenta)) + " " + unidad(m.pestana, r.Cuenta))
			}
			b.WriteString(linea + "\n")
		}
	}
	b.WriteString("\n")
	switch {
	case m.confirmar != nil:
		b.WriteString(estError.Render(fmt.Sprintf("¿Purgar %s? Se borran %d mensajes. (s/n)", m.confirmar.Nombre, m.confirmar.Cuenta)) + "\n")
	case m.filtrando || m.filtro.Value() != "":
		b.WriteString(m.filtro.View() + "\n")
	case m.aviso != "":
		b.WriteString(m.aviso + "\n")
	default:
		b.WriteString("\n")
	}
	b.WriteString(m.ayuda.ShortHelpView([]key.Binding{kb.Pestana, kb.Arriba, kb.Filtrar, kb.Refrescar, kb.Purgar, kb.Salir}))
	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// causa saca el error de red del fondo de la cadena: el SDK lo envuelve en tres capas y la de
// fuera («exceeded maximum number of attempts») no dice qué pasó.
func causa(err error) string {
	var op *net.OpError
	if errors.As(err, &op) {
		return op.Op + " " + op.Addr.String() + ": " + op.Err.Error()
	}
	return err.Error()
}

func recortar(s string, n int) string {
	if n < 10 || lipgloss.Width(s) <= n {
		return s
	}
	r := []rune(s)
	if len(r) > n-1 {
		r = r[:n-1]
	}
	return string(r) + "…"
}
