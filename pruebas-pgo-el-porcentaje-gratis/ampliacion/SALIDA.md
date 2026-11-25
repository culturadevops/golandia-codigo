# SALIDA — PGO en Go — ampliación (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-10 18:07 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## 1 · Lo que dice go help build de -pgo

`$ go help build   # el flag -pgo`

```
-pgo file
		specify the file path of a profile for profile-guided optimization (PGO).
		When the special name "auto" is specified, for each main package in the
		build, the go command selects a file named "default.pgo" in the package's
		directory if that file exists, and applies it to the (transitive)
		dependencies of the main package (other packages are not affected).
		Special name "off" turns off PGO. The default is "auto".
```

## 2 · Qué hay dentro de un perfil

`$ ls -l default.pgo && go tool pprof -top -nodecount=5 default.pgo`

```
2891 default.pgo
File: sin-pgo
Build ID: b2ff012e601662210bd2cf538d61290579c73465
Type: cpu
Time: 2026-10-10 12:46:41 UTC
Duration: 3.28s, Total samples = 3280ms (99.91%)
Showing nodes accounting for 2980ms, 90.85% of 3280ms total
Showing top 5 nodes out of 11
      flat  flat%   sum%        cum   cum%
    1360ms 41.46% 41.46%     3270ms 99.70%  main.Suma.Eval
     670ms 20.43% 61.89%      970ms 29.57%  main.suavizar
     400ms 12.20% 74.09%     2150ms 65.55%  main.Mul.Eval
     280ms  8.54% 82.62%      300ms  9.15%  math.IsInf (inline)
```

## 3 · Lo que hace el compilador con el perfil del programa

`$ go build -gcflags=-m -pgo=off   vs   -pgo=default.pgo   # contar lo que dice el compilador`

```
sin perfil:  0 desvirtualizadas · 17 llamadas inlineadas
con perfil:  5 desvirtualizadas · 30 llamadas inlineadas
```

## 4 · Fusionar varios perfiles

`$ ./sin -perfil p1.pprof; ./sin -perfil p2.pprof
$ go tool pprof -proto p1.pprof p2.pprof > fusion.pgo`

```
4360 /tmp/fusion.pgo
3212 /tmp/p1.pprof
3131 /tmp/p2.pprof
p1.pprof: Duration: 3.69s, Total samples = 3680ms (99.70%)
p2.pprof: Duration: 3.63s, Total samples = 3.62s (99.66%)
fusion.pgo: Duration: 7.32s, Total samples = 7.30s (99.68%)
```

## 5 · El perfil de un servicio en marcha: net/http/pprof

`$ import _ "net/http/pprof"   # en el servicio
$ curl 'localhost:6060/debug/pprof/profile?seconds=3' -o cpu.pprof
$ go tool pprof -top -nodecount=3 cpu.pprof`

```
2789 cpu.pprof
      flat  flat%   sum%        cum   cum%
     230ms 76.67% 76.67%      230ms 76.67%  main.fib
      30ms 10.00% 86.67%       30ms 10.00%  runtime.futex
      10ms  3.33% 90.00%       10ms  3.33%  internal/runtime/syscall/linux.Syscall6
```
