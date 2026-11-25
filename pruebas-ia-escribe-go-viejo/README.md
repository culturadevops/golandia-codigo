# pruebas-ia-escribe-go-viejo

Código de estos vídeos del canal Golandia:

- [La IA escribe Go viejo: lo medí con go fix (Claude Code y un modelo local, Go 1.27)](https://youtu.be/8GCkkBfBayc)
- [5 cambios en Go que están pasando ahora (2026): go fix, IA, JSON v2, goroutineleak y genéricos](https://youtu.be/hYiObRvadKg)

## Qué es

El código, los scripts de medición y los resultados (`SALIDA*.md`) que se ejecutaron para el vídeo. Las cifras del vídeo salen de estas ejecuciones.

## Cómo ejecutarlo

```bash
python ejecutar_todo.py
python medir.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.

## Resultados

Los resultados de la ejecución que sale en el vídeo están en los archivos `SALIDA*.md`.
