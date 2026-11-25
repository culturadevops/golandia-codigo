# SALIDA — demos de pruebas-go-vs-ruby (golandia)

> Corrida del 2026-11-05 11:08. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py ruby --con-candado`. Cada bloque es la salida literal del guion de `demos/`.

## Ruby: hilos, el GVL y Ractor (2 CPU)

Imagen `ruby:4.0.9-slim`, `--cpus 2`, guion `demos/hilos.sh`.

```
### ruby hilos.rb
ruby 4.0.9 · núcleos visibles: 24
en serie:        1014 ms
2 hilos:         1041 ms
hilos.rb:28: warning: Ractor API is experimental and may change in future versions of Ruby.
2 Ractors:       480 ms
[código de salida: 0]
### ruby --yjit hilos.rb
ruby 4.0.9 · núcleos visibles: 24
en serie:        174 ms
2 hilos:         201 ms
hilos.rb:28: warning: Ractor API is experimental and may change in future versions of Ruby.
2 Ractors:       92 ms
[código de salida: 0]
```

## Go: goroutines con el mismo cálculo (2 CPU)

Imagen `golang:1.27.2`, `--cpus 2`, guion `demos/hilos_go.sh`.

```
### ./hilos
go go1.27.2 · GOMAXPROCS: 2
en serie:        50 ms
2 goroutines:    25 ms
[código de salida: 0]
```

