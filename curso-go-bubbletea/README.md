# 🐹 La otra pantalla: interfaces de terminal en Go con Bubble Tea — material del curso

Código del vídeo [«Interfaces de terminal en Go con Bubble Tea: 10 trampas que no dan error (curso completo)»](https://youtu.be/C-hp8CN-FnY) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/C-hp8CN-FnY**

---

> 🎬 Acompaña al vídeo **«Interfaces de terminal en Go con Bubble Tea: 10 trampas que no dan error (curso completo)»** del canal **Golandia** (solo miembros).

Código del curso **«La otra pantalla: interfaces de terminal en Go con Bubble Tea»** de
[Golandia](https://www.youtube.com/@golandia), exclusivo para miembros del canal.

> Antes de una ventana, lo que escribe un programador de Go es una herramienta de terminal. Y una
> de verdad —con lista, filtro, refresco y ayuda— no es un `fmt.Println`: es una interfaz, con su
> bucle de eventos y su modelo de estado. Cada lección enseña cómo se hace y, sobre todo, **la
> trampa que no da error**. Al final se construye `nubepanel`, un panel para
> [nubecita](https://github.com/) servido también por ssh.

Bubble Tea **v2** (`charm.land/bubbletea/v2`). Si vienes de un tutorial de la v1, la lección 2
te ahorra una tarde.

## Cómo usar esto

```bash
go version                 # 1.26.8 o superior (con GOTOOLCHAIN=auto, Go lo baja solo)
python ejecutar_todo.py    # corre las 10 lecciones y regenera SALIDA.md
go run ./03_comandos       # o una sola lección
go run ./10_proyecto       # el panel, en tu terminal, contra una nubecita en 127.0.0.1:4566
```

La lección 10 necesita el código de nubecita (`python ejecutar_todo.py --nubecita ruta`). La
prueba de carreras de la lección 4 corre en Docker, porque `-race` necesita cgo y en Windows no
suele haber compilador de C.

Las lecciones corren **sin terminal**: arrancan un `tea.Program` de verdad con
`WithInput(nil)` y le mandan las teclas con `p.Send`. Así la salida se puede guardar literal en
`SALIDA.md`, que es de donde sale cada línea del vídeo.

## Las 10 lecciones

| # | directorio | cómo se hace | la trampa |
|---:|---|---|---|
| 1 | `01_el_bucle` | `Init` / `Update` / `View` | leer datos en `View`: 41 teclas, 46 lecturas |
| 2 | `02_teclas` | `tea.KeyPressMsg`, `String()` | `case " "` no salta en la v2 (es `"space"`); sin `ctrl+c` no se puede salir |
| 3 | `03_comandos` | `tea.Cmd` | llamar a la API dentro de `Update`: la pantalla se congela 1 s |
| 4 | `04_gorrutinas` | `p.Send` | tocar el estado desde otra gorrutina: la pantalla no se entera, y `-race` lo caza |
| 5 | `05_batch_sequence` | `tea.Batch`, `tea.Sequence` | `Batch` devuelve en orden de llegada, no de escritura |
| 6 | `06_componer` | bubbles dentro de tu modelo | no reenviar mensajes: spinner quieto; sin `Focus()`: caja muda |
| 7 | `07_ancho` | `lipgloss.Width` | `%-15s` cuenta runas: la tilde de macOS (NFD) tuerce la tabla |
| 8 | `08_slices` | `slices.Clone` | copiar el modelo no copia sus slices: el «deshacer» no deshace |
| 9 | `09_probar` | `go test` sobre `Update` | `cmd != nil` pasa en verde aunque `q` no salga (**falla a propósito**) |
| 10 | `10_proyecto` | `nubepanel` + wish | sin `-autorizadas`, por ssh entra **cualquiera** sin credenciales |

## El proyecto: nubepanel

```bash
go run ./10_proyecto -endpoint http://127.0.0.1:4566                   # en tu terminal
go run ./10_proyecto -ssh 127.0.0.1:23234                              # por ssh, solo local
go run ./10_proyecto -ssh 0.0.0.0:23234 -autorizadas ~/.ssh/authorized_keys
```

Buckets de S3, colas de SQS y tablas de DynamoDB; se refresca solo cada 2 s, filtra con `/`,
purga una cola con confirmación y avisa si se queda sin servidor. Habla con cualquier endpoint
compatible: nubecita, LocalStack, moto o AWS de verdad.

⚠️ Con AWS de verdad, el número de elementos de una tabla (`ItemCount` de `DescribeTable`) no es
en vivo: Amazon lo actualiza cada cierto tiempo. Contra nubecita sí lo es.

---

<!-- ficha-repos-miembros -->
## Ficha rápida

### Requisitos

- Go 1.26 o superior.
- Docker (lo usa `ejecutar_todo.py`).
- Python 3.10 o superior (para `ejecutar_todo.py`; solo biblioteca estándar salvo que se diga)

### Cómo ejecutarlo

```bash
git clone <este-repo> && cd curso-go-bubbletea
python ejecutar_todo.py
```

El detalle de cada lección está arriba, en este mismo README.

### Estructura de carpetas

```
curso-go-bubbletea/
├── 01_el_bucle/
│   └── main.go
├── 02_teclas/
│   └── main.go
├── 03_comandos/
│   └── main.go
├── 04_gorrutinas/
│   └── main.go
├── 05_batch_sequence/
│   └── main.go
├── 06_componer/
│   └── main.go
├── 07_ancho/
│   └── main.go
├── 08_slices/
│   └── main.go
├── 09_probar/
│   ├── visor.go
│   └── visor_test.go
├── 10_proyecto/
│   ├── cliente_ssh/
│   ├── demo/
│   ├── panel/
│   └── main.go
├── comun/
│   └── comun.go
├── .gitignore
├── ejecutar_todo.py
├── go.mod
├── go.sum
├── README.md
└── SALIDA.md
```
