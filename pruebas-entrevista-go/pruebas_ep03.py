# -*- coding: utf-8 -*-
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from probar import episodio

P1 = '''package main

import "fmt"

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

// El error esta AQUI, y parece correcto.
func hacerAlgo(fallar bool) error {
	var err *MiError // nil
	if fallar {
		err = &MiError{Codigo: 500}
	}
	return err // devuelve la variable, no nil literal
}

func main() {
	err := hacerAlgo(false) // le decimos que NO falle
	fmt.Println("err vale:", err)
	if err != nil {
		fmt.Println(">> ENTRO EN EL IF. No fallo nada y estamos tratando un error.")
	} else {
		fmt.Println(">> todo bien")
	}
}
'''

P2 = '''package main

import "fmt"

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

func main() {
	var puntero *MiError // nil
	var iface error = puntero

	fmt.Println("el puntero es nil?          ", puntero == nil)
	fmt.Println("la interfaz es nil?         ", iface == nil, "  <-- LA TRAMPA")
	fmt.Println()
	fmt.Println("una interfaz son DOS palabras: (tipo, valor)")
	fmt.Printf("  tipo  = %T\\n", iface)
	fmt.Printf("  valor = %v\\n", iface)
	fmt.Println()

	var vacia error // nunca se le asigno nada
	fmt.Printf("interfaz de verdad vacia: tipo=%T valor=%v  nil? %v\\n", vacia, vacia, vacia == nil)
}
'''

P3 = '''package main

import "fmt"

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

// LA CORRECCION: no declarar la variable con tipo concreto para devolverla como error.
func bien(fallar bool) error {
	if fallar {
		return &MiError{Codigo: 500}
	}
	return nil // nil LITERAL: las dos palabras vacias
}

func main() {
	for _, fallar := range []bool{false, true} {
		err := bien(fallar)
		fmt.Printf("fallar=%-5v  err=%-9v  err != nil -> %v\\n", fallar, err, err != nil)
	}
}
'''

P4 = '''package main

import (
	"errors"
	"fmt"
)

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

func malo() error {
	var err *MiError
	return err
}

func main() {
	err := malo()

	// La repregunta: y errors.As, que hace con esto?
	var destino *MiError
	fmt.Println("errors.As lo reconoce?", errors.As(err, &destino))
	fmt.Println("y destino vale:       ", destino)
	fmt.Println()
	fmt.Println("O sea: As dice que SI es un *MiError, y ese *MiError es nil.")
	fmt.Println("Si ahora lees destino.Codigo, panic:")
	fmt.Println(destino.Codigo)
}
'''

P5 = r'''package main

import "fmt"

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

func main() {
	// No es cosa de `error`: le pasa a CUALQUIER interfaz, incluida `any`.
	var p *MiError
	var cualquiera any = p
	fmt.Println("any con un puntero nil dentro == nil?", cualquiera == nil)

	// Y tambien con un slice o un mapa nil dentro.
	var s []int
	var m map[string]int
	var a1 any = s
	var a2 any = m
	fmt.Println("any con un slice nil dentro == nil? ", a1 == nil)
	fmt.Println("any con un mapa  nil dentro == nil? ", a2 == nil)
	fmt.Println()
	fmt.Println("Pero el slice de dentro SI es nil:", s == nil)
}
'''

P6 = r'''package main

import (
	"fmt"
	"reflect"
)

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return "fallo" }

// Lo que la gente escribe cuando descubre el problema. Funciona... y es la respuesta MALA.
func esRealmenteNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return rv.IsNil()
	}
	return false
}

func main() {
	var p *MiError
	var iface error = p
	fmt.Println("iface == nil?          ", iface == nil)
	fmt.Println("esRealmenteNil(iface)? ", esRealmenteNil(iface))
	fmt.Println()
	fmt.Println("Funciona. Pero ahora cada comprobacion de error de tu programa")
	fmt.Println("pasa por reflect, y el bug sigue estando en la funcion que lo devuelve.")
}
'''


episodio(
    "03-interfaces-y-el-nil-que-no-es-nil",
    "Episodio 3/8 — el nil que no es nil",
    [
        {"nombre": "1. El caso, tal cual aparece en codigo real",
         "pregunta": "La funcion no fallo y devuelve nil. Por que entra en el if?",
         "archivos": {"p01/main.go": P1},
         "correr": [(["go", "run", "./p01"], False)]},
        {"nombre": "2. Por que: una interfaz son DOS palabras",
         "pregunta": "Que hay dentro de una interfaz?",
         "archivos": {"p02/main.go": P2},
         "correr": [(["go", "run", "./p02"], False)],
         "nota": "El visual del episodio: las dos casillas (tipo, valor) y el %T delatando al culpable."},
        {"nombre": "3. La correccion: nil LITERAL",
         "pregunta": "Como se arregla?",
         "archivos": {"p03/main.go": P3},
         "correr": [(["go", "run", "./p03"], False)]},
        {"nombre": "4. La repregunta: y errors.As?",
         "pregunta": "Que pasa si intentas extraer el tipo con errors.As?",
         "archivos": {"p04/main.go": P4},
         "correr": [(["go", "run", "./p04"], True)],
         "nota": "Esta es la vuelta de tuerca que casi nadie ve venir y que remata el episodio."},
        {"nombre": "5. No es cosa de `error`: le pasa a cualquier interfaz",
         "pregunta": "Solo pasa con error, o con any tambien?",
         "archivos": {"p05/main.go": P5},
         "correr": [(["go", "run", "./p05"], False)]},
        {"nombre": "6. La solucion con reflect, y por que es la MALA",
         "pregunta": "Y si compruebo con reflect?",
         "archivos": {"p06/main.go": P6},
         "correr": [(["go", "run", "./p06"], False)],
         "nota": "Funciona, y es la respuesta equivocada: mete reflect en cada comprobacion de error y deja el bug donde estaba. La buena es no crear el error envenenado."},
    ])
