# SALIDA — PGO en Go (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-10 12:47 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## La versión de Go

`$ go version`

```
go version go1.27.2 linux/amd64
```

## 1 · Benchmark sin PGO

`$ go test -run x -bench Trabajo -count 6 -pgo=off`

```
BenchmarkTrabajo-24    	    3510	    339064 ns/op
BenchmarkTrabajo-24    	    3577	    341543 ns/op
BenchmarkTrabajo-24    	    3255	    341578 ns/op
BenchmarkTrabajo-24    	    3466	    341705 ns/op
BenchmarkTrabajo-24    	    3560	    341658 ns/op
BenchmarkTrabajo-24    	    3530	    339388 ns/op
```

## 2 · Perfil de CPU del benchmark

`$ go test -run x -bench Trabajo -cpuprofile test.pprof`

```
2696 test.pprof
```

## 3 · Benchmark con PGO

`$ go test -run x -bench Trabajo -count 6 -pgo=test.pprof`

```
BenchmarkTrabajo-24    	    4326	    270745 ns/op
BenchmarkTrabajo-24    	    4468	    271575 ns/op
BenchmarkTrabajo-24    	    4526	    270683 ns/op
BenchmarkTrabajo-24    	    4369	    272111 ns/op
BenchmarkTrabajo-24    	    4406	    271124 ns/op
BenchmarkTrabajo-24    	    4443	    270409 ns/op
```

`$ benchstat sin.txt con.txt`

```
│   sin.txt   │              con.txt               │
           │   sec/op    │   sec/op     vs base               │
Trabajo-24   341.6µ ± 1%   270.9µ ± 0%  -20.68% (p=0.002 n=6)
```

## 4 · Por qué: lo que cambia en el compilador

`$ diff  <(go test -c -pgo=off -gcflags=-m)  <(go test -c -pgo=test.pprof -gcflags=-m)`

```
./evaluador.go:24:58: PGO devirtualizing interface call s.A.Eval to Suma.Eval
./evaluador.go:24:58: inlining call to Suma.Eval
./evaluador.go:24:72: PGO devirtualizing interface call s.B.Eval to Suma.Eval
./evaluador.go:24:72: inlining call to Suma.Eval
./evaluador.go:25:58: PGO devirtualizing interface call m.A.Eval to Suma.Eval
./evaluador.go:25:58: inlining call to Suma.Eval
./evaluador.go:25:72: PGO devirtualizing interface call m.B.Eval to Suma.Eval
./evaluador.go:25:72: inlining call to Suma.Eval
./evaluador.go:26:67: PGO devirtualizing interface call f.A.Eval to Suma.Eval
./evaluador.go:26:67: inlining call to Suma.Eval
```

## 5 · La trampa: ese perfil no encaja con el binario

`$ go tool pprof -top test.pprof | grep Eval`

```
ejemplo.com/evaluador.Suma.Eval
ejemplo.com/evaluador.Mul.Eval
ejemplo.com/evaluador.Var.Eval
```

`$ go build -pgo=test.pprof -gcflags=-m . | grep -c devirtualizing`

```
0
```

## 6 · El perfil del programa de verdad

`$ go build -pgo=off -o sin-pgo . && ./sin-pgo -perfil default.pgo`

```
resultado 483813589.249 en 3.283s
```

`$ go tool pprof -top default.pgo | grep Eval`

```
main.Suma.Eval
main.Mul.Eval
main.Func.Eval
```

`$ go build -o con-pgo .      # default.pgo está en el paquete main
$ go version -m con-pgo | grep pgo`

```
con-pgo: go1.27.2
	build	-pgo=/w/default.pgo
```

`$ for i in 1 2 3; do ./sin-pgo; done;  for i in 1 2 3; do ./con-pgo; done`

```
resultado 483813589.249 en 3.259s
resultado 483813589.249 en 3.264s
resultado 483813589.249 en 3.26s
---
resultado 483813589.249 en 2.591s
resultado 483813589.249 en 2.583s
resultado 483813589.249 en 2.582s
```

## 7 · Lo que cuesta

`$ ls -l sin-pgo con-pgo`

```
2966334 con-pgo
2966228 sin-pgo
```

`$ time go build -pgo=off   vs   time go build -pgo=default.pgo   # caché vacía`

```
sin PGO, caché vacía: 1654 ms
con PGO, caché vacía: 1740 ms
```

## 8 · Cuándo no sirve: un perfil que no es representativo

`$ go test -bench OtraCosa -cpuprofile otro.pprof
$ go test -run x -bench Trabajo -count 6 -pgo=otro.pprof`

```
BenchmarkTrabajo-24    	    3469	    339315 ns/op
BenchmarkTrabajo-24    	    3572	    339027 ns/op
BenchmarkTrabajo-24    	    3555	    338805 ns/op
BenchmarkTrabajo-24    	    3457	    339315 ns/op
BenchmarkTrabajo-24    	    3536	    339075 ns/op
BenchmarkTrabajo-24    	    3508	    342487 ns/op
```

`$ benchstat sin.txt mal.txt`

```
│   sin.txt   │           mal.txt            │
           │   sec/op    │   sec/op     vs base         │
Trabajo-24   341.6µ ± 1%   339.2µ ± 1%  ~ (p=0.162 n=6)
```
