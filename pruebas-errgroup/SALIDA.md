# SALIDA — errgroup en Go (golandia)

> Corrida del 2026-10-20 17:20. golang:1.27.2 + golang.org/x/sync v0.24.0, contenedor con --cpus 2. Tiempos redondeados a 10 ms.

```
$ go version; grep x/sync go.mod
go version go1.27.2 linux/amd64
require golang.org/x/sync v0.24.0

$ ./grupo waitgroup
   100 ms  precios falla
   300 ms  stock   terminó
   300 ms  fotos   terminó
   300 ms  total · error: precios: la API respondió 503

$ ./grupo errgroup
   100 ms  precios falla
   100 ms  stock   cancelada (precios: la API respondió 503)
   100 ms  fotos   cancelada (precios: la API respondió 503)
   100 ms  total · error: precios: la API respondió 503

$ ./grupo sin-contexto
   100 ms  precios falla
   300 ms  stock   terminó
   300 ms  fotos   terminó
   300 ms  total · error: precios: la API respondió 503

$ ./grupo dos-errores
   100 ms  precios falla
   150 ms  stock   falla
   150 ms  Wait devolvió: precios: la API respondió 503

$ ./grupo http
   100 ms  cliente: /precios respondió 503
   100 ms  cliente: /fotos cancelada
   100 ms  cliente: /stock cancelada
   100 ms  servidor: el cliente cortó /stock
   100 ms  servidor: el cliente cortó /fotos
   100 ms  total (Wait) · error: /precios: la API respondió 503

$ ./grupo http-sin-ctx
   100 ms  cliente: /precios respondió 503
   300 ms  cliente: /fotos respondió 200
   300 ms  cliente: /stock respondió 200
   300 ms  total (Wait) · error: /precios: la API respondió 503

$ ./grupo pipeline
  productor: paro en el item 105 (item 100: dato corrupto)
  Wait: item 100: dato corrupto · procesados: 103 de 1000

$ ./grupo pipeline-atascado
  a los 2 s Wait sigue bloqueado · procesados: 103 de 1000 · el productor quedó atascado

$ ./grupo limite
  10 tareas de 50 ms con SetLimit(3): máximo a la vez = 3 · total = 200 ms
  10 tareas de 50 ms sin límite:     máximo a la vez = 10 · total = 50 ms
```
