# Pruebas ejecutadas — Episodio 3/8 — el nil que no es nil

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. El caso, tal cual aparece en codigo real

**La pregunta:** La funcion no fallo y devuelve nil. Por que entra en el if?

`p01/main.go`

```go
package main

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
```

```console
$ go run ./p01
err vale: <nil>
>> ENTRO EN EL IF. No fallo nada y estamos tratando un error.
[codigo de salida: 0]
```

---

## 2. Por que: una interfaz son DOS palabras

**La pregunta:** Que hay dentro de una interfaz?

`p02/main.go`

```go
package main

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
	fmt.Printf("  tipo  = %T\n", iface)
	fmt.Printf("  valor = %v\n", iface)
	fmt.Println()

	var vacia error // nunca se le asigno nada
	fmt.Printf("interfaz de verdad vacia: tipo=%T valor=%v  nil? %v\n", vacia, vacia, vacia == nil)
}
```

```console
$ go run ./p02
el puntero es nil?           true
la interfaz es nil?          false   <-- LA TRAMPA

una interfaz son DOS palabras: (tipo, valor)
  tipo  = *main.MiError
  valor = <nil>

interfaz de verdad vacia: tipo=<nil> valor=<nil>  nil? true
[codigo de salida: 0]
```

> El visual del episodio: las dos casillas (tipo, valor) y el %T delatando al culpable.

---

## 3. La correccion: nil LITERAL

**La pregunta:** Como se arregla?

`p03/main.go`

```go
package main

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
		fmt.Printf("fallar=%-5v  err=%-9v  err != nil -> %v\n", fallar, err, err != nil)
	}
}
```

```console
$ go run ./p03
fallar=false  err=<nil>      err != nil -> false
fallar=true   err=fallo 500  err != nil -> true
[codigo de salida: 0]
```

---

## 4. La repregunta: y errors.As?

**La pregunta:** Que pasa si intentas extraer el tipo con errors.As?

`p04/main.go`

```go
package main

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
```

```console
$ go run ./p04   # ⚠️ falla A PROPÓSITO
errors.As lo reconoce? true
y destino vale:        <nil>

O sea: As dice que SI es un *MiError, y ese *MiError es nil.
Si ahora lees destino.Codigo, panic:
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x4a29c6]

goroutine 1 [running]:
main.main()
	/app/p04/main.go:27 +0x1a6
exit status 2
[codigo de salida: 1]
```

> Esta es la vuelta de tuerca que casi nadie ve venir y que remata el episodio.

---

## 5. No es cosa de `error`: le pasa a cualquier interfaz

**La pregunta:** Solo pasa con error, o con any tambien?

`p05/main.go`

```go
package main

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
```

```console
$ go run ./p05
any con un puntero nil dentro == nil? false
any con un slice nil dentro == nil?  false
any con un mapa  nil dentro == nil?  false

Pero el slice de dentro SI es nil: true
[codigo de salida: 0]
```

---

## 6. La solucion con reflect, y por que es la MALA

**La pregunta:** Y si compruebo con reflect?

`p06/main.go`

```go
package main

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
```

```console
$ go run ./p06
iface == nil?           false
esRealmenteNil(iface)?  true

Funciona. Pero ahora cada comprobacion de error de tu programa
pasa por reflect, y el bug sigue estando en la funcion que lo devuelve.
[codigo de salida: 0]
```

> Funciona, y es la respuesta equivocada: mete reflect en cada comprobacion de error y deja el bug donde estaba. La buena es no crear el error envenenado.
