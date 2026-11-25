# pruebas-go-vs

Código de estos vídeos del canal Golandia:

- [¿Go o C#? Lo medí con .NET 10 (también compilado nativo)](https://youtu.be/EFcnH4gahlE)
- [¿Go o Kotlin? Lo medí, y el JIT da la sorpresa](https://youtu.be/G1DfJ-VL-hI)
- [¿Go o Zig? Lo medí: más chico, más rápido… y más caro de escribir](https://youtu.be/_I6uG3utnm4)
- [¿Go o Elixir? Lo medí: lo que gana cada uno](https://youtu.be/ejEN_Vk4hnE)

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

Los scripts del banco común de la serie «¿Go o X?» están en [`../banco-go-o-x`](../banco-go-o-x/).

Lee los archivos de la carpeta: cada script indica en su cabecera cómo se lanza.

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
