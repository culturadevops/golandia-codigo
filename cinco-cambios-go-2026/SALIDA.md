# SALIDA — 714 «5 cambios en Go que están pasando ahora (2026)»

> Corrida del 2026-10-30 05:55 con `python ejecutar_todo.py`. Lo nuevo, en `golang:1.27.2` (Docker, --cpus 2). Lo reutilizado se copia abajo de los SALIDA.md de 700, 703, 705 y 545, con su ruta.


# Lo nuevo del 714 · golang:1.27.2

```
go version: go version go1.27.2 linux/amd64

## 01-go-fix-servicio · go.mod go 1.27
$ go build ./...
[código de salida: 0]
$ go vet ./...
[código de salida: 0]
$ go test ./...
ok  	pedidos	0.002s
[código de salida: 0]
$ go fix -diff ./...
# pedidos
# [pedidos]
fix: omitzero: ignoring alternative fix "Replace omitempty with omitzero (behavior change)"
--- …/01-go-fix-servicio/pedidos.go (old)
+++ …/01-go-fix-servicio/pedidos.go (new)
@@ -5,7 +5,8 @@
 import (
 	"errors"
 	"io/fs"
-	"sort"
+	"maps"
+	"slices"
 	"strings"
 	"sync"
 	"sync/atomic"
@@ -16,15 +17,15 @@
 	ID      string    `json:"id"`
 	Cliente string    `json:"cliente"`
 	Total   int       `json:"total"`
-	Alta    time.Time `json:"alta,omitempty"`
+	Alta    time.Time `json:"alta"`
 }
 
-var procesados int64
+var procesados atomic.Int64
 
 // Etiquetas convierte "a,b,c" en una lista sin espacios.
 func Etiquetas(s string) []string {
 	var res []string
-	for _, p := range strings.Split(s, ",") {
+	for p := range strings.SplitSeq(s, ",") {
 		res = append(res, strings.TrimSpace(p))
 	}
 	return res
@@ -32,38 +27,31 @@
 
 // Cliente saca el cliente de una línea "cliente=ana".
 func Cliente(linea string) (string, bool) {
-	if strings.HasPrefix(linea, "cliente=") {
-		return strings.TrimPrefix(linea, "cliente="), true
+	if after, ok := strings.CutPrefix(linea, "cliente="); ok {
+		return after, true
 	}
 	return "", false
 }
 
 // EsVIP dice si el cliente está en la lista.
 func EsVIP(vips []string, cliente string) bool {
-	for _, v := range vips {
-		if v == cliente {
-			return true
-		}
-	}
-	return false
+	return slices.Contains(vips, cliente)
 }
 
 // Copiar pasa los totales de un mapa a otro.
 func Copiar(dst, src map[string]int) {
-	for k, v := range src {
-		dst[k] = v
-	}
+	maps.Copy(dst, src)
 }
 
 // Ordenar ordena los totales de menor a mayor.
 func Ordenar(totales []int) {
-	sort.Slice(totales, func(i, j int) bool { return totales[i] < totales[j] })
+	slices.Sort(totales)
 }
 
 // Reintentos devuelve los tiempos de espera: 100 ms, 200 ms, 400 ms...
 func Reintentos(n int) []time.Duration {
 	var res []time.Duration
-	for i := 0; i < n; i++ {
+	for i := range n {
 		res = append(res, time.Duration(100<<i)*time.Millisecond)
 	}
 	return res
@@ -74,31 +43,27 @@
 	res := make([]int, len(ps))
 	var wg sync.WaitGroup
 	for i, p := range ps {
-		i, p := i, p
-		wg.Add(1)
-		go func() {
-			defer wg.Done()
+		wg.Go(func() {
 			res[i] = f(p)
-			atomic.AddInt64(&procesados, 1)
-		}()
+			procesados.Add(1)
+		})
 	}
 	wg.Wait()
 	return res
 }
 
 // Procesados cuenta los pedidos procesados desde que arrancó.
-func Procesados() int64 { return atomic.LoadInt64(&procesados) }
+func Procesados() int64 { return procesados.Load() }
 
 // RutaFallida saca la ruta de un error de archivo.
 func RutaFallida(err error) (string, bool) {
-	var pe *fs.PathError
-	if errors.As(err, &pe) {
+	if pe, ok := errors.AsType[*fs.PathError](err); ok {
 		return pe.Path, true
 	}
 	return "", false
 }
 
 // Campos devuelve los datos de un pedido como mapa genérico.
-func Campos(p Pedido) map[string]interface{} {
-	return map[string]interface{}{"id": p.ID, "total": p.Total}
+func Campos(p Pedido) map[string]any {
+	return map[string]any{"id": p.ID, "total": p.Total}
 }
[código de salida: 1]
# líneas quitadas / añadidas por go fix: 28 / 18
$ go fix ./...
# pedidos
# [pedidos]
fix: omitzero: ignoring alternative fix "Replace omitempty with omitzero (behavior change)"
[código de salida: 0]
$ go build ./...
[código de salida: 0]
$ go vet ./...
[código de salida: 0]
$ go test ./...
ok  	pedidos	0.002s
[código de salida: 0]
$ go fix -diff ./...
[código de salida: 0]

## 01-go-fix-servicio · el mismo código con go.mod go 1.21
$ go test ./...
ok  	pedidos	0.001s
[código de salida: 0]
# líneas quitadas / añadidas por go fix: 15 / 10
$ go fix -diff ./...
--- …/servicio121/pedidos.go (old)
+++ …/servicio121/pedidos.go (new)
@@ -5,7 +5,7 @@
 import (
 	"errors"
 	"io/fs"
-	"sort"
+	"slices"
 	"strings"
 	"sync"
 	"sync/atomic"
@@ -19,7 +19,7 @@
 	Alta    time.Time `json:"alta,omitempty"`
 }
 
-var procesados int64
+var procesados atomic.Int64
 
 // Etiquetas convierte "a,b,c" en una lista sin espacios.
 func Etiquetas(s string) []string {
@@ -32,20 +32,15 @@
 
 // Cliente saca el cliente de una línea "cliente=ana".
 func Cliente(linea string) (string, bool) {
-	if strings.HasPrefix(linea, "cliente=") {
-		return strings.TrimPrefix(linea, "cliente="), true
+	if after, ok := strings.CutPrefix(linea, "cliente="); ok {
+		return after, true
 	}
 	return "", false
 }
 
 // EsVIP dice si el cliente está en la lista.
 func EsVIP(vips []string, cliente string) bool {
-	for _, v := range vips {
-		if v == cliente {
-			return true
-		}
-	}
-	return false
+	return slices.Contains(vips, cliente)
 }
 
 // Copiar pasa los totales de un mapa a otro.
@@ -57,7 +46,7 @@
 
 // Ordenar ordena los totales de menor a mayor.
 func Ordenar(totales []int) {
-	sort.Slice(totales, func(i, j int) bool { return totales[i] < totales[j] })
+	slices.Sort(totales)
 }
 
 // Reintentos devuelve los tiempos de espera: 100 ms, 200 ms, 400 ms...
@@ -79,7 +68,7 @@
 		go func() {
 			defer wg.Done()
 			res[i] = f(p)
-			atomic.AddInt64(&procesados, 1)
+			procesados.Add(1)
 		}()
 	}
 	wg.Wait()
@@ -87,7 +76,7 @@
 }
 
 // Procesados cuenta los pedidos procesados desde que arrancó.
-func Procesados() int64 { return atomic.LoadInt64(&procesados) }
+func Procesados() int64 { return procesados.Load() }
 
 // RutaFallida saca la ruta de un error de archivo.
 func RutaFallida(err error) (string, bool) {
@@ -99,6 +88,6 @@
 }
 
 // Campos devuelve los datos de un pedido como mapa genérico.
-func Campos(p Pedido) map[string]interface{} {
-	return map[string]interface{}{"id": p.ID, "total": p.Total}
+func Campos(p Pedido) map[string]any {
+	return map[string]any{"id": p.ID, "total": p.Total}
 }
[código de salida: 1]

## 02-go-fix-inline · //go:fix inline
$ cat viejo/viejo.go
// Paquete de una biblioteca interna que cambió de API.
package viejo

import "strings"

// Normalizar es la función nueva.
func Normalizar(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// Limpiar se queda por compatibilidad.
//
// Deprecated: usa Normalizar.
//
//go:fix inline
func Limpiar(s string) string { return Normalizar(s) }
[código de salida: 0]
$ go run ./app
ana@ejemplo.com
luis
[código de salida: 0]
$ go fix -diff ./...
--- …/02-go-fix-inline/app/main.go (old)
+++ …/02-go-fix-inline/app/main.go (new)
@@ -8,6 +8,6 @@
 )
 
 func main() {
-	fmt.Println(viejo.Limpiar("  ANA@Ejemplo.com "))
-	fmt.Println(viejo.Limpiar(" Luis "))
+	fmt.Println(viejo.Normalizar("  ANA@Ejemplo.com "))
+	fmt.Println(viejo.Normalizar(" Luis "))
 }
[código de salida: 1]
$ go fix ./...
[código de salida: 0]
$ cat app/main.go
// Código de un equipo que usa la biblioteca: llama a la función vieja.
package main

import (
	"fmt"

	"demo/viejo"
)

func main() {
	fmt.Println(viejo.Normalizar("  ANA@Ejemplo.com "))
	fmt.Println(viejo.Normalizar(" Luis "))
}
[código de salida: 0]
$ go run ./app
ana@ejemplo.com
luis
[código de salida: 0]

## 03-go-tool-fix-help · cuántos analizadores
$ go tool fix help | grep -cE '^    [a-z]+ '
26
```

# Avisos de go fix por analizador (go vet -vettool=fix -json)

- go.mod go 1.27: 13 avisos en 12 analizadores: {'any': 2, 'atomictypes': 1, 'errorsastype': 1, 'forvar': 1, 'mapsloop': 1, 'omitzero': 1, 'rangeint': 1, 'slicescontains': 1, 'slicessort': 1, 'stringscutprefix': 1, 'stringsseq': 1, 'waitgroupgo': 1}
- go.mod go 1.21: 6 avisos en 5 analizadores: {'any': 2, 'atomictypes': 1, 'slicescontains': 1, 'slicessort': 1, 'stringscutprefix': 1}
- Solo con go 1.27: ['errorsastype', 'forvar', 'mapsloop', 'omitzero', 'rangeint', 'stringsseq', 'waitgroupgo']

# Fuentes con fecha

- go.dev/blog/gofix (Alan Donovan, 17-feb-2026): copia en recursos/pruebas-ia-escribe-go-viejo/fuente-go-blog-gofix.txt.
- go.dev/doc/go1.27 (bajada el 2026-10-29): recursos/novedades-go-1-27/fuente-go1.27.txt.

# Reutilizado (copias literales)


### De `recursos/novedades-go-1-27/SALIDA.md`

```
## 01-metodos-genericos · go1.27.2
$ go run .
puerto: 8080 true
host: localhost true
puerto como string: false
r.N(10 * time.Second): 7.693732693s
[código de salida: 0]
```

### De `recursos/novedades-go-1-27/SALIDA.md`

```
## 01-metodos-genericos · go1.26.10
$ go run .
# demo
./main.go:13:20: syntax error: method must have no type parameters
[código de salida: 1]
```

### De `recursos/novedades-go-1-27/SALIDA.md`

```
## Go 1.27 con «go 1.26» en go.mod (¿dependen de esa línea?)
$ go run .
# demo
./main.go:13:21: generic method requires go1.27 or later (-lang was set to go1.26; check go.mod)
[código de salida: 1]
```

### De `recursos/novedades-go-1-27/SALIDA.md`

```
## 06-goroutineleak · go1.27.2
$ go run .
buscar: timeout
buscar: timeout
buscar: timeout
goroutineleak profile: total 3
3 @ 0x47f06a 0x41421c 0x413e17 0x4e0776 0x484ec1
#	0x4e0775	main.buscar.func1+0x35	…/06-goroutineleak/main.go:17

[código de salida: 0]
# Go 1.27 con el experimento viejo:
$ env GOEXPERIMENT=goroutineleakprofile go run .
go: unknown GOEXPERIMENT goroutineleakprofile
[código de salida: 2]
```

### De `recursos/novedades-go-1-27/SALIDA.md`

```
## 06-goroutineleak · go1.26.10
$ go run .
buscar: timeout
buscar: timeout
buscar: timeout
pprof.Lookup("goroutineleak") = nil: este Go no tiene el perfil
[código de salida: 0]
```

### De `recursos/novedades-go-1-27/SALIDA.md`

```
| BenchmarkGzipDefault | 90664229.00 | 23514894.00 | — | -74.1 % | — |
```

### De `recursos/pruebas-ia-escribe-go-viejo/SALIDA.md`

```
## Tabla principal (por modelo y prompt)
| modelo | prompt | muestras | compila | go vet limpio* | con ≥1 aviso de go fix* | avisos* | usa el modismo nuevo de su tarea* |
|---|---|---|---|---|---|---|---|
| claude-code | A | 15 | 15/15 (100 %) | 15/15 (100 %) | 5/15 (33 %) | 5 | 4/15 (27 %) |
| claude-code | B | 15 | 15/15 (100 %) | 15/15 (100 %) | 1/15 (7 %) | 1 | 13/15 (87 %) |
| qwen2.5-coder_7b | A | 72 | 57/72 (79 %) | 56/57 (98 %) | 28/57 (49 %) | 33 | 0/57 (0 %) |
| qwen2.5-coder_7b | B | 72 | 56/72 (78 %) | 55/56 (98 %) | 27/56 (48 %) | 32 | 1/56 (2 %) |
| qwen2.5-coder_7b | C | 72 | 24/72 (33 %) | 24/24 (100 %) | 5/24 (21 %) | 5 | 11/24 (46 %) |

*sobre las que compilan.
```

### De `recursos/pruebas-ia-escribe-go-viejo/SALIDA.md`

```
## ¿Cambió algo la frase «Usa Go 1.27»? (mismo modelo, misma tarea, misma seed)
- qwen2.5-coder_7b: 46 de 72 respuestas B idénticas byte a byte a su A.
- claude-code: 2 de 15 respuestas B idénticas byte a byte a su A.
```

### De `recursos/pruebas-ia-goroutines-que-no-terminan/SALIDA.md`

```
## Tabla principal (por modelo y prompt)
| modelo | prompt | muestras | compila | go vet limpio* | termina la prueba | correctas | con fuga | goroutines fugadas | con data race* | sin ningún problema |
|---|---|---|---|---|---|---|---|---|---|---|
| claude-code | A | 12 | 12 | 12/12 | 12 | 12 | 0 | 0 | 0/12 | 12/12 |
| claude-code | B | 12 | 12 | 12/12 | 12 | 12 | 0 | 0 | 0/12 | 12/12 |
| qwen2.5-coder_7b | A | 48 | 37 | 37/37 | 30 | 26 | 7 | 21 | 5/37 | 19/48 |
| qwen2.5-coder_7b | B | 48 | 33 | 33/33 | 27 | 22 | 4 | 9 | 6/33 | 17/48 |

*sobre las que compilan.
```

### De `recursos/pruebas-ia-goroutines-que-no-terminan/SALIDA.md`

```
goroutineleak profile: total 2
1 @ 0x4866ea 0x41531c 0x414f17 0x545aac 0x48cce1
#	0x545aab	tarea.ElMasRapido.func1+0x6b	/w/m/m0138/tarea.go:15

1 @ 0x4866ea 0x461c32 0x461c09 0x487bae 0x491ca5 0x5459e5 0x48cce1
#	0x487bad	sync.runtime_SemacquireWaitGroup+0x2d	/usr/local/go/src/runtime/sema.go:114
#	0x491ca4	sync.(*WaitGroup).Wait+0x84		/usr/local/go/src/sync/waitgroup.go:206
#	0x5459e4	tarea.ElMasRapido.func2+0x24		/w/m/m0138/tarea.go:20
```

### De `recursos/pruebas-json-v2/SALIDA.md`

```
1 · Slice y map sin inicializar (nil):
  v1  {"nombre":"ana","roles":null,"extra":null}
  v2  {"nombre":"ana","roles":[],"extra":{},"edad":0}

2 · Mayúsculas distintas en la clave ({"NOMBRE":"ana"}):
  v1  nombre="ana"
  v2  nombre=""

3 · Clave repetida ({"nombre":"ana","nombre":"admin"}):
  v1  nombre="admin"
  v2  error: jsontext: duplicate object member name "nombre"
```

### De `recursos/pruebas-json-v2/SALIDA.md`

```
$ GOEXPERIMENT=nojsonv2 go test -bench . -benchmem -count 5   # encoding/json con su implementación vieja
  BenchmarkMarshalSinMapas_v1API       1.29 ms/op   1.04 MB/op       2 allocs/op
  BenchmarkMarshal_v1API               2.92 ms/op   2.93 MB/op   50002 allocs/op
  BenchmarkUnmarshal_v1API            13.55 ms/op   8.83 MB/op  150026 allocs/op

$ go test -bench . -benchmem -count 5   # Go 1.27 por defecto: encoding/json ya va sobre v2
  BenchmarkMarshalSinMapas_v1API       2.08 ms/op   1.06 MB/op       3 allocs/op
  BenchmarkMarshalSinMapas_v2API       2.19 ms/op   1.04 MB/op       3 allocs/op
  BenchmarkMarshal_v1API               4.69 ms/op   2.13 MB/op   40004 allocs/op
  BenchmarkMarshal_v2API               4.22 ms/op   1.84 MB/op   20003 allocs/op
  BenchmarkUnmarshal_v1API            10.03 ms/op   8.44 MB/op   83191 allocs/op
  BenchmarkUnmarshal_v2API             8.58 ms/op   8.44 MB/op   83191 allocs/op
```
