// parser.go — un lector de mensajes con longitud declarada.
//
// Es el formato más común que existe: un byte que dice el tipo, dos bytes que dicen cuánto mide
// el contenido, y el contenido. Así hablan medio protocolo de red y medio formato de fichero.
//
// ⛔ Y tiene un fallo plantado a propósito, de una clase muy real: SE FÍA DE LA LONGITUD QUE
//    VIENE EN EL MENSAJE. Si alguien dice «mido 500» y solo manda 3 bytes, el corte se sale.
//    Así nacen media docena de vulnerabilidades al año en parsers de verdad.
package parser

import "encoding/binary"

// Mensaje es lo que sale de leer un paquete.
type Mensaje struct {
	Tipo      byte
	Contenido []byte
}

// Parsea lee un mensaje del formato: [tipo:1][longitud:2][contenido:longitud]
//
// ⛔ EL FALLO ESTÁ AQUÍ y es de libro: se lee la longitud que viene DENTRO del mensaje y se usa
//    para cortar, sin comprobar que de verdad haya tantos bytes detrás.
func Parsea(datos []byte) (*Mensaje, bool) {
	if len(datos) < 3 {
		return nil, false
	}
	tipo := datos[0]
	n := binary.BigEndian.Uint16(datos[1:3])

	// Aquí debería ir:  if int(n) > len(datos)-3 { return nil, false }
	// No está. A propósito.
	contenido := datos[3 : 3+n]

	return &Mensaje{Tipo: tipo, Contenido: contenido}, true
}

// ParseaBien es la versión corregida, para poder enseñar las dos.
func ParseaBien(datos []byte) (*Mensaje, bool) {
	if len(datos) < 3 {
		return nil, false
	}
	tipo := datos[0]
	n := binary.BigEndian.Uint16(datos[1:3])
	if int(n) > len(datos)-3 {
		return nil, false
	}
	return &Mensaje{Tipo: tipo, Contenido: datos[3 : 3+n]}, true
}

// Serializa hace el camino contrario. Sirve para la propiedad de ida y vuelta del fuzz.
func Serializa(m *Mensaje) []byte {
	out := make([]byte, 3+len(m.Contenido))
	out[0] = m.Tipo
	binary.BigEndian.PutUint16(out[1:3], uint16(len(m.Contenido)))
	copy(out[3:], m.Contenido)
	return out
}
