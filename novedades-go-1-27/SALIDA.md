# SALIDA — Lo nuevo de Go 1.27, ejecutado (golandia 700)

> Corrida del 2026-10-29 07:47. Contenedores golang:1.27.2 y golang:1.26 con --cpus 2. Notas oficiales: fuente-go1.27.html / fuente-go1.27.txt (go.dev/doc/go1.27, bajadas el 2026-10-29).


# Demos · golang:1.27.2

```
go version: go version go1.27.2 linux/amd64

## 01-metodos-genericos · go1.27.2
$ go run .
puerto: 8080 true
host: localhost true
puerto como string: false
r.N(10 * time.Second): 7.693732693s
[código de salida: 0]

## 02-campos-promovidos · go1.27.2
$ go run .
{Base:{ID:7 Creado:2026-10-29} Nombre:ana}
[código de salida: 0]

## 03-inferencia · go1.27.2
$ go run .
conversión: 9
slice: 1.5
struct: rust MAYOR
[código de salida: 0]

## 04-uuid · go1.27.2
$ go run .
New():   5bb7761a-6e39-40cb-acc0-b7eaa154e23d (versión 4)
NewV7(): 01a187b3-f69c-7508-93d6-888b22761aa4
NewV7(): 01a187b3-f69c-750a-95f1-2c9a758bd22c
NewV7(): 01a187b3-f69c-750b-84e5-19c581329c71
NewV7(): 01a187b3-f69c-750c-b512-9ee672087032
NewV7(): 01a187b3-f69c-750d-a656-7c7007d914a5
¿v7 ya ordenados? true
Parse con llaves y mayúsculas: f81d4fae-7dec-11d0-a765-00a0c91e6bf6 <nil>
Parse inválido: invalid uuid
¿comparables con ==? true
[código de salida: 0]

## 05-cutlast · go1.27.2
$ go run .
Cut:     "respaldo" "2026-10-29.tar.gz" true
CutLast: "respaldo.2026-10-29.tar" "gz" true
LastIndex: "respaldo.2026-10-29.tar" "gz"
IPv6: host="[::1]" puerto="8080"
bytes.CutLast: "a/b" "c" true
sin separador: "sin-punto" "" false
[código de salida: 0]

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

## 08-vet-w-puntero · go1.27.2
$ go run .
buscar: no encontrado: usuario:42
errors.As(err, &NoEncontrado): false
[código de salida: 0]
$ go vet .
main.go:16:29: %w wants operand of error type NoEncontrado, not pointer type *NoEncontrado (defeats errors.Is)
[código de salida: 1]

## 09-asynctimerchan · go1.27.2
$ cat go.mod
module demo

go 1.27
[código de salida: 0]
$ go run .
cap(t.C) = 0
[código de salida: 0]
$ cat go.mod
module demo

go 1.27

godebug asynctimerchan=1
[código de salida: 0]
$ go run .
go: error loading go.mod:
go.mod:5: removed GODEBUG "asynctimerchan" set to old value "1" (https://go.dev/doc/godebug#go-127)
[código de salida: 1]
$ cat go.mod
module demo

go 1.27

godebug asynctimerchan=0
[código de salida: 0]
$ go run .
cap(t.C) = 0
[código de salida: 0]

## 11-traceback-etiquetas · go1.27.2
$ cat go.mod
module demo

go 1.27
[código de salida: 0]
$ env GOTRACEBACK=all …/tb
panic: algo salió mal

goroutine 1 [running]:
main.main()
	…/11-traceback-etiquetas/main.go:22 +0x89

goroutine 19 [chan receive] {cliente: acme, ruta: /api/pedidos}:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x5f

goroutine 20 [chan receive] {cliente: globex, ruta: /api/pagos}:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x5f
[código de salida: 2]
# El mismo programa con GODEBUG=tracebacklabels=0:
$ env GOTRACEBACK=all GODEBUG=tracebacklabels=0 …/tb
panic: algo salió mal

goroutine 1 [running]:
main.main()
	…/11-traceback-etiquetas/main.go:22 +0x89

goroutine 19 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x5f

goroutine 20 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x5f
[código de salida: 2]

## Go 1.27 con «go 1.26» en go.mod (¿dependen de esa línea?)
$ go run .
# demo
./main.go:13:21: generic method requires go1.27 or later (-lang was set to go1.26; check go.mod)
[código de salida: 1]
$ env GOTRACEBACK=all …/tb126
panic: algo salió mal

goroutine 1 [running]:
main.main()
	…/11-traceback-etiquetas/main.go:22 +0x89

goroutine 19 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x5f

goroutine 20 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x5f
[código de salida: 2]

## 10-flate · go1.27.2
$ go test -run TestTamano -v .
=== RUN   TestTamano
    flate_test.go:37: nivel 1: 4194323 -> 1200731 bytes
    flate_test.go:37: nivel -1: 4194323 -> 1064217 bytes
    flate_test.go:37: nivel 9: 4194323 -> 932388 bytes
--- PASS: TestTamano (0.30s)
PASS
ok  	demo	0.326s
[código de salida: 0]

## 07-malloc · go1.27.2
binario hola (por defecto): 2342585 bytes
binario hola (GOEXPERIMENT=nosizespecializedmalloc): 2301248 bytes
```

# Demos · golang:1.26

```
go version: go version go1.26.10 linux/amd64

## 01-metodos-genericos · go1.26.10
$ go run .
# demo
./main.go:13:20: syntax error: method must have no type parameters
[código de salida: 1]

## 02-campos-promovidos · go1.26.10
$ go run .
# demo
./main.go:18:15: unknown field ID in struct literal of type Usuario
./main.go:18:22: unknown field Creado in struct literal of type Usuario
[código de salida: 1]

## 03-inferencia · go1.26.10
$ go run .
# demo
./main.go:24:28: cannot use generic function Mayor without instantiation
./main.go:28:42: cannot use generic function Mayor without instantiation
./main.go:32:38: cannot use generic function Mayor without instantiation
[código de salida: 1]

## 04-uuid · go1.26.10
$ go run .
main.go:7:2: package uuid is not in std (/usr/local/go/src/uuid)
[código de salida: 1]

## 05-cutlast · go1.26.10
$ go run .
# demo
./main.go:16:31: undefined: strings.CutLast
./main.go:24:29: undefined: strings.CutLast
./main.go:27:20: undefined: bytes.CutLast
./main.go:30:25: undefined: strings.CutLast
[código de salida: 1]

## 06-goroutineleak · go1.26.10
$ go run .
buscar: timeout
buscar: timeout
buscar: timeout
pprof.Lookup("goroutineleak") = nil: este Go no tiene el perfil
[código de salida: 0]
# Go 1.26 con el experimento:
$ env GOEXPERIMENT=goroutineleakprofile go run .
buscar: timeout
buscar: timeout
buscar: timeout
goroutineleak profile: total 3
3 @ 0x47f8ee 0x41491c 0x414517 0x4dbcf6 0x485ee1
#	0x4dbcf5	main.buscar.func1+0x35	…/06-goroutineleak/main.go:17

[código de salida: 0]

## 08-vet-w-puntero · go1.26.10
$ go run .
buscar: no encontrado: usuario:42
errors.As(err, &NoEncontrado): false
[código de salida: 0]
$ go vet .
[código de salida: 0]

## 09-asynctimerchan · go1.26.10
$ cat go.mod
module demo

go 1.26
[código de salida: 0]
$ go run .
cap(t.C) = 0
[código de salida: 0]
$ cat go.mod
module demo

go 1.26

godebug asynctimerchan=1
[código de salida: 0]
$ go run .
cap(t.C) = 1
[código de salida: 0]
$ cat go.mod
module demo

go 1.26

godebug asynctimerchan=0
[código de salida: 0]
$ go run .
cap(t.C) = 0
[código de salida: 0]

## 11-traceback-etiquetas · go1.26.10
$ cat go.mod
module demo

go 1.26
[código de salida: 0]
$ env GOTRACEBACK=all …/tb
panic: algo salió mal

goroutine 1 [running]:
main.main()
	…/11-traceback-etiquetas/main.go:22 +0x89

goroutine 19 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x56

goroutine 20 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x56
[código de salida: 2]
# El mismo programa con GODEBUG=tracebacklabels=0:
$ env GOTRACEBACK=all GODEBUG=tracebacklabels=0 …/tb
panic: algo salió mal

goroutine 1 [running]:
main.main()
	…/11-traceback-etiquetas/main.go:22 +0x89

goroutine 19 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x56

goroutine 20 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x56
[código de salida: 2]

## 10-flate · go1.26.10
$ go test -run TestTamano -v .
=== RUN   TestTamano
    flate_test.go:37: nivel 1: 4194323 -> 1241894 bytes
    flate_test.go:37: nivel -1: 4194323 -> 987875 bytes
    flate_test.go:37: nivel 9: 4194323 -> 949822 bytes
--- PASS: TestTamano (0.39s)
PASS
ok  	demo	0.416s
[código de salida: 0]

## 07-malloc · go1.26.10
binario hola (por defecto): 2397438 bytes
```

# Bancos · golang:1.26 (salida cruda)

```
go version: go version go1.26.10 linux/amd64 · nproc=24

## 07-malloc · go1.26.10
$ go test -run ^$ -bench Alloc -benchmem -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkAlloc16-2    	155138625	         7.666 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	161342732	         7.506 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	151775647	         7.932 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	155014970	         7.683 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	154700215	         7.834 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	128725718	         9.295 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	131645474	         9.213 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	122869768	         9.717 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	131050444	         9.167 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	128318954	         9.449 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	81957164	        13.48 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	88588357	        13.74 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	87544119	        13.87 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	87387321	        13.32 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	93551925	        13.73 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	57491270	        23.04 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	51798937	        23.34 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	51601568	        22.75 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	52178758	        23.08 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	50553914	        22.95 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	24.092s
[código de salida: 0]

## 10-flate · go1.26.10
$ go test -run ^$ -bench Gzip -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkGzipDefault-2     	      13	  90719440 ns/op	  46.23 MB/s
BenchmarkGzipDefault-2     	      13	  92607719 ns/op	  45.29 MB/s
BenchmarkGzipDefault-2     	      13	  88415480 ns/op	  47.44 MB/s
BenchmarkGzipDefault-2     	      13	  90101304 ns/op	  46.55 MB/s
BenchmarkGzipDefault-2     	      13	  90664229 ns/op	  46.26 MB/s
BenchmarkGzipBestSpeed-2   	      60	  18110384 ns/op	 231.60 MB/s
BenchmarkGzipBestSpeed-2   	      61	  17796306 ns/op	 235.69 MB/s
BenchmarkGzipBestSpeed-2   	      62	  17699871 ns/op	 236.97 MB/s
BenchmarkGzipBestSpeed-2   	      63	  17833511 ns/op	 235.19 MB/s
BenchmarkGzipBestSpeed-2   	      62	  17808761 ns/op	 235.52 MB/s
PASS
ok  	demo	11.411s
[código de salida: 0]
```

# Bancos · golang:1.27.2 (salida cruda)

```
go version: go version go1.27.2 linux/amd64 · nproc=24

## 07-malloc · go1.27.2
$ go test -run ^$ -bench Alloc -benchmem -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkAlloc16-2    	181647136	         6.330 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	211237460	         5.731 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	198406784	         6.137 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	199592366	         5.947 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	205772804	         5.977 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	153554725	         7.863 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	146604753	         8.172 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	148639075	         8.067 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	149012428	         8.006 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	146956590	         8.147 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	89394024	        13.50 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	91782147	        13.67 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	87372591	        13.19 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	92516900	        12.93 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	88789180	        13.22 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	52775551	        23.44 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	49122493	        22.39 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	57787339	        21.28 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	48294134	        22.74 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	60952467	        23.11 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	24.057s
[código de salida: 0]
# Go 1.27 sin las rutinas nuevas:
$ env GOEXPERIMENT=nosizespecializedmalloc go test -run ^$ -bench Alloc -benchmem -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkAlloc16-2    	151553061	         7.897 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	149769896	         7.877 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	155597406	         7.802 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	155215874	         7.733 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	154558698	         7.860 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	123914575	         9.623 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	124206074	         9.701 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	124549046	         9.742 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	121652412	         9.959 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	124233943	         9.602 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	90789770	        13.75 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	92069642	        14.73 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	89279124	        14.32 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	72040650	        14.35 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	77185616	        13.91 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	54745228	        22.07 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	56583314	        22.49 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	49888440	        22.61 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	49714647	        22.58 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	53070530	        21.96 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	23.920s
[código de salida: 0]

## 10-flate · go1.27.2
$ go test -run ^$ -bench Gzip -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkGzipDefault-2     	      48	  22946622 ns/op	 182.79 MB/s
BenchmarkGzipDefault-2     	      51	  23421856 ns/op	 179.08 MB/s
BenchmarkGzipDefault-2     	      49	  23514894 ns/op	 178.37 MB/s
BenchmarkGzipDefault-2     	      49	  24077008 ns/op	 174.20 MB/s
BenchmarkGzipDefault-2     	      49	  24167137 ns/op	 173.55 MB/s
BenchmarkGzipBestSpeed-2   	      88	  12152578 ns/op	 345.14 MB/s
BenchmarkGzipBestSpeed-2   	      97	  12382936 ns/op	 338.72 MB/s
BenchmarkGzipBestSpeed-2   	      85	  12441984 ns/op	 337.11 MB/s
BenchmarkGzipBestSpeed-2   	      84	  12476808 ns/op	 336.17 MB/s
BenchmarkGzipBestSpeed-2   	      88	  12182871 ns/op	 344.28 MB/s
PASS
ok  	demo	11.291s
[código de salida: 0]
```

# Medianas de 5 pasadas (ns/op; flate en ns/op por 4 MB)

| benchmark | Go 1.26 | Go 1.27 | Go 1.27 nosizespecializedmalloc | 1.27 vs 1.26 | 1.27 vs nosize |
|---|---|---|---|---|---|
| BenchmarkAlloc128 | 23.04 | 22.74 | 22.49 | -1.3 % | +1.1 % |
| BenchmarkAlloc16 | 7.68 | 5.98 | 7.86 | -22.2 % | -24.0 % |
| BenchmarkAlloc32 | 9.29 | 8.07 | 9.70 | -13.2 % | -16.8 % |
| BenchmarkAlloc64 | 13.73 | 13.22 | 14.32 | -3.7 % | -7.7 % |
| BenchmarkGzipBestSpeed | 17808761.00 | 12382936.00 | — | -30.5 % | — |
| BenchmarkGzipDefault | 90664229.00 | 23514894.00 | — | -74.1 % | — |
