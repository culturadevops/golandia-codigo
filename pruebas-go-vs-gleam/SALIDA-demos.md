# SALIDA — demos de pruebas-go-vs-gleam (golandia)

> Corrida del 2026-11-05 08:19. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py gleam`. Cada bloque es la salida literal del guion de `demos/`.

## Gleam: un case al que le falta un caso, y enteros sin desbordamiento

Imagen `ghcr.io/gleam-lang/gleam:v1.19.2-erlang-alpine`, `--cpus 1`, guion `demos/tipos.sh`.

```
### gleam build
error: Inexhaustive patterns
   ┌─ /tmp/t/src/tipos.gleam:13:3
   │  
13 │ ╭   case e {
14 │ │     Pendiente -> "esperando el pago"
15 │ │     Pagado -> "listo para enviar"
16 │ │     Enviado -> "en camino"
17 │ │   }
   │ ╰───^

This case expression does not have a pattern for all possible values. If it
is run on one of the values without a pattern then it will crash.

The missing patterns are:

    Devuelto

[código de salida: 1]
### (agrego el caso Devuelto) gleam run
en camino
2^64 + 1 = 18446744073709551617
[código de salida: 0]
```

## Go: el mismo switch y el mismo desbordamiento

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/tipos_go.sh`.

```
### go vet .
[código de salida: 0]
### go run .
Devuelto: ""
2^64 + 1 = 1
[código de salida: 0]
```

## Gleam: un proceso que se cae no tumba al resto

Imagen `ghcr.io/gleam-lang/gleam:v1.19.2-erlang-alpine`, `--cpus 1`, guion `demos/caidas.sh`.

```
### gleam run
proceso 3: terminé
proceso 1: terminé
=CRASH REPORT==== 5-Nov-2026::13:19:11.066458 ===
  crasher:
    initial call: caidas:'-main/0-anonymous-1-'/0
    pid: <0.85.0>
    registered_name: []
    exception error: #{function => <<"trabajar">>,line => 9,
                       message =>
                           <<"división entre cero en el pedido 42"/utf8>>,
                       module => <<"caidas">>,file => <<"src/caidas.gleam">>,
                       gleam_error => panic}
      in function  caidas:'-main/0-anonymous-1-'/0 (src/caidas.gleam:9)
    ancestors: [<0.83.0>]
    message_queue_len: 0
    messages: []
    links: []
    dictionary: []
    trap_exit: false
    status: running
    heap_size: 233
    stack_size: 29
    reductions: 73
  neighbours:

main: sigo vivo, 3 procesos lanzados
[código de salida: 0]
```

## Go: una goroutine que entra en pánico

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/caidas_go.sh`.

```
### ./caidas
panic: división entre cero en el pedido 42
[código de salida: 2]
```

## Líneas de cada servidor

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/lineas.sh`.

```
90 líneas  go/main.go
90 líneas  servidor/src/servidor.gleam
```


## Nota de la construcción (2026-11-05 08:0x, primer intento del Dockerfile.gleam)

Con `mist = ">= 5.0.0 and < 6.0.0"` en gleam.toml, `gleam deps download` resolvió mist 5.0.4 + glisten 8.0.3 + gleam_stdlib 0.71.0,
y la compilación falló (salida literal de `docker build`). Con `mist >= 6.0.0` (mist 6, glisten 9) compila.

```
1.306 error: Unknown module value
1.306     ┌─ /app/build/packages/glisten/src/glisten/internal/acceptor.gleam:121:24
1.306 121 │   let acceptors = list.range(from: 0, to: pool.pool_count)
1.306 The module `gleam/list` does not have a `range` value.
```
