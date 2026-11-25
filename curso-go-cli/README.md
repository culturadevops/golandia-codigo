# 🐹 Herramientas de línea de comandos en Go, de cero a release — material del curso

Código del vídeo [«Herramientas de línea de comandos en Go: de flag a GoReleaser (curso completo)»](https://youtu.be/VNgyEcqeLVk) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/VNgyEcqeLVk**

---

Código del curso **«Herramientas de línea de comandos en Go, de cero a release»** de
[Golandia](https://www.youtube.com/@golandia), exclusivo para miembros del canal.

A lo largo de 12 lecciones se construye **vigia**, una CLI para DevOps que revisa endpoints
HTTP: flags, subcomandos, configuración por capas, stdout/stderr, códigos de salida, `-json`
para `jq`, Ctrl+C, tests, versión embebida, compilación cruzada y release con GoReleaser.
Cada lección enseña cómo se hace y **la trampa que no da error**, con su medida.

> No es el curso de Bubble Tea (eso son interfaces de terminal) ni el vídeo de Bash contra Go
> para administradores. Aquí se hace una CLI de las que se meten en un pipe, en un cron o en
> un pipeline de CI.

## Requisitos

- **Docker** y **Python 3.10+**. Nada más: Go, Bash, jq y GoReleaser van en la imagen.
- Si prefieres tu máquina: Go 1.26+, Bash, `jq`, `file`, y GoReleaser v2 para la lección 12.

| pieza | versión con la que se grabó |
|---|---|
| Go | 1.27.2 (imagen `golang:1.27.2`) |
| Cobra | v1.10.2 (con pflag v1.0.9) |
| testscript | `github.com/rogpeppe/go-internal` v1.16.0 |
| golang.org/x/term | v0.47.0 |
| GoReleaser | v2.18.1 (imagen `goreleaser/goreleaser:v2.18.1`) |
| jq | 1.7 (Debian) |

## Cómo usar esto

```bash
git clone <este-repo> && cd curso-go-cli
python ejecutar_todo.py           # las 12 lecciones en Docker; reescribe SALIDA.md
python ejecutar_todo.py 05 08     # solo esas, por pantalla (no toca SALIDA.md)
```

`SALIDA.md` es la salida literal de la última corrida: todo número que sale en el vídeo
está ahí. Las latencias y los tiempos de arranque cambian de una máquina a otra.

Cada lección es un `demo.sh` que se puede correr suelto, en Linux o macOS con Go instalado
(lo de la lección 11 y 12 espera las herramientas de Linux: `file`, `ldd`, `sha256sum`):

```bash
bash 06_salida/demo.sh
```

El primer `demo.sh` que corras levanta el servidor de pruebas (`./banco`, en
`127.0.0.1:8080`) y lo deja corriendo; los binarios van a `/tmp/curso-cli-bin`.
Todo lo que escribe archivos lo hace en un directorio de `mktemp -d` y lo borra al terminar.

## Las 12 lecciones

| # | directorio | cómo se hace | la trampa |
|---:|---|---|---|
| 1 | `01_flag` | el paquete `flag` | un flag después de una URL deja de ser un flag; `-v false` |
| 2 | `02_subcomandos` | un `flag.FlagSet` por subcomando | un flag delante del subcomando no es de nadie |
| 3 | `03_cobra` | lo mismo con Cobra, **medido** | en pflag, `-timeout 1s` es `-t imeout` |
| 4 | `04_config` | flags > entorno > archivo > defecto, con `fs.Visit` | el flag con su valor por defecto pisa al archivo |
| 5 | `05_stdio` | datos a stdout, progreso a stderr, URLs por stdin | colores en un archivo; `/dev/null` «es una terminal» |
| 6 | `06_salida` | 0, 1, 2 y 130, probados desde Bash | `os.Exit` se salta el `defer`; un pipe se come el código |
| 7 | `07_json` | `-json` para `jq` | `time.Duration` en nanosegundos; `null` en vez de `[]` |
| 8 | `08_senales` | `signal.NotifyContext` + informe atómico | el informe de ayer, truncado por un Ctrl+C |
| 9 | `09_tests` | `run()` en el proceso, testscript y un golden | el test que lee tu `os.Getenv` |
| 10 | `10_version` | `-ldflags -X` y `debug.ReadBuildInfo` | un nombre mal escrito en `-X` no da error |
| 11 | `11_cruzada` | `GOOS`/`GOARCH`, `-trimpath -s -w` | cgo encendido: el binario sale dinámico |
| 12 | `12_release` | GoReleaser en snapshot, con versión fija | en snapshot no hay changelog |

Las lecciones 4 a 8 tienen una carpeta `ingenuo/` con la versión que compila, no da error y
se equivoca. La versión buena es la de `vigia/`.

## vigia, la CLI del curso

```bash
go build -o vigia ./vigia
./vigia revisar https://ejemplo.com/salud https://ejemplo.com/api
./vigia revisar -json -solo-caidos - < urls.txt | jq -r '.[].url'
./vigia revisar -informe informe.txt -timeout 3s URL...
./vigia config        # la configuración efectiva, y de dónde sale cada valor
./vigia version
echo $?   # 0 todo arriba · 1 hay caídos · 2 uso o configuración mal · 130 interrumpido
```

Configuración, de más fuerte a más débil: flags (`-timeout`, `-concurrencia`, `-config`),
entorno (`VIGIA_TIMEOUT`, `VIGIA_CONCURRENCIA`, `VIGIA_CONFIG`), archivo (`vigia.json`) y
valores por defecto (10s, 4).

```json
{"timeout": "5s", "concurrencia": 8}
```

Tests: `go test ./vigia` (tabla de códigos de salida en el propio proceso, scripts de
testscript en `vigia/testdata/script` y la ayuda contra `vigia/testdata/ayuda.golden`;
si cambias la ayuda a propósito, `go test ./vigia -run Ayuda -update`).

Release: `.goreleaser.yaml` en la raíz. `goreleaser release --snapshot --clean` construye
6 binarios (Linux, Windows y macOS, amd64 y arm64), sus archivos y `checksums.txt` sin
publicar nada.

## Estructura

```
banco/        servidor de pruebas: /ok, /roto, /lento?ms=N
medir/        lanza un programa N veces: mediana y p90 del arranque
comun/        revisar un endpoint (lo que comparten todas las lecciones)
vigia/        la CLI del curso, con sus tests
01_… 12_…     una carpeta por lección, cada una con su demo.sh
comun.sh      funciones de los demo.sh
correr.sh     corre las lecciones dentro del contenedor
Dockerfile    el laboratorio: Go, jq, file y GoReleaser con versión fija
```

## La regla de la casa

Nada de lo que sale en el vídeo está escrito a mano: sale de `SALIDA.md`, y `SALIDA.md`
sale de `ejecutar_todo.py`. Si algo no te da lo mismo, abre un comentario en el vídeo con
tu salida.
