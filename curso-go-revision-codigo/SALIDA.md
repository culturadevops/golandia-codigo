# SALIDA — curso «Revisión de código en Go: 20 fallos reales que compilan»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-10-28 14:15:54Z`
- go version go1.27.2 linux/amd64 (imagen `golang:1.27.2`)
- gofmt -l: `limpio`

El PR compila: `go build -tags malo ./...` · código `0` · compila

## Fallo 01 — Error ignorado

`go test -race -tags malo ./01_error_ignorado/` · código de salida `1` (el PR)

```
    fallo_test.go:9: cantidad "tres" -> {Cliente:ana Cantidad:0}, err = nil: el pedido entra con cantidad 0
--- FAIL: TestDecodificarRechazaLoIlegible (0.00s)
FAIL	golandia/revision/01_error_ignorado	0.011s
```

`go test -race ./01_error_ignorado/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:11: cantidad "tres" -> pedido ilegible: json: cannot unmarshal string into Go struct field Pedido.cantidad of type int
--- PASS: TestDecodificarRechazaLoIlegible (0.00s)
ok  	golandia/revision/01_error_ignorado	1.013s
```

## Fallo 02 — El nil que no es nil

`go test -race -tags malo ./02_nil_no_es_nil/` · código de salida `1` (el PR)

```
    fallo_test.go:8: pedido bueno -> err != nil es true; tipo *pedidos.ErrValidacion; err = <nil>
--- FAIL: TestValidarPedidoBueno (0.00s)
FAIL	golandia/revision/02_nil_no_es_nil	0.011s
```

`go test -race ./02_nil_no_es_nil/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:10: pedido bueno -> err = <nil>
--- PASS: TestValidarPedidoBueno (0.00s)
ok  	golandia/revision/02_nil_no_es_nil	1.012s
```

## Fallo 03 — Errores comparados con ==

`go test -race -tags malo ./03_errores_con_igual/` · código de salida `1` (el PR)

```
    fallo_test.go:7: GET /pedidos/99 -> 500; esperaba 404 (el error va envuelto con %w)
--- FAIL: TestPedidoQueNoExisteDa404 (0.00s)
FAIL	golandia/revision/03_errores_con_igual	0.010s
```

`go test -race ./03_errores_con_igual/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:9: GET /pedidos/99 -> 404
--- PASS: TestPedidoQueNoExisteDa404 (0.00s)
ok  	golandia/revision/03_errores_con_igual	1.014s
```

## Fallo 04 — Shadowing con :=

`go test -race -tags malo ./04_shadowing/` · código de salida `1` (el PR)

```
    fallo_test.go:8: Total("10", "x", "5") = 10, err = nil: el error se quedó dentro del bucle
--- FAIL: TestTotalConUnaLineaMala (0.00s)
FAIL	golandia/revision/04_shadowing	0.010s
```

`go test -race ./04_shadowing/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:10: Total("10", "x", "5") -> línea 2: strconv.Atoi: parsing "x": invalid syntax
--- PASS: TestTotalConUnaLineaMala (0.00s)
ok  	golandia/revision/04_shadowing	1.014s
```

## Fallo 05 — recover que se traga el pánico

`go test -race -tags malo ./05_recover_traga/` · código de salida `1` (el PR)

```
2026/10/28 14:16:09 recuperado: assignment to entry in nil map
    fallo_test.go:14: Procesar -> err = nil, pero solo se hicieron 1 de 3 pasos: cobrado y sin enviar
--- FAIL: TestProcesarAvisaDelPanico (0.00s)
FAIL	golandia/revision/05_recover_traga	0.011s
```

`go test -race ./05_recover_traga/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:16: Procesar -> pánico procesando el pedido: assignment to entry in nil map (pasos hechos: 1 de 3)
--- PASS: TestProcesarAvisaDelPanico (0.00s)
ok  	golandia/revision/05_recover_traga	1.011s
```

## Fallo 06 — La goroutine que se fuga

`go test -race -tags malo ./06_goroutine_fuga/` · código de salida `1` (el PR)

```
    fallo_test.go:18: 200 llamadas con plazo vencido -> 200 goroutines siguen vivas, bloqueadas para siempre
--- FAIL: TestPrecioNoDejaGoroutines (0.54s)
FAIL	golandia/revision/06_goroutine_fuga	0.547s
```

`go test -race ./06_goroutine_fuga/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:20: 200 llamadas con plazo vencido -> goroutines de más: 0
--- PASS: TestPrecioNoDejaGoroutines (0.53s)
ok  	golandia/revision/06_goroutine_fuga	1.545s
```

## Fallo 07 — El context que se ignora

`go test -race -tags malo ./07_context_ignorado/` · código de salida `1` (el PR)

```
    fallo_test.go:16: el cliente colgó a los 50ms y Reservar siguió hasta los 2s (err = <nil>)
--- FAIL: TestReservarSeCortaSiElClienteCuelga (2.00s)
FAIL	golandia/revision/07_context_ignorado	2.012s
```

`go test -race ./07_context_ignorado/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:18: el cliente colgó a los 50ms -> Reservar volvió a los 50ms: context canceled
--- PASS: TestReservarSeCortaSiElClienteCuelga (0.05s)
ok  	golandia/revision/07_context_ignorado	1.062s
```

## Fallo 08 — defer dentro de un bucle

`go test -race -tags malo ./08_defer_en_bucle/` · código de salida `1` (el PR)

```
    fallo_test.go:8: 10 pedidos, pool de 3 -> pedido 104: no quedan conexiones libres (3 de 3 en uso)
--- FAIL: TestExportarDiezConPoolDeTres (0.00s)
FAIL	golandia/revision/08_defer_en_bucle	0.011s
```

`go test -race ./08_defer_en_bucle/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:10: 10 pedidos, pool de 3 -> exportados los 10
--- PASS: TestExportarDiezConPoolDeTres (0.00s)
ok  	golandia/revision/08_defer_en_bucle	1.011s
```

## Fallo 09 — El slice que comparte memoria

`go test -race -tags malo ./09_slice_compartido/` · código de salida `1` (el PR)

```
    fallo_test.go:12: cap(carrito) = 4; la cotización normal dice "envío exprés": la pisó la exprés
--- FAIL: TestDosCotizacionesDelMismoCarrito (0.00s)
FAIL	golandia/revision/09_slice_compartido	0.011s
```

`go test -race ./09_slice_compartido/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:14: normal: "envío normal" · exprés: "envío exprés"
--- PASS: TestDosCotizacionesDelMismoCarrito (0.00s)
ok  	golandia/revision/09_slice_compartido	1.013s
```

## Fallo 10 — range modifica una copia

`go test -race -tags malo ./10_range_copia/` · código de salida `1` (el PR)

```
    fallo_test.go:9: el banco confirmó 1 y 3 -> [{1 pendiente} {2 pendiente} {3 pendiente}]: se cambió la copia, no el pedido
--- FAIL: TestMarcarPagados (0.00s)
FAIL	golandia/revision/10_range_copia	0.011s
```

`go test -race ./10_range_copia/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:11: el banco confirmó 1 y 3 -> [{1 pagado} {2 pendiente} {3 pagado}]
--- PASS: TestMarcarPagados (0.00s)
ok  	golandia/revision/10_range_copia	1.011s
```

## Fallo 11 — Mapa sin candado

`go test -race -tags malo ./11_mapa_sin_candado/` · código de salida `1` (el PR)

```
WARNING: DATA RACE
Read at 0x00c0000dc900 by goroutine 14:
Previous write at 0x00c0000dc900 by goroutine 10:
  golandia/revision/11_mapa_sin_candado.(*Inventario).Vender()
      /src/11_mapa_sin_candado/malo.go:12 +0x98
fatal error: concurrent map read and map write
FAIL	golandia/revision/11_mapa_sin_candado	0.013s
(... y 3 carreras más del mismo tipo)
```

`go test -tags malo ./11_mapa_sin_candado/` (sin -race) · código de salida `1`

```
fatal error: concurrent map writes
FAIL	golandia/revision/11_mapa_sin_candado	0.006s
```

`go test -race ./11_mapa_sin_candado/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:22: 8 handlers x 1000 ventas -> stock -8000
--- PASS: TestVenderDesdeOchoHandlers (0.00s)
ok  	golandia/revision/11_mapa_sin_candado	1.013s
```

## Fallo 12 — sync.Mutex copiado

`go test -race -tags malo ./12_mutex_copiado/` · código de salida `1` (el PR)

```
    fallo_test.go:16: 100 pedidos atendidos -> el contador dice 0
--- FAIL: TestContadorDeCienPedidos (0.00s)
FAIL	golandia/revision/12_mutex_copiado	0.010s
```

`go test -race ./12_mutex_copiado/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:18: 100 pedidos atendidos -> el contador dice 100
--- PASS: TestContadorDeCienPedidos (0.00s)
ok  	golandia/revision/12_mutex_copiado	1.011s
```

## Fallo 13 — WaitGroup.Add dentro de la goroutine

`go test -race -tags malo ./13_waitgroup_add/` · código de salida `1` (el PR)

```
    fallo_test.go:15: 100 productos de 10, una CPU -> total 0: Wait volvió antes de que las goroutines hicieran su Add
--- FAIL: TestTotalDeCienProductos (0.00s)
FAIL	golandia/revision/13_waitgroup_add	0.010s
```

`go test -race ./13_waitgroup_add/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:17: 100 productos de 10, una CPU -> total 1000
--- PASS: TestTotalDeCienProductos (0.00s)
ok  	golandia/revision/13_waitgroup_add	1.017s
```

## Fallo 14 — La variable de bucle (Go 1.21 contra 1.22)

`go test -race -tags malo ./14_variable_de_bucle/` · código de salida `1` (el PR)

```
    fallo_test.go:14: avisos de 7, 8 y 9 -> ["pedido 9 listo" "pedido 9 listo" "pedido 9 listo"]
--- FAIL: TestUnAvisoPorPedido (0.00s)
FAIL	golandia/revision/14_variable_de_bucle	0.009s
```

`go test -race ./14_variable_de_bucle/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:16: avisos de 7, 8 y 9 -> ["pedido 7 listo" "pedido 8 listo" "pedido 9 listo"]
--- PASS: TestUnAvisoPorPedido (0.00s)
ok  	golandia/revision/14_variable_de_bucle	1.011s
```

## Fallo 15 — time.After dentro del bucle

`go test -race -tags malo ./15_time_after_bucle/` · código de salida `1` (el PR)

```
    fallo_test.go:23: plazo de 100ms -> Recoger tardó 3.1s (293 confirmaciones): cada mensaje reinicia el plazo
--- FAIL: TestRecogerRespetaElPlazo (3.10s)
FAIL	golandia/revision/15_time_after_bucle	3.115s
```

`go test -race ./15_time_after_bucle/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:25: plazo de 100ms -> Recoger volvió a los 100ms con 10 confirmaciones
--- PASS: TestRecogerRespetaElPlazo (0.10s)
ok  	golandia/revision/15_time_after_bucle	1.115s
```

## Fallo 16 — Body sin cerrar

`go test -race -tags malo ./16_body_sin_cerrar/` · código de salida `1` (el PR)

```
    fallo_test.go:30: 50 consultas al almacén -> 50 conexiones TCP nuevas
--- FAIL: TestDisponibleReutilizaLaConexion (0.07s)
FAIL	golandia/revision/16_body_sin_cerrar	0.075s
```

`go test -race ./16_body_sin_cerrar/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:32: 50 consultas al almacén -> 1 conexión TCP
--- PASS: TestDisponibleReutilizaLaConexion (0.01s)
ok  	golandia/revision/16_body_sin_cerrar	1.023s
```

## Fallo 17 — Cliente HTTP sin timeout

`go test -race -tags malo ./17_cliente_sin_timeout/` · código de salida `1` (el PR)

```
    fallo_test.go:25: paquetería colgada -> Cotizar sigue esperando tras 3s (y sin Timeout, esperaría para siempre)
--- FAIL: TestCotizarConUnaPaqueteriaColgada (3.01s)
FAIL	golandia/revision/17_cliente_sin_timeout	3.018s
```

`go test -race ./17_cliente_sin_timeout/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:23: paquetería colgada -> Cotizar volvió a los 1s: Get "http://127.0.0.1:40359": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
--- PASS: TestCotizarConUnaPaqueteriaColgada (1.00s)
ok  	golandia/revision/17_cliente_sin_timeout	2.024s
```

## Fallo 18 — JSON con campos sin exportar

`go test -race -tags malo ./18_json_sin_exportar/` · código de salida `1` (el PR)

```
    fallo_test.go:8: recibo del pedido 7 -> {} (err = <nil>)
--- FAIL: TestReciboEnJSON (0.00s)
FAIL	golandia/revision/18_json_sin_exportar	0.009s
```

`go test -race ./18_json_sin_exportar/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:10: recibo del pedido 7 -> {"id":7,"total":12.5}
--- PASS: TestReciboEnJSON (0.00s)
ok  	golandia/revision/18_json_sin_exportar	1.010s
```

## Fallo 19 — Conversión que desborda

`go test -race -tags malo ./19_overflow_conversion/` · código de salida `1` (el PR)

```
    fallo_test.go:8: Cantidad("40000") = -25536, err = nil: dio la vuelta y quedó negativa
--- FAIL: TestCantidadQueNoCabe (0.00s)
FAIL	golandia/revision/19_overflow_conversion	0.009s
```

`go test -race ./19_overflow_conversion/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:10: Cantidad("40000") -> cantidad "40000": strconv.ParseInt: parsing "40000": value out of range
--- PASS: TestCantidadQueNoCabe (0.00s)
ok  	golandia/revision/19_overflow_conversion	1.013s
```

## Fallo 20 — Concatenar strings en un bucle

`go test -race -tags malo ./20_concatenar_en_bucle/` · código de salida `1` (el PR)

```
    fallo_test.go:26: CSV de 5000 pedidos (106 KB) -> reservó 279384 KB de memoria
--- FAIL: TestCSVNoReservaDeMas (0.11s)
FAIL	golandia/revision/20_concatenar_en_bucle	0.125s
```

`go test -race ./20_concatenar_en_bucle/` · código de salida `0` (con el arreglo)

```
    fallo_test.go:28: CSV de 5000 pedidos (106 KB) -> reservó 316 KB de memoria
--- PASS: TestCSVNoReservaDeMas (0.00s)
ok  	golandia/revision/20_concatenar_en_bucle	1.021s
```

## Medida — el reporte CSV de 5000 pedidos, con benchmark

el PR (+=): `go test -bench CSV -benchmem -tags malo ./20_concatenar_en_bucle/`

```
goos: linux
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkCSV-24    	      32	  34172355 ns/op	285962490 B/op	   15027 allocs/op
ok  	golandia/revision/20_concatenar_en_bucle	1.099s
```

el arreglo (strings.Builder): `go test -bench CSV -benchmem  ./20_concatenar_en_bucle/`

```
goos: linux
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkCSV-24    	   10000	    116465 ns/op	  203847 B/op	   10001 allocs/op
ok  	golandia/revision/20_concatenar_en_bucle	1.171s
```

## Los linters, contra el PR y contra el arreglo

- staticcheck 2026.2.2 (0.8.2)
- golangci-lint has version 2.15.0 built with go1.27.0 from 25c5369a on 2026-10-16T11:03:09Z

### go vet

Con el PR (`malo.go`), lo que avisa y desaparece con el arreglo:

```
12  Sumar passes lock by value: golandia/revision/12_mutex_copiado.Contador contains sync.Mutex
13  WaitGroup.Add called from inside new goroutine
18  struct field id has json tag but is not exported
18  struct field total has json tag but is not exported
```

Avisos que siguen en el arreglo (`bien.go`): 0

### staticcheck

Con el PR (`malo.go`), lo que avisa y desaparece con el arreglo:

```
13  should call wg.Add(1) before starting the goroutine to avoid a race (SA2000)
18  unexported struct field cannot have non-ignored `json:"id"` tag (SA5008)
18  unexported struct field cannot have non-ignored `json:"total"` tag (SA5008)
18  struct type 'golandia/revision/18_json_sin_exportar.Recibo' doesn't have any exported fields, nor custom marshaling (SA9005)
```

Avisos que siguen en el arreglo (`bien.go`): 0

### golangci-lint (de serie)

Con el PR (`malo.go`), lo que avisa y desaparece con el arreglo:

```
01  Error return value of `json.Unmarshal` is not checked (errcheck)
12  copylocks: Sumar passes lock by value: golandia/revision/12_mutex_copiado.Contador contains sync.Mutex (govet)
13  waitgroup: WaitGroup.Add called from inside new goroutine (govet)
18  structtag: struct field id has json tag but is not exported (govet)
18  structtag: struct field total has json tag but is not exported (govet)
18  SA9005: struct type 'golandia/revision/18_json_sin_exportar.Recibo' doesn't have any exported fields, nor custom marshaling (staticcheck)
```

Avisos que siguen en el arreglo (`bien.go`): 2

```
16  Error return value of `resp.Body.Close` is not checked (errcheck)
17  Error return value of `resp.Body.Close` is not checked (errcheck)
```

### golangci-lint (.golangci.yml del curso)

Con el PR (`malo.go`), lo que avisa y desaparece con el arreglo:

```
01  Error return value of `json.Unmarshal` is not checked (errcheck)
03  comparing with == will fail on wrapped errors. Use errors.Is to check for a specific error (errorlint)
04  shadow: declaration of "err" shadows declaration at line 9 (govet)
04  error is not nil (line 12) but it returns nil (nilerr)
07  Non-inherited new context, use function like `context.WithXXX` instead (contextcheck)
07  unused-parameter: parameter 'ctx' seems to be unused, consider removing or renaming it as _ (revive)
08  defer: prefer not to defer inside loops (revive)
10  unusedwrite: unused write to field Estado (govet)
12  copylocks: Sumar passes lock by value: golandia/revision/12_mutex_copiado.Contador contains sync.Mutex (govet)
12  modifies-value-receiver: suspicious assignment to a by-value method receiver (revive)
13  waitgroup: WaitGroup.Add called from inside new goroutine (govet)
16  response body must be closed (bodyclose)
18  structtag: struct field id has json tag but is not exported (govet)
18  structtag: struct field total has json tag but is not exported (govet)
18  SA9005: struct type 'golandia/revision/18_json_sin_exportar.Recibo' doesn't have any exported fields, nor custom marshaling (staticcheck)
19  G109: Potential Integer overflow made by strconv.Atoi result conversion to int16/32 (gosec)
20  concat-loop: string concatenation in a loop (perfsprint)
```

Avisos que siguen en el arreglo (`bien.go`): 2

```
16  (*net/http.Client).Get must not be called. use (*net/http.Client).Do(*http.Request) (noctx)
17  (*net/http.Client).Get must not be called. use (*net/http.Client).Do(*http.Request) (noctx)
```

### La tabla: quién ve cada fallo

| # | fallo | go vet | staticcheck | golangci-lint (de serie) | golangci-lint (.golangci.yml del curso) | go test -race |
|---|---|---|---|---|---|---|
| 01 | Error ignorado | · | · | sí | sí | · |
| 02 | El nil que no es nil | · | · | · | · | · |
| 03 | Errores comparados con == | · | · | · | sí | · |
| 04 | Shadowing con := | · | · | · | sí | · |
| 05 | recover que se traga el pánico | · | · | · | · | · |
| 06 | La goroutine que se fuga | · | · | · | · | · |
| 07 | El context que se ignora | · | · | · | sí | · |
| 08 | defer dentro de un bucle | · | · | · | sí | · |
| 09 | El slice que comparte memoria | · | · | · | · | · |
| 10 | range modifica una copia | · | · | · | sí | · |
| 11 | Mapa sin candado | · | · | · | · | sí |
| 12 | sync.Mutex copiado | sí | · | sí | sí | · |
| 13 | WaitGroup.Add dentro de la goroutine | sí | sí | sí | sí | · |
| 14 | La variable de bucle (Go 1.21 contra 1.22) | · | · | · | · | · |
| 15 | time.After dentro del bucle | · | · | · | · | · |
| 16 | Body sin cerrar | · | · | · | sí | · |
| 17 | Cliente HTTP sin timeout | · | · | · | · | · |
| 18 | JSON con campos sin exportar | sí | sí | sí | sí | · |
| 19 | Conversión que desborda | · | · | · | sí | · |
| 20 | Concatenar strings en un bucle | · | · | · | sí | · |
| | **detectados, de 20** | **3** | **2** | **4** | **12** | **1** |

Se escapan a todas las herramientas: **7 de 20**

- 02 El nil que no es nil
- 05 recover que se traga el pánico
- 06 La goroutine que se fuga
- 09 El slice que comparte memoria
- 14 La variable de bucle (Go 1.21 contra 1.22)
- 15 time.After dentro del bucle
- 17 Cliente HTTP sin timeout

«go test -race» solo cuenta si hay una prueba que ejecute ese código a la vez.

