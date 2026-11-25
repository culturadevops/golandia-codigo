package parser

import (
	"bytes"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────
// LAS PRUEBAS NORMALES, las que escribiría cualquiera. Y que PASAN.
// Ese es medio vídeo: el parser está probado, y aun así tiene un agujero.
// ─────────────────────────────────────────────────────────────────────────

func TestParseaMensajeNormal(t *testing.T) {
	datos := []byte{0x01, 0x00, 0x05, 'h', 'o', 'l', 'a', '!'}
	m, ok := Parsea(datos)
	if !ok {
		t.Fatal("debería parsear")
	}
	if m.Tipo != 1 {
		t.Errorf("tipo: quería 1, tengo %d", m.Tipo)
	}
	if !bytes.Equal(m.Contenido, []byte("hola!")) {
		t.Errorf("contenido: quería «hola!», tengo %q", m.Contenido)
	}
}

func TestParseaMensajeVacio(t *testing.T) {
	datos := []byte{0x02, 0x00, 0x00}
	m, ok := Parsea(datos)
	if !ok || len(m.Contenido) != 0 {
		t.Errorf("un mensaje vacío debería parsear con contenido vacío")
	}
}

func TestParseaDemasiadoCorto(t *testing.T) {
	if _, ok := Parsea([]byte{0x01}); ok {
		t.Error("un mensaje de 1 byte no debería parsear")
	}
}

func TestIdaYVuelta(t *testing.T) {
	orig := &Mensaje{Tipo: 7, Contenido: []byte("un contenido cualquiera")}
	m, ok := Parsea(Serializa(orig))
	if !ok || m.Tipo != orig.Tipo || !bytes.Equal(m.Contenido, orig.Contenido) {
		t.Error("ida y vuelta debería devolver lo mismo")
	}
}

// ─────────────────────────────────────────────────────────────────────────
// Y AHORA EL FUZZ. Son seis líneas.
// ─────────────────────────────────────────────────────────────────────────

func FuzzParsea(f *testing.F) {
	// El corpus de semillas: ejemplos VÁLIDOS de los que partir. No hay que darle
	// ejemplos que fallen: para eso está.
	f.Add([]byte{0x01, 0x00, 0x05, 'h', 'o', 'l', 'a', '!'})
	f.Add([]byte{0x02, 0x00, 0x00})
	f.Add([]byte{0x07, 0x00, 0x03, 'a', 'b', 'c'})

	f.Fuzz(func(t *testing.T, datos []byte) {
		// Lo único que se le pide: que no reviente. Devolver «no se puede» es correcto.
		Parsea(datos)
	})
}

// El mismo fuzz contra la versión corregida, para enseñar que aguanta.
func FuzzParseaBien(f *testing.F) {
	f.Add([]byte{0x01, 0x00, 0x05, 'h', 'o', 'l', 'a', '!'})
	f.Fuzz(func(t *testing.T, datos []byte) {
		ParseaBien(datos)
	})
}

// ⛔ Y este es el que demuestra lo que el fuzz NO encuentra: un fallo de LÓGICA que no
// provoca pánico. La versión corregida rechaza mensajes con longitud declarada mayor que
// los datos, pero acepta datos de SOBRA al final y los ignora en silencio — que en un
// protocolo de red es un agujero clásico (contrabando de peticiones). No revienta nunca.
func FuzzSobrasIgnoradas(f *testing.F) {
	f.Add([]byte{0x01, 0x00, 0x02, 'a', 'b'})
	f.Fuzz(func(t *testing.T, datos []byte) {
		ParseaBien(datos)
	})
}
