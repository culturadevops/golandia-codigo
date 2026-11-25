// Las mismas entradas por encoding/json (v1) y encoding/json/v2, lado a lado.
package main

import (
	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"strings"
	"time"
)

type Usuario struct {
	Nombre   string            `json:"nombre"`
	Roles    []string          `json:"roles"`
	Extra    map[string]string `json:"extra"`
	Alta     time.Time         `json:"alta,omitzero"`
	Edad     int               `json:"edad,omitempty"`
}

func linea(etiqueta string, b []byte, err error) {
	if err != nil {
		fmt.Printf("  %-3s error: %v\n", etiqueta, err)
		return
	}
	fmt.Printf("  %-3s %s\n", etiqueta, b)
}

func leer(etiqueta, entrada string, v2 bool) {
	var u Usuario
	var err error
	if v2 {
		err = json.Unmarshal([]byte(entrada), &u)
	} else {
		err = jsonv1.Unmarshal([]byte(entrada), &u)
	}
	if err != nil {
		fmt.Printf("  %-3s error: %v\n", etiqueta, err)
		return
	}
	fmt.Printf("  %-3s nombre=%q\n", etiqueta, u.Nombre)
}

func main() {
	fmt.Println("1 · Slice y map sin inicializar (nil):")
	u := Usuario{Nombre: "ana"}
	b, err := jsonv1.Marshal(u)
	linea("v1", b, err)
	b, err = json.Marshal(u)
	linea("v2", b, err)

	fmt.Println("\n2 · Mayúsculas distintas en la clave ({\"NOMBRE\":\"ana\"}):")
	leer("v1", `{"NOMBRE":"ana"}`, false)
	leer("v2", `{"NOMBRE":"ana"}`, true)

	fmt.Println("\n3 · Clave repetida ({\"nombre\":\"ana\",\"nombre\":\"admin\"}):")
	leer("v1", `{"nombre":"ana","nombre":"admin"}`, false)
	leer("v2", `{"nombre":"ana","nombre":"admin"}`, true)

	fmt.Println("\n4 · UTF-8 inválido en el texto (byte 0xff):")
	leer("v1", "{\"nombre\":\"an\xffa\"}", false)
	leer("v2", "{\"nombre\":\"an\xffa\"}", true)

	fmt.Println("\n5 · Opciones en la llamada (v2):")
	b, err = json.Marshal(map[string]int{"z": 1, "a": 2, "m": 3}, json.Deterministic(true), jsontext.WithIndent("  "))
	fmt.Printf("  Deterministic + WithIndent:\n%s\n", b)
	b, err = json.Marshal(Usuario{Nombre: "ana"}, json.FormatNilSliceAsNull(true), json.FormatNilMapAsNull(true))
	linea("v2", b, err)
	fmt.Println("      ^ con FormatNilSliceAsNull y FormatNilMapAsNull: igual que v1")

	fmt.Println("\n6 · Leer directo de un io.Reader:")
	var v Usuario
	err = json.UnmarshalRead(strings.NewReader(`{"nombre":"luis","roles":["admin"]}`), &v)
	fmt.Printf("  UnmarshalRead -> %+v err=%v\n", v.Roles, err)
}
