# Pruebas ejecutadas — Episodio 2/8 — mapas, valor cero y las tres cosas que no te esperas

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. Leer de un mapa NUNCA falla

**La pregunta:** Lees una clave que no existe. Por que no explota?

`p01/main.go`

```go
package main

import "fmt"

func main() {
	var m map[string]int // NIL. Nunca se hizo make.

	fmt.Println("m == nil?", m == nil)
	fmt.Println("len(m):", len(m))
	fmt.Println("leer una clave que no existe:", m["loquesea"], "<- no explota, devuelve el cero")

	edades := map[string]int{"ana": 34}
	fmt.Println("edades[pedro]:", edades["pedro"], "<- pedro no existe y aun asi hay un numero")
}
```

```console
$ go run ./p01
m == nil? true
len(m): 0
leer una clave que no existe: 0 <- no explota, devuelve el cero
edades[pedro]: 0 <- pedro no existe y aun asi hay un numero
[codigo de salida: 0]
```

> Aqui esta la mitad de la respuesta: el valor cero. La otra mitad es la prueba 2.

---

## 2. Escribir en un mapa NIL sí falla

**La pregunta:** Y si escribo en vez de leer?

`p02/main.go`

```go
package main

import "fmt"

func main() {
	var m map[string]int
	fmt.Println("leer de un mapa nil va bien:", m["x"])
	fmt.Println("ahora escribo...")
	m["x"] = 1 // <-- aqui
}
```

```console
$ go run ./p02   # ⚠️ falla A PROPÓSITO
leer de un mapa nil va bien: 0
ahora escribo...
panic: assignment to entry in nil map

goroutine 1 [running]:
main.main()
	/app/p02/main.go:9 +0xdf
exit status 2
[codigo de salida: 1]
```

> «leer va, escribir no» es la frase exacta que hay que decir en la entrevista.

---

## 3. El comma-ok: distinguir «no esta» de «esta y vale cero»

**La pregunta:** Como sabes si la clave existia?

`p03/main.go`

```go
package main

import "fmt"

func main() {
	edades := map[string]int{"ana": 34, "recien": 0}

	// Sin comma-ok no puedes distinguir estas dos situaciones:
	fmt.Println("edades[recien] =", edades["recien"])
	fmt.Println("edades[nadie]  =", edades["nadie"])

	for _, k := range []string{"recien", "nadie"} {
		v, ok := edades[k]
		fmt.Printf("%-7s valor=%d existe=%v\n", k, v, ok)
	}

	// El idiom que esto habilita: contar sin inicializar nada.
	cuenta := map[rune]int{}
	for _, r := range "abracadabra" {
		cuenta[r]++ // el valor cero hace que esto funcione a la primera
	}
	fmt.Println("a aparece", cuenta['a'], "veces")
}
```

```console
$ go run ./p03
edades[recien] = 0
edades[nadie]  = 0
recien  valor=0 existe=true
nadie   valor=0 existe=false
a aparece 5 veces
[codigo de salida: 0]
```

---

## 4. El orden del range cambia en cada vuelta

**La pregunta:** Puedo confiar en el orden de un range sobre un mapa?

`p04/main.go`

```go
package main

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
```

```console
$ go run ./p04
vuelta 1: h a b c d e f g 
vuelta 2: e f g h a b c d 
vuelta 3: c d e f g h a b 
vuelta 4: e f g h a b c d
[codigo de salida: 0]
```

> No es un fallo: Go lo aleatoriza A PROPOSITO para que nadie dependa de un orden
> que no esta garantizado.

---

## 5. Dos goroutines escribiendo: fatal error, no panic

**La pregunta:** El mapa es seguro entre goroutines?

`p05/main.go`

```go
package main

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
```

```console
$ go run ./p05   # ⚠️ falla A PROPÓSITO
fatal error: concurrent map writes

goroutine 8 [running]:
internal/runtime/maps.fatal({0x493c79?, 0x0?})
	/usr/local/go/src/runtime/panic.go:1181 +0x18
main.main.func1(0x3)
	/app/p05/main.go:13 +0x6f
created by main.main in goroutine 1
	/app/p05/main.go:10 +0x45
exit status 2
[codigo de salida: 1]
```

> El detalle que separa la respuesta buena de la correcta: es un FATAL ERROR, no un
> panic. No hay recover que lo salve, el proceso se muere.

---

## 7. Borrar: la tercera asimetria del nil

**La pregunta:** Que pasa si borras de un mapa nil? Y durante un range?

`p07/main.go`

```go
package main

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
```

```console
$ go run ./p07
delete sobre un mapa nil: no pasa nada
tras borrar la b: 2 claves
borrar algo que no esta tampoco falla: 2 claves
tras borrar los pares durante el range, quedan: 5
[codigo de salida: 0]
```

> delete sobre nil NO explota, al reves que escribir. Y borrar durante el range
> esta permitido por la especificacion, no es una casualidad.

---

## 8. Que puede ser clave de un mapa

**La pregunta:** Puedo usar cualquier cosa como clave?

`p08/main.go`

```go
package main

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
```

```console
$ go run ./p08
struct como clave: true
array como clave: uno-dos
lo que NO puede ser clave: slices, mapas y funciones
[codigo de salida: 0]
```

---

## 9. Y que NO puede: el error del compilador

**La pregunta:** Que dice el compilador si lo intentas con un slice?

`p09/main.go`

```go
package main

type Ruta struct{ Pasos []int }

func main() {
	// Un slice NO es comparable, asi que no puede ser clave.
	m := map[[]int]string{}
	_ = m
	// Y un struct que contiene un slice, tampoco.
	m2 := map[Ruta]string{}
	_ = m2
}
```

```console
$ go build -o /tmp/salida-p09 ./p09   # ⚠️ falla A PROPÓSITO
# pruebas/p09
p09/main.go:7:11: invalid map key type []int
p09/main.go:10:12: invalid map key type Ruta
[codigo de salida: 1]
```
