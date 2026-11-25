# SALIDA — «Concurrencia en Go: goroutines, canales y select, medido» (golandia)

> Generado por `correr.py`. **No se escribe a mano.**

- fecha: `2026-10-12 20:10:37Z`
- golang:1.27.2 en Docker, limitado a 8 CPU (`--cpus 8`); una sola máquina, los tiempos son de esta.

```
== 1 · Cuánto cuesta una goroutine ==
$ go run . 1000
  1000 goroutines · arrancadas en 1ms · vivas: 1000 · pila: 2.1 MB (2.1 KB cada una)
$ go run . 100000
  100000 goroutines · arrancadas en 100ms · vivas: 100000 · pila: 204.9 MB (2.0 KB cada una)
$ go run . 1000000
  1000000 goroutines · arrancadas en 1.309s · vivas: 1000000 · pila: 2048.2 MB (2.0 KB cada una)

== 2 · Canales sin búfer y con búfer ==
$ go run .
  búfer    0: 1.000.000 mensajes en 104ms (104 ns por mensaje)
  búfer    1: 1.000.000 mensajes en 78ms (78 ns por mensaje)
  búfer  100: 1.000.000 mensajes en 27ms (27 ns por mensaje)
  búfer 1000: 1.000.000 mensajes en 28ms (28 ns por mensaje)
  --- contrapresión: el productor manda cada 1 ms, el consumidor tarda 5 ms ---
  búfer   0: 40 mensajes en 215ms · el productor pasó 166ms bloqueado en el envío
  búfer  10: 40 mensajes en 214ms · el productor pasó 113ms bloqueado en el envío
  búfer 100: 40 mensajes en 210ms · el productor pasó 0s bloqueado en el envío

== 3 · Un pool de trabajadores ==
$ go run .
  [GOMAXPROCS=8 · núcleos que ve Go: 24]
  --- 200 tareas que esperan 20 ms (E/S) ---
    1 trabajadores: 4.05s
    4 trabajadores: 1.012s
   16 trabajadores: 263ms
   64 trabajadores: 81ms
  200 trabajadores: 21ms
  --- 64 tareas que solo calculan (CPU) ---
    1 trabajadores: 217ms
    2 trabajadores: 115ms (1.9× más rápido)
    4 trabajadores: 63ms (3.4× más rápido)
    8 trabajadores: 47ms (4.6× más rápido)
   16 trabajadores: 45ms (4.8× más rápido)
   32 trabajadores: 43ms (5.0× más rápido)

== 4 · Una carrera de datos ==
$ go run . sin
  sin    → total = 342533 (esperado 1000000)
$ go run . mutex
  mutex  → total = 1000000 (esperado 1000000)
$ go run . atomic
  atomic → total = 1000000 (esperado 1000000)
$ go run -race . sin
  sin    → total = 533187 (esperado 1000000)
  ==================
  WARNING: DATA RACE
  Read at 0x00c0000180c0 by goroutine 12:
  main.main.func1()
      /w/carrera/main.go:25 +0x104
  
  Previous write at 0x00c0000180c0 by goroutine 13:
  main.main.func1()
      /w/carrera/main.go:25 +0x119
  
  Goroutine 12 (running) created at:
  main.main()
      /w/carrera/main.go:20 +0x12a
  … (25 líneas más)
  [código de salida 1]

== 5 · Un envío que nadie recibe ==
$ go run .
  enviando 42 a un canal sin búfer…
  fatal error: all goroutines are asleep - deadlock!
  
  goroutine 1 [chan send]:
  main.main()
  	/w/bloqueo/main.go:9 +0x7b
  exit status 2
  [código de salida 1]

== 6 · La goroutine que se queda colgada ==
$ go run . fuga
  fuga     · tras 1.000 llamadas: 1001 goroutines vivas
  fuga     · 200 ms después:     1001 goroutines vivas
$ go run . bufer
  bufer    · tras 1.000 llamadas: 5 goroutines vivas
  bufer    · 200 ms después:     1 goroutines vivas
$ go run . contexto
  contexto · tras 1.000 llamadas: 2 goroutines vivas
  contexto · 200 ms después:     1 goroutines vivas

== 7 · Cancelar con context ==
$ go run .
  trabajo de 1ms entre comprobaciones: los 8 pararon 1.1ms después de cancel() · 784 vueltas hechas
  trabajo de 20ms entre comprobaciones: los 8 pararon 14.3ms después de cancel() · 48 vueltas hechas
```
