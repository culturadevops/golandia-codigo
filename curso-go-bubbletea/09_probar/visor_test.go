package visor

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func tecla(s string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
}

// Probar el estado es directo: mensaje dentro, modelo fuera.
func TestPausaParaElContador(t *testing.T) {
	var m tea.Model = Visor{}
	m, _ = m.Update(refrescoMsg{})
	m, _ = m.Update(tecla("p"))
	m, _ = m.Update(refrescoMsg{})
	if got := m.(Visor).Lineas; got != 1 {
		t.Fatalf("con pausa no debería contar: Lineas = %d, quería 1", got)
	}
}

// ⛔ La prueba de la trampa: pasa, y el programa no sale.
func TestQSale_Mal(t *testing.T) {
	_, cmd := Visor{}.Update(tecla("q"))
	if cmd == nil {
		t.Fatal("q no devolvió ningún comando")
	}
}

// ✅ La buena: ejecuta el comando y mira QUÉ mensaje produce.
func TestQSale_Bien(t *testing.T) {
	_, cmd := Visor{}.Update(tecla("q"))
	if cmd == nil {
		t.Fatal("q no devolvió ningún comando")
	}
	if msg := cmd(); msg != (tea.QuitMsg{}) {
		t.Fatalf("q tenía que producir tea.QuitMsg y produjo %T", msg)
	}
}
