# pruebas-entrevista-go

Código de estos vídeos del canal Golandia:

- [¿Qué imprime este append? La pregunta de Go que tumba a la mayoría](https://youtu.be/xndgjRACT4A)
- [¿Por qué tu mapa en Go devuelve algo que nunca guardaste?](https://youtu.be/f7wwp02rOkM)
- [Devuelves nil y el if dice que no: la pregunta trampa de Go](https://youtu.be/PpyuEykYHek)
- [¿Puntero o valor? La pregunta de Go que separa junior de senior](https://youtu.be/gJTyWm5hE0o)
- [Este programa en Go se cuelga: ¿sabes por qué?](https://youtu.be/HjUqyIhyhS0)
- [¿Mutex o canal en Go? La respuesta que quieren oír](https://youtu.be/zWZMloGuGII)
- [¿errors.Is o errors.As? ¿Y por qué el context va primero?](https://youtu.be/WbyJ1k7NV0k)
- [Te dan un problema y 45 minutos: así se resuelve en Go delante de alguien](https://youtu.be/S-6_Q2uKdwY)
- [¿Te compensa gRPC en Go? Lo medí, y el argumento del tamaño se cae](https://youtu.be/e2H81hr2kMk)
- [¿Go o TypeScript compilado a nativo? Bun AOT y scriptc, lo medí (2026)](https://youtu.be/94D-EpJk5jw)

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

```bash
python pruebas_ep01.py
python pruebas_ep02.py
python pruebas_ep03.py
python pruebas_ep04.py
python pruebas_ep05.py
python pruebas_ep06.py
python pruebas_ep07.py
python pruebas_ep08.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.
