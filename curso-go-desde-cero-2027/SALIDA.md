# SALIDA — Curso de Go desde cero 2027

> Generado por `ejecutar_todo.py` (golang:1.27.2 en Docker) en dos tandas el 2026-11-16: partes 1-3 y partes 4-6.
> **No se escribe a mano.** Es la corrida de la que salen las cifras de los vídeos. Los tiempos, el orden de los maps,
> las direcciones y los contadores de la carrera cambian en cada corrida: `python ejecutar_todo.py` regenera este archivo.

## Partes 1-3

```
@@VERSIONES
go version go1.27.2 linux/amd64
Linux 6.18.33.2-microsoft-standard-WSL2 x86_64
@@INICIO parte1/01_instalar
$ go version
go version go1.27.2 linux/amd64
$ go env GOOS GOARCH GOPATH
linux
amd64
/go
$ go mod init hola
go: creating new go.mod: module hola
$ cat go.mod
module hola

go 1.27.2
$ cat main.go
package main

import "fmt"

func main() {
	fmt.Println("hola, Golandia")
}
$ go run .
hola, Golandia
$ go build -o hola .
$ ls -l hola
2342441 hola
$ ./hola
hola, Golandia
$ go build -ldflags="-s -w" -o hola-chico .
$ ls -l hola-chico
1511584 hola-chico
$ GOOS=windows go build -o hola.exe .
$ file hola hola.exe
0000000 177   E   L   F
0000000   M   Z
@@FIN parte1/01_instalar 0
@@INICIO parte1/02_variables
--- a_vars.go
package main

import "fmt"

func main() {
	var nombre string = "Ana"
	var edad int
	vidas := 3
	precio := 19.99
	var activo bool
	var x, y = 1, "dos"

	fmt.Println(nombre, edad, vidas, precio, activo, x, y)
	fmt.Printf("%T %T %T %T %T\n", nombre, edad, vidas, precio, activo)
	fmt.Printf("valores cero: %q %d %v %v\n", "", edad, activo, 0.0)
}
---
$ go run . 
Ana 0 3 19.99 false 1 dos
string int int float64 bool
valores cero: "" 0 false 0
[codigo de salida: 0]
--- b_no_usada.go
package main

import "fmt"

func main() {
	total := 10
	descuento := 2
	fmt.Println(total)
}
---
$ go run . 
# p1_vars
./main.go:7:2: declared and not used: descuento
[codigo de salida: 1]
--- c_import.go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hola")
}
---
$ go run . 
# p1_vars
./main.go:5:2: "os" imported and not used
[codigo de salida: 1]
--- d_redeclarar.go
package main

import "fmt"

func main() {
	vidas := 3
	vidas := 2
	fmt.Println(vidas)
}
---
$ go run . 
# p1_vars
./main.go:7:8: no new variables on left side of :=
[codigo de salida: 1]
@@FIN parte1/02_variables 0
@@INICIO parte1/03_tipos
--- a_mezcla.go
package main

import "fmt"

func main() {
	var tamano int64 = 4096
	bloques := 3
	total := tamano * bloques
	fmt.Println(total)
}
---
$ go run . 
# p1_tipos
./main.go:8:11: invalid operation: tamano * bloques (mismatched types int64 and int)
[codigo de salida: 1]
--- b_conversion.go
package main

import "fmt"

func main() {
	var tamano int64 = 4096
	bloques := 3
	fmt.Println("total:", tamano*int64(bloques))

	precio := 19.99
	fmt.Println("int(19.99) =", int(precio))

	var nivel int8 = 127
	nivel++
	fmt.Println("int8 127 + 1 =", nivel)

	var grande int64 = 300
	fmt.Println("uint8(300) =", uint8(grande))
}
---
$ go run . 
total: 12288
int(19.99) = 19
int8 127 + 1 = -128
uint8(300) = 44
[codigo de salida: 0]
--- c_constante.go
package main

import "fmt"

func main() {
	var b uint8 = 300
	fmt.Println(b)
}
---
$ go run . 
# p1_tipos
./main.go:6:16: cannot use 300 (untyped int constant) as uint8 value in variable declaration (overflows)
[codigo de salida: 1]
@@FIN parte1/03_tipos 0
@@INICIO parte1/04_constantes
--- a_iota.go
package main

import "fmt"

type Estado int

const (
	Pendiente Estado = iota
	Enviado
	Entregado
	Devuelto
)

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
)

func main() {
	fmt.Println(Pendiente, Enviado, Entregado, Devuelto)
	fmt.Println(KB, MB, GB)
}
---
$ go run . 
0 1 2 3
1024 1048576 1073741824
[codigo de salida: 0]
--- b_sin_tipo.go
package main

import "fmt"

const Enorme = 1 << 100

func main() {
	fmt.Println(Enorme >> 98)
	fmt.Println(Enorme / 1e30)
	var n int = Enorme
	fmt.Println(n)
}
---
$ go run . 
# p1_const
./main.go:10:14: cannot use Enorme (untyped int constant 1267650600228229401496703205376) as int value in variable declaration (overflows)
[codigo de salida: 1]
@@FIN parte1/04_constantes 0
@@INICIO parte1/05_fmt
--- a_verbos.go
package main

import "fmt"

type Punto struct {
	X, Y int
}

func main() {
	p := Punto{3, 4}
	fmt.Printf("%v | %+v | %#v | %T\n", p, p, p, p)
	fmt.Printf("%d | %5d | %05d | %x | %b\n", 42, 42, 42, 255, 5)
	fmt.Printf("%f | %.2f | %8.3f | %e\n", 3.14159, 3.14159, 3.14159, 123456.789)
	fmt.Printf("%s | %q | %t\n", "hola", "hola", true)
	s := fmt.Sprintf("%s tiene %d años", "Ana", 30)
	fmt.Println(s)
}
---
$ go run . 
{3 4} | {X:3 Y:4} | main.Punto{X:3, Y:4} | main.Punto
42 |    42 | 00042 | ff | 101
3.141590 | 3.14 |    3.142 | 1.234568e+05
hola | "hola" | true
Ana tiene 30 años
[codigo de salida: 0]
--- b_verbo_malo.go
package main

import "fmt"

func main() {
	nombre := "Ana"
	edad := 30
	fmt.Printf("%d tiene %s años\n", nombre, edad)
	fmt.Printf("faltan: %s y %s\n", nombre)
}
---
$ go run . 
%!d(string=Ana) tiene %!s(int=30) años
faltan: Ana y %!s(MISSING)
[codigo de salida: 0]
$ go vet .
main.go:8:14: fmt.Printf format %d has arg nombre of wrong type string
main.go:9:27: fmt.Printf format %s reads arg #2, but call has 1 arg
[codigo de salida: 1]
@@FIN parte1/05_fmt 0
@@INICIO parte2/01_if
--- a_if.go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, entrada := range []string{"42", "cuarenta"} {
		if n, err := strconv.Atoi(entrada); err != nil {
			fmt.Println("no es un número:", entrada)
		} else if n > 18 {
			fmt.Println(n, "es mayor de edad")
		} else {
			fmt.Println(n, "es menor de edad")
		}
	}
}
---
$ go run . 
42 es mayor de edad
no es un número: cuarenta
[codigo de salida: 0]
--- b_alcance.go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	if n, err := strconv.Atoi("42"); err == nil {
		fmt.Println("dentro:", n)
	}
	fmt.Println("fuera:", n)
}
---
$ go run . 
# p2_if
./main.go:12:24: undefined: n
[codigo de salida: 1]
--- c_parentesis.go
package main

import "fmt"

func main() {
	x := 5
	if x > 3
	{
		fmt.Println("mayor")
	}
}
---
$ go run . 
# p2_if
./main.go:7:10: syntax error: unexpected newline, expected { after if clause
[codigo de salida: 1]
@@FIN parte2/01_if 0
@@INICIO parte2/02_for
--- a_formas.go
package main

import "fmt"

func main() {
	// 1. clásico
	for i := 0; i < 3; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println("<- clásico")

	// 2. como un while
	n := 1
	for n < 100 {
		n *= 3
	}
	fmt.Println(n, "<- como while")

	// 3. infinito con break
	intentos := 0
	for {
		intentos++
		if intentos == 4 {
			break
		}
	}
	fmt.Println(intentos, "<- infinito con break")

	// 4. range sobre un número (Go 1.22+)
	for i := range 3 {
		fmt.Print(i, " ")
	}
	fmt.Println("<- range 3")

	// 5. range sobre un slice, y continue
	for i, f := range []string{"mango", "guayaba", "parchita"} {
		if i == 1 {
			continue
		}
		fmt.Println(i, f)
	}
}
---
$ go run . 
0 1 2 <- clásico
243 <- como while
4 <- infinito con break
0 1 2 <- range 3
0 mango
2 parchita
[codigo de salida: 0]
--- b_texto.go
package main

import "fmt"

func main() {
	ave := "ñandú"
	fmt.Println("len:", len(ave))
	for i, r := range ave {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()
	for i := 0; i < len(ave); i++ {
		fmt.Printf("%d ", ave[i])
	}
	fmt.Println()
}
---
$ go run . 
len: 7
0:ñ 2:a 3:n 4:d 5:ú 
195 177 97 110 100 195 186 
[codigo de salida: 0]
--- c_while.go
package main

func main() {
	n := 0
	while n < 3 {
		n++
	}
}
---
$ go run . 
# p2_for
./main.go:5:8: syntax error: unexpected name n at end of statement
./main.go:8:1: syntax error: non-declaration statement outside function body
[codigo de salida: 1]
@@FIN parte2/02_for 0
@@INICIO parte2/03_switch
--- a_switch.go
package main

import "fmt"

func dia(n int) string {
	switch n {
	case 1, 2, 3, 4, 5:
		return "laborable"
	case 6, 7:
		return "fin de semana"
	default:
		return "no existe"
	}
}

func nota(x int) string {
	switch {
	case x >= 90:
		return "excelente"
	case x >= 70:
		return "aprobado"
	}
	return "a repasar"
}

func main() {
	fmt.Println(dia(3), dia(7), dia(9))
	fmt.Println(nota(95), nota(72), nota(10))
}
---
$ go run . 
laborable fin de semana no existe
excelente aprobado a repasar
[codigo de salida: 0]
--- b_fallthrough.go
package main

import "fmt"

func main() {
	nivel := 2
	fmt.Println("sin fallthrough:")
	switch nivel {
	case 2:
		fmt.Println("  nivel 2")
	case 1:
		fmt.Println("  nivel 1")
	}
	fmt.Println("con fallthrough:")
	switch nivel {
	case 2:
		fmt.Println("  nivel 2")
		fallthrough
	case 1:
		fmt.Println("  nivel 1")
	case 0:
		fmt.Println("  nivel 0")
	}
}
---
$ go run . 
sin fallthrough:
  nivel 2
con fallthrough:
  nivel 2
  nivel 1
[codigo de salida: 0]
@@FIN parte2/03_switch 0
@@INICIO parte2/04_funciones
--- a_funciones.go
package main

import "fmt"

func dividir(a, b int) (int, int) {
	return a / b, a % b
}

func estadisticas(nums ...int) (min, max int) {
	min, max = nums[0], nums[0]
	for _, n := range nums {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return
}

func aplicar(nums []int, f func(int) int) []int {
	out := []int{}
	for _, n := range nums {
		out = append(out, f(n))
	}
	return out
}

func main() {
	c, r := dividir(17, 5)
	fmt.Println("17 / 5 =", c, "resto", r)
	_, resto := dividir(10, 3)
	fmt.Println("solo el resto:", resto)
	fmt.Println(estadisticas(7, 2, 9, 4))
	doble := func(n int) int { return n * 2 }
	fmt.Println(aplicar([]int{1, 2, 3}, doble))
}
---
$ go run . 
17 / 5 = 3 resto 2
solo el resto: 1
2 9
[2 4 6]
[codigo de salida: 0]
--- b_uno_de_dos.go
package main

import "fmt"

func dividir(a, b int) (int, int) {
	return a / b, a % b
}

func main() {
	c := dividir(17, 5)
	fmt.Println(c)
}
---
$ go run . 
# p2_fn
./main.go:10:7: assignment mismatch: 1 variable but dividir returns 2 values
[codigo de salida: 1]
@@FIN parte2/04_funciones 0
@@INICIO parte2/05_errores
--- a_atoi.go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, s := range []string{"42", "4x2", "99999999999999999999", ""} {
		n, err := strconv.Atoi(s)
		if err != nil {
			fmt.Printf("%-22q error: %v\n", s, err)
			continue
		}
		fmt.Printf("%-22q ok: %d\n", s, n)
	}
}
---
$ go run . 
"42"                   ok: 42
"4x2"                  error: strconv.Atoi: parsing "4x2": invalid syntax
"99999999999999999999" error: strconv.Atoi: parsing "99999999999999999999": value out of range
""                     error: strconv.Atoi: parsing "": invalid syntax
[codigo de salida: 0]
--- b_ignorar.go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	stock, _ := strconv.Atoi("12 unidades")
	fmt.Println("stock:", stock)
	if stock == 0 {
		fmt.Println("¡sin stock! cancelo el pedido")
	}
}
---
$ go run . 
stock: 0
¡sin stock! cancelo el pedido
[codigo de salida: 0]
--- c_propio.go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

func leerEdad(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("edad %q: %w", s, err)
	}
	if n < 0 {
		return 0, errors.New("la edad no puede ser negativa")
	}
	return n, nil
}

func main() {
	for _, s := range []string{"30", "treinta", "-5"} {
		edad, err := leerEdad(s)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println("edad:", edad)
	}
}
---
$ go run . 
edad: 30
error: edad "treinta": strconv.Atoi: parsing "treinta": invalid syntax
error: la edad no puede ser negativa
[codigo de salida: 0]
@@FIN parte2/05_errores 0
@@INICIO parte2/06_defer
--- a_orden.go
package main

import "fmt"

func main() {
	fmt.Println("abro el archivo")
	defer fmt.Println("defer 1: cierro el archivo")
	defer fmt.Println("defer 2: libero el candado")
	defer fmt.Println("defer 3: anoto en el log")
	fmt.Println("trabajo...")
}
---
$ go run . 
abro el archivo
trabajo...
defer 3: anoto en el log
defer 2: libero el candado
defer 1: cierro el archivo
[codigo de salida: 0]
--- b_argumentos.go
package main

import "fmt"

func main() {
	contador := 1
	defer fmt.Println("defer con argumento:", contador)
	defer func() { fmt.Println("defer con closure:", contador) }()
	contador = 100
	fmt.Println("al final de main, contador =", contador)
}
---
$ go run . 
al final de main, contador = 100
defer con closure: 100
defer con argumento: 1
[codigo de salida: 0]
--- c_panico.go
package main

import "fmt"

func procesar(datos []int) {
	defer fmt.Println("defer: limpio aunque haya pánico")
	fmt.Println("el tercero es", datos[2])
}

func main() {
	procesar([]int{1, 2})
	fmt.Println("esto ya no sale")
}
---
$ go run . 
defer: limpio aunque haya pánico
panic: runtime error: index out of range [2] with length 2

goroutine 1 [running]:
main.procesar({0x119e90088e88?, 0x0?, 0x0?})
	/tmp/p/p2_defer/main.go:7 +0xd1
main.main()
	/tmp/p/p2_defer/main.go:11 +0x31
exit status 2
[codigo de salida: 1]
@@FIN parte2/06_defer 0
@@INICIO parte3/01_arrays
--- a_valor.go
package main

import "fmt"

func main() {
	a := [3]int{1, 2, 3}
	b := a
	b[0] = 99
	fmt.Println("a =", a, " b =", b)
	fmt.Println("a == [3]int{1, 2, 3}:", a == [3]int{1, 2, 3})

	s := []int{1, 2, 3}
	t := s
	t[0] = 99
	fmt.Println("s =", s, " t =", t)
}
---
$ go run . 
a = [1 2 3]  b = [99 2 3]
a == [3]int{1, 2, 3}: true
s = [99 2 3]  t = [99 2 3]
[codigo de salida: 0]
--- b_largo_tipo.go
package main

import "fmt"

func suma(nums [3]int) int {
	return nums[0] + nums[1] + nums[2]
}

func main() {
	cuatro := [4]int{1, 2, 3, 4}
	fmt.Println(suma(cuatro))
}
---
$ go run . 
# p3_arr
./main.go:11:19: cannot use cuatro (variable of type [4]int) as [3]int value in argument to suma
[codigo de salida: 1]
@@FIN parte3/01_arrays 0
@@INICIO parte3/02_append
--- a_cap.go
package main

import "fmt"

func main() {
	var s []int
	fmt.Println("nil:", s == nil, "len", len(s), "cap", cap(s))
	ultimaCap := cap(s)
	for i := 0; i < 2000; i++ {
		s = append(s, i)
		if cap(s) != ultimaCap {
			fmt.Printf("len %4d -> cap %4d\n", len(s), cap(s))
			ultimaCap = cap(s)
		}
	}
}
---
$ go run . 
nil: true len 0 cap 0
len    1 -> cap    4
len    5 -> cap    8
len    9 -> cap   16
len   17 -> cap   32
len   33 -> cap   64
len   65 -> cap  128
len  129 -> cap  256
len  257 -> cap  512
len  513 -> cap  848
len  849 -> cap 1280
len 1281 -> cap 1792
len 1793 -> cap 2560
[codigo de salida: 0]
--- b_make.go
package main

import (
	"fmt"
	"testing"
)

func sinReservar() []int {
	var s []int
	for i := 0; i < 100000; i++ {
		s = append(s, i)
	}
	return s
}

func reservando() []int {
	s := make([]int, 0, 100000)
	for i := 0; i < 100000; i++ {
		s = append(s, i)
	}
	return s
}

func main() {
	for _, c := range []struct {
		nombre string
		f      func() []int
	}{{"sin reservar", sinReservar}, {"make con cap", reservando}} {
		r := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				c.f()
			}
		})
		fmt.Printf("%-13s %8d ns/op  %3d reservas/op\n", c.nombre, r.NsPerOp(), r.AllocsPerOp())
	}
}
---
$ go run . 
sin reservar   1098158 ns/op   25 reservas/op
make con cap     67806 ns/op    1 reservas/op
[codigo de salida: 0]
@@FIN parte3/02_append 0
@@INICIO parte3/03_compartido
--- a_trampa.go
package main

import "fmt"

func main() {
	notas := []int{10, 20, 30, 40, 50}
	primeras := notas[:2]
	fmt.Println("notas:", notas, " primeras:", primeras, "len", len(primeras), "cap", cap(primeras))

	primeras[0] = 99
	fmt.Println("tras primeras[0] = 99 ->", notas)

	primeras = append(primeras, 7)
	fmt.Println("tras append(primeras, 7) ->", notas)
}
---
$ go run . 
notas: [10 20 30 40 50]  primeras: [10 20] len 2 cap 5
tras primeras[0] = 99 -> [99 20 30 40 50]
tras append(primeras, 7) -> [99 20 7 40 50]
[codigo de salida: 0]
--- b_arreglo.go
package main

import (
	"fmt"
	"slices"
)

func main() {
	notas := []int{10, 20, 30, 40, 50}
	a := notas[:2:2]
	a = append(a, 7)
	fmt.Println("con notas[:2:2]:", notas, a)

	notas2 := []int{10, 20, 30, 40, 50}
	b := slices.Clone(notas2[:2])
	b = append(b, 7)
	fmt.Println("con slices.Clone:", notas2, b)
}
---
$ go run . 
con notas[:2:2]: [10 20 30 40 50] [10 20 7]
con slices.Clone: [10 20 30 40 50] [10 20 7]
[codigo de salida: 0]
@@FIN parte3/03_compartido 0
@@INICIO parte3/04_copy
--- a_copy.go
package main

import "fmt"

func main() {
	origen := []int{1, 2, 3, 4, 5}

	var destino []int
	n := copy(destino, origen)
	fmt.Println("copy a un slice nil:", n, "copiados", destino)

	destino = make([]int, 3)
	n = copy(destino, origen)
	fmt.Println("copy a make(3):     ", n, "copiados", destino)

	destino = make([]int, len(origen))
	n = copy(destino, origen)
	destino[0] = 99
	fmt.Println("copia completa:     ", n, "copiados", destino, "origen", origen)
}
---
$ go run . 
copy a un slice nil: 0 copiados []
copy a make(3):      3 copiados [1 2 3]
copia completa:      5 copiados [99 2 3 4 5] origen [1 2 3 4 5]
[codigo de salida: 0]
@@FIN parte3/04_copy 0
@@INICIO parte3/05_maps
--- a_maps.go
package main

import "fmt"

func main() {
	stock := map[string]int{"café": 10, "arepa": 4}
	stock["té"] = 7
	stock["café"] += 2
	fmt.Println(stock, "len", len(stock))

	fmt.Println("sushi:", stock["sushi"])
	if n, ok := stock["sushi"]; !ok {
		fmt.Println("sushi no está (n =", n, ")")
	}
	stock["agua"] = 0
	n, ok := stock["agua"]
	fmt.Println("agua:", n, "existe:", ok)

	delete(stock, "té")
	fmt.Println("tras delete:", stock)
}
---
$ go run . 
map[arepa:4 café:12 té:7] len 3
sushi: 0
sushi no está (n = 0 )
agua: 0 existe: true
tras delete: map[agua:0 arepa:4 café:12]
[codigo de salida: 0]
--- b_nil.go
package main

import "fmt"

func main() {
	var visitas map[string]int
	fmt.Println("leer de un map nil:", visitas["inicio"], "len", len(visitas))
	visitas["inicio"]++
	fmt.Println("esto ya no sale")
}
---
$ go run . 
leer de un map nil: 0 len 0
panic: assignment to entry in nil map

goroutine 1 [running]:
main.main()
	/tmp/p/p3_map/main.go:8 +0xdf
exit status 2
[codigo de salida: 1]
@@FIN parte3/05_maps 0
@@INICIO parte3/06_orden
--- a_orden.go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	m := map[string]int{"uno": 1, "dos": 2, "tres": 3, "cuatro": 4, "cinco": 5}
	for vuelta := 1; vuelta <= 5; vuelta++ {
		fmt.Print("vuelta ", vuelta, ": ")
		for k := range m {
			fmt.Print(k, " ")
		}
		fmt.Println()
	}
	fmt.Println("ordenadas:", slices.Sorted(maps.Keys(m)))
}
---
$ go run . 
vuelta 1: uno dos tres cuatro cinco 
vuelta 2: cuatro cinco uno dos tres 
vuelta 3: cuatro cinco uno dos tres 
vuelta 4: cinco uno dos tres cuatro 
vuelta 5: uno dos tres cuatro cinco 
ordenadas: [cinco cuatro dos tres uno]
[codigo de salida: 0]
@@FIN parte3/06_orden 0
```

## Partes 4-6

```
@@VERSIONES
go version go1.27.2 linux/amd64
Linux 6.18.33.2-microsoft-standard-WSL2 x86_64
@@INICIO parte4/01_structs
--- a_struct.go
package main

import "fmt"

type Usuario struct {
	Nombre string
	Edad   int
	Activo bool
}

func main() {
	ana := Usuario{Nombre: "Ana", Edad: 31, Activo: true}
	var vacio Usuario
	beto := Usuario{Nombre: "Beto"}
	fmt.Println(ana.Nombre, ana.Edad, ana.Activo)
	fmt.Printf("vacío: %+v\n", vacio)
	fmt.Printf("beto:  %+v\n", beto)
	ana.Edad++
	fmt.Println("ana cumple años:", ana.Edad)
	otra := Usuario{Nombre: "Ana", Edad: 32, Activo: true}
	fmt.Println("¿ana == otra?", ana == otra)
}
---
$ go run . 
Ana 31 true
vacío: {Nombre: Edad:0 Activo:false}
beto:  {Nombre:Beto Edad:0 Activo:false}
ana cumple años: 32
¿ana == otra? true
[codigo de salida: 0]
--- b_error.go
package main

import "fmt"

type Punto struct {
	X, Y int
}

func main() {
	p := Punto{3}
	fmt.Println(p)
}
---
$ go run . 
# p4_structs
./main.go:10:14: too few values in struct literal of type Punto
[codigo de salida: 1]
@@FIN parte4/01_structs 0
@@INICIO parte4/02_punteros
--- a_valor.go
package main

import "fmt"

type Cuenta struct {
	Titular string
	Saldo   int
}

func (c Cuenta) Depositar(monto int) {
	c.Saldo += monto
	fmt.Println("  dentro de Depositar, saldo =", c.Saldo)
}

func main() {
	c := Cuenta{Titular: "Ana", Saldo: 100}
	c.Depositar(50)
	fmt.Println("después de Depositar, saldo =", c.Saldo)
}
---
$ go run . 
  dentro de Depositar, saldo = 150
después de Depositar, saldo = 100
[codigo de salida: 0]
--- b_puntero.go
package main

import "fmt"

type Cuenta struct {
	Titular string
	Saldo   int
}

func (c *Cuenta) Depositar(monto int) {
	c.Saldo += monto
}

func duplicar(n *int) {
	*n = *n * 2
}

func main() {
	c := Cuenta{Titular: "Ana", Saldo: 100}
	c.Depositar(50)
	fmt.Println("con receptor puntero, saldo =", c.Saldo)

	x := 21
	p := &x
	duplicar(p)
	fmt.Println("x =", x, "· *p =", *p)

	var q *int
	fmt.Println("un puntero sin asignar es", q)
	fmt.Println(*q)
}
---
$ go run . 
con receptor puntero, saldo = 150
x = 42 · *p = 42
un puntero sin asignar es <nil>
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x499f9b]

goroutine 1 [running]:
main.main()
	/tmp/p/p4_punteros/main.go:30 +0x1bb
exit status 2
[codigo de salida: 1]
@@FIN parte4/02_punteros 0
@@INICIO parte4/03_interfaces
--- a_interfaz.go
package main

import (
	"fmt"
	"math"
)

type Figura interface {
	Area() float64
}

type Rectangulo struct{ Ancho, Alto float64 }
type Circulo struct{ Radio float64 }

func (r Rectangulo) Area() float64 { return r.Ancho * r.Alto }
func (c Circulo) Area() float64    { return math.Pi * c.Radio * c.Radio }

func total(fs []Figura) float64 {
	suma := 0.0
	for _, f := range fs {
		suma += f.Area()
	}
	return suma
}

func main() {
	figuras := []Figura{Rectangulo{3, 4}, Circulo{1}}
	for _, f := range figuras {
		fmt.Printf("%T: %.2f\n", f, f.Area())
	}
	fmt.Printf("total: %.2f\n", total(figuras))
}
---
$ go run . 
main.Rectangulo: 12.00
main.Circulo: 3.14
total: 15.14
[codigo de salida: 0]
--- b_falta.go
package main

import "fmt"

type Figura interface {
	Area() float64
	Perimetro() float64
}

type Rectangulo struct{ Ancho, Alto float64 }

func (r Rectangulo) Area() float64 { return r.Ancho * r.Alto }

func main() {
	var f Figura = Rectangulo{3, 4}
	fmt.Println(f.Area())
}
---
$ go run . 
# p4_interfaces
./main.go:15:17: cannot use Rectangulo{…} (value of struct type Rectangulo) as Figura value in variable declaration: Rectangulo does not implement Figura (missing method Perimetro)
[codigo de salida: 1]
--- c_receptor.go
package main

import "fmt"

type Figura interface {
	Area() float64
}

type Cuadrado struct{ Lado float64 }

func (c *Cuadrado) Area() float64 { return c.Lado * c.Lado }

func main() {
	var f Figura = Cuadrado{2}
	fmt.Println(f.Area())
}
---
$ go run . 
# p4_interfaces
./main.go:14:17: cannot use Cuadrado{…} (value of struct type Cuadrado) as Figura value in variable declaration: Cuadrado does not implement Figura (method Area has pointer receiver)
[codigo de salida: 1]
@@FIN parte4/03_interfaces 0
@@INICIO parte4/04_nil
--- a_nil.go
package main

import "fmt"

type ErrorValidacion struct {
	Campo string
}

func (e *ErrorValidacion) Error() string { return "campo inválido: " + e.Campo }

func validar(nombre string) error {
	var err *ErrorValidacion
	if nombre == "" {
		err = &ErrorValidacion{Campo: "nombre"}
	}
	return err
}

func main() {
	err := validar("Ana")
	fmt.Printf("err = %v · tipo %T\n", err, err)
	if err != nil {
		fmt.Println("¿err != nil? sí: la validación falló... con un nombre válido")
	}
}
---
$ go run . 
err = <nil> · tipo *main.ErrorValidacion
¿err != nil? sí: la validación falló... con un nombre válido
[codigo de salida: 0]
--- b_bien.go
package main

import "fmt"

type ErrorValidacion struct {
	Campo string
}

func (e *ErrorValidacion) Error() string { return "campo inválido: " + e.Campo }

func validar(nombre string) error {
	if nombre == "" {
		return &ErrorValidacion{Campo: "nombre"}
	}
	return nil
}

func main() {
	fmt.Println("validar(\"Ana\") =", validar("Ana"))
	fmt.Println("validar(\"\")    =", validar(""))
}
---
$ go run . 
validar("Ana") = <nil>
validar("")    = campo inválido: nombre
[codigo de salida: 0]
@@FIN parte4/04_nil 0
@@INICIO parte4/05_embedding
--- a_embed.go
package main

import (
	"fmt"
	"time"
)

type Registro struct {
	Creado time.Time
}

func (r Registro) Antiguedad(ahora time.Time) time.Duration { return ahora.Sub(r.Creado) }

type Producto struct {
	Registro
	Nombre string
	Precio int
}

func (p Producto) String() string {
	return fmt.Sprintf("%s ($%d.%02d)", p.Nombre, p.Precio/100, p.Precio%100)
}

func main() {
	base := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	p := Producto{Registro: Registro{Creado: base}, Nombre: "café", Precio: 250}
	fmt.Println(p)
	fmt.Println("antigüedad:", p.Antiguedad(base.Add(36*time.Hour)))
	fmt.Println("creado:", p.Creado.Format("2006-01-02"))
	fmt.Printf("%v · %+v\n", p, p.Registro)
}
---
$ go run . 
café ($2.50)
antigüedad: 36h0m0s
creado: 2027-01-01
café ($2.50) · {Creado:2027-01-01 00:00:00 +0000 UTC}
[codigo de salida: 0]
@@FIN parte4/05_embedding 0
@@INICIO parte5/01_errores
--- a_wrap.go
package main

import (
	"errors"
	"fmt"
)

var ErrSinStock = errors.New("sin stock")

func reservar(producto string, cantidad int) error {
	if cantidad > 3 {
		return ErrSinStock
	}
	return nil
}

func comprarConW(cantidad int) error {
	if err := reservar("café", cantidad); err != nil {
		return fmt.Errorf("comprar %d cafés: %w", cantidad, err)
	}
	return nil
}

func comprarConV(cantidad int) error {
	if err := reservar("café", cantidad); err != nil {
		return fmt.Errorf("comprar %d cafés: %v", cantidad, err)
	}
	return nil
}

func main() {
	e1 := comprarConW(5)
	e2 := comprarConV(5)
	fmt.Println("con %w:", e1)
	fmt.Println("con %v:", e2)
	fmt.Println("errors.Is(con %w, ErrSinStock):", errors.Is(e1, ErrSinStock))
	fmt.Println("errors.Is(con %v, ErrSinStock):", errors.Is(e2, ErrSinStock))
	fmt.Println("e1 == ErrSinStock:", e1 == ErrSinStock)
	fmt.Println("comprar 2:", comprarConW(2))
}
---
$ go run . 
con %w: comprar 5 cafés: sin stock
con %v: comprar 5 cafés: sin stock
errors.Is(con %w, ErrSinStock): true
errors.Is(con %v, ErrSinStock): false
e1 == ErrSinStock: false
comprar 2: <nil>
[codigo de salida: 0]
@@FIN parte5/01_errores 0
@@INICIO parte5/02_as
--- a_as.go
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func leerConfig(ruta string) ([]byte, error) {
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("leer configuración: %w", err)
	}
	return datos, nil
}

func main() {
	_, err := leerConfig("/etc/no-existe.toml")
	fmt.Println("error:", err)
	fmt.Println("¿es fs.ErrNotExist?", errors.Is(err, fs.ErrNotExist))
	var pe *fs.PathError
	if errors.As(err, &pe) {
		fmt.Printf("errors.As: operación=%q ruta=%q causa=%v\n", pe.Op, pe.Path, pe.Err)
	}
	datos, err := leerConfig("/etc/hostname")
	fmt.Println("con un archivo que sí existe:", err == nil, len(datos) > 0)
}
---
$ go run . 
error: leer configuración: open /etc/no-existe.toml: no such file or directory
¿es fs.ErrNotExist? true
errors.As: operación="open" ruta="/etc/no-existe.toml" causa=no such file or directory
con un archivo que sí existe: true true
[codigo de salida: 0]
@@FIN parte5/02_as 0
@@INICIO parte5/03_paquetes
--- tienda/tienda.go
package tienda

// Precio es exportado: empieza con mayúscula.
func Precio(producto string) int {
	return precioBase(producto) * 116 / 100
}

// precioBase no se exporta: empieza con minúscula.
func precioBase(producto string) int {
	switch producto {
	case "café":
		return 250
	case "arepa":
		return 400
	}
	return 0
}
---
$ find . -name "*.go" -o -name go.mod
./go.mod
./main.go
./tienda/tienda.go
--- a_main.go
package main

import (
	"fmt"

	"tiendita/tienda"
)

func main() {
	fmt.Println("precio con impuesto:", tienda.Precio("café"))
}
---
$ go run . 
precio con impuesto: 290
[codigo de salida: 0]
--- b_main.go
package main

import (
	"fmt"

	"tiendita/tienda"
)

func main() {
	fmt.Println("precio sin impuesto:", tienda.precioBase("café"))
}
---
$ go run . 
# tiendita
./main.go:10:45: name precioBase not exported by package tienda
[codigo de salida: 1]
@@FIN parte5/03_paquetes 0
@@INICIO parte5/04_dependencias
--- a_uuid.go
package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	id := uuid.NewString()
	fmt.Println("largo del id:", len(id), "· versión:", uuid.MustParse(id).Version())
	fijo := uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://golandia.dev"))
	fmt.Println("id determinista:", fijo)
}
---
$ go run .   # sin pedir la dependencia
main.go:6:2: no required module provides package github.com/google/uuid; to add it:
	go get github.com/google/uuid
[codigo de salida: 1]
$ go get github.com/google/uuid@v1.6.0
added github.com/google/uuid v1.6.0
[codigo de salida: 0]
$ cat go.mod
module p5_dep

go 1.27.2

require github.com/google/uuid v1.6.0 // indirect
$ go mod tidy && cat go.mod
module p5_dep

go 1.27.2

require github.com/google/uuid v1.6.0
$ wc -l go.sum
2
$ go run . 
largo del id: 36 · versión: VERSION_4
id determinista: ba7fb919-7a1c-577d-a742-e2f77f23848c
[codigo de salida: 0]
@@FIN parte5/04_dependencias 0
@@INICIO parte5/05_tests
--- precio.go
package main

// Descuento aplica un porcentaje a un precio en centavos, redondeando hacia abajo.
func Descuento(precio, porcentaje int) int {
	if porcentaje > 100 {
		porcentaje = 100
	}
	return precio - precio*porcentaje/100
}

func main() {}
---
--- precio_test.go
package main

import "testing"

func TestDescuento(t *testing.T) {
	casos := []struct {
		nombre     string
		precio     int
		porcentaje int
		quiero     int
	}{
		{"sin descuento", 1000, 0, 1000},
		{"diez por ciento", 1000, 10, 900},
		{"todo gratis", 1000, 100, 0},
		{"más de cien", 1000, 150, 0},
		{"porcentaje negativo", 1000, -20, 1000},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := Descuento(c.precio, c.porcentaje); got != c.quiero {
				t.Errorf("Descuento(%d, %d) = %d, quiero %d", c.precio, c.porcentaje, got, c.quiero)
			}
		})
	}
}
---
$ go test
--- FAIL: TestDescuento (X.XXXs)
    --- FAIL: TestDescuento/porcentaje_negativo (X.XXXs)
        precio_test.go:21: Descuento(1000, -20) = 1200, quiero 1000
FAIL
exit status 1
FAIL	p5_tests	X.XXXs
[codigo de salida: 1]
$ go test -v -run "TestDescuento/negativo"
=== RUN   TestDescuento
=== RUN   TestDescuento/porcentaje_negativo
    precio_test.go:21: Descuento(1000, -20) = 1200, quiero 1000
--- FAIL: TestDescuento (X.XXs)
    --- FAIL: TestDescuento/porcentaje_negativo (X.XXs)
FAIL
exit status 1
FAIL	p5_tests	X.XXXs
[codigo de salida: 1]
$ # arreglo: if porcentaje < 0 { porcentaje = 0 }
$ go test -cover
PASS
coverage: 100.0% of statements
ok  	p5_tests	X.XXXs
[codigo de salida: 0]
@@FIN parte5/05_tests 0
@@INICIO parte6/01_goroutines
--- a_sin_esperar.go
package main

import "fmt"

func saludar(n int) {
	fmt.Println("hola desde la goroutine", n)
}

func main() {
	for i := 1; i <= 3; i++ {
		go saludar(i)
	}
	fmt.Println("main terminó")
}
---
$ go run . 
hola desde la goroutine 1
main terminó
[codigo de salida: 0]
$ go run . 
main terminó
[codigo de salida: 0]
--- b_waitgroup.go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	inicio := time.Now()
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Go(func() {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("terminó la tarea", i)
		})
	}
	wg.Wait()
	fmt.Printf("main terminó en %d ms\n", time.Since(inicio).Milliseconds()/10*10)
}
---
$ go run . 
terminó la tarea 2
terminó la tarea 3
terminó la tarea 1
main terminó en 100 ms
[codigo de salida: 0]
@@FIN parte6/01_goroutines 0
@@INICIO parte6/02_channels
--- a_canal.go
package main

import "fmt"

func cuadrados(nums []int, salida chan<- int) {
	for _, n := range nums {
		salida <- n * n
	}
	close(salida)
}

func main() {
	c := make(chan int)
	go cuadrados([]int{1, 2, 3, 4}, c)
	for v := range c {
		fmt.Println("recibí", v)
	}

	b := make(chan string, 2)
	b <- "uno"
	b <- "dos"
	fmt.Println("con buffer:", <-b, <-b, "· largo", len(b), "capacidad", cap(b))
}
---
$ go run . 
recibí 1
recibí 4
recibí 9
recibí 16
con buffer: uno dos · largo 0 capacidad 2
[codigo de salida: 0]
--- b_deadlock.go
package main

import "fmt"

func main() {
	c := make(chan int)
	c <- 42
	fmt.Println(<-c)
}
---
$ go run . 
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/tmp/p/p6_chan/main.go:7 +0x36
exit status 2
[codigo de salida: 1]
--- c_range_sin_close.go
package main

import "fmt"

func main() {
	c := make(chan int)
	go func() {
		for i := 1; i <= 3; i++ {
			c <- i
		}
	}()
	for v := range c {
		fmt.Println("recibí", v)
	}
}
---
$ go run . 
recibí 1
recibí 2
recibí 3
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan receive]:
main.main()
	/tmp/p/p6_chan/main.go:12 +0xf4
exit status 2
[codigo de salida: 1]
@@FIN parte6/02_channels 0
@@INICIO parte6/03_select
--- a_select.go
package main

import (
	"fmt"
	"time"
)

func consultar(demora time.Duration) <-chan string {
	c := make(chan string, 1)
	go func() {
		time.Sleep(demora)
		c <- fmt.Sprintf("respuesta tras %v", demora)
	}()
	return c
}

func pedir(demora time.Duration) {
	select {
	case r := <-consultar(demora):
		fmt.Println("ok:", r)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("timeout: me rindo a los 500ms")
	}
}

func main() {
	pedir(100 * time.Millisecond)
	pedir(2 * time.Second)
}
---
$ go run . 
ok: respuesta tras 100ms
timeout: me rindo a los 500ms
[codigo de salida: 0]
@@FIN parte6/03_select 0
@@INICIO parte6/04_carrera
--- a_carrera.go
package main

import (
	"fmt"
	"sync"
)

func main() {
	contador := 0
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			for range 1000 {
				contador++
			}
		})
	}
	wg.Wait()
	fmt.Println("contador =", contador, "(esperaba 1000000)")
}
---
$ go run . 
contador = 497536 (esperaba 1000000)
[codigo de salida: 0]
$ go run . 
contador = 346450 (esperaba 1000000)
[codigo de salida: 0]
$ go run . 
contador = 385015 (esperaba 1000000)
[codigo de salida: 0]
$ go run -race .
==================
WARNING: DATA RACE
Read at 0x… by goroutine 8:
  main.main.func1()
      /tmp/p/p6_race/main.go:14 +0x…

Previous write at 0x… by goroutine 10:
  main.main.func1()
      /tmp/p/p6_race/main.go:14 +0x…

Goroutine 8 (running) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x…
...
contador = 436630 (esperaba 1000000)
Found 2 data race(s)
exit status 66
[codigo de salida: 1]
--- b_mutex.go
package main

import (
	"fmt"
	"sync"
)

func main() {
	contador := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			for range 1000 {
				mu.Lock()
				contador++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	fmt.Println("contador =", contador, "(esperaba 1000000)")
}
---
$ go run . 
contador = 1000000 (esperaba 1000000)
[codigo de salida: 0]
$ go run -race .
contador = 1000000 (esperaba 1000000)
[codigo de salida: 0]
@@FIN parte6/04_carrera 0
@@INICIO parte6/05_proyecto
--- main.go
// huellas: calcula el sha256 de varios archivos con un grupo de workers.
//
//	huellas [-workers N] archivo...
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type resultado struct {
	indice int
	ruta   string
	huella string
	bytes  int64
	err    error
}

func huella(ruta string) (string, int64, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", n, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func main() {
	workers := flag.Int("workers", runtime.NumCPU(), "cuántos archivos a la vez")
	flag.Parse()
	rutas := flag.Args()
	if len(rutas) == 0 {
		fmt.Fprintln(os.Stderr, "uso: huellas [-workers N] archivo...")
		os.Exit(2)
	}

	inicio := time.Now()
	trabajos := make(chan int)
	resultados := make([]resultado, len(rutas))
	var wg sync.WaitGroup
	for range *workers {
		wg.Go(func() {
			for i := range trabajos {
				h, n, err := huella(rutas[i])
				resultados[i] = resultado{i, rutas[i], h, n, err}
			}
		})
	}
	for i := range rutas {
		trabajos <- i
	}
	close(trabajos)
	wg.Wait()

	var total int64
	fallos := 0
	for _, r := range resultados {
		if r.err != nil {
			fmt.Fprintln(os.Stderr, "error:", r.err)
			fallos++
			continue
		}
		total += r.bytes
		fmt.Printf("%s…  %s\n", r.huella[:16], r.ruta)
	}
	fmt.Fprintf(os.Stderr, "%d archivos · %d MiB · %d workers · %d ms\n",
		len(rutas)-fallos, total>>20, *workers, time.Since(inicio).Milliseconds())
	if fallos > 0 {
		os.Exit(1)
	}
}
---
$ ls datos | wc -l && du -sh datos
16
1.1G
$ nproc
24
$ ./huellas datos/f01.txt datos/f02.txt datos/no-existe.txt
764713c5c29a7a0e…  datos/f01.txt
292200dcd6df2d24…  datos/f02.txt
error: open datos/no-existe.txt: no such file or directory
2 archivos · 128 MiB · 24 workers · … ms
[codigo de salida: 1]
$ ./huellas
uso: huellas [-workers N] archivo...
[codigo de salida: 2]
$ ./huellas -workers 1 datos/* > /dev/null   (3 veces)
16 archivos · 1024 MiB · 1 workers · 322 ms
16 archivos · 1024 MiB · 1 workers · 310 ms
16 archivos · 1024 MiB · 1 workers · 307 ms
$ ./huellas -workers 2 datos/* > /dev/null   (3 veces)
16 archivos · 1024 MiB · 2 workers · 153 ms
16 archivos · 1024 MiB · 2 workers · 152 ms
16 archivos · 1024 MiB · 2 workers · 152 ms
$ ./huellas -workers 4 datos/* > /dev/null   (3 veces)
16 archivos · 1024 MiB · 4 workers · 79 ms
16 archivos · 1024 MiB · 4 workers · 79 ms
16 archivos · 1024 MiB · 4 workers · 80 ms
$ ./huellas -workers 8 datos/* > /dev/null   (3 veces)
16 archivos · 1024 MiB · 8 workers · 42 ms
16 archivos · 1024 MiB · 8 workers · 44 ms
16 archivos · 1024 MiB · 8 workers · 44 ms
$ ./huellas -workers 16 datos/* > /dev/null   (3 veces)
16 archivos · 1024 MiB · 16 workers · 28 ms
16 archivos · 1024 MiB · 16 workers · 29 ms
16 archivos · 1024 MiB · 16 workers · 29 ms
$ ./huellas -workers 1 datos/* | head -3
764713c5c29a7a0e…  datos/f01.txt
292200dcd6df2d24…  datos/f02.txt
9063768dd66bc2c9…  datos/f03.txt
$ diff <(./huellas -workers 1 datos/*) <(./huellas -workers 16 datos/*) && echo iguales
iguales
@@FIN parte6/05_proyecto 0

```
