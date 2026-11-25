# SALIDA — demos de pruebas-tinygo (golandia)

> Corrida del 2026-11-05 08:31. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py tinygo`. Cada bloque es la salida literal del guion de `demos/`.

## Tamaños: hola mundo y el programa del banco, nativo y WebAssembly

Imagen `tinygo/tinygo:0.44.0`, `--cpus 1`, guion `demos/tamanos.sh`.

```
### tinygo version
tinygo version 0.44.0 linux/amd64 (using go version go1.27.3 and LLVM version 22.1.4)
### ls -l (bytes)
   1511584  hola-go
   2430668  hola-go.wasm
    246056  hola-tinygo
    183642  hola-tinygo.wasm
   5873824  servidor-go
   1265000  servidor-tinygo
   1421240  servidor-tinygo-opt2
### ./hola-go && ./hola-tinygo
hola, Golandia
hola, Golandia
```

## Lo que compila y lo que no

Imagen `tinygo/tinygo:0.44.0`, `--cpus 1`, guion `demos/compila.sh`.

```
### json · go run .
{"id":42,"cliente":"Ana","total":99.5,"items":["café","pan"]} <nil>
{ID:7 Cliente:Luis Total:10 Items:[té]} <nil>
[código de salida: 0]
### json · tinygo run .
{"id":42,"cliente":"Ana","total":99.5,"items":["café","pan"]} <nil>
{ID:7 Cliente:Luis Total:10 Items:[té]} <nil>
[código de salida: 0]
### plantilla · go run .
hola, Ana: 3 items
[código de salida: 0]
### plantilla · tinygo run .
hola, panic: unimplemented: (reflect.Value).MethodByName()
failed to run compiled binary /tmp/tinygo939047283/main: signal: aborted (core dumped)
[código de salida: 1]
### exec · go run .
"Linux\n" <nil>
[código de salida: 0]
### exec · tinygo run .
"" files setting not implemented
[código de salida: 0]
### metodo · go run .
hola, Ana
[código de salida: 0]
### metodo · tinygo run .
panic: unimplemented: (reflect.Value).MethodByName()
failed to run compiled binary /tmp/tinygo611896762/main: signal: aborted (core dumped)
[código de salida: 1]
### plugin · go run .
plugin.Open("x.so"): realpath failed
[código de salida: 0]
### plugin · tinygo run .
ld.lld: error: unable to find library -ldl
[código de salida: 1]
```

