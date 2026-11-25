# Los picos del recolector de Go: reproduje el caso Discord (Go 1.9 vs Go 1.26)

Código del vídeo [«Los picos del recolector de Go: reproduje el caso Discord (Go 1.9 vs Go 1.26)»](https://youtu.be/pACc7qBaCqY) del canal Golandia.

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

```bash
python pruebas_gc_discord.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
