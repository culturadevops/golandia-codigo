# ¿Go o Zig? La REVANCHA: un comentario tenía razón (medido)

Código del vídeo [«¿Go o Zig? La REVANCHA: un comentario tenía razón (medido)»](https://youtu.be/gi2NjDjEwUE) del canal Golandia.

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

Los scripts del banco común de la serie «¿Go o X?» están en [`../banco-go-o-x`](../banco-go-o-x/).

```bash
python pruebas_zig_revancha.py
python pruebas_zig_revancha2.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
