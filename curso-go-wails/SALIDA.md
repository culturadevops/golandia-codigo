# SALIDA — curso «Wails: tu frontend web con un backend de Go de verdad»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Capturas en `capturas/`.

- fecha: `2026-10-05 16:22:25Z` · todo compilado y abierto en Docker (Debian 12) sobre Xvfb
- `go version go1.26.8 linux/amd64`
- `github.com/wailsapp/wails/v2 v2.16.0`
- `libwebkit2gtk-4.1-0:amd64 2.50.6-1~deb12u2`

## Leccion 1 — La primera ventana

`bash capturar.sh 01_ventana 6`  · codigo de salida `0`

```
  ventana viva · captura: capturas/01_ventana.png
  app> Ctrl+C detected. Shutting down...
```

## Leccion 2 — Los tipos, por el puente

`bash capturar.sh 02_tipos 8`  · codigo de salida `0`

```
  ventana viva · captura: capturas/02_tipos.png
  app>   JS recibió  ID: 9007199254740992 (number) · ¿igual a 9007199254740993? false
  app>   JS recibió  Lineas: null · Lineas.length -> ERROR: null is not an object (evaluating 'p.Lineas.length')
  app>   JS recibió  Creado: "2026-10-05T12:00:00-06:00" (string)
  app>   JS recibió  PorPais: {"52":"México","54":"Argentina"} · claves: string, string
  app>   JS recibió  interno: (no llegó)
  app>   corregido   ID: 9007199254740993 (string) · ¿igual a 9007199254740993? true
  app>   corregido   Lineas: [] · Lineas.length -> 0
  app>   corregido   Creado: new Date(...) -> 2026-10-05T18:00:00.000Z
  app> Ctrl+C detected. Shutting down...
```

## Leccion 3 — Los errores, por el puente

`bash capturar.sh 03_errores 7`  · codigo de salida `0`

```
  ventana viva · captura: capturas/03_errores.png
  app>   Comprar("taza", 2) -> 1
  app>   Comprar("taza", 5) -> rechazada: {} · typeof object
  app>   ¿es «sin stock»? e === "sin stock": false · e.includes("sin stock"): true
  app> ERR | process message error: C{"name":"main.App.Romper","args":[],"callbackID":"main.App.Romper-3530806174"} -> assignment to entry in nil map
  app> ERR | assignment to entry in nil map
  app>   Romper() (entra en pánico) -> 3 s después: ni respuesta ni error
  app>   la ventana sigue viva después del pánico
  app> Ctrl+C detected. Shutting down...
  app> wails.Run terminó: <nil>
```

## Leccion 4 — Eventos: los que se pierden al arrancar

`bash capturar.sh 04_eventos 7`  · codigo de salida `0`

```
  ventana viva · captura: capturas/04_eventos.png
  app>   trabajo «en OnStartup»: la ventana recibió 9 de 10 avisos de progreso
  app> Ctrl+C detected. Shutting down...
```

## Leccion 4b — Eventos: esperando a que la ventana diga «listo»

`bash capturar.sh 04_eventos 7` MODO=bueno NOMBRE=04_eventos-bueno · codigo de salida `0`

```
  ventana viva · captura: capturas/04_eventos-bueno.png
  app>   trabajo «tras Listo()»: la ventana recibió 10 de 10 avisos de progreso
  app> Ctrl+C detected. Shutting down...
```

## Leccion 5 — El contexto de Wails

`bash capturar.sh 05_contexto 3` DESPUES=7 · codigo de salida `0`

```
  ventana viva · captura: capturas/05_contexto.png
  ⛔ después de la captura, la app terminó sola (código 1)
  app>   con el ctx de OnStartup:  cambiado
  app>   ahora, con context.Background():
  app> 2026/10/05 16:23:17 cannot call 'github.com/wailsapp/wails/v2/pkg/runtime.WindowSetTitle': An invalid context was passed. This method requires the specific context given in the lifecycle hooks:
  app> https://wails.io/docs/reference/runtime/intro
```

## Leccion 6 — El motor es del sistema

`bash capturar.sh 06_motor 6`  · codigo de salida `0`

```
  ventana viva · captura: capturas/06_motor.png
  app>   userAgent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/60.5 Safari/605.1.15 wails.io/605.1.15
  app>   CSS: selector :has()             sí
  app>   CSS: color-mix()                 sí
  app>   CSS: text-wrap: balance          sí
  app>   CSS: anidamiento (&)             sí
  app>   CSS: container queries           sí
  app>   JS: Array.prototype.toSorted     sí
  app>   JS: Object.groupBy               sí
  app>   API: dialog.showModal()          sí
  app>   API: Popover                     sí
  app> Ctrl+C detected. Shutting down...
```

## Leccion 7 — Empaquetar

`bash 07_empaquetar.sh` · codigo de salida `0`

```
== 7.1 · go build a secas ==
  go build: código 0
  sin etiquetas                                     5456101 bytes ·   3 librerías del sistema
  al ejecutarlo: código 0 en menos de 8 s

== 7.2 · con las etiquetas de Wails ==
  -tags desktop,production                         12285336 bytes · 169 librerías del sistema
  al ejecutarlo: sigue abierta a los 8 s (la cerró timeout)
    + -ldflags="-s -w"                              8610536 bytes · 169 librerías del sistema

== 7.3 · Para Windows, desde Linux ==
  GOOS=windows: código 0 · 11236864 bytes
  (en Windows usa WebView2, el motor de Edge: aquí no se puede abrir, y no se dice que se vea igual)
```

## Leccion 8 — El proyecto: un inspector de colas para nubecita

`bash 08_proyecto.sh` · codigo de salida `0`

```
== 8.1 · nubecita viva, y llegan mensajes con la ventana abierta ==
  sembrado: map[correos:0 pedidos:4 pedidos-muertos:1]
  sembrado: map[pedidos:12]
  ventana viva · captura: capturas/08_proyecto-vivo.png
  app>   refresco 1: correos=0 pedidos=4 pedidos-muertos=1
  app>   refresco 2: correos=0 pedidos=4 pedidos-muertos=1
  app>   refresco 3: correos=0 pedidos=4 pedidos-muertos=1
  app>   refresco 4: correos=0 pedidos=4 pedidos-muertos=1
  app>   refresco 5: correos=0 pedidos=4 pedidos-muertos=1
  app>   refresco 6: correos=0 pedidos=16 pedidos-muertos=1
  app>   refresco 7: correos=0 pedidos=16 pedidos-muertos=1
  app>   refresco 8: correos=0 pedidos=16 pedidos-muertos=1
  app>   refresco 9: correos=0 pedidos=16 pedidos-muertos=1
  app>   refresco 10: correos=0 pedidos=16 pedidos-muertos=1
  app> Ctrl+C detected. Shutting down...

== 8.2 · nubecita apagada ==
  ventana viva · captura: capturas/08_proyecto-sin-servidor.png
  app> Ctrl+C detected. Shutting down...
```

