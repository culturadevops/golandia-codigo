# pruebas-go-vs-comun

Código de estos vídeos del canal Golandia:

- [¿Go o C++? Lo medí: C++ calcula más rápido… y su servidor atendió menos](https://youtu.be/krzdSCnEd7E)
- [¿Go o PHP? Lo medí con PHP 8.5 y FrankenPHP (y el modo worker no ganó)](https://youtu.be/xaqTZ6v9Po8)
- [¿Go o Swift en el servidor? Lo medí, y esta vez Go perdió](https://youtu.be/zWaYcJ8XJwA)

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

Los scripts del banco común de la serie «¿Go o X?» están en [`../banco-go-o-x`](../banco-go-o-x/).

Lee los archivos de la carpeta: cada script indica en su cabecera cómo se lanza.

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
