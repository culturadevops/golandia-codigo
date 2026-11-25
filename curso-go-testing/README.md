# 🐹 Testing en Go para producción — material del curso

Código del vídeo [«Testing en Go para producción: curso completo (testcontainers, fuzzing, synctest y CI)»](https://youtu.be/8Te42fOTUDE) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/8Te42fOTUDE**

---

Código del curso **«Testing en Go para producción»** de
[Golandia](https://www.youtube.com/@golandia), exclusivo para miembros del canal.

No es un curso de «cómo se escribe un test». Es la **estrategia completa** sobre un servicio
pequeño de verdad (pedidos con descuento y envío, un repositorio Postgres y una pasarela de
cobros por HTTP): qué tipo de test va en cada capa, cuánto cuesta cada uno **medido**, y en cada
lección la trampa que no da error: el test que pasa en verde y no prueba nada.

## Cómo usar esto

```bash
git clone <este-repo> && cd curso-go-testing
docker version             # solo hace falta Docker (y Python 3.10+)
python ejecutar_todo.py    # corre las 12 lecciones en golang:1.27.2 y regenera SALIDA.md
python ejecutar_todo.py 05 07          # solo algunas lecciones
python ejecutar_todo.py --sin-docker   # con el Go de tu máquina (Linux o macOS)
```

Con Go instalado (1.25 o superior, por `testing/synctest`):

```bash
go test ./...                          # lo rápido: sin Docker, en segundos
go test -tags integracion ./...        # + Postgres de verdad con testcontainers
go test -tags trampa -v ./04_httptest  # las trampas de una lección (fallan A PROPÓSITO)
```

## Las etiquetas de compilación del repo

| etiqueta | qué añade | ¿falla? |
|---|---|---|
| *(ninguna)* | los tests buenos: rápidos, sin Docker | no |
| `integracion` | los de Postgres con testcontainers-go (lecciones 5 y 11) | no |
| `trampa` | las demostraciones de cada trampa | **sí, a propósito** |
| `refactor`, `conbug` | otras versiones del servicio de la lección 3 | según la lección |
| `paso2` | la versión intermedia de la lección 7 | sí (lo encuentra el fuzzing) |

## El servicio: `tienda/`

| paquete | qué es |
|---|---|
| `tienda` | pedidos, `Total` (10 % en el primer pedido, envío gratis desde 500,00), `Servicio.Crear`, `Recibo`, montos en centavos |
| `tienda/memoria` | el **fake**: un repositorio en memoria y una pasarela que aprueba |
| `tienda/postgres` | el repositorio de verdad, con pgx |
| `tienda/cobros` | el cliente HTTP de la pasarela: reintenta los 5xx con espera exponencial y respeta el `context` |
| `tienda/api` | la API HTTP (`POST /pedidos`, `GET /pedidos/{id}`) |

## Las 12 lecciones

| # | directorio | cómo se hace | la trampa que no da error |
|---:|---|---|---|
| 1 | `01_tabla` | tests de tabla, `t.Run`, `t.Parallel` | un fake compartido que «reinicias» en cada subtest; con `go 1.21` en el go.mod, todos los subtests prueban el último caso |
| 2 | `02_ayudantes` | `t.Cleanup`, `t.TempDir`, `t.Setenv`, `t.Context` | `defer` en el padre cierra antes de que corran los subtests paralelos; `t.Context()` ya está cancelado en el Cleanup |
| 3 | `03_fakes_mocks` | fake en memoria frente a gomock | un refactor sin cambio de comportamiento rompe 5 de 6 tests con mocks y 0 de 6 con el fake |
| 4 | `04_httptest` | `NewRecorder`, `NewServer`, `ConnState` | el handler pasa y la ruta da 404; `http.DefaultClient` sin timeout; el `Body` sin cerrar |
| 5 | `05_testcontainers` | Postgres real, `Snapshot`/`Restore`, `WithReuseByName` | la hora que guardas no es la que lees (microsegundos); Ryuk borra el contenedor «reutilizable» |
| 6 | `06_golden` | golden files con `-update` | `-update` acepta el fallo; un golden con CRLF falla sin diferencia visible (`.gitattributes`) |
| 7 | `07_fuzzing` | la propiedad «ida y vuelta» | dos fallos de signo en montos negativos que las pruebas de tabla no ven |
| 8 | `08_synctest` | `testing/synctest` con un `RoundTripper` en memoria | `time.Sleep` no se entera de que el context se canceló |
| 9 | `09_cobertura` | `-cover`, `-coverprofile`, `-coverpkg` | 100 % de cobertura con el fallo de la frontera dentro |
| 10 | `10_flaky` | `-count`, `-race`, `-shuffle` | el orden de un map, el estado global, la carrera de datos |
| 11 | `11_etiquetas` | `//go:build integracion` | renombras una función y el test con etiqueta se queda roto sin que nadie lo compile |
| 12 | `12_ci` | `go test -json` + `12_ci/resumen` | sin `pipefail`, una tubería se come el código de salida del test |

`.github/workflows/pruebas.yml` es una **plantilla** de CI con dos trabajos (rápidos e
integración). No se ha ejecutado en GitHub para el curso: lo ejecutado está en `SALIDA.md`.

## testcontainers dentro de un contenedor

`ejecutar_todo.py` corre Go dentro de `golang:1.27.2`. Para que testcontainers-go pueda arrancar
Postgres desde ahí, monta el socket de Docker del anfitrión (`-v /var/run/docker.sock:/var/run/docker.sock`)
y le dice cómo llegar a los puertos publicados:
`--add-host host.docker.internal:host-gateway` y `TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal`.
Funciona igual con Docker Desktop (Windows, macOS) y con Docker en Linux.

## Qué deja en tu Docker: nada

Cada Postgres declara un volumen anónimo para sus datos. Si un contenedor se borra sin `-v`, ese
volumen se queda huérfano para siempre, y cientos de ejecuciones son decenas de GB. Por eso:

- El contenedor de Go va con `docker run --rm` y **sin volúmenes con nombre**: la caché de
  módulos y de compilación vive dentro de él y se va con él.
- Los Postgres de testcontainers los borra `TerminateContainer` (en `TestMain` y en cada test) o,
  si el proceso muere antes, **Ryuk**. Los dos borran el contenedor **y su volumen**.
- El de `WithReuseByName` (lección 5) existe para sobrevivir entre ejecuciones, así que nadie lo
  termina desde el test. `ejecutar_todo.py` lo borra al final. Si lo usas a mano:

```bash
docker rm -f -v curso-go-testing-pg-reuso    # -v: también su volumen
# y si cortaste ejecutar_todo.py a mitad:
docker rm -f -v curso-go-testing-ejecutor
```

Con `TESTCONTAINERS_RYUK_DISABLED=true` nadie limpia por ti: cada contenedor que no termines se
queda corriendo, con su volumen. Compruébalo con `docker ps -a` y `docker volume ls -f dangling=true`.

## Versiones

Go 1.27.2 (el `go.mod` pide 1.25) · testcontainers-go v0.44.0 · pgx v5.11.0 · go.uber.org/mock v0.6.0 ·
Postgres `postgres:18.6-alpine`. Las exactas de cada ejecución están al principio de `SALIDA.md`.

## La regla de la casa

`SALIDA.md` es la salida literal de `ejecutar_todo.py`. Los números del vídeo salen de ahí.
Los tiempos dependen de tu máquina y los fallos flaky de la suerte: si en la tuya salen otros, es
un dato, dilo.
