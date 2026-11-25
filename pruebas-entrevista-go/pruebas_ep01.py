# -*- coding: utf-8 -*-
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from probar import episodio

P1 = '''package main

import "fmt"

func main() {
	original := []int{1, 2, 3, 4, 5}
	trozo := original[1:3]

	fmt.Println("original:", original, "len", len(original), "cap", cap(original))
	fmt.Println("trozo:   ", trozo, "len", len(trozo), "cap", cap(trozo))

	trozo = append(trozo, 99)

	fmt.Println("--- despues del append ---")
	fmt.Println("trozo:   ", trozo)
	fmt.Println("original:", original, "  <-- nadie lo toco, y cambio")
}
'''

P2 = '''package main

import "fmt"

// El MISMO append. Lo unico que cambia es si cabia en la capacidad.
func main() {
	for _, corte := range [][2]int{{1, 3}, {3, 5}} {
		base := []int{1, 2, 3, 4, 5}
		t := base[corte[0]:corte[1]]
		fmt.Printf("base[%d:%d] -> len %d cap %d ", corte[0], corte[1], len(t), cap(t))
		antes := fmt.Sprint(base)
		t = append(t, 99)
		despues := fmt.Sprint(base)
		if antes == despues {
			fmt.Printf("| append REASIGNO, base intacta: %s\\n", despues)
		} else {
			fmt.Printf("| append PISO en sitio: %s -> %s\\n", antes, despues)
		}
	}
}
'''

P3 = '''package main

import "fmt"

func main() {
	base := []int{1, 2, 3, 4, 5}

	// Corte de tres indices: [inicio:fin:tope]. El tope recorta la CAPACIDAD.
	seguro := base[1:3:3]
	fmt.Println("base[1:3:3] -> len", len(seguro), "cap", cap(seguro))
	seguro = append(seguro, 99)
	fmt.Println("tras append -> base:", base, " (intacta)")

	// La otra salida: copy(), que no comparte nada desde el primer momento.
	copia := make([]int, 2)
	n := copy(copia, base[1:3])
	copia = append(copia, 99)
	fmt.Println("copy copio", n, "elementos -> copia:", copia, " base:", base)
}
'''

P4 = '''package main

import "fmt"

// El caso real: una funcion que hace append y NO devuelve el slice.
func agregar(s []int, v int) {
	s = append(s, v)
	fmt.Println("  dentro de agregar:", s)
}

func main() {
	s := make([]int, 0, 4)
	s = append(s, 1, 2)
	fmt.Println("antes:", s, "len", len(s), "cap", cap(s))
	agregar(s, 3)
	fmt.Println("despues:", s, "  <-- el 3 no esta: len sigue siendo", len(s))
	fmt.Println("pero SI se escribio en el array de debajo:", s[:cap(s)])
}
'''

P5 = r'''package main

import "fmt"

func main() {
	// Como CRECE la capacidad: solo se imprime cuando cambia.
	var s []int
	prev := -1
	for i := 0; i < 2000; i++ {
		if cap(s) != prev {
			fmt.Printf("len %-5d cap %d\n", len(s), cap(s))
			prev = cap(s)
		}
		s = append(s, i)
	}
	fmt.Println("final -> len", len(s), "cap", cap(s))
}
'''

P6 = '''package main

import "fmt"

func main() {
	var nulo []int          // slice NIL: nunca se le asigno nada
	vacio := []int{}        // slice vacio, pero existe

	fmt.Println("nulo  == nil?", nulo == nil, " len", len(nulo), "cap", cap(nulo))
	fmt.Println("vacio == nil?", vacio == nil, "len", len(vacio), "cap", cap(vacio))
	fmt.Println()
	fmt.Println("Y aun asi, los dos hacen lo mismo:")
	nulo = append(nulo, 1)  // append sobre nil FUNCIONA
	fmt.Println("append sobre el nil ->", nulo)
	for range nulo[:0] {
		fmt.Println("nunca")
	}
	fmt.Println("recorrer un slice nil tampoco explota")
}
'''

episodio(
    "01-slices-y-arrays",
    "Episodio 1/8 — slices, capacidad y el append que pisa",
    [
        {"nombre": "1. El append que NO pisa (porque reasigna)",
         "pregunta": "trozo := original[1:3]; trozo = append(trozo, 99). Que imprime original?",
         "archivos": {"p01/main.go": P1},
         "correr": [(["go", "run", "./p01"], False)]},
        {"nombre": "2. El mismo append, dos resultados distintos",
         "pregunta": "Por que a veces pisa el original y a veces no?",
         "archivos": {"p02/main.go": P2},
         "correr": [(["go", "run", "./p02"], False)]},
        {"nombre": "3. Las dos salidas: corte de tres indices y copy",
         "pregunta": "Como lo evitas en produccion?",
         "archivos": {"p03/main.go": P3},
         "correr": [(["go", "run", "./p03"], False)]},
        {"nombre": "4. El caso que muerde de verdad: append dentro de una funcion",
         "pregunta": "Por que el elemento no aparece fuera, pero el array SI cambio?",
         "archivos": {"p04/main.go": P4},
         "correr": [(["go", "run", "./p04"], False)]},
        {"nombre": "5. Como crece la capacidad",
         "pregunta": "Cuanto pide append cuando no cabe?",
         "archivos": {"p05/main.go": P5},
         "correr": [(["go", "run", "./p05"], False)],
         "nota": "El crecimiento NO es un contrato del lenguaje: es como esta implementado hoy. "
                 "Decirlo asi en la entrevista es lo correcto."},
        {"nombre": "6. Slice nil contra slice vacio",
         "pregunta": "Un slice nil y uno vacio son lo mismo?",
         "archivos": {"p06/main.go": P6},
         "correr": [(["go", "run", "./p06"], False)]},
    ])
