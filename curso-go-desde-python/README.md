# 🐹 Go para quien ya sabe Python — material del curso

Código del vídeo [«Curso de Go para quien ya sabe Python»](https://youtu.be/LuYqer5bxxU) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/LuYqer5bxxU**

---

> 🎬 Acompaña al vídeo **«Curso de Go para quien ya sabe Python»** del canal **Golandia** (solo miembros).

Código del curso **«Go para quien ya sabe Python»** de [Golandia](https://www.youtube.com/@golandia),
exclusivo para miembros del canal.

> **No es un curso de Go desde cero.** Es para quien ya escribe Python y quiere pasarse a Go
> sin tropezar. Cada lección toma algo que en Python funciona de una manera, enseña cómo se hace
> en Go, y sobre todo **dónde te vas a tropezar**: casi todas las trampas de esta lista compilan,
> no dan error y dan un resultado mal.

## Cómo usar esto

```bash
git clone <este-repo> && cd curso-go-desde-python
go version                # 1.22 o superior
python --version          # 3.10 o superior, para las comparaciones
python ejecutar_todo.py   # corre las 12 lecciones, su lado Python, y regenera SALIDA.md
go run ./07_nil           # o una sola lección
```

Cero dependencias: solo la biblioteca estándar de Go y la de Python.

## Las 12 lecciones

| # | directorio | en Python | en Go, y la trampa |
|---:|---|---|---|
| 1 | `01_binario` | `python script.py` + intérprete | un binario que se copia y ya |
| 2 | `02_valor_cero` | `KeyError` | el mapa devuelve 0 sin avisar: la coma ok |
| 3 | `03_slices` | `hosts[:2]` copia | el trozo es una ventana, y `append` a veces pisa al vecino |
| 4 | `04_mapas` | el dict tiene orden | el orden de un map es aleatorio a propósito |
| 5 | `05_cadenas` | `len()` cuenta letras | `len()` cuenta bytes |
| 6 | `06_errores` | la excepción te para | el error es un valor, y `_` lo tira sin avisar |
| 7 | `07_nil` | `None` es `None` | un `error` que imprime `<nil>` y no es nil |
| 8 | `08_structs` | `self` es el objeto | el método con receptor por valor no modifica nada |
| 9 | `09_defer` | `with` cierra al salir del bloque | `defer` cierra al salir de la función |
| 10 | `10_goroutines` | — | main no espera, y el contador que pierde sumas |
| 11 | `11_paralelo` | hilos para esperar | 30 chequeos: en fila, a la vez y con límite |
| 12 | `12_proyecto` | el `revisar_servicios.py` | el mismo, en Go: tamaño y arranque medidos |

## La regla de la casa

`SALIDA.md` es la salida literal de `ejecutar_todo.py`, con la versión de Go, la de Python y la
fecha. Los números del vídeo salen de ahí. Si en tu máquina salen otros, es un dato: dilo.

---

<!-- ficha-repos-miembros -->
## Ficha rápida

### Requisitos

- Go 1.24 o superior.
- Python 3.10 o superior (el lado Python de cada lección y `ejecutar_todo.py`).

### Cómo ejecutarlo

```bash
git clone <este-repo> && cd curso-go-desde-python
python ejecutar_todo.py
```

El detalle de cada lección está arriba, en este mismo README.

### Estructura de carpetas

```
curso-go-desde-python/
├── 01_binario/
│   ├── main.go
│   └── python.py
├── 02_valor_cero/
│   ├── main.go
│   └── python.py
├── 03_slices/
│   ├── main.go
│   └── python.py
├── 04_mapas/
│   ├── main.go
│   └── python.py
├── 05_cadenas/
│   ├── main.go
│   └── python.py
├── 06_errores/
│   ├── main.go
│   └── python.py
├── 07_nil/
│   ├── main.go
│   └── python.py
├── 08_structs/
│   ├── main.go
│   └── python.py
├── 09_defer/
│   ├── main.go
│   └── python.py
├── 10_goroutines/
│   └── main.go
├── 11_paralelo/
│   ├── main.go
│   └── python.py
├── 12_proyecto/
│   ├── main.go
│   └── revisar_servicios.py
├── .gitignore
├── ejecutar_todo.py
├── go.mod
├── README.md
└── SALIDA.md
```
