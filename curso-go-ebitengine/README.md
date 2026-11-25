# 🎮 Ebitengine: haz un juego en Go y entiende el bucle de fotogramas — material del curso

Código del vídeo [«Ebitengine: haz un juego en Go y entiende el bucle de fotogramas (curso completo)»](https://youtu.be/Dc-zZWyL7oI) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/Dc-zZWyL7oI**

---

> 🎬 Acompaña al vídeo **«Ebitengine: haz un juego en Go y entiende el bucle de fotogramas (curso completo)»** del canal **Golandia** (solo miembros).

Código del curso de [Golandia](https://www.youtube.com/@golandia), exclusivo para miembros del canal.

> No va de hacer juegos: va del **bucle de fotogramas**, el modelo mental que le falta a casi todo el
> programador de servidores. La lógica y el dibujo, separados; tiempo fijo; y por qué tu lógica no
> puede depender de lo que tarde el dibujado. Todo medido.

Ebitengine **v2.10.4**. Compilado y abierto en Docker sobre una pantalla virtual (Xvfb), con OpenGL por
software: las capturas son ventanas de verdad, y los fotogramas por segundo son de ese entorno.

```bash
python ejecutar_todo.py                         # todo, y SALIDA.md
go test -v ./mundo                              # la lógica del juego, sin ventana
```

| # | carpeta | lo medido |
|---:|---|---|
| 1 | `01_bucle` | Update ~59 veces por segundo; Draw, cientos: dos relojes distintos |
| 2 | `02_mover_en_draw` | mover en Draw: 131 px con un monitor de 60 Hz, 271 con uno de 144 |
| 3 | `03_tps` | con `SetTPS(30)`, «1 px por tick» va a la mitad; «60 px por segundo», igual |
| 4 | `04_lento` | Draw lento: menos fotogramas, misma lógica. Update lento: todo a cámara lenta |
| 5 | `05_entrada` | dos pulsaciones: 62 disparos con `IsKeyPressed`, 2 con `IsKeyJustPressed` |
| 6 | `mundo` | 35,8 s de partida calculados en microsegundos, y la misma semilla da la misma partida |
| 7 | `07_esquiva` | el proyecto: la lógica de `mundo` en una ventana |

---

<!-- ficha-repos-miembros -->
## Ficha rápida

### Requisitos

- Go 1.26 o superior (para abrir la ventana en tu máquina, las bibliotecas gráficas de tu sistema).
- Docker (`ejecutar_todo.py` compila y prueba dentro de un contenedor).
- Python 3.10 o superior (para `ejecutar_todo.py`; solo biblioteca estándar salvo que se diga)

### Cómo ejecutarlo

```bash
git clone <este-repo> && cd curso-go-ebitengine
python ejecutar_todo.py
```

El detalle de cada lección está arriba, en este mismo README.

### Estructura de carpetas

```
curso-go-ebitengine/
├── 01_bucle/
│   └── main.go
├── 02_mover_en_draw/
│   └── main.go
├── 03_tps/
│   └── main.go
├── 04_lento/
│   └── main.go
├── 05_entrada/
│   ├── main.go
│   └── pulsar.sh
├── 07_esquiva/
│   └── main.go
├── capturas/
│   ├── 01_bucle-sin-vsync.png
│   ├── 01_bucle.png
│   ├── 02_mover-144hz.png
│   ├── 02_mover-60hz.png
│   └── 07_esquiva.png
├── comun/
│   └── comun.go
├── mundo/
│   ├── mundo.go
│   └── mundo_test.go
├── .gitattributes
├── .gitignore
├── correr.sh
├── Dockerfile
├── ejecutar_todo.py
├── FUENTES.md
├── go.mod
├── go.sum
├── README.md
└── SALIDA.md
```
