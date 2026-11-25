# Lista de revisión de un PR en Go

Para imprimir y tener al lado. Cada línea remite al fallo del curso que la demuestra.
Las marcadas con ⚠ no las ve ninguna de las herramientas medidas (`SALIDA.md`): solo tú.

## Antes de leer el código

- [ ] `gofmt`, `go vet` y `go test -race ./...` en verde. Si el CI no corre `-race`, pídelo.
- [ ] `golangci-lint` con una configuración que vaya más allá de la de serie (`.golangci.yml`).
- [ ] ¿Qué dice el `go` del `go.mod`? Por debajo de 1.22, la variable de bucle es una sola (14) ⚠

## Errores

- [ ] Ninguna llamada que devuelve `error` se queda sin mirar, tampoco con `_ =` (01)
- [ ] Ninguna función que devuelve `error` devuelve un puntero tipado que puede ser nil (02) ⚠
- [ ] Los errores se comparan con `errors.Is` / `errors.As`, nunca con `==` (03)
- [ ] Ningún `:=` dentro de un bloque tapa un `err` que se devuelve fuera (04)
- [ ] Cada `recover` convierte el pánico en un error que llega a quien llama (05) ⚠

## Goroutines, canales y context

- [ ] Cada goroutine tiene un camino para terminar aunque nadie lea su resultado (06) ⚠
- [ ] Las funciones que reciben un `ctx` lo usan, y no lo cambian por `context.Background()` (07)
- [ ] `wg.Add` va antes del `go`, o mejor `wg.Go` (Go 1.25) (13)
- [ ] Un plazo total se crea una vez, fuera del bucle; `time.After` dentro del `select` lo reinicia (15) ⚠

## Memoria compartida

- [ ] Un mapa o una variable que tocan varias goroutines tiene su candado (11)
- [ ] Ningún tipo con `sync.Mutex` dentro se copia: receptores por puntero (12)
- [ ] Un `append` sobre un slice que no es tuyo puede pisar el de otro: `slices.Clip` o copia (09) ⚠
- [ ] `for _, v := range` da copias: para modificar, `for i := range` y `xs[i]` (10)

## Recursos

- [ ] Ningún `defer` dentro de un `for`: va en una función aparte (08)
- [ ] Cada `resp.Body` se cierra, y se lee entero si quieres reutilizar la conexión (16)
- [ ] Ningún `http.Client` sin `Timeout` (ni `http.Get`, que usa el de serie) (17) ⚠

## Datos

- [ ] Los campos que van a JSON empiezan con mayúscula (18)
- [ ] Toda conversión a un entero más pequeño comprueba el rango: `strconv.ParseInt(s, 10, 16)` (19)
- [ ] Nada de `s +=` en un bucle: `strings.Builder` (20)

## Y la pregunta que ninguna herramienta hace

- [ ] Para cada fallo que encuentres: ¿hay una prueba que falla sin el arreglo? Si no, no está arreglado.
