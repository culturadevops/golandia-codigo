# ¿Go, Bun o Node? Lo medí en 2026 (y la trampa de los núcleos)

Código del vídeo [«¿Go, Bun o Node? Lo medí en 2026 (y la trampa de los núcleos)»](https://youtu.be/o7Ihsz9izcY) del canal Golandia.

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

Los scripts del banco común de la serie «¿Go o X?» están en [`../banco-go-o-x`](../banco-go-o-x/).

```bash
python pruebas_go_bun_node.py
python pruebas_go_bun_node_cluster.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
