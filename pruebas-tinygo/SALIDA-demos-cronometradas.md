# SALIDA — demos de pruebas-tinygo (golandia)
> ⚠️ La sección de goroutines de esta corrida NO vale (el resultado no se leía y LLVM borró el cálculo: «0 ms»). La demo ya imprime el resultado; falta repetirla con el candado. Lo válido de aquí está copiado en SALIDA-wasm.md.

> Corrida del 2026-11-05 11:09. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py tinygo --con-candado`. Cada bloque es la salida literal del guion de `demos/`.

## Goroutines: 2 a la vez con go y con tinygo (2 CPU)

Imagen `tinygo/tinygo:0.44.0`, `--cpus 2`, guion `demos/gorutinas.sh`.

```
### ./con-go
go go1.27.3 · GOMAXPROCS: 2
en serie:        50 ms
2 goroutines:    25 ms
[código de salida: 0]
### ./con-tinygo
go 0.44.0 · GOMAXPROCS: 1
en serie:        0 ms
2 goroutines:    0 ms
[código de salida: 0]
```

## WebAssembly y nativo: el cálculo de la serie, sin servidor (2 CPU)

Imagen `v808-banco:1`, `--cpus 2`, guion `demos/wasm.sh`.

```
### wasmtime --version
wasmtime 50.0.1 (c307ec81d 2026-10-29)
### ls -l *.wasm (bytes)
   2522268  go.wasm
    302908  tinygo-opt2.wasm
    225318  tinygo.wasm
### wasmtime go.wasm suma 300
suma resultado=99863963591 mediana_ns=280998 p99_ns=362528 compilador=gc
memoria máx: 84732 KiB
### wasmtime go.wasm trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=9716846 p99_ns=14228658 compilador=gc
memoria máx: 33700 KiB
### wasmtime tinygo.wasm suma 300
suma resultado=99863963591 mediana_ns=130412 p99_ns=174667 compilador=tinygo
memoria máx: 37520 KiB
### wasmtime tinygo.wasm trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=7421436 p99_ns=10703040 compilador=tinygo
memoria máx: 27300 KiB
### wasmtime tinygo-opt2.wasm suma 300
suma resultado=99863963591 mediana_ns=327800 p99_ns=360409 compilador=tinygo
memoria máx: 47008 KiB
### wasmtime tinygo-opt2.wasm trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=6739461 p99_ns=9813032 compilador=tinygo
memoria máx: 26956 KiB
### ./go-nativo suma 300
suma resultado=99863963591 mediana_ns=124488 p99_ns=139926 compilador=gc
memoria máx: 1968 KiB
### ./go-nativo trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=3210579 p99_ns=3928031 compilador=gc
memoria máx: 7180 KiB
### ./tinygo-nativo suma 300
suma resultado=99863963591 mediana_ns=381184 p99_ns=444954 compilador=tinygo
memoria máx: 620 KiB
### ./tinygo-nativo trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=4098441 p99_ns=4987603 compilador=tinygo
memoria máx: 6144 KiB
### ./tinygo-nativo-opt2 suma 300
suma resultado=99863963591 mediana_ns=110436 p99_ns=116837 compilador=tinygo
memoria máx: 812 KiB
### ./tinygo-nativo-opt2 trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=3624803 p99_ns=4145734 compilador=tinygo
memoria máx: 5572 KiB
```

