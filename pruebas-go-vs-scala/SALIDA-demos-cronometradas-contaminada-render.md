# SALIDA — demos de pruebas-go-vs-scala (golandia)

> ⛔ CONTAMINADA: desde las 09:48 un render (producir_y_subir.py) corrió a la vez, después de que el candado se borrara por huérfano. No se usa; se repite.

> Corrida del 2026-11-05 09:54. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py scala --con-candado`. Cada bloque es la salida literal del guion de `demos/`.

## Recompilar tras cambiar una línea (2 CPU)

Imagen `virtuslab/scala-cli:1.18.0`, `--cpus 2`, guion `demos/recompilar.sh`.

```
### scala-cli compile (en frío, sin caché del proyecto): 2746 ms [código de salida: 0]
### scala-cli compile (tras cambiar una línea): 3060 ms [código de salida: 0]
### scala-cli compile con el servidor Bloop ya caliente (tras cambiar una línea): 1052 ms [código de salida: 0]
```

## Go: recompilar tras cambiar una línea (2 CPU)

Imagen `golang:1.27.2`, `--cpus 2`, guion `demos/recompilar_go.sh`.

```
### go build (en frío, sin caché): 8926 ms [código de salida: 0]
### go build (tras cambiar una línea): 245 ms [código de salida: 0]
```

