# pruebas-lambda-aws-real

Código de estos vídeos del canal Golandia:

- [Go en AWS Lambda de verdad: contra Rust, Java SnapStart, GraalVM y .NET AOT (lo medí, 2026)](https://youtu.be/qHiuCtW5mLA)
- [Go en Azure Functions con custom handler: lo probé en local (y lo que no se sabe sin la nube)](https://youtu.be/BCph_fxZCAE)

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

```bash
python banco.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
