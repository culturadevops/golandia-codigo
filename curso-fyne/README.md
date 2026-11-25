# Fyne a fondo — material del curso

Código de estos vídeos del canal Golandia:

- [Tu primera ventana en Go no compila, y el error no menciona ni Go ni Fyne](https://youtu.be/6k4s_grihRg) (miembros)
- [Le pusiste un tamaño a tu botón y no hizo caso: los contenedores de Fyne](https://youtu.be/dPTSV1tp08A) (miembros)
- [El enlace de datos en Fyne: tocas el dato y la interfaz se entera sola](https://youtu.be/3eXhAWyOUIM) (miembros)
- [Dos ventanas idénticas y una está mal: la regla de hilos de Fyne](https://youtu.be/QK91oW0WaHY) (miembros)
- [50.000 objetos en una lista de Fyne, y solo construye nueve filas](https://youtu.be/tT0QQ1bZ3PA) (miembros)
- [Tu ventana se congela al llamar a la red: el patrón que lo arregla](https://youtu.be/VOkYcRf0y5s) (miembros)
- [Un botón que se apaga solo, y un diálogo que no bloquea (curso 7/10)](https://youtu.be/a0KyWE3syCo) (miembros)
- [Tu app arranca sin iconos en otro ordenador: empotrar con go:embed](https://youtu.be/KYzTH4RNIuY) (miembros)
- [Con cgo se acabó cruzar gratis: el precio real de repartir tu app en Go](https://youtu.be/858DECHnZ1o) (miembros)
- [Cuándo NO usar Fyne: los tres precios, y la que tiene más estrellas](https://youtu.be/WpLvfqf5VfI) (miembros)

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/6k4s_grihRg**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/dPTSV1tp08A**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/3eXhAWyOUIM**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/QK91oW0WaHY**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/tT0QQ1bZ3PA**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/VOkYcRf0y5s**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/a0KyWE3syCo**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/KYzTH4RNIuY**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/858DECHnZ1o**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/WpLvfqf5VfI**

---

Canal **Golandia** (YouTube) · curso exclusivo para miembros.

Un solo módulo de Go (`laboratorio/`) con una etapa por lección. Se compiló y se capturó dentro de Docker sobre Xvfb: las imágenes de `laboratorio/capturas/` son ventanas de verdad. Salida literal en `laboratorio/salida-real.txt`.

## Los vídeos

| # | carpeta | vídeo (título en YouTube) | id BD |
|---:|---|---|---:|
| 1 | [`01-la-primera-ventana/`](01-la-primera-ventana/) | Tu primera ventana en Go no compila, y el error no menciona ni Go ni Fyne | 188 |
| 2 | [`02-contenedores/`](02-contenedores/) | Le pusiste un tamaño a tu botón y no hizo caso: los contenedores de Fyne | 189 |
| 3 | [`03-binding/`](03-binding/) | El enlace de datos en Fyne: tocas el dato y la interfaz se entera sola | 190 |
| 4 | [`04-hilos/`](04-hilos/) | Dos ventanas idénticas y una está mal: la regla de hilos de Fyne | 191 |
| 5 | [`05-listas-grandes/`](05-listas-grandes/) | 50.000 objetos en una lista de Fyne, y solo construye nueve filas | 192 |
| 6 | [`06-conectar-nubecita/`](06-conectar-nubecita/) | Tu ventana se congela al llamar a la red: el patrón que lo arregla | 193 |
| 7 | [`07-formularios-y-dialogos/`](07-formularios-y-dialogos/) | Un botón que se apaga solo, y un diálogo que no bloquea (curso 7/10) | 194 |
| 8 | [`08-temas-e-iconos/`](08-temas-e-iconos/) | Tu app arranca sin iconos en otro ordenador: empotrar con go:embed | 195 |
| 9 | [`09-empaquetar/`](09-empaquetar/) | Con cgo se acabó cruzar gratis: el precio real de repartir tu app en Go | 196 |
| 10 | [`10-cuando-no/`](10-cuando-no/) | Cuándo NO usar Fyne: los tres precios, y la que tiene más estrellas | 197 |

## Requisitos

- Docker (la forma reproducible), o bien Go 1.24+ con un compilador de C y las cabeceras de X11 **y de Wayland** (`libgl1-mesa-dev xorg-dev libwayland-dev wayland-protocols libxkbcommon-dev`): sin las de Wayland, Fyne 2.8 no compila (es el gancho de la lección 1).
- Lección 6: nubecita corriendo (repo `curso-go-nube`).

## Cómo ejecutarlo

```bash
cd laboratorio
docker volume create gomodcache
docker build -t fynelab .
MSYS_NO_PATHCONV=1 docker run --rm -v gomodcache:/go/pkg/mod -v "$(pwd):/app" -w /app \
    fynelab sh -c "go mod tidy; sh correr.sh"
```

O en tu máquina, una etapa suelta: `cd laboratorio && go run ./etapas/e01-ventana`.

⚠️ Guarda los `.sh` con saltos de línea LF: con CRLF, `dash` falla con `set: Illegal option -`.

## Estructura

```
curso-fyne/
├── 01-la-primera-ventana/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 02-contenedores/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 03-binding/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 04-hilos/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 05-listas-grandes/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 06-conectar-nubecita/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 07-formularios-y-dialogos/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 08-temas-e-iconos/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 09-empaquetar/
│   ├── EN-PANTALLA.md
│   └── README.md
├── 10-cuando-no/
│   ├── EN-PANTALLA.md
│   └── README.md
├── laboratorio/
│   ├── capturas/
│   ├── etapas/
│   ├── correr.sh
│   ├── correr_e06.sh
│   ├── correr_e08.sh
│   ├── correr_e09.sh
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── LEEME.md
│   └── salida-real.txt
└── .gitignore
```

Cada carpeta de lección trae su `README.md` y, si el vídeo enseña terminales, un `EN-PANTALLA.md` con los comandos y salidas tal cual aparecen en el vídeo.
