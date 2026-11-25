# SALIDA — curso «La otra pantalla: interfaces de terminal en Go con Bubble Tea»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-10-05 02:00:25Z`
- go version go1.26.8 windows/amd64
- `charm.land/bubbletea/v2 v2.0.10`
- `charm.land/bubbles/v2 v2.2.1`
- `charm.land/lipgloss/v2 v2.0.6`
- `charm.land/wish/v2 v2.0.5`
- sistema: `Windows 10`

## Leccion 01 — El bucle: mensaje, Update, View

`go run ./01_el_bucle` · codigo de salida `0`

```
== 1.1 · El modelo es un valor: Update devuelve el siguiente ==
  ┌─ tras pulsar ↑ tres veces ─┐
  │Contador: 3                 │
  │                            │
  │↑ suma · q sale             │
  └────────────────────────────┘

== 1.2 · La trampa: leer datos dentro de View ==
  41 pulsaciones -> Update: 44 veces · View: 46 veces · inventario leído: 46 veces

== 1.3 · Lo correcto: leer una vez y que View solo pinte ==
  41 pulsaciones -> Update: 44 veces · View: 46 veces · inventario leído: 1 vez
  View puede llamarse muchas veces y en cualquier momento: tiene que ser barato.
```

## Leccion 02 — Las teclas

`go run ./02_teclas` · codigo de salida `0`

```
== 2.1 · La barra espaciadora, escrita como en la v1 ==
  3 pulsaciones de espacio -> marcados: 0   (Bubble Tea la llama "space")

== 2.2 · La misma, escrita como en la v2 ==
  3 pulsaciones de espacio -> marcados: 3

== 2.3 · Cómo se llama cada tecla ==
  q        -> String() = "q"
  space    -> String() = "space"
  enter    -> String() = "enter"
  esc      -> String() = "esc"
  down     -> String() = "down"
  ctrl+c   -> String() = "ctrl+c"

== 2.4 · La trampa: un programa que no atiende Ctrl+C ==
  3 veces Ctrl+C -> ¿terminó? false   (hubo que matarlo a los 2 s)

== 2.5 · Lo correcto: Ctrl+C también sale ==
  1 vez Ctrl+C   -> ¿terminó? true
```

## Leccion 03 — Los comandos: el trabajo lento va fuera de Update

`go run ./03_comandos` · codigo de salida `0`

```
== 3.1 · La trampa: llamar a la API dentro de Update ==
  la API tardó 1000 ms · servidores: 42
  durante la espera: 0 fotogramas del reloj · 0 teclas atendidas

== 3.2 · Lo correcto: devolver un tea.Cmd ==
  la API tardó 1000 ms · servidores: 42
  durante la espera: 10 fotogramas del reloj · 5 teclas atendidas

  Los dos traen los mismos datos. Uno congela la pantalla un segundo entero.
```

## Leccion 04 — No se toca el estado desde otra gorrutina

`go run ./04_gorrutinas` · codigo de salida `0`

```
== 4.1 · La trampa: la gorrutina suma directamente ==
  errores de verdad: 10
  en pantalla al salir: "errores en el log: 10"
  lo que se vio, y cuándo (los errores llegan de t=100 a t=1000 ms; sale en t=1500):
    t=   0 ms -> errores en el log: 0
    t=1500 ms -> errores en el log: 10

== 4.2 · Lo correcto: la gorrutina manda un mensaje ==
  errores de verdad: 10
  en pantalla al salir: "errores en el log: 10"
  lo que se vio, y cuándo (los errores llegan de t=100 a t=1000 ms; sale en t=1500):
    t=   0 ms -> errores en el log: 0
    t= 100 ms -> errores en el log: 1
    t= 200 ms -> errores en el log: 2
    t= 301 ms -> errores en el log: 3
    t= 401 ms -> errores en el log: 4
    t= 501 ms -> errores en el log: 5
    t= 602 ms -> errores en el log: 6
    t= 702 ms -> errores en el log: 7
    t= 802 ms -> errores en el log: 8
    t= 903 ms -> errores en el log: 9
    t=1003 ms -> errores en el log: 10
```

### La carrera, con -race (contenedor golang:1.26)

`go run -race ./04_gorrutinas -carrera` · codigo de salida `1`

```
==================
WARNING: DATA RACE
Write at 0x00c000018188 by goroutine 9:
  main.vigilarMal()
      /src/04_gorrutinas/main.go:64 +0x56
  main.correr.gowrap2()
      /src/04_gorrutinas/main.go:84 +0xe

Previous read at 0x00c000018188 by main goroutine:
  main.monitor.View()
      /src/04_gorrutinas/main.go:52 +0x73
  charm.land/bubbletea/v2.(*Program).render()
      /go/pkg/mod/charm.land/bubbletea/v2@v2.0.10/tea.go:896 +0x8a
  charm.land/bubbletea/v2.(*Program).eventLoop()
      /go/pkg/mod/charm.land/bubbletea/v2@v2.0.10/tea.go:888 +0x1531
  charm.land/bubbletea/v2.(*Program).Run()
      /go/pkg/mod/charm.land/bubbletea/v2@v2.0.10/tea.go:1153 +0x14f4
  main.correr()
      /src/04_gorrutinas/main.go:86 +0x3f7
  main.main()
      /src/04_gorrutinas/main.go:94 +0xfa

Goroutine 9 (running) created at:
  main.correr()
      /src/04_gorrutinas/main.go:84 +0x3ed
  main.main()
      /src/04_gorrutinas/main.go:94 +0xfa
==================
Found 1 data race(s)
exit status 66
```

## Leccion 05 — tea.Batch no tiene orden

`go run ./05_batch_sequence` · codigo de salida `0`

```
  escritos en este orden: 1-leer-config (300 ms) · 2-conectar (100 ms) · 3-pedir-lista (200 ms)

== 5.1 · La trampa: tea.Batch ==
  llegó 1.º: 2-conectar (t=100 ms)
  llegó 2.º: 3-pedir-lista (t=200 ms)
  llegó 3.º: 1-leer-config (t=301 ms)

== 5.2 · Lo correcto cuando importa el orden: tea.Sequence ==
  llegó 1.º: 1-leer-config (t=300 ms)
  llegó 2.º: 2-conectar (t=401 ms)
  llegó 3.º: 3-pedir-lista (t=601 ms)

  Batch: más rápido, sin orden. Sequence: en orden, la suma de los tiempos.
```

## Leccion 06 — Componer modelos

`go run ./06_componer` · codigo de salida `0`

```
== 6.1 · La trampa: no reenviar los mensajes a los hijos ==
  tecleado «web» en 1 s -> filtro: "" · fotogramas distintos del spinner: 1
  ┌─ al final ─────────────┐
  │| buscando · filtro: [] │
  └────────────────────────┘

== 6.2 · Reenviados, pero la caja sin Focus() ==
  tecleado «web» en 1 s -> filtro: "" · fotogramas distintos del spinner: 4
  ┌─ al final ─────────────┐
  │- buscando · filtro: [] │
  └────────────────────────┘

== 6.3 · Lo correcto: reenviar, devolver sus comandos y enfocar ==
  tecleado «web» en 1 s -> filtro: "web" · fotogramas distintos del spinner: 4
  ┌─ al final ────────────────┐
  │- buscando · filtro: [web] │
  └───────────────────────────┘
```

## Leccion 07 — El ancho de lo que se ve no es len()

`go run ./07_ancho` · codigo de salida `0`

```
== 7.1 · Tres formas de medir lo mismo ==
  texto                   len()  runas  columnas
  web-01                      6      6         6
  caché (é de una runa)       6      5         5
  caché (e + acento)          7      6         5
  ✅ (icono)                  3      1         2
  ¿«caché» == «caché»? false   (se ven iguales; no lo son)

== 7.2 · La trampa: alinear con %-15s ==
  |web-01         |ok      |   12 ms|
  |almacén-norte |lento   |  870 ms|
  |facturación   |caído  |       —|
  |caché          |ok      |    3 ms|

== 7.3 · Lo correcto: rellenar por columnas con lipgloss.Width ==
  |web-01         |ok      |   12 ms|
  |almacén-norte  |lento   |  870 ms|
  |facturación    |caído   |       —|
  |caché          |ok      |    3 ms|

== 7.4 · Lo mismo, con un estilo de lipgloss ==
  |web-01         |ok      |   12 ms|
  |almacén-norte  |lento   |  870 ms|
  |facturación    |caído   |       —|
  |caché          |ok      |    3 ms|
```

## Leccion 08 — El modelo es un valor, sus slices no

`go run ./08_slices` · codigo de salida `0`

```
== 8.1 · La trampa: modificar el slice en sitio ==
  ┌─ despues (marcado web-01) ─┐
  │> [x] web-01                │
  │  [ ] web-02                │
  │  [ ] db-01                 │
  └────────────────────────────┘
  ┌─ antes (para deshacer) ─┐
  │> [x] web-01             │
  │  [ ] web-02             │
  │  [ ] db-01              │
  └─────────────────────────┘
  ¿el estado anterior sigue sin marcar? false

== 8.2 · Lo correcto: slices.Clone antes de modificar ==
  ┌─ despues (marcado web-01) ─┐
  │> [x] web-01                │
  │  [ ] web-02                │
  │  [ ] db-01                 │
  └────────────────────────────┘
  ┌─ antes (para deshacer) ─┐
  │> [ ] web-01             │
  │  [ ] web-02             │
  │  [ ] db-01              │
  └─────────────────────────┘
  ¿el estado anterior sigue sin marcar? true
```

## Leccion 09 — Probar el modelo sin terminal

`go test -v -count=1 ./09_probar` · codigo de salida `1`

```
=== RUN   TestPausaParaElContador
--- PASS: TestPausaParaElContador (0.00s)
=== RUN   TestQSale_Mal
--- PASS: TestQSale_Mal (0.00s)
=== RUN   TestQSale_Bien
    visor_test.go:39: q tenía que producir tea.QuitMsg y produjo visor.refrescoMsg
--- FAIL: TestQSale_Bien (1.00s)
FAIL
FAIL	curso-tui/09_probar	1.238s
FAIL
```

## Leccion 10 — El proyecto: nubepanel

### El panel contra nubecita, conducido sin teclado

`go run ./10_proyecto/demo -nubecita nubecita` · codigo de salida `0`

```
  nubecita sembrada: 3 buckets, 3 colas, 2 tablas
  ┌─ 1 · al arrancar ──────────────────────────────────────────────────────┐
  │nubepanel  refrescos: 1                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> copias-db                          2 objetos                          │
  │  fotos                              3 objetos                          │
  │  logs-2026                          5 objetos                          │
  │                                                                        │
  │                                                                        │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  ┌─ 2 · tab: las colas ───────────────────────────────────────────────────┐
  │nubepanel  refrescos: 1                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> correos                            0 mensajes                         │
  │  pedidos                            4 mensajes                         │
  │  pedidos-muertos                    1 mensaje                          │
  │                                                                        │
  │                                                                        │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  (otro proceso mete 6 mensajes en «correos»; nadie toca el panel)
  ┌─ 3 · se refresca solo ─────────────────────────────────────────────────┐
  │nubepanel  refrescos: 2                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> correos                            6 mensajes                         │
  │  pedidos                            4 mensajes                         │
  │  pedidos-muertos                    1 mensaje                          │
  │                                                                        │
  │                                                                        │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  ┌─ 4 · p: purgar «correos» ──────────────────────────────────────────────┐
  │nubepanel  refrescos: 3                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> correos                            6 mensajes                         │
  │  pedidos                            4 mensajes                         │
  │  pedidos-muertos                    1 mensaje                          │
  │                                                                        │
  │¿Purgar correos? Se borran 6 mensajes. (s/n)                            │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  ┌─ 5 · s: purgada ───────────────────────────────────────────────────────┐
  │nubepanel  refrescos: 4                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> correos                            0 mensajes                         │
  │  pedidos                            4 mensajes                         │
  │  pedidos-muertos                    1 mensaje                          │
  │                                                                        │
  │purgada: correos                                                        │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  ┌─ 6 · /muer: filtrando ─────────────────────────────────────────────────┐
  │nubepanel  refrescos: 4                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> pedidos-muertos                    1 mensaje                          │
  │                                                                        │
  │/muer                                                                   │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  ┌─ 7 · tab: las tablas ──────────────────────────────────────────────────┐
  │nubepanel  refrescos: 5                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> sesiones                           2 elementos                        │
  │  usuarios                           3 elementos                        │
  │                                                                        │
  │                                                                        │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  (se apaga nubecita)
  ┌─ 8 · sin servidor ────────────────────────────────────────────────────────────┐
  │nubepanel  refrescos: 6                                                        │
  │                                                                               │
  │ Buckets (3)  Colas (3)  Tablas (2)                                            │
  │                                                                               │
  │sin conexión: dial 127.0.0.1:60013: connectex: No se puede establecer una con… │
  │se reintenta solo cada 2s                                                      │
  │                                                                               │
  │                                                                               │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir        │
  └───────────────────────────────────────────────────────────────────────────────┘
  el panel terminó con q · error: <nil>
```

### Servido por ssh, sin -autorizadas

`go run ./10_proyecto/cliente_ssh   # sin ninguna credencial` · codigo de salida `0`

```
  entrando como admin con: ninguna credencial
  ┌─ lo que ve quien hace ssh ─────────────────────────────────────────────┐
  │nubepanel  refrescos: 1                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> copias-db                          2 objetos                          │
  │  fotos                              3 objetos                          │
  │  logs-2026                          5 objetos                          │
  │                                                                        │
  │                                                                        │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  q -> la sesión se cerró (err = <nil>)
```

### Con -autorizadas: sin credenciales

`go run ./10_proyecto/cliente_ssh` · codigo de salida `1`

```
  entrando como admin con: ninguna credencial
  rechazado: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none], no supported methods remain
exit status 3
```

### Con -autorizadas: otra clave

`go run ./10_proyecto/cliente_ssh -clave otra` · codigo de salida `1`

```
  entrando como admin con: la clave C:\Users\tu-usuario\AppData\Local\Temp\curso-tui-_0g2r230\otra
  rechazado: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain
exit status 3
```

### Con -autorizadas: la clave autorizada

`go run ./10_proyecto/cliente_ssh -clave buena` · codigo de salida `0`

```
  entrando como admin con: la clave C:\Users\tu-usuario\AppData\Local\Temp\curso-tui-_0g2r230\buena
  ┌─ lo que ve quien hace ssh ─────────────────────────────────────────────┐
  │nubepanel  refrescos: 1                                                 │
  │                                                                        │
  │ Buckets (3)  Colas (3)  Tablas (2)                                     │
  │                                                                        │
  │> copias-db                          2 objetos                          │
  │  fotos                              3 objetos                          │
  │  logs-2026                          5 objetos                          │
  │                                                                        │
  │                                                                        │
  │tab pestaña • ↑↓ moverse • / filtrar • r refrescar • p purgar • q salir │
  └────────────────────────────────────────────────────────────────────────┘
  q -> la sesión se cerró (err = <nil>)
```

### Escuchar en 0.0.0.0 sin -autorizadas

`nubepanel -ssh 0.0.0.0:23299` · codigo de salida `2`

```
sin -autorizadas entraría cualquiera: solo se permite 127.0.0.1
```

### El binario

`go build ./10_proyecto  (CGO_ENABLED=0)` · codigo de salida `0`

```
  normal                  17.2 MiB
  -ldflags="-s -w"        12.2 MiB
  linux/amd64 -s -w       11.9 MiB
```
