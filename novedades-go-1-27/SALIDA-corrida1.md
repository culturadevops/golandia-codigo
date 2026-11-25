# SALIDA — Lo nuevo de Go 1.27, ejecutado (golandia 700)

> Corrida del 2026-10-29 07:39. Contenedores golang:1.27.2 y golang:1.26 con --cpus 2. Notas oficiales: fuente-go1.27.html / fuente-go1.27.txt (go.dev/doc/go1.27, bajadas el 2026-10-29).


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
New():   9d5c14e6-295a-401a-9e26-500a6160f921 (versión 4)
NewV7(): 01a187ac-4452-79aa-8665-3f901f7fb8d6
NewV7(): 01a187ac-4452-79b2-8eb2-89ba43a6c188
NewV7(): 01a187ac-4452-79b3-b1b6-9d1679f43d54
NewV7(): 01a187ac-4452-79b4-8f4c-5d4504444c1e
NewV7(): 01a187ac-4452-79b5-9bab-64726047283e
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
$ env GOTRACEBACK=all /tmp/tmp.kY0HnzwAwG/tb
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
$ env GOTRACEBACK=all GODEBUG=tracebacklabels=0 /tmp/tmp.kY0HnzwAwG/tb
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
--- PASS: TestTamano (0.28s)
PASS
ok  	demo	0.310s
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
$ env GOTRACEBACK=all /tmp/tmp.rwuTZQUUFs/tb
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
$ env GOTRACEBACK=all GODEBUG=tracebacklabels=0 /tmp/tmp.rwuTZQUUFs/tb
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
ok  	demo	0.389s
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
BenchmarkAlloc16-2    	154863608	         7.715 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	153037916	         7.907 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	149642918	         7.943 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	152697430	         8.000 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	152401725	         7.938 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.43 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.20 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	121524434	         9.847 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.07 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	120794523	         9.888 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	91249852	        13.97 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	84786280	        14.37 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	91468886	        13.87 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	99745963	        12.88 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	80897018	        13.55 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	52403551	        20.90 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	63742676	        20.60 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	59233695	        22.07 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	55501531	        22.05 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	52431909	        23.08 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	23.787s
[código de salida: 0]

## 10-flate · go1.26.10
$ go test -run ^$ -bench Gzip -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkGzipDefault-2     	      13	  87839358 ns/op	  47.75 MB/s
BenchmarkGzipDefault-2     	      13	  89121564 ns/op	  47.06 MB/s
BenchmarkGzipDefault-2     	      13	  87458040 ns/op	  47.96 MB/s
BenchmarkGzipDefault-2     	      13	  88618574 ns/op	  47.33 MB/s
BenchmarkGzipDefault-2     	      12	  87690494 ns/op	  47.83 MB/s
BenchmarkGzipBestSpeed-2   	      62	  17898547 ns/op	 234.34 MB/s
BenchmarkGzipBestSpeed-2   	      61	  18227513 ns/op	 230.11 MB/s
BenchmarkGzipBestSpeed-2   	      63	  17812062 ns/op	 235.48 MB/s
BenchmarkGzipBestSpeed-2   	      64	  18149960 ns/op	 231.09 MB/s
BenchmarkGzipBestSpeed-2   	      64	  17797255 ns/op	 235.67 MB/s
PASS
ok  	demo	11.318s
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
BenchmarkAlloc16-2    	204084753	         5.848 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	208575685	         5.764 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	196360892	         6.018 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	199857978	         6.008 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	196290610	         6.015 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	145795718	         8.305 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	143388921	         8.441 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	142071897	         8.453 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	147454881	         8.339 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	138914602	         8.617 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	89637085	        13.36 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	83887392	        12.69 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	87764790	        13.06 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	92466164	        12.91 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	95179324	        13.56 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	49438951	        24.59 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	54960835	        24.95 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	49815556	        24.00 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	50857225	        24.81 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	48114861	        23.96 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	24.107s
[código de salida: 0]
# Go 1.27 sin las rutinas nuevas:
$ env GOEXPERIMENT=nosizespecializedmalloc go test -run ^$ -bench Alloc -benchmem -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkAlloc16-2    	153277164	         7.947 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	147782973	         8.146 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	148588674	         8.009 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	148705660	         8.027 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc16-2    	145123148	         8.187 ns/op	      16 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.10 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.04 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	121493593	         9.991 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	120459508	         9.948 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc32-2    	100000000	        10.07 ns/op	      32 B/op	       1 allocs/op
BenchmarkAlloc64-2    	83953633	        14.13 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	92913301	        14.39 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	97444470	        14.13 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	85137333	        14.29 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc64-2    	72573517	        14.55 ns/op	      64 B/op	       1 allocs/op
BenchmarkAlloc128-2   	51386378	        22.96 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	45352239	        24.93 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	51471549	        27.33 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	35651836	        31.38 ns/op	     128 B/op	       1 allocs/op
BenchmarkAlloc128-2   	42977011	        26.69 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	demo	23.593s
[código de salida: 0]

## 10-flate · go1.27.2
$ go test -run ^$ -bench Gzip -count 5 .
goos: linux
goarch: amd64
pkg: demo
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkGzipDefault-2     	      48	  23739723 ns/op	 176.68 MB/s
BenchmarkGzipDefault-2     	      50	  24846785 ns/op	 168.81 MB/s
BenchmarkGzipDefault-2     	      49	  25417134 ns/op	 165.02 MB/s
BenchmarkGzipDefault-2     	      45	  26643602 ns/op	 157.42 MB/s
BenchmarkGzipDefault-2     	      46	  24352292 ns/op	 172.24 MB/s
BenchmarkGzipBestSpeed-2   	      87	  12848656 ns/op	 326.44 MB/s
BenchmarkGzipBestSpeed-2   	      79	  13032834 ns/op	 321.83 MB/s
BenchmarkGzipBestSpeed-2   	      84	  12469404 ns/op	 336.37 MB/s
BenchmarkGzipBestSpeed-2   	      87	  13146066 ns/op	 319.06 MB/s
BenchmarkGzipBestSpeed-2   	      86	  12843416 ns/op	 326.57 MB/s
PASS
ok  	demo	11.423s
[código de salida: 0]
```

# Medianas de 5 pasadas (ns/op; flate en ns/op por 4 MB)

| benchmark | Go 1.26 | Go 1.27 | Go 1.27 nosizespecializedmalloc | 1.27 vs 1.26 | 1.27 vs nosize |
|---|---|---|---|---|---|
| BenchmarkAlloc128 | 22.05 | 24.59 | 26.69 | +11.5 % | -7.9 % |
| BenchmarkAlloc16 | 7.94 | 6.01 | 8.03 | -24.3 % | -25.2 % |
| BenchmarkAlloc32 | 10.07 | 8.44 | 10.04 | -16.2 % | -15.9 % |
| BenchmarkAlloc64 | 13.87 | 13.06 | 14.29 | -5.8 % | -8.6 % |
| BenchmarkGzipBestSpeed | 17898547.00 | 12848656.00 | — | -28.2 % | — |
| BenchmarkGzipDefault | 87839358.00 | 24846785.00 | — | -71.7 % | — |
