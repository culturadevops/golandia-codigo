# pruebas-go-java-revancha

Código de estos vídeos del canal Golandia:

- [¿Go o Java? La revancha: GraalVM, jlink y una carga de verdad (medido)](https://youtu.be/BTjUtsPLrH8)
- [¿Go o Rust? Lo medí: velocidad, memoria y lo que cuesta escribirlo (axum vs net/http)](https://youtu.be/7spxAC8JTDw)

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

Los scripts del banco común de la serie «¿Go o X?» están en [`../banco-go-o-x`](../banco-go-o-x/).

```bash
python pruebas_go_java_revancha.py
python pruebas_go_java_revancha_go.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
