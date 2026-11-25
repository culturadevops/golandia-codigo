# SALIDA — PGO en Go — segunda ampliación (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-11 02:52 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## 1 · Estabilidad de fuente: el código cambia, el perfil no

`$ go build -gcflags=-m -pgo=default.pgo   # el mismo perfil, con el código original y cambiado`

```
original:                    5 desvirtualizadas
10 líneas añadidas arriba:    5 desvirtualizadas
Suma.Eval en 3 líneas:       3 desvirtualizadas
evaluador.go:28:58: PGO devirtualizing interface call m.A.Eval to Suma.Eval
evaluador.go:28:72: PGO devirtualizing interface call m.B.Eval to Suma.Eval
evaluador.go:29:67: PGO devirtualizing interface call f.A.Eval to Suma.Eval
```

`$ # Suma.Eval reescrita en 3 líneas`

```
func (s Suma) Eval(x []float64) float64 {
	a := s.A.Eval(x)
	return a + s.B.Eval(x)
}
```

## 2 · Estabilidad iterativa: perfilar el binario ya optimizado

`$ ./gen1 -perfil g1.pprof; go build -pgo=g1.pprof -o gen2 .; ./gen2 -perfil g2.pprof; go build -pgo=g2.pprof -o gen3 .
$ for x in sin gen1 gen2 gen3; do for i in 1 2 3; do ./$x; done; done`

```
gen1 (perfil del binario sin PGO): 5 desvirtualizadas · 30 inlineadas
gen2 (perfil de gen1):             5 desvirtualizadas · 30 inlineadas
gen3 (perfil de gen2):             6 desvirtualizadas · 30 inlineadas
---
sin: en 3.69s en 3.296s en 3.293s 
gen1: en 2.616s en 2.601s en 2.602s 
gen2: en 2.929s en 2.958s en 2.915s 
gen3: en 2.906s en 2.916s en 2.909s
```

`$ go tool pprof -top default.pgo   vs   go tool pprof -top g1.pprof`

```
== default.pgo
      flat  flat%   sum%        cum   cum%
    1360ms 41.46% 41.46%     3270ms 99.70%  main.Suma.Eval
     670ms 20.43% 61.89%      970ms 29.57%  main.suavizar
     400ms 12.20% 74.09%     2150ms 65.55%  main.Mul.Eval
     280ms  8.54% 82.62%      300ms  9.15%  math.IsInf (inline)
     270ms  8.23% 90.85%     2920ms 89.02%  main.Func.Eval
     270ms  8.23% 99.09%      270ms  8.23%  main.Var.Eval
== g1.pprof
      flat  flat%   sum%        cum   cum%
    1050ms 40.23% 40.23%     2600ms 99.62%  main.Suma.Eval (inline)
     440ms 16.86% 57.09%      870ms 33.33%  main.suavizar (inline)
     390ms 14.94% 72.03%      430ms 16.48%  math.IsInf (inline)
     270ms 10.34% 82.38%     1710ms 65.52%  main.Mul.Eval
     270ms 10.34% 92.72%      270ms 10.34%  main.Var.Eval
     140ms  5.36% 98.08%     2250ms 86.21%  main.Func.Eval
```

`$ diff  <(go build -gcflags=-m -pgo=default.pgo)  <(go build -gcflags=-m -pgo=g1.pprof)   # qué cambia`

```
gen1: 24:58: PGO devirtualizing interface call s.A.Eval to Suma.Eval
gen1: 24:58: inlining call to Suma.Eval
gen1: 24:72: PGO devirtualizing interface call s.B.Eval to Suma.Eval
gen1: 24:72: inlining call to Suma.Eval
gen2: 24:58: PGO devirtualizing interface call s.A.Eval to Func.Eval
gen2: 24:58: inlining call to Func.Eval
gen2: 24:72: PGO devirtualizing interface call s.B.Eval to Func.Eval
gen2: 24:72: inlining call to Func.Eval
gen2: 69:18: PGO devirtualizing interface call t.Eval to Suma.Eval
```
