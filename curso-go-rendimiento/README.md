# 🐹 Rendimiento en Go medido: pprof, trace, GC y race — material del curso

Código del vídeo [«Rendimiento en Go medido: pprof, trace, GC y race (curso completo)»](https://youtu.be/eDlwclBwvDI) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/eDlwclBwvDI**

---

Código del curso **«Rendimiento en Go medido»** de [Golandia](https://www.youtube.com/@golandia),
exclusivo para miembros del canal.

Un servicio que resume registros de acceso (una línea JSON por petición) y que «funciona»...
despacio. Se arregla paso a paso, **midiendo antes de tocar y después de tocar**: benchmarks
con `benchstat`, perfiles de CPU y de memoria con `pprof`, análisis de escape, `sync.Pool`,
`go tool trace`, el recolector (`GOGC`, `GOMEMLIMIT`), el detector de carreras y `pprof` en un
servicio vivo. Cada lección enseña cómo se hace, **la trampa que no da error** y la medida.

## Cómo usar esto

```bash
git clone <este-repo> && cd curso-go-rendimiento
docker version               # todo corre en golang:1.27.2, con la CPU fija (--cpus=2)
python ejecutar_todo.py      # las 10 lecciones (~10 min) y regenera SALIDA.md
python ejecutar_todo.py --solo 06       # una lección, a la pantalla
python ejecutar_todo.py --cpus 4        # con otra cantidad de CPU
```

No necesitas Go instalado: `ejecutar_todo.py` levanta un contenedor `golang:1.27.2` con
`--cpus=2`, monta el repositorio en `/curso` y corre el `comandos.sh` de cada lección. Al terminar
borra el contenedor y los volúmenes de caché (`--conservar-cache` para que la siguiente vez vaya
más rápido). Perfiles, trazas y binarios van a `/tmp` del contenedor, nunca al repositorio.

¿Prefieres tu Go (1.27 o superior)? Casi todo funciona igual (`go test ./resumen -bench .`,
`go run ./06_trace`...), salvo la lección 7, que lee `/proc` y solo compila en Linux.

**Los números son de una máquina.** Con `--cpus=2`, Go 1.25+ ajusta `GOMAXPROCS` a 2 aunque el
anfitrión tenga más núcleos (por eso los benchmarks salen como `-2`). En tu máquina saldrán
otros milisegundos: lo que vale es una versión contra otra **en la misma máquina**, y con
`benchstat` diciendo si la diferencia es real.

## El servicio

`POST /resumen` recibe un registro y devuelve, por método y ruta (`/api/v1/pedidos/48213` →
`/api/v1/pedidos/:id`), cuántas peticiones hubo, cuántas fallaron (5xx) y la latencia p50 y p95.
El registro lo fabrica `datos/` con semilla fija: siempre los mismos bytes.

| versión | archivo | qué cambia |
|---|---|---|
| V1 | `resumen/v1.go` | la que «funciona»: `regexp.MustCompile` en el bucle, `map[string]any`, `Sprintf` |
| V2 | `resumen/v2.go` | la expresión regular, compilada una vez |
| V3 | `resumen/v3.go` | struct en vez de `map[string]any`, sin copiar el cuerpo, clave armada en la pila |
| V4 | `resumen/servicio.go` | búferes en `sync.Pool`... **con una condición de carrera** |
| V5 | `resumen/servicio.go` | V4 arreglada: escribe la respuesta antes de devolver el búfer |

`TestTodasIguales` comprueba que las cinco devuelven exactamente los mismos bytes.

## Las 10 lecciones

| # | directorio | cómo se hace | la trampa |
|---:|---|---|---|
| 1 | `01_benchmarks` | `b.Loop`, `-benchmem`, `-count 10`, `benchstat` | 0,14 ns/op: el compilador quitó lo que medías; con `-count 3`, benchstat no puede afirmar nada |
| 2 | `02_pprof_cpu` | `-cpuprofile`, `top -cum`, `list`, `peek` | el `top` por defecto (flat) reparte la culpa entre el runtime y no señala tu línea |
| 3 | `03_memoria` | `-sample_index=alloc_space` / `inuse_space` | un trozo de string retiene el cuerpo entero: el perfil señala dónde se asignó, no quién lo retiene |
| 4 | `04_escape` | `-gcflags=-m` | `var e` dentro del bucle + `&e` a un `any` = una asignación por línea |
| 5 | `05_pool` | `sync.Pool` para los búferes | si el objeto no escapa, la pila es más barata que el pool; el pool se vacía en dos GC |
| 6 | `06_trace` | `runtime/trace`, `go tool trace -pprof=sync/sched` | el candado alrededor del trabajo caro: 8 goroutines, casi todo el tiempo esperando |
| 7 | `07_gc` | `GOGC`, `GOMEMLIMIT`, `GODEBUG=gctrace=1` | Go no lee el límite de memoria del contenedor; un `GOMEMLIMIT` por debajo de lo vivo |
| 8 | `08_carrera` | `go test -race` | el búfer devuelto al pool antes de usar sus bytes: pasa 200 de 200 veces sin `-race` |
| 9 | `09_vivo` | `net/http/pprof` en `127.0.0.1:6060` | `import _ "net/http/pprof"` + `ListenAndServe(":8080", nil)` = perfiles y argumentos en internet |
| 10 | `10_carga` | carga con clientes concurrentes | el antes y el después del servicio entero |

Cada directorio tiene su `comandos.sh`, que es exactamente lo que corre `ejecutar_todo.py`.

## Versiones

- Go 1.27.2 (imagen `golang:1.27.2`). `b.Loop` necesita Go 1.24; el `GOMAXPROCS` que respeta el
  límite de CPU del contenedor, Go 1.25.
- `benchstat` de `golang.org/x/perf`, fijado en `go.mod` como herramienta (`go tool benchstat`).
- Ninguna otra dependencia: el servicio solo usa la biblioteca estándar.

## La regla de la casa

`SALIDA.md` es la salida literal de `ejecutar_todo.py`, con la versión de Go, la CPU y la fecha.
Los números del vídeo salen de ahí. Si en tu máquina salen otros, es un dato: dilo.
