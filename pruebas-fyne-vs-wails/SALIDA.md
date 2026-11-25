# Salida ejecutada — vídeo 232 «Fyne o Wails: lo medí»

> Corrida del 2026-09-10. **La misma app** en los dos: una ventana con un contador, un botón
> que suma y una lista de cinco elementos. Misma imagen base (`golang:1.25-bookworm`), misma
> caché de módulos, mismo método de medida. Reproducible con `bash demo.sh`.

---

## ⛔ PASO 0, BLOQUEANTE: ¿abren los dos una ventana de verdad?

```
   Fyne : ventana VIVA tras 8 s · enlaza 16 librerias del sistema
   Wails: ventana VIVA tras 8 s · enlaza 169 librerias del sistema
```

**Este paso existe por un fallo mío que casi publica un dato falso.** Ver el final.

---

## 🥇 LA TABLA

| | Fyne | Wails | |
|---|---|---|---|
| líneas de código (misma app) | **28** | 40 | + `wails.json` y `frontend/dist` |
| paquetes de sistema para compilar | **153** · 67 MB | 558 · 318 MB | 4,7× más |
| entorno de compilación | **1,6 GB** | 3,02 GB | |
| módulos en el grafo | **65** | 111 | |
| compilar (módulos ya bajados) | 82.905 ms | **20.156 ms** | Fyne 4,1× más lento |
| **tamaño del binario** | 31.928.984 B (30,4 MB) | **12.292.496 B (11,7 MB)** | **Wails 2,6× menos** |
| **librerías del sistema en marcha** | **16** | 169 | **Wails 10,6× más** |
| **memoria con la ventana abierta** | **152 MB** · 1 proceso | 377 MB · 3 procesos | **Fyne 2,5× menos** |
| cruzar a Windows desde Linux | **no** | **sí** | |

---

## 🥇 HALLAZGO 1 — el que «usa un navegador» pesa MENOS

```
   Fyne : 31928984 bytes
   Wails: 12292496 bytes
```

Va contra el tópico. Fyne se vende como «un solo archivo, sin navegador, sin JavaScript» y
produce un binario **2,6 veces más grande**. Porque Fyne se lleva dentro todo lo que dibuja:
el motor de trazado, las fuentes, los temas, los iconos. Wails no dibuja nada: **le pide el
navegador al sistema operativo**.

## 🥇 HALLAZGO 2 — y ese navegador prestado se paga

```
   Fyne : 16 librerias del sistema
      libGL.so.1
      libX11.so.6
      libwayland-client.so.0
   Wails: 169 librerias del sistema
      libgtk-3.so.0
      libjavascriptcoregtk-4.1.so.0
      libsoup-3.0.so.0
      libwebkit2gtk-4.1.so.0
```

**169 contra 16.** El binario de Wails es pequeño porque las 150 librerías que le faltan las
pone la máquina del usuario. Y en memoria:

```
   Fyne : 152 MB · arbol de 1 procesos
   Wails: 377 MB · arbol de 3 procesos
```

Lo que Wails se ahorra en disco lo gasta en RAM, con creces: **un motor de navegador entero**.

## 🥇 HALLAZGO 3 — Wails cruza de Linux a Windows. Fyne no

```
   Fyne : NO cruza:  imports fyne.io/fyne/v2/app  (necesita un toolchain de mingw)
   Wails: cruza
```

El motivo no es trivial y es lo bonito del hallazgo:

```
   CGO_ENABLED al cruzar Fyne: 0   -> y Fyne NECESITA cgo, asi que se cae
   CGO_ENABLED al cruzar Wails: 0  -> y a Wails NO le hace falta en Windows

   backend que Wails compila PARA WINDOWS:
     github.com/wailsapp/go-webview2/internal/w32
     github.com/wailsapp/go-webview2/pkg/combridge
     github.com/wailsapp/go-webview2/webviewloader
     github.com/wailsapp/go-webview2/pkg/edge
     github.com/wailsapp/wails/v2/internal/frontend/desktop/windows/winc
     github.com/wailsapp/wails/v2/internal/frontend/desktop/windows/win32
     github.com/wailsapp/wails/v2/internal/webview2runtime
```

En Linux, Wails usa WebKit **por cgo**. En Windows usa WebView2 con enlaces de **Go puro**
(llamadas al sistema, sin C). Por eso el binario de Windows sale sin cgo, y por eso cruza.

### Y lo ejecuté de verdad en Windows

El `.exe` salió de un contenedor de Linux. Cabecera `MZ`, 15.703.552 bytes, 1.471 apariciones
de «WebView2» dentro. Arrancado nativo en Windows 11:

```
   ventana: 'contador'
   procesos hijos: msedgewebview2.exe
```

Memida contra línea base, porque Windows ya tenía WebView2 corriendo por otras apps:

```
   LINEA BASE (sin mi app): 12 procesos msedgewebview2 · 756.0 MB
   CON MI APP:              18 procesos msedgewebview2 · 1.062,1 MB
   binario de Go:           27.2 MB

   => ATRIBUIBLE A MI APP: 6 procesos de WebView2 · 306,1 MB + 27,2 MB = 333,4 MB
```

**333 MB en Windows, 377 MB en Linux.** El número se sostiene en los dos sistemas.

## HALLAZGO 4 — Fyne compila 4 veces más lento

```
   Fyne : 82905 ms
   Wails: 20156 ms
```

Con los módulos ya descargados en los dos, así que esto no mide la red. Fyne compila **C**:
glfw y los enlaces de OpenGL. Wails enlaza contra librerías del sistema ya compiladas.

## HALLAZGO 5 — pero el código de Fyne es más corto, y es un solo fichero

```
   Fyne  (main.go):                      28
   Wails (main.go + app.go + index.html): 16 + 9 + 15 = 40
   ⚠️ y Wails necesita ademas wails.json y la carpeta frontend/dist
```

---

## ⛔ CINCO FALLOS MÍOS MIDIENDO ESTO

1. **Git Bash convierte `-w /app` en `C:/Program Files/Git/app`.** Docker fallaba... y el
   script **salía con código 0**. La primera corrida «pasó» sin haber compilado nada.
2. **Fyne no compila con lo que dice su propia documentación.** `libgl1-mesa-dev xorg-dev` no
   basta: glfw pide `wayland-client-core.h`. Faltan `libwayland-dev libxkbcommon-dev
   wayland-protocols`.
3. **Wails 2.15 exige Go ≥ 1.25.** Con 1.24 ni resuelve el módulo.
4. 🥇 **EL GRANDE, y por poco publico un dato falso.** `go build` a secas sobre un proyecto de
   Wails compila **sin el frontend de escritorio**. Sale un binario de 4.9 MB que enlaza solo
   `libc`, no abre ninguna ventana y **termina con código 0**. Yo lo di por bueno y ya había
   escrito «Wails pesa 6 veces menos que Fyne»: estaba comparando una app contra un programa
   vacío. Hacen falta las etiquetas `desktop,production`. Por eso el paso 0 de este banco
   comprueba que la ventana **se sostiene** antes de dar ningún número.
5. **La compilación «en frío» incluía descargar los módulos** — 26 contra 6. Eso medía mi red,
   no el compilador. Ahora la descarga va aparte y el cronómetro empieza con todo ya en disco.

## ⛔ Lo que NO mide este banco

- **Una app real.** Aquí hay un contador y una lista. El reparto cambia cuando la interfaz
  crece: Wails hereda todo el ecosistema web, Fyne tiene los widgets que tiene.
- **macOS.** No hay máquina. En macOS los dos usan lo nativo (Cocoa / WKWebView) y el equilibrio
  vuelve a moverse.
- **El arranque en frío percibido**, que depende del WebView2 del usuario y de si ya está en
  memoria: en esta máquina había 12 procesos suyos antes de empezar.
- **Que el usuario TENGA WebView2.** En Windows 11 viene; en instalaciones viejas no, y Wails
  entonces tiene que instalarlo.
