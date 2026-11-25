# SALIDA — demos de pruebas-go-vs-scala (golandia)

> Corrida del 2026-11-05 11:02. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py scala --con-candado`. Cada bloque es la salida literal del guion de `demos/`.

## Recompilar tras cambiar una línea (2 CPU)

Imagen `virtuslab/scala-cli:1.18.0`, `--cpus 2`, guion `demos/recompilar.sh`.

```
### scala-cli compile (en frío, sin caché del proyecto): 2135 ms [código de salida: 0]
### scala-cli compile (tras cambiar una línea): 2209 ms [código de salida: 0]
### scala-cli compile con el servidor Bloop ya caliente (tras cambiar una línea): 762 ms [código de salida: 0]
```

## Go: recompilar tras cambiar una línea (2 CPU)

Imagen `golang:1.27.2`, `--cpus 2`, guion `demos/recompilar_go.sh`.

```
### go build (en frío, sin caché): 7640 ms [código de salida: 0]
### go build (tras cambiar una línea): 204 ms [código de salida: 0]
```

