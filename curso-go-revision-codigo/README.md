# 🐹 Revisión de código en Go: 20 fallos reales que compilan — material del curso

Código del vídeo [«Revisión de código en Go: 20 fallos reales que compilan (curso completo)»](https://youtu.be/FClWiJJ7IPE) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/FClWiJJ7IPE**

---

Código del curso **«Revisión de código en Go: 20 fallos reales que compilan»** de
[Golandia](https://www.youtube.com/@golandia), exclusivo para miembros del canal.

Un compañero manda el **PR #42** del servicio de pedidos. Compila, `gofmt` no se queja y las
pruebas del PR pasan. Tiene 20 fallos. Cada carpeta de este repo es uno: el código tal como
llegó, la prueba que lo demuestra (falla con el código del PR y pasa con el arreglo) y el
arreglo. Al final, qué ven `go vet`, `staticcheck` y `golangci-lint`, medido, y cuáles se les
escapan.

## Cómo usar esto

```bash
git clone <este-repo> && cd curso-go-revision-codigo
docker version              # lo único que hace falta, además de Python 3.10+
python ejecutar_todo.py     # los 20 fallos, con el PR y con el arreglo, y los linters -> SALIDA.md
```

¿Con tu propio Go (1.25 o superior)?

```bash
go test -tags malo ./06_goroutine_fuga/   # el PR: la prueba FALLA
go test ./06_goroutine_fuga/              # el arreglo: la prueba pasa
go test -race -tags malo ./11_mapa_sin_candado/
diff 06_goroutine_fuga/malo.go 06_goroutine_fuga/bien.go
git log --oneline                         # un commit por arreglo
git checkout pr-42                        # el PR tal como llegó, sin pruebas ni arreglos
```

## Cómo está hecho cada fallo

| archivo | qué es |
|---|---|
| `malo.go` | el código del PR (`//go:build malo`) |
| `bien.go` | el arreglo (`//go:build !malo`) |
| `fallo_test.go` | la prueba: con `-tags malo` falla, sin él pasa |

Todos los paquetes se llaman `pedidos`: son piezas del mismo servicio, separadas para que cada
fallo tenga su prueba aislada.

## Los 20 fallos

| # | directorio | el fallo | lo que se mide |
|---:|---|---|---|
| 1 | `01_error_ignorado` | `json.Unmarshal` sin mirar el error | el pedido entra con cantidad 0 |
| 2 | `02_nil_no_es_nil` | un `*ErrValidacion` nil devuelto como `error` | `err != nil` con un pedido bueno |
| 3 | `03_errores_con_igual` | `err == ErrNoExiste` con un error envuelto | 500 en vez de 404 |
| 4 | `04_shadowing` | `:=` dentro del bucle tapa el `err` de fuera | total 10 y `err = nil` |
| 5 | `05_recover_traga` | `recover` que no devuelve el error | cobrado y sin enviar, con `nil` |
| 6 | `06_goroutine_fuga` | canal sin hueco + plazo vencido | 200 goroutines colgadas |
| 7 | `07_context_ignorado` | `context.Background()` en vez del de la petición | 2 s trabajando para nadie |
| 8 | `08_defer_en_bucle` | `defer` dentro del `for` | pool agotado en el pedido 4 |
| 9 | `09_slice_compartido` | `append` sobre el array de otro | una cotización pisa a la otra |
| 10 | `10_range_copia` | `range` por valor modifica una copia | ningún pedido queda pagado |
| 11 | `11_mapa_sin_candado` | mapa escrito desde varias goroutines | ventas perdidas o `fatal error` |
| 12 | `12_mutex_copiado` | receptor por valor con un `sync.Mutex` | el contador dice 0 |
| 13 | `13_waitgroup_add` | `wg.Add` dentro de la goroutine | `Wait` vuelve antes de tiempo |
| 14 | `14_variable_de_bucle` | la variable de bucle con reglas de Go 1.21 | tres avisos del pedido 9 |
| 15 | `15_time_after_bucle` | `time.After` dentro del `select` del bucle | plazo de 100 ms, tarda 3 s |
| 16 | `16_body_sin_cerrar` | `resp.Body` sin cerrar | 50 conexiones TCP en vez de 1 |
| 17 | `17_cliente_sin_timeout` | `&http.Client{}` sin `Timeout` | espera para siempre |
| 18 | `18_json_sin_exportar` | campos en minúscula con etiqueta `json` | el recibo sale `{}` |
| 19 | `19_overflow_conversion` | `int16(n)` sin comprobar | 40000 se vuelve -25536 |
| 20 | `20_concatenar_en_bucle` | `s +=` en un bucle | cientos de MB para un CSV de 106 KB |

Las cifras exactas están en `SALIDA.md`.

## Los linters

`ejecutar_todo.py` pasa cuatro herramientas por el PR y por el arreglo, con versión fija:

- `go vet` de Go 1.27.2 (imagen `golang:1.27.2`)
- `staticcheck` 2026.2.2
- `golangci-lint` v2.15.0 de serie (`--no-config`)
- `golangci-lint` v2.15.0 con el `.golangci.yml` de este repo (cópialo a tu proyecto)

Cuenta un fallo como «detectado» si la herramienta avisa en `malo.go` con algo que ya no dice
en `bien.go`. La tabla, y la lista de los que se escapan a todas, está al final de `SALIDA.md`.
Lo que ninguna herramienta ve es para lo que sirve la revisión: está en `CHECKLIST.md`.

## La regla de la casa

`SALIDA.md` es la salida literal de `ejecutar_todo.py`, con las versiones y la fecha. Los
números del vídeo salen de ahí. Los tiempos y la memoria dependen de tu máquina: si en la
tuya salen otros, es un dato.
