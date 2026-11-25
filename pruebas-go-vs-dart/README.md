# pruebas-go-vs-dart

Código de estos vídeos del canal Golandia:

- [¿Go o Dart? Lo medí: dart compile contra go build (2026)](https://youtu.be/aTzYHR-Nrts)
- [¿Gin + GORM o Serverpod? Lo medí: la misma API con Postgres en Go y en Dart](https://youtu.be/fsoKeGqW-I0)

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

Los scripts del banco común de la serie «¿Go o X?» están en [`../banco-go-o-x`](../banco-go-o-x/).

Lee los archivos de la carpeta: cada script indica en su cabecera cómo se lanza.

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
