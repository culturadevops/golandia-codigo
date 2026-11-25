# 🐹 Observabilidad en Go: slog, Prometheus y OpenTelemetry — material del curso

Código del vídeo [«Observabilidad en Go: slog, Prometheus y OpenTelemetry, y las trampas que no dan error»](https://youtu.be/yMvWuqYfprU) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/yMvWuqYfprU**

---

> 🎬 Acompaña al vídeo **«Observabilidad en Go: slog, Prometheus y OpenTelemetry, y las trampas que no dan error»** del canal **Golandia** (solo miembros).

Código del curso **«Observabilidad en Go»** de [Golandia](https://www.youtube.com/@golandia),
exclusivo para miembros del canal.

> Contado desde el lado del que **escribe** la aplicación, no del que monta el stack (eso es el
> curso de observabilidad de Cultura DevOps). Cada lección enseña cómo se instrumenta y, sobre
> todo, **la trampa que no da error**: el registro que filtra una clave, la métrica que no existe
> hasta el primer fallo, el p99 que miente y la traza que se parte en dos.

Go **1.26**, `client_golang` **v1.24.1**, OpenTelemetry **v1.47.0**, `otelhttp` **v0.72.0**.

## Cómo usar esto

Todo corre en Docker (`golang:1.26`). Las lecciones 3, 4 y 5 corren junto a un **Prometheus de
verdad** (`prom/prometheus:v3.1.0`, ver `prometheus.yml`) que las raspa cada segundo: cada PromQL
que sale en el vídeo es la respuesta de ese Prometheus.

```bash
python ejecutar_todo.py          # las 8 lecciones y SALIDA.md
python ejecutar_todo.py 05       # solo una
```

## Las 8 lecciones

| # | directorio | cómo se hace | la trampa |
|---:|---|---|---|
| 1 | `01_slog` | atributos, `slog.LogValuer` | una clave olvidada sale como `!BADKEY`; el struct entero filtra la clave |
| 2 | `02_niveles` | `Enabled`, `LogValuer` perezoso | 10.000 `Debug` que no escriben nada serializan 10.000 veces |
| 3 | `03_sin_iniciar` | crear las series al arrancar | la serie de errores no existe; y los 5 primeros errores no se ven en `increase` |
| 4 | `04_cardinalidad` | la plantilla de la ruta | 1.000 ids = 1.000 series |
| 5 | `05_histogramas` | buckets para tu rango | con `DefBuckets`, p99 = 10 s cuando es 15 s |
| 6 | `06_trazas` | pasar el `ctx`, `otelhttp`, propagador | un `context.Background()` parte la traza; sin propagador, `otelhttp` no propaga |
| 7 | `07_spans_perdidos` | `defer tp.Shutdown`, `defer span.End` | sin `Shutdown`, 0 de 100 spans llegan |
| 8 | `08_todo_junto` | `trace_id` en cada registro, `r.Pattern` | — el servicio entero, con las tres señales enlazadas |

`01_slog/malo` tiene a propósito la llamada que `go vet` rechaza (lección 1.2).

## Lo que hay que saber de las medidas

- En la lección 5 las latencias se **observan** (`h.Observe`), no se esperan: lo que se mide es lo
  que Prometheus calcula a partir de los buckets.
- `increase()` extrapola a los bordes de la ventana: por eso la lección 3 da «cerca de 5».

---

<!-- ficha-repos-miembros -->
## Ficha rápida

### Requisitos

- Go 1.26 o superior.
- Docker (`ejecutar_todo.py` levanta Prometheus y la app en contenedores).
- Python 3.10 o superior (para `ejecutar_todo.py`; solo biblioteca estándar salvo que se diga)

### Cómo ejecutarlo

```bash
git clone <este-repo> && cd curso-go-observabilidad
python ejecutar_todo.py
```

El detalle de cada lección está arriba, en este mismo README.

### Estructura de carpetas

```
curso-go-observabilidad/
├── 01_slog/
│   ├── malo/
│   └── main.go
├── 02_niveles/
│   └── main.go
├── 03_sin_iniciar/
│   └── main.go
├── 04_cardinalidad/
│   └── main.go
├── 05_histogramas/
│   └── main.go
├── 06_trazas/
│   └── main.go
├── 07_spans_perdidos/
│   └── main.go
├── 08_todo_junto/
│   └── main.go
├── comun/
│   └── comun.go
├── .gitattributes
├── .gitignore
├── ejecutar_todo.py
├── go.mod
├── go.sum
├── prometheus.yml
├── README.md
└── SALIDA.md
```
