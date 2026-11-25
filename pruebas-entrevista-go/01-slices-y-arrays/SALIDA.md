# Pruebas ejecutadas — Episodio 1/8 — slices, capacidad y el append que pisa

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. El append que NO pisa (porque reasigna)

**La pregunta:** trozo := original[1:3]; trozo = append(trozo, 99). Que imprime original?

`p01/main.go`

```go
package main

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
```

```console
$ go run ./p01
original: [1 2 3 4 5] len 5 cap 5
trozo:    [2 3] len 2 cap 4
--- despues del append ---
trozo:    [2 3 99]
original: [1 2 3 99 5]   <-- nadie lo toco, y cambio
[codigo de salida: 0]
```

---

## 2. El mismo append, dos resultados distintos

**La pregunta:** Por que a veces pisa el original y a veces no?

`p02/main.go`

```go
package main

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
			fmt.Printf("| append REASIGNO, base intacta: %s\n", despues)
		} else {
			fmt.Printf("| append PISO en sitio: %s -> %s\n", antes, despues)
		}
	}
}
```

```console
$ go run ./p02
base[1:3] -> len 2 cap 4 | append PISO en sitio: [1 2 3 4 5] -> [1 2 3 99 5]
base[3:5] -> len 2 cap 2 | append REASIGNO, base intacta: [1 2 3 4 5]
[codigo de salida: 0]
```

---

## 3. Las dos salidas: corte de tres indices y copy

**La pregunta:** Como lo evitas en produccion?

`p03/main.go`

```go
package main

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
```

```console
$ go run ./p03
base[1:3:3] -> len 2 cap 2
tras append -> base: [1 2 3 4 5]  (intacta)
copy copio 2 elementos -> copia: [2 3 99]  base: [1 2 3 4 5]
[codigo de salida: 0]
```

---

## 4. El caso que muerde de verdad: append dentro de una funcion

**La pregunta:** Por que el elemento no aparece fuera, pero el array SI cambio?

`p04/main.go`

```go
package main

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
```

```console
$ go run ./p04
antes: [1 2] len 2 cap 4
  dentro de agregar: [1 2 3]
despues: [1 2]   <-- el 3 no esta: len sigue siendo 2
pero SI se escribio en el array de debajo: [1 2 3 0]
[codigo de salida: 0]
```

---

## 5. Como crece la capacidad

**La pregunta:** Cuanto pide append cuando no cabe?

`p05/main.go`

```go
package main

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
```

```console
$ go run ./p05
len 0     cap 0
len 1     cap 4
len 5     cap 8
len 9     cap 16
len 17    cap 32
len 33    cap 64
len 65    cap 128
len 129   cap 256
len 257   cap 512
len 513   cap 848
len 849   cap 1280
len 1281  cap 1792
len 1793  cap 2560
final -> len 2000 cap 2560
[codigo de salida: 0]
```

> El crecimiento NO es un contrato del lenguaje: es como esta implementado hoy. Decirlo asi en la entrevista es lo correcto.

---

## 6. Slice nil contra slice vacio

**La pregunta:** Un slice nil y uno vacio son lo mismo?

`p06/main.go`

```go
package main

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
```

```console
$ go run ./p06
nulo  == nil? true  len 0 cap 0
vacio == nil? false len 0 cap 0

Y aun asi, los dos hacen lo mismo:
append sobre el nil -> [1]
recorrer un slice nil tampoco explota
[codigo de salida: 0]
```
