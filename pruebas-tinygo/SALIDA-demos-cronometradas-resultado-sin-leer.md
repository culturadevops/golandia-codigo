# SALIDA — demos de pruebas-tinygo (golandia)
> ⛔ NO SE USA: TinyGo midió «0 ms» porque nadie leía el resultado (sumidero) y LLVM borró el cálculo. Se repitió imprimiendo el resultado.

> Corrida del 2026-11-05 10:49. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py tinygo --con-candado`. Cada bloque es la salida literal del guion de `demos/`.

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
   2519894  go.wasm
    299312  tinygo-opt2.wasm
    222058  tinygo.wasm
### wasmtime go.wasm suma 300
suma resultado=99863963591 mediana_ns=284363 p99_ns=307365 compilador=gc
memoria máx: 84324 KiB
### wasmtime go.wasm trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=9532555 p99_ns=14574808 compilador=gc
memoria máx: 34936 KiB
### wasmtime tinygo.wasm suma 300
suma resultado=99863963591 mediana_ns=116170 p99_ns=165853 compilador=tinygo
memoria máx: 38188 KiB
### wasmtime tinygo.wasm trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=7822351 p99_ns=10894042 compilador=tinygo
memoria máx: 26956 KiB
### wasmtime tinygo-opt2.wasm suma 300
suma resultado=99863963591 mediana_ns=104826 p99_ns=132873 compilador=tinygo
memoria máx: 46088 KiB
### wasmtime tinygo-opt2.wasm trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=6900352 p99_ns=9857051 compilador=tinygo
memoria máx: 26968 KiB
### ./go-nativo suma 300
suma resultado=99863963591 mediana_ns=123211 p99_ns=171665 compilador=gc
memoria máx: 1772 KiB
### ./go-nativo trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=3215332 p99_ns=3639212 compilador=gc
memoria máx: 7172 KiB
### ./tinygo-nativo suma 300
suma resultado=99863963591 mediana_ns=381328 p99_ns=492347 compilador=tinygo
memoria máx: 616 KiB
### ./tinygo-nativo trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=3849963 p99_ns=4606754 compilador=tinygo
memoria máx: 5376 KiB
### ./tinygo-nativo-opt2 suma 300
suma resultado=99863963591 mediana_ns=78361 p99_ns=85232 compilador=tinygo
memoria máx: 620 KiB
### ./tinygo-nativo-opt2 trabajo 300
trabajo resultado={"total":9990000,"primeros":[0,1000,2000,3000,4000]} mediana_ns=3650817 p99_ns=4329801 compilador=tinygo
memoria máx: 5568 KiB
```

