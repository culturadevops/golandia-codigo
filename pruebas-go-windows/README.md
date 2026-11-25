# Go en Windows: qué se rompe al cruzar (lo ejecuté en los dos)

Código del vídeo [«Go en Windows: qué se rompe al cruzar (lo ejecuté en los dos)»](https://youtu.be/yVp1Tyt2yo0) del canal Golandia.

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

```bash
go test ./...
go run .
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
