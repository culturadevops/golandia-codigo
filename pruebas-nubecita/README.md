# pruebas-nubecita

Código de estos vídeos del canal Golandia:

- [LocalStack ya no es gratis: escribí mi propio S3 en Go y la CLI de AWS no lo nota](https://youtu.be/v56zfbNnYgs)
- [Escribí DynamoDB en Go y descubrí que no habla el mismo idioma que S3](https://youtu.be/j4YVpo-xdLE)
- [Escribí SQS en Go: la cola que no saca nada, y por qué eso lo cambia todo](https://youtu.be/q-qdRTutVdg)

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

Más detalle en [LEEME.md](LEEME.md).

## Cómo ejecutarlo

```bash
python pruebas_nubecita.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
