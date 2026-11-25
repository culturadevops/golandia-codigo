# pruebas-serie-prohibido

Código de estos vídeos del canal Golandia:

- [Devuelves nil y el if dice que no: la pregunta trampa de Go](https://youtu.be/PpyuEykYHek)
- [En Go 0.1 + 0.2 SÍ da 0.3. Y por eso es más peligroso](https://youtu.be/YtS9UBI6_5I)
- [El nil de Go que no es nil (y se lleva por delante tu manejo de errores)](https://youtu.be/VmmImEMEHRk)
- [Go te recorre los mapas en orden ALEATORIO, y lo hace a propósito](https://youtu.be/wtudqF3s3bU)
- [Dos slices de Go comparten memoria, y uno te corrompe el otro](https://youtu.be/giL9eKkNDYg)
- [Go no te deja compilar por una variable sin usar. Y es a propósito](https://youtu.be/xkRAwks8TPY)
- [Go no tiene excepciones, y fue a propósito](https://youtu.be/B7GDAIAyrOE)
- [Go no tiene herencia. Y no la echas de menos](https://youtu.be/DpfPqI00MaU)
- [Go no tiene operador ternario, y el equipo de Go explicó por qué](https://youtu.be/Xt0zkg7yLUU)
- [Los imports circulares en Go no compilan. Y eso te ordena la arquitectura](https://youtu.be/U4d0LpNPBpg)
- [En Go está prohibido hacer panic en una librería (y no lo impide el compilador)](https://youtu.be/rgRQ8Wevxdg)
- [Las goroutines que nunca mueren: la fuga que Go no te avisa](https://youtu.be/SmQJLMxWLSI)
- [Go arrastró este bug del bucle for durante 10 años, y lo arregló en 1.22](https://youtu.be/jUgitpSg5KA)
- [init() en Go: la función que se ejecuta sin que la llames](https://youtu.be/vmvS9n5wyAA)
- [any mata el tipado de Go que fuiste a buscar](https://youtu.be/GMKD6fgypZY)
- [Los struct tags de Go son cadenas mágicas que nadie comprueba](https://youtu.be/C3IxIvFBy1Y)
- [defer dentro de un bucle en Go no hace lo que crees](https://youtu.be/nyBvz9BiwWo)

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
