# 🖼️ Wails: tu frontend web con un backend de Go de verdad — material del curso

Código del vídeo [«Wails: tu frontend web con un backend de Go de verdad (curso completo)»](https://youtu.be/tI6zlZU6K9g) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/tI6zlZU6K9g**

---

> 🎬 Acompaña al vídeo **«Wails: tu frontend web con un backend de Go de verdad (curso completo)»** del canal **Golandia** (solo miembros).

Código del curso de [Golandia](https://www.youtube.com/@golandia), exclusivo para miembros del canal.

> Wails es la interfaz gráfica de Go con más estrellas en GitHub, y no lleva navegador: le pide el
> suyo al sistema operativo. Este curso va de lo que no cuenta ningún tutorial: **qué les pasa a tus
> tipos al cruzar el puente** entre Go y JavaScript, qué pasa con los errores y los pánicos, los eventos
> que se pierden, el contexto que mata la aplicación, y cómo se empaqueta.

Wails **v2.16.0** (la estable; la v3 está en beta). Todo se compila y se abre en Docker, sobre una
pantalla virtual (Xvfb): las capturas de `capturas/` son ventanas de verdad. Lo de Windows y macOS se
dice, no se enseña.

## Cómo usar esto

```bash
docker build -t nubecita:curso <ruta-de-nubecita>    # para la lección 8
python ejecutar_todo.py                               # compila, abre, fotografía y regenera SALIDA.md
```

## Las 8 lecciones

| # | carpeta | lo que pasa sin dar error |
|---:|---|---|
| 1 | `01_ventana` | la ventana es el WebKit del sistema; Go presta métodos con `Bind` |
| 2 | `02_tipos` | un `int64` grande llega redondeado; un slice nil llega como `null`; las claves enteras, como texto |
| 3 | `03_errores` | el error llega como texto (adiós a `errors.Is`); un **pánico deja la promesa colgada para siempre** |
| 4 | `04_eventos` | emitir en `OnStartup` pierde eventos; con un «listo» de la ventana, llegan todos |
| 5 | `05_contexto` | el runtime con `context.Background()` **mata la aplicación** (código 1) |
| 6 | `06_motor` | el motor lo pone el sistema: se pregunta por capacidades, no por el userAgent |
| 7 | `07_empaquetar.sh` | `go build` a secas: código 0, 3 librerías y ninguna ventana. Hacen falta las etiquetas |
| 8 | `08_proyecto` | un inspector de colas para nubecita, con todo lo anterior |

---

<!-- ficha-repos-miembros -->
## Ficha rápida

### Requisitos

- Go 1.26 o superior.
- Docker (compila y fotografía las ventanas dentro de un contenedor).
- Python 3.10 o superior (para `ejecutar_todo.py`; solo biblioteca estándar salvo que se diga)

### Cómo ejecutarlo

```bash
git clone <este-repo> && cd curso-go-wails
python ejecutar_todo.py
```

El detalle de cada lección está arriba, en este mismo README.

### Estructura de carpetas

```
curso-go-wails/
├── 01_ventana/
│   ├── frontend/
│   └── main.go
├── 02_tipos/
│   ├── frontend/
│   └── main.go
├── 03_errores/
│   ├── frontend/
│   └── main.go
├── 04_eventos/
│   ├── frontend/
│   └── main.go
├── 05_contexto/
│   ├── frontend/
│   └── main.go
├── 06_motor/
│   ├── frontend/
│   └── main.go
├── 08_proyecto/
│   ├── frontend/
│   ├── sembrar/
│   └── main.go
├── capturas/
│   ├── 01_ventana.png
│   ├── 02_tipos.png
│   ├── 03_errores.png
│   ├── 04_eventos-bueno.png
│   ├── 04_eventos.png
│   ├── 05_contexto.png
│   ├── 06_motor.png
│   ├── 08_proyecto-sin-servidor.png
│   └── 08_proyecto-vivo.png
├── comun/
│   └── comun.go
├── .gitattributes
├── .gitignore
├── 07_empaquetar.sh
├── 08_proyecto.sh
├── capturar.sh
├── Dockerfile
├── ejecutar_todo.py
├── FUENTES.md
├── go.mod
├── go.sum
├── README.md
└── SALIDA.md
```
