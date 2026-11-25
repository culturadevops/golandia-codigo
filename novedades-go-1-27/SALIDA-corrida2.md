# SALIDA — Lo nuevo de Go 1.27, ejecutado (golandia 700)

> Corrida del 2026-10-29 07:42. Contenedores golang:1.27.2 y golang:1.26 con --cpus 2. Notas oficiales: fuente-go1.27.html / fuente-go1.27.txt (go.dev/doc/go1.27, bajadas el 2026-10-29).


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
New():   17c75e7d-98ea-4766-86ed-2ce73ec0afce (versión 4)
NewV7(): 01a187ae-b814-711a-a31b-742c5dc22cd9
NewV7(): 01a187ae-b814-711b-a1d7-1c2267ad66aa
NewV7(): 01a187ae-b814-711c-ad07-a448ac904bf4
NewV7(): 01a187ae-b814-711d-aff2-0f060c764417
NewV7(): 01a187ae-b814-711e-8f1d-15921f238e60
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

goroutine 7 [chan receive] {cliente: acme, ruta: /api/pedidos}:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x5f

goroutine 8 [chan receive] {cliente: globex, ruta: /api/pagos}:
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

## 10-flate · go1.27.2
$ go test -run TestTamano -v .
=== RUN   TestTamano
    flate_test.go:37: nivel 1: 4194323 -> 1200731 bytes
    flate_test.go:37: nivel -1: 4194323 -> 1064217 bytes
    flate_test.go:37: nivel 9: 4194323 -> 932388 bytes
--- PASS: TestTamano (0.29s)
PASS
ok  	demo	0.320s
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

goroutine 7 [chan receive]:
main.atender.func1.1()
	…/11-traceback-etiquetas/main.go:13 +0x19
created by main.atender.func1 in goroutine 1
	…/11-traceback-etiquetas/main.go:13 +0x56

goroutine 8 [chan receive]:
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
--- PASS: TestTamano (0.36s)
PASS
ok  	demo	0.392s
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
BenchmarkAlloc16-2    	147842214	         8.105 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	151219459	         7.874 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	154521024	         7.790 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	151308252	         7.908 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	149514831	         7.970 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	126844642	         9.425 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	127515717	         9.459 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	120086500	        10.08 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.13 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.27 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	73007168	        14.80 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	85594690	        14.74 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	72462997	        15.21 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	84841754	        14.88 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	81281086	        14.73 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	54222757	        24.27 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	55373959	        24.46 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	47492030	        23.45 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	48028621	        23.78 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	51455007	        24.62 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	23.742s
[código de salida: 0]

## 10-flate · go1.26.10
$ go test -run ^$ -bench Gzip -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkGzipDefault-2     	      12	  92858442 ns/op	  45.17 MB/s
BenchmarkGzipDefault-2     	      13	  92459629 ns/op	  45.36 MB/s
BenchmarkGzipDefault-2     	      12	  92329954 ns/op	  45.43 MB/s
BenchmarkGzipDefault-2     	      13	  91593878 ns/op	  45.79 MB/s
BenchmarkGzipDefault-2     	      12	  92770190 ns/op	  45.21 MB/s
BenchmarkGzipBestSpeed-2   	      58	  18829928 ns/op	 222.75 MB/s
BenchmarkGzipBestSpeed-2   	      60	  19395379 ns/op	 216.25 MB/s
BenchmarkGzipBestSpeed-2   	      60	  19482166 ns/op	 215.29 MB/s
BenchmarkGzipBestSpeed-2   	      60	  18987592 ns/op	 220.90 MB/s
BenchmarkGzipBestSpeed-2   	      57	  19419893 ns/op	 215.98 MB/s
PASS
ok  	demo	11.435s
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
BenchmarkAlloc16-2    	197216301	         6.086 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	197190141	         6.070 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	193113739	         6.157 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	193088496	         6.259 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	196754022	         6.075 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	143311539	         8.438 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	144355573	         8.599 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	132071894	         8.961 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	135723745	         8.778 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	139988323	         8.579 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	97525136	        13.94 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	89433711	        13.78 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	93556658	        14.22 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	96165589	        13.82 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	90082927	        13.85 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	47532357	        24.00 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	46667596	        23.87 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	46528880	        23.96 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	52622000	        23.91 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	53065516	        23.32 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	24.390s
[código de salida: 0]
# Go 1.27 sin las rutinas nuevas:
$ env GOEXPERIMENT=nosizespecializedmalloc go test -run ^$ -bench Alloc -benchmem -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkAlloc16-2    	150827156	         8.131 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	153036702	         7.907 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	144779773	         8.186 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	147111398	         8.201 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	137495881	         8.686 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.97 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	111321026	        10.89 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        11.38 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        11.09 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        11.03 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	76008301	        15.07 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	82161993	        14.99 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	83071278	        14.42 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	78048622	        14.88 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	74551044	        15.42 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	52077558	        24.61 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	40079398	        25.35 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	47717472	        27.16 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	45164977	        28.56 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	42937314	        28.46 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	23.685s
[código de salida: 0]

## 10-flate · go1.27.2
$ go test -run ^$ -bench Gzip -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkGzipDefault-2     	      40	  26776833 ns/op	 156.64 MB/s
BenchmarkGzipDefault-2     	      45	  25500614 ns/op	 164.48 MB/s
BenchmarkGzipDefault-2     	      46	  25306858 ns/op	 165.74 MB/s
BenchmarkGzipDefault-2     	      44	  25866248 ns/op	 162.15 MB/s
BenchmarkGzipDefault-2     	      50	  24127894 ns/op	 173.84 MB/s
BenchmarkGzipBestSpeed-2   	      98	  12522284 ns/op	 334.95 MB/s
BenchmarkGzipBestSpeed-2   	      91	  12530551 ns/op	 334.73 MB/s
BenchmarkGzipBestSpeed-2   	      87	  12792634 ns/op	 327.87 MB/s
BenchmarkGzipBestSpeed-2   	      85	  12565985 ns/op	 333.78 MB/s
BenchmarkGzipBestSpeed-2   	      92	  12304877 ns/op	 340.87 MB/s
PASS
ok  	demo	11.446s
[código de salida: 0]
```

# Medianas de 5 pasadas (ns/op; flate en ns/op por 4 MB)

| benchmark | Go 1.26 | Go 1.27 | Go 1.27 nosizespecializedmalloc | 1.27 vs 1.26 | 1.27 vs nosize |
|---|---|---|---|---|---|
| BenchmarkAlloc128 | 24.27 | 23.91 | 27.16 | -1.5 % | -12.0 % |
| BenchmarkAlloc16 | 7.91 | 6.09 | 8.19 | -23.0 % | -25.7 % |
| BenchmarkAlloc32 | 10.08 | 8.60 | 11.03 | -14.7 % | -22.0 % |
| BenchmarkAlloc64 | 14.80 | 13.85 | 14.99 | -6.4 % | -7.6 % |
| BenchmarkGzipBestSpeed | 19395379.00 | 12530551.00 | — | -35.4 % | — |
| BenchmarkGzipDefault | 92459629.00 | 25500614.00 | — | -72.4 % | — |
