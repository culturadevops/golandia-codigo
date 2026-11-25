# -*- coding: utf-8 -*-
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from probar import episodio

P1 = '''package main

import "fmt"

func main() {
	var m map[string]int // NIL. Nunca se hizo make.

	fmt.Println("m == nil?", m == nil)
	fmt.Println("len(m):", len(m))
	fmt.Println("leer una clave que no existe:", m["loquesea"], "<- no explota, devuelve el cero")

	edades := map[string]int{"ana": 34}
	fmt.Println("edades[pedro]:", edades["pedro"], "<- pedro no existe y aun asi hay un numero")
}
'''

P2 = '''package main

import "fmt"

func main() {
	var m map[string]int
	fmt.Println("leer de un mapa nil va bien:", m["x"])
	fmt.Println("ahora escribo...")
	m["x"] = 1 // <-- aqui
}
'''

P3 = '''package main

import "fmt"

func main() {
	edades := map[string]int{"ana": 34, "recien": 0}

	// Sin comma-ok no puedes distinguir estas dos situaciones:
	fmt.Println("edades[recien] =", edades["recien"])
	fmt.Println("edades[nadie]  =", edades["nadie"])

	for _, k := range []string{"recien", "nadie"} {
		v, ok := edades[k]
		fmt.Printf("%-7s valor=%d existe=%v\\n", k, v, ok)
	}

	// El idiom que esto habilita: contar sin inicializar nada.
	cuenta := map[rune]int{}
	for _, r := range "abracadabra" {
		cuenta[r]++ // el valor cero hace que esto funcione a la primera
	}
	fmt.Println("a aparece", cuenta['a'], "veces")
}
'''

P4 = '''package main

import "fmt"

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8}
	for vuelta := 1; vuelta <= 4; vuelta++ {
		fmt.Print("vuelta ", vuelta, ": ")
		for k := range m {
			fmt.Print(k, " ")
		}
		fmt.Println()
	}
}
'''

P5 = '''package main

import "sync"

func main() {
	m := map[int]int{}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				m[n*10000+j] = j // dos goroutines escribiendo el mismo mapa
			}
		}(i)
	}
	wg.Wait()
}
'''

P7 = r'''package main

import "fmt"

func main() {
	// delete sobre un mapa NIL: no explota. Al reves que escribir.
	var nulo map[string]int
	delete(nulo, "x")
	fmt.Println("delete sobre un mapa nil: no pasa nada")

	m := map[string]int{"a": 1, "b": 2, "c": 3}
	delete(m, "b")
	fmt.Println("tras borrar la b:", len(m), "claves")
	delete(m, "no-existe")
	fmt.Println("borrar algo que no esta tampoco falla:", len(m), "claves")

	// Borrar DURANTE el recorrido esta permitido explicitamente.
	g := map[int]int{}
	for i := 0; i < 10; i++ {
		g[i] = i
	}
	for k := range g {
		if k%2 == 0 {
			delete(g, k)
		}
	}
	fmt.Println("tras borrar los pares durante el range, quedan:", len(g))
}
'''

P8 = r'''package main

import "fmt"

type Punto struct{ X, Y int }

func main() {
	// Un struct SI puede ser clave: es comparable.
	visitados := map[Punto]bool{}
	visitados[Punto{1, 2}] = true
	fmt.Println("struct como clave:", visitados[Punto{1, 2}])

	// Un array tambien.
	porPar := map[[2]int]string{{1, 2}: "uno-dos"}
	fmt.Println("array como clave:", porPar[[2]int{1, 2}])

	fmt.Println("lo que NO puede ser clave: slices, mapas y funciones")
}
'''

P9 = r'''package main

type Ruta struct{ Pasos []int }

func main() {
	// Un slice NO es comparable, asi que no puede ser clave.
	m := map[[]int]string{}
	_ = m
	// Y un struct que contiene un slice, tampoco.
	m2 := map[Ruta]string{}
	_ = m2
}
'''

episodio(
    "02-mapas-y-el-valor-cero",
    "Episodio 2/8 — mapas, valor cero y las tres cosas que no te esperas",
    [
        {"nombre": "1. Leer de un mapa NUNCA falla",
         "pregunta": "Lees una clave que no existe. Por que no explota?",
         "archivos": {"p01/main.go": P1},
         "correr": [(["go", "run", "./p01"], False)],
         "nota": "Aqui esta la mitad de la respuesta: el valor cero. La otra mitad es la prueba 2."},
        {"nombre": "2. Escribir en un mapa NIL sí falla",
         "pregunta": "Y si escribo en vez de leer?",
         "archivos": {"p02/main.go": P2},
         "correr": [(["go", "run", "./p02"], True)],
         "nota": "«leer va, escribir no» es la frase exacta que hay que decir en la entrevista."},
        {"nombre": "3. El comma-ok: distinguir «no esta» de «esta y vale cero»",
         "pregunta": "Como sabes si la clave existia?",
         "archivos": {"p03/main.go": P3},
         "correr": [(["go", "run", "./p03"], False)]},
        {"nombre": "4. El orden del range cambia en cada vuelta",
         "pregunta": "Puedo confiar en el orden de un range sobre un mapa?",
         "archivos": {"p04/main.go": P4},
         "correr": [(["go", "run", "./p04"], False)],
         "nota": "No es un fallo: Go lo aleatoriza A PROPOSITO para que nadie dependa de un orden\nque no esta garantizado."},
        {"nombre": "5. Dos goroutines escribiendo: fatal error, no panic",
         "pregunta": "El mapa es seguro entre goroutines?",
         "archivos": {"p05/main.go": P5},
         "correr": [(["go", "run", "./p05"], True)],
         "nota": "El detalle que separa la respuesta buena de la correcta: es un FATAL ERROR, no un\npanic. No hay recover que lo salve, el proceso se muere."},
        {"nombre": "7. Borrar: la tercera asimetria del nil",
         "pregunta": "Que pasa si borras de un mapa nil? Y durante un range?",
         "archivos": {"p07/main.go": P7},
         "correr": [(["go", "run", "./p07"], False)],
         "nota": "delete sobre nil NO explota, al reves que escribir. Y borrar durante el range\n"
                 "esta permitido por la especificacion, no es una casualidad."},
        {"nombre": "8. Que puede ser clave de un mapa",
         "pregunta": "Puedo usar cualquier cosa como clave?",
         "archivos": {"p08/main.go": P8},
         "correr": [(["go", "run", "./p08"], False)]},
        {"nombre": "9. Y que NO puede: el error del compilador",
         "pregunta": "Que dice el compilador si lo intentas con un slice?",
         "archivos": {"p09/main.go": P9},
         "correr": [(["go", "build", "-o", "/tmp/salida-p09", "./p09"], True)]},
    ])
