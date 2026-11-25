# SALIDA — demos de pruebas-go-vs-scala (golandia)

> Corrida del 2026-11-05 08:17. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py scala`. Cada bloque es la salida literal del guion de `demos/`.

## Scala: un match al que le falta un caso, y Option en vez de nil

Imagen `virtuslab/scala-cli:1.18.0`, `--cpus 1`, guion `demos/tipos.sh`.

```
### scala-cli compile Estado.scala
-- [E029] Pattern Match Exhaustivity Warning: /tmp/Estado.scala:6:35 -----------
6 |def describir(e: Estado): String = e match
  |                                   ^
  |                                   match may not be exhaustive.
  |
  |                                   It would fail on pattern case: Devuelto
  |
  | longer explanation available when compiling with `-explain`
1 warning found
[código de salida: 0]
### scala-cli compile -Werror Estado.scala
-- [E029] Pattern Match Exhaustivity Warning: /tmp/Estado.scala:6:35 -----------
6 |def describir(e: Estado): String = e match
  |                                   ^
  |                                   match may not be exhaustive.
  |
  |                                   It would fail on pattern case: Devuelto
  |
  | longer explanation available when compiling with `-explain`
No warnings can be incurred under -Werror
1 warning found
1 error found
Compilation failed
[código de salida: 1]
### scala-cli run Estado.scala
en camino
no existe
[código de salida: 0]
```

## Go: el mismo switch y el mismo nil

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/tipos_go.sh`.

```
### go vet .
[código de salida: 0]
### go run .
Devuelto: ""
[código de salida: 0]
```

## Líneas de cada servidor

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/lineas.sh`.

```
90 líneas  go/main.go
45 líneas  src/Servidor.scala
```

