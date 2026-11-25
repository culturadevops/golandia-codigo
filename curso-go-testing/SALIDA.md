# SALIDA — curso «Testing en Go para producción»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-10-28 18:14:59Z`
- go version go1.27.2 linux/amd64 (contenedor `golang:1.27.2`, --cpus 4)
- Docker `29.6.2` · Postgres `postgres:18.6-alpine` (descargadas antes de medir)
- `github.com/testcontainers/testcontainers-go v0.44.0`
- `github.com/jackc/pgx/v5 v5.11.0`
- `go.uber.org/mock v0.6.0`
- anfitrión: `Windows 10`
- testcontainers desde DENTRO del contenedor: socket montado (`-v /var/run/docker.sock`) y `TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal`

Las trampas fallan a propósito: cada paso dice el código de salida que se esperaba.

**0 pasos con un código distinto del esperado.**


## Lección 01 — Tests de tabla, subtests y t.Parallel

### Tabla con subtests paralelos

`$ go test -count=1 -v -run TestTotal ./01_tabla | grep -v '^=== '`

código de salida `0` (esperado `0`) · 4.1 s de reloj

```
--- PASS: TestTotal (0.00s)
    --- PASS: TestTotal/una_taza,_cliente_nuevo (0.00s)
    --- PASS: TestTotal/descuento_redondea_hacia_abajo (0.00s)
    --- PASS: TestTotal/cantidades (0.00s)
    --- PASS: TestTotal/una_taza,_cliente_de_siempre (0.00s)
    --- PASS: TestTotal/justo_500,00:_envío_gratis (0.00s)
    --- PASS: TestTotal/499,99:_paga_envío (0.00s)
PASS
ok  	curso-testing/01_tabla	0.002s
```

### La trampa: un fake compartido, 50 veces

`$ go test -tags trampa -count=50 -json -run TestCompartido ./01_tabla | go run ./12_ci/resumen -max-fallos 1 -lentos 0`

código de salida `1` (esperado `1`) · 0.3 s de reloj

```
1 paquetes · 50 tests: 0 pasan, 50 fallan, 0 saltados

FALLAN:
  curso-testing/01_tabla TestCompartido
      compartido_test.go:39: se cobró 3 veces; quiero 1
      compartido_test.go:39: se cobró 3 veces; quiero 1
      compartido_test.go:39: se cobró 4 veces; quiero 1
      compartido_test.go:39: se cobró 2 veces; quiero 1
      ... (146 líneas más)
  ... y 49 fallos más
exit status 1
```

### El mismo test, de uno en uno (-parallel 1)

`$ go test -tags trampa -count=50 -parallel 1 -json -run TestCompartido ./01_tabla | go run ./12_ci/resumen -lentos 0`

código de salida `0` (esperado `0`) · 0.4 s de reloj

```
1 paquetes · 50 tests: 50 pasan, 0 fallan, 0 saltados
```

### Con -race

`$ go test -tags trampa -count=1 -race -run TestCompartido ./01_tabla`

código de salida `1` (esperado `1`) · 5.7 s de reloj

```
==================
WARNING: DATA RACE
Write at 0x00c000018268 by goroutine 9:
  curso-testing/01_tabla.TestCompartido.func1()
      /src/01_tabla/compartido_test.go:35 +0x64
  testing.tRunner()
      /usr/local/go/src/testing/testing.go:2195 +0x21c
  testing.(*T).Run.gowrap1()
      /usr/local/go/src/testing/testing.go:2260 +0x38

Previous write at 0x00c000018268 by goroutine 12:
  curso-testing/01_tabla.TestCompartido.func1()
... (87 líneas más)
```

### El arreglo: un fake por subtest, 50 veces y con -race

`$ go test -tags trampa -count=50 -race -json -run TestPorSubtest ./01_tabla | go run ./12_ci/resumen -lentos 0`

código de salida `0` (esperado `0`) · 1.3 s de reloj

```
1 paquetes · 50 tests: 50 pasan, 0 fallan, 0 saltados
```

### go.mod con «go 1.21»: la variable del for compartida

`$ cd 01_tabla/go121 && go test -count=1 -v . | grep -v '^=== '`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
    abs_test.go:32: pruebo el caso "cero": Abs(0)
    abs_test.go:32: pruebo el caso "cero": Abs(0)
    abs_test.go:32: pruebo el caso "cero": Abs(0)
--- PASS: TestAbs (0.00s)
    --- PASS: TestAbs/positivo (0.00s)
    --- PASS: TestAbs/negativo (0.00s)
    --- PASS: TestAbs/cero (0.00s)
PASS
ok  	go121	0.002s
```

### El mismo test con «go 1.22»

`$ (copia temporal) sed -i 's/^go 1.21$/go 1.22/' go.mod && go test -v .`

código de salida `1` (esperado `1`) · 0.3 s de reloj

```
    abs_test.go:32: pruebo el caso "positivo": Abs(3)
    abs_test.go:32: pruebo el caso "cero": Abs(0)
    abs_test.go:32: pruebo el caso "negativo": Abs(-3)
    abs_test.go:34: Abs(-3) = -3; quiero 3
--- FAIL: TestAbs (0.00s)
    --- PASS: TestAbs/positivo (0.00s)
    --- PASS: TestAbs/cero (0.00s)
    --- FAIL: TestAbs/negativo (0.00s)
FAIL
FAIL	go121	0.002s
FAIL
```

## Lección 02 — t.Cleanup, t.TempDir, t.Setenv y t.Context

### Los ayudantes, bien usados

`$ go test -count=1 -v ./02_ayudantes | grep -v '^=== '`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
    ayudantes_test.go:18: dentro del test:  existe=true  /tmp/TestTempDirusa1974333037/001
    ayudantes_test.go:21: después del test: existe=false
--- PASS: TestTempDir (0.00s)
    --- PASS: TestTempDir/usa (0.00s)
    ayudantes_test.go:32: dentro:  TIENDA_MONEDA=USD
    ayudantes_test.go:34: después: TIENDA_MONEDA=MXN
--- PASS: TestSetenv (0.00s)
    --- PASS: TestSetenv/cambia (0.00s)
--- PASS: TestCleanupConParalelos (0.00s)
    --- PASS: TestCleanupConParalelos/creado (0.00s)
    --- PASS: TestCleanupConParalelos/enviado (0.00s)
    --- PASS: TestCleanupConParalelos/cobrado (0.00s)
    ayudantes_test.go:74: durante el test: t.Context().Err() = <nil>
--- PASS: TestContextEnCleanup (0.00s)
PASS
ok  	curso-testing/02_ayudantes	0.002s
```

### Las trampas: defer con subtests paralelos y t.Context en un Cleanup

`$ go test -tags trampa -count=1 -v -run 'TestDefer|TestContextCancelado' ./02_ayudantes | grep -v '^=== '`

código de salida `1` (esperado `1`) · 0.2 s de reloj

```
    trampa_test.go:26: escribiendo "creado": write /tmp/TestDefer1060196305/001/eventos.log: file already closed
    trampa_test.go:26: escribiendo "enviado": write /tmp/TestDefer1060196305/001/eventos.log: file already closed
    trampa_test.go:26: escribiendo "cobrado": write /tmp/TestDefer1060196305/001/eventos.log: file already closed
--- FAIL: TestDefer (0.00s)
    --- FAIL: TestDefer/creado (0.00s)
    --- FAIL: TestDefer/enviado (0.00s)
    --- FAIL: TestDefer/cobrado (0.00s)
    trampa_test.go:35: en el Cleanup: t.Context().Err() = context canceled
    trampa_test.go:37: la limpieza con t.Context() no llega a hacerse
--- FAIL: TestContextCancelado (0.00s)
FAIL
FAIL	curso-testing/02_ayudantes	0.004s
FAIL
```

### t.Setenv con t.Parallel

`$ go test -tags trampa -count=1 -run TestSetenvParalelo ./02_ayudantes`

código de salida `1` (esperado `1`) · 0.2 s de reloj

```
--- FAIL: TestSetenvParalelo (0.00s)
panic: testing: test using t.Setenv, t.Chdir, or cryptotest.SetGlobalRandom can not use t.Parallel [recovered, repanicked]

... (19 líneas más)
```

## Lección 03 — Fakes frente a mocks: lo que rompe un refactor

### Las 12 pruebas contra la versión «antes»

`$ go test -count=1 -json ./03_fakes_mocks | go run ./12_ci/resumen -lentos 0 -max-fallos 12`

código de salida `0` (esperado `0`) · 0.3 s de reloj

```
1 paquetes · 12 tests: 12 pasan, 0 fallan, 0 saltados
```

### Las 12 pruebas contra la versión «refactor sin cambio de comportamiento»

`$ go test -tags refactor -count=1 -json ./03_fakes_mocks | go run ./12_ci/resumen -lentos 0 -max-fallos 12`

código de salida `1` (esperado `1`) · 0.3 s de reloj

```
1 paquetes · 12 tests: 7 pasan, 5 fallan, 0 saltados

FALLAN:
  curso-testing/03_fakes_mocks TestMock_PrimerPedido
      servicio_refactor.go:42: Unexpected call to *fakesmocks.MockRepositorio.Guardar([context.Background.WithCancel &{0 ana [{TAZA 1 150,00}] 184,00 pagado cob-ana 0001-01-01 00:00:00 +0000 UTC}]) at /src/03_fakes_mocks/servicio_refactor.go:42 because:
      expected call at /src/03_fakes_mocks/mock_test.go:30 has already been called the max number of times
  curso-testing/03_fakes_mocks TestMock_SegundoPedido
      servicio_refactor.go:42: Unexpected call to *fakesmocks.MockRepositorio.Guardar([context.Background.WithCancel &{0 ana [{TAZA 1 150,00}] 199,00 pagado cob-ana 0001-01-01 00:00:00 +0000 UTC}]) at /src/03_fakes_mocks/servicio_refactor.go:42 because:
      expected call at /src/03_fakes_mocks/mock_test.go:41 has already been called the max number of times
  curso-testing/03_fakes_mocks TestMock_CobroRechazado
      servicio_refactor.go:32: Unexpected call to *fakesmocks.MockRepositorio.Guardar([context.Background.WithCancel &{0 moroso [{TAZA 1 150,00}] 184,00 pendiente  0001-01-01 00:00:00 +0000 UTC}]) at /src/03_fakes_mocks/servicio_refactor.go:32 because: there are no expected calls of the method "Guardar" for that receiver
      controller.go:97: missing call(s) to *fakesmocks.MockCobrador.Cobrar(is anything, is equal to moroso (string), is anything) /src/03_fakes_mocks/mock_test.go:51
      controller.go:97: aborting test due to missing call(s)
  curso-testing/03_fakes_mocks TestMock_EnvioGratis
      servicio_refactor.go:42: Unexpected call to *fakesmocks.MockRepositorio.Guardar([context.Background.WithCancel &{0 ana [{MOLINO 1 500,00}] 500,00 pagado cob-ana 0001-01-01 00:00:00 +0000 UTC}]) at /src/03_fakes_mocks/servicio_refactor.go:42 because:
      expected call at /src/03_fakes_mocks/mock_test.go:69 has already been called the max number of times
  curso-testing/03_fakes_mocks TestMock_QuedaPagado
      servicio_refactor.go:42: Unexpected call to *fakesmocks.MockRepositorio.Guardar([context.Background.WithCancel &{0 ana [{TAZA 1 150,00}] 184,00 pagado cob-ana 0001-01-01 00:00:00 +0000 UTC}]) at /src/03_fakes_mocks/servicio_refactor.go:42 because:
      expected call at /src/03_fakes_mocks/mock_test.go:81 has already been called the max number of times
exit status 1
```

### Las 12 pruebas contra la versión «con un fallo de verdad»

`$ go test -tags conbug -count=1 -json ./03_fakes_mocks | go run ./12_ci/resumen -lentos 0 -max-fallos 12`

código de salida `1` (esperado `1`) · 0.2 s de reloj

```
1 paquetes · 12 tests: 8 pasan, 4 fallan, 0 saltados

FALLAN:
  curso-testing/03_fakes_mocks TestFake_PrimerPedido
      fake_test.go:38: guardado: total 184,00, estado pendiente; quiero 184,00 pagado
  curso-testing/03_fakes_mocks TestFake_QuedaPagado
      fake_test.go:92: guardado &{ID:1 Cliente:ana Items:[{SKU:TAZA Cantidad:1 Precio:150,00}] Total:184,00 Estado:pendiente CobroID: Creado:0001-01-01 00:00:00 +0000 UTC}, cobros [184,00]
  curso-testing/03_fakes_mocks TestMock_CobroRechazado
      servicio_conbug.go:31: Unexpected call to *fakesmocks.MockRepositorio.Guardar([context.Background.WithCancel &{0 moroso [{TAZA 1 150,00}] 184,00 pendiente  0001-01-01 00:00:00 +0000 UTC}]) at /src/03_fakes_mocks/servicio_conbug.go:31 because: there are no expected calls of the method "Guardar" for that receiver
      controller.go:97: missing call(s) to *fakesmocks.MockCobrador.Cobrar(is anything, is equal to moroso (string), is anything) /src/03_fakes_mocks/mock_test.go:51
      controller.go:97: aborting test due to missing call(s)
  curso-testing/03_fakes_mocks TestMock_QuedaPagado
      mock_test.go:88: se guardó con estado "pendiente"; quiero pagado
exit status 1
```

## Lección 04 — httptest para servidores y clientes

### El handler, con NewRecorder

`$ go test -count=1 -v ./04_httptest | grep -v '^=== '`

código de salida `0` (esperado `0`) · 4.4 s de reloj

```
    handler_test.go:19: handler directo: 200 {"estado":"pagado","id":"42"}
--- PASS: TestVerPedido_Handler (0.00s)
PASS
ok  	curso-testing/04_httptest	0.003s
```

### El servidor entero, el cliente sin timeout y el Body sin cerrar

`$ go test -tags trampa -count=1 -v ./04_httptest | grep -v '^=== '`

código de salida `1` (esperado `1`) · 2.7 s de reloj

```
    handler_test.go:19: handler directo: 200 {"estado":"pagado","id":"42"}
--- PASS: TestVerPedido_Handler (0.00s)
    trampa_test.go:25: servidor entero: GET /pedidos/42 -> 404
    trampa_test.go:27: código 404; quiero 200
--- FAIL: TestVerPedido_Servidor (0.00s)
    trampa_test.go:52: http.DefaultClient       2.00 s  err=false
    trampa_test.go:52: Client{Timeout: 300ms}   0.30 s  err=true
--- PASS: TestClienteSinTimeout (2.30s)
    trampa_test.go:82: 20 peticiones, cerrar Body=false -> conexiones TCP abiertas: 20
    trampa_test.go:82: 20 peticiones, cerrar Body=true  -> conexiones TCP abiertas: 1
--- PASS: TestBodySinCerrar (0.00s)
FAIL
FAIL	curso-testing/04_httptest	2.315s
FAIL
```

### La API y el cliente de cobros de la tienda, con httptest.NewServer

`$ go test -count=1 -json ./tienda/api ./tienda/cobros | go run ./12_ci/resumen -lentos 3`

código de salida `0` (esperado `0`) · 0.4 s de reloj

```
2 paquetes · 4 tests: 4 pasan, 0 fallan, 0 saltados

Los 3 más lentos:
    0.05 s  curso-testing/tienda/cobros TestCobrarRespetaElContext
    0.00 s  curso-testing/tienda/api TestCrearYVer
    0.00 s  curso-testing/tienda/api TestAPI
```

## Lección 05 — testcontainers-go con Postgres de verdad

### Un contenedor para el paquete, Snapshot y Restore por test

`$ go test -tags integracion -count=1 -v ./05_testcontainers | grep -v '^=== '`

código de salida `0` (esperado `0`) · 9.7 s de reloj

```
arranque del contenedor (postgres:18.6-alpine): 2.18 s
Snapshot de la base migrada: 63 ms
    main_test.go:81: Restore: 35 ms
--- PASS: TestGuardarYBuscar (0.06s)
    main_test.go:81: Restore: 26 ms
--- PASS: TestActualizar (0.05s)
    main_test.go:81: Restore: 30 ms
--- PASS: TestNoExiste (0.05s)
    main_test.go:81: Restore: 31 ms
--- PASS: TestAislados (0.06s)
    repo_test.go:87: Guardar -> item FILTRO: ERROR: new row for relation "items" violates check constraint "items_cantidad_check" (SQLSTATE 23514)
    repo_test.go:88: pedidos de ana después del fallo: 0
    main_test.go:81: Restore: 34 ms
--- PASS: TestCantidadCeroDeshaceTodo (0.06s)
PASS
ok  	curso-testing/05_testcontainers	2.899s
```

### La hora: fake contra Postgres, y lo que cuesta un contenedor por test

`$ go test -tags 'integracion trampa' -count=1 -v -run 'TestHora|TestTresContenedores' ./05_testcontainers | grep -v '^=== '`

código de salida `1` (esperado `1`) · 7.1 s de reloj

```
arranque del contenedor (postgres:18.6-alpine): 1.63 s
Snapshot de la base migrada: 49 ms
    trampa_test.go:30: fake en memoria  guardé 12:00:00.123456789  leí 12:00:00.123456789  iguales=true
    trampa_test.go:30: Postgres         guardé 12:00:00.123456789  leí 12:00:00.123456000  iguales=false
    trampa_test.go:34: Postgres: la hora no vuelve igual
    main_test.go:81: Restore: 32 ms
--- FAIL: TestHoraIdaYVuelta (0.05s)
    trampa_test.go:47: contenedor 1: 1.25 s
    trampa_test.go:47: contenedor 2: 1.26 s
    trampa_test.go:47: contenedor 3: 1.12 s
--- PASS: TestTresContenedores (4.44s)
FAIL
FAIL	curso-testing/05_testcontainers	6.533s
FAIL
```

### WithReuseByName, primera vez

`$ go test -tags 'integracion trampa' -count=1 -v -run TestReuso ./05_testcontainers | grep -E 'Reuse|ok|FAIL'`

código de salida `0` (esperado `0`) · 3.7 s de reloj

```
    trampa_test.go:66: WithReuseByName: 1.26 s · contenedor 98c1690a3c2e
ok  	curso-testing/05_testcontainers	3.201s
```

### WithReuseByName, otra vez enseguida

`$ go test -tags 'integracion trampa' -count=1 -v -run TestReuso ./05_testcontainers | grep -E 'Reuse|ok|FAIL'`

código de salida `0` (esperado `0`) · 2.8 s de reloj

```
    trampa_test.go:66: WithReuseByName: 0.10 s · contenedor 98c1690a3c2e
ok  	curso-testing/05_testcontainers	2.240s
```

```
docker ps -a --filter name=curso-go-testing-pg-reuso  ->  desaparecido a los 8 s (lo borró Ryuk)
```

### WithReuseByName, después de esperar

`$ go test -tags 'integracion trampa' -count=1 -v -run TestReuso ./05_testcontainers | grep -E 'Reuse|ok|FAIL'`

código de salida `0` (esperado `0`) · 4.0 s de reloj

```
    trampa_test.go:66: WithReuseByName: 1.31 s · contenedor 8f351ba523f3
ok  	curso-testing/05_testcontainers	3.509s
```

### Sin Ryuk, primera vez

`$ TESTCONTAINERS_RYUK_DISABLED=true go test -tags 'integracion trampa' -count=1 -v -run TestReuso ./05_testcontainers | grep -E 'Reuse|ok|FAIL'`

código de salida `0` (esperado `0`) · 3.9 s de reloj

```
    trampa_test.go:66: WithReuseByName: 1.41 s · contenedor 8b4ff7cfa580
ok  	curso-testing/05_testcontainers	3.428s
```

### Sin Ryuk, 30 s después

`$ TESTCONTAINERS_RYUK_DISABLED=true go test -tags 'integracion trampa' -count=1 -v -run TestReuso ./05_testcontainers | grep -E 'Reuse|ok|FAIL'`

código de salida `0` (esperado `0`) · 2.0 s de reloj

```
    trampa_test.go:66: WithReuseByName: 0.08 s · contenedor 8b4ff7cfa580
ok  	curso-testing/05_testcontainers	1.605s
```

```
docker rm -f -v curso-go-testing-pg-reuso  ->  borrado, con su volumen
```

## Lección 06 — Golden files con -update

### Los recibos contra testdata/*.golden

`$ go test -count=1 -v ./06_golden | grep -v '^=== '`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
--- PASS: TestRecibo (0.00s)
    --- PASS: TestRecibo/envio_gratis (0.01s)
    --- PASS: TestRecibo/primer_pedido (0.01s)
    --- PASS: TestRecibo/rechazado (0.01s)
PASS
ok  	curso-testing/06_golden	0.010s
```

### Un cambio en Recibo (el descuento pierde el signo)

`$ (copia temporal, con el signo del descuento quitado) go test ./06_golden`

código de salida `1` (esperado `1`) · 0.6 s de reloj

```
--- FAIL: TestRecibo (0.00s)
    --- FAIL: TestRecibo/primer_pedido (0.00s)
        recibo_test.go:55: testdata/primer_pedido.golden no coincide (364 bytes, el golden 364):
              línea 7
                - "Descuento primer pedido           -15,00"
                + "Descuento primer pedido            15,00"
FAIL
FAIL	curso-testing/06_golden	0.001s
FAIL
```

### El mismo cambio, con -update

`$ (copia temporal) go test ./06_golden -update && go test ./06_golden && diff -u <antes> <después>`

código de salida `1` (esperado `1`) · 0.5 s de reloj

```
ok  	curso-testing/06_golden	0.002s
ok  	curso-testing/06_golden	0.001s
--- /src/06_golden/testdata/primer_pedido.golden	2026-10-28 18:13:23.313591500 +0000
+++ 06_golden/testdata/primer_pedido.golden	2026-10-28 18:16:46.137522967 +0000
@@ -4,7 +4,7 @@
 TAZA         1 x     150,00       150,00
 ----------------------------------------
 Subtotal                          150,00
-Descuento primer pedido           -15,00
+Descuento primer pedido            15,00
 Envío                              49,00
 TOTAL                             184,00
 Estado: pagado · cobro: cob-1
```

### El golden con finales de línea CRLF (lo que hace git en Windows sin .gitattributes)

`$ (copia temporal) sed -i 's/$/\r/' primer_pedido.golden && go test -run TestRecibo/primer ./06_golden`

código de salida `1` (esperado `1`) · 0.3 s de reloj

```
--- FAIL: TestRecibo (0.00s)
    --- FAIL: TestRecibo/primer_pedido (0.00s)
        recibo_test.go:55: testdata/primer_pedido.golden no coincide (364 bytes, el golden 374):
              línea 1
                - "Pedido #1 · ana\r"
                + "Pedido #1 · ana"
              línea 2
                - "Fecha: 2026-10-28 12:00 UTC\r"
                + "Fecha: 2026-10-28 12:00 UTC"
              línea 3
                - "----------------------------------------\r"
                + "----------------------------------------"
... (24 líneas más)
```

## Lección 07 — Fuzzing que encuentra un fallo real

### Las pruebas de tabla de la versión original

`$ go test -count=1 -v ./07_fuzzing | grep -v '^=== '`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
--- PASS: TestFormatear (0.00s)
--- PASS: TestParsear (0.00s)
--- PASS: FuzzIdaYVuelta (0.00s)
    --- PASS: FuzzIdaYVuelta/seed#0 (0.00s)
    --- PASS: FuzzIdaYVuelta/seed#1 (0.00s)
PASS
ok  	curso-testing/07_fuzzing	0.004s
```

### Fuzzing de la propiedad «ida y vuelta»

`$ (copia temporal) go test -fuzz=FuzzIdaYVuelta -fuzztime=60s ./07_fuzzing; cat testdata/...; go test ./07_fuzzing`

código de salida `1` (esperado `1`) · 4.4 s de reloj

```
fuzz: elapsed: 0s, gathering baseline coverage: 0/2 completed
fuzz: elapsed: 0s, gathering baseline coverage: 2/2 completed, now fuzzing with 4 workers
fuzz: elapsed: 0s, execs: 7 (66/sec), new interesting: 1 (total: 3)
--- FAIL: FuzzIdaYVuelta (0.11s)
    --- FAIL: FuzzIdaYVuelta (0.00s)
        montos_test.go:42: ida y vuelta: -122 -> "-1,-22" -> -78
    
    Failing input written to testdata/fuzz/FuzzIdaYVuelta/5948820d5aee3734
    To re-run:
    go test -run=FuzzIdaYVuelta/5948820d5aee3734
FAIL
exit status 1
FAIL	curso-testing/07_fuzzing	0.109s
--- testdata/fuzz/FuzzIdaYVuelta/
go test fuzz v1
int64(-122)
--- y ahora, go test sin -fuzz:
--- FAIL: FuzzIdaYVuelta (0.00s)
    --- FAIL: FuzzIdaYVuelta/5948820d5aee3734 (0.00s)
        montos_test.go:42: ida y vuelta: -122 -> "-1,-22" -> -78
FAIL
FAIL	curso-testing/07_fuzzing	0.001s
FAIL
```

### Arreglado Formatear (-tags paso2), otra vez

`$ (copia temporal) go test -tags paso2 -fuzz=FuzzIdaYVuelta -fuzztime=60s ./07_fuzzing`

código de salida `1` (esperado `1`) · 0.3 s de reloj

```
fuzz: elapsed: 0s, gathering baseline coverage: 0/3 completed
fuzz: elapsed: 0s, gathering baseline coverage: 0/3 completed
--- FAIL: FuzzIdaYVuelta (0.01s)
    --- FAIL: FuzzIdaYVuelta (0.00s)
        montos_test.go:42: ida y vuelta: -90 -> "-0,90" -> 90
    
    Failing input written to testdata/fuzz/FuzzIdaYVuelta/3a77333f630d984f
    To re-run:
    go test -run=FuzzIdaYVuelta/3a77333f630d984f
FAIL
exit status 1
FAIL	curso-testing/07_fuzzing	0.006s
```

### La versión de tienda/, 30 segundos de fuzzing

`$ (copia temporal) go test -fuzz=FuzzMontoIdaYVuelta -fuzztime=30s ./tienda`

código de salida `0` (esperado `0`) · 30.4 s de reloj

```
fuzz: elapsed: 30s, execs: 10136757 (0/sec), new interesting: 1 (total: 6)
PASS
ok  	curso-testing/tienda	30.085s
```

## Lección 08 — testing/synctest para código con tiempo

### Reintentos dentro de la burbuja

`$ go test -count=1 -v ./08_synctest | grep -v '^=== '`

código de salida `0` (esperado `0`) · 0.4 s de reloj

```
    cobros_test.go:50: llamadas: 4 · tiempo dentro de la burbuja: 7s
--- PASS: TestReintentos (0.00s)
    cobros_test.go:67: se rinde tras 4 llamadas, a los 7s
--- PASS: TestSeRinde (0.00s)
    cobros_test.go:79: cobros.Cliente: sale a los 2.5s con context deadline exceeded
--- PASS: TestCancelaDuranteLaEspera (0.00s)
PASS
ok  	curso-testing/08_synctest	0.003s
```

### Con reloj de verdad, y la versión con time.Sleep

`$ go test -tags trampa -count=1 -v -run 'Verdad|Sleep' ./08_synctest | grep -v '^=== '`

código de salida `1` (esperado `1`) · 7.3 s de reloj

```
    trampa_test.go:21: reloj de verdad: 7.00 s
--- PASS: TestReintentosRelojDeVerdad (7.00s)
    trampa_test.go:32: ReintentarConSleep: sale a los 3s con context deadline exceeded
    trampa_test.go:34: quiero salir a los 2.5s, cuando vence el context
--- FAIL: TestSleepNoSeEntera (0.00s)
FAIL
FAIL	curso-testing/08_synctest	7.008s
FAIL
```

## Lección 09 — Cobertura: -coverprofile, -coverpkg y el 100 % que miente

### 100 % de cobertura

`$ go test -count=1 -cover ./09_cobertura`

código de salida `0` (esperado `0`) · 0.4 s de reloj

```
ok  	curso-testing/09_cobertura	0.005s	coverage: 100.0% of statements
```

### Por función

`$ go test -count=1 -coverprofile=/tmp/c09.out ./09_cobertura >/dev/null && go tool cover -func=/tmp/c09.out`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
curso-testing/09_cobertura/envio.go:14:	Envio		100.0%
curso-testing/09_cobertura/envio.go:22:	Etiqueta	100.0%
total:					(statements)	100.0%
```

### El caso de la frontera

`$ go test -tags trampa -count=1 ./09_cobertura`

código de salida `1` (esperado `1`) · 0.2 s de reloj

```
--- FAIL: TestEnvioFrontera (0.00s)
    frontera_test.go:14: Envio(50000) = 4900; quiero 0
FAIL
FAIL	curso-testing/09_cobertura	0.003s
FAIL
```

### Etiqueta rota, y el test sin aserciones sigue en verde

`$ (copia temporal, Etiqueta siempre dice «envío gratis») go test -cover ./09_cobertura`

código de salida `0` (esperado `0`) · 0.3 s de reloj

```
ok  	curso-testing/09_cobertura	0.002s	coverage: 100.0% of statements
```

### La cobertura de tienda/ con sus propios tests

`$ go test -count=1 -cover ./tienda`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
ok  	curso-testing/tienda	0.003s	coverage: 32.2% of statements
```

### La de tienda/ con TODOS los tests del repo (-coverpkg)

`$ go test -count=1 -coverpkg=./tienda/... -coverprofile=/tmp/todo.out ./... >/dev/null && go tool cover -func=/tmp/todo.out | grep -E 'tienda/(monto|pedido).go|total'`

código de salida `0` (esperado `0`) · 1.5 s de reloj

```
curso-testing/tienda/monto.go:13:		String		100.0%
curso-testing/tienda/monto.go:19:		FormatearMonto	100.0%
curso-testing/tienda/monto.go:30:		ParsearMonto	86.4%
curso-testing/tienda/monto.go:62:		digitos		66.7%
curso-testing/tienda/pedido.go:41:		Validar		83.3%
curso-testing/tienda/pedido.go:53:		Subtotal	100.0%
curso-testing/tienda/pedido.go:62:		Envio		100.0%
curso-testing/tienda/pedido.go:70:		Descuento	100.0%
curso-testing/tienda/pedido.go:77:		Total		100.0%
curso-testing/tienda/pedido.go:100:		ahora		66.7%
curso-testing/tienda/pedido.go:110:		Crear		84.6%
curso-testing/tienda/pedido.go:132:		Recibo		100.0%
total:						(statements)	72.8%
```

## Lección 10 — Tests flaky: -count, -race y -shuffle

### El orden de un map, 100 veces

`$ go test -tags trampa -count=100 -json -run 'TestSKUs$' ./10_flaky | go run ./12_ci/resumen -max-fallos 1 -lentos 0`

código de salida `1` (esperado `1`) · 0.2 s de reloj

```
1 paquetes · 100 tests: 56 pasan, 44 fallan, 0 saltados

FALLAN:
  curso-testing/10_flaky TestSKUs
      trampa_test.go:17: got [FILTRO MOLINO CAFE TAZA]
      trampa_test.go:17: got [CAFE TAZA FILTRO MOLINO]
      trampa_test.go:17: got [CAFE TAZA FILTRO MOLINO]
      trampa_test.go:17: got [MOLINO CAFE TAZA FILTRO]
      ... (40 líneas más)
  ... y 43 fallos más
exit status 1
```

### Estado global, en el orden del archivo

`$ go test -tags trampa -count=1 -v -run 'TestMoneda|TestCambia' ./10_flaky | grep -E '^(---|ok|FAIL)'`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
--- PASS: TestMonedaPorDefecto (0.00s)
--- PASS: TestCambiaMoneda (0.00s)
ok  	curso-testing/10_flaky	0.005s
```

### Estado global, barajado con 10 semillas

`$ for s in 1 2 3 4 5 6 7 8 9 10; do go test -tags trampa -count=1 -shuffle=$s -run 'TestMoneda|TestCambia' ./10_flaky >/dev/null && echo "-shuffle=$s  pasa" || echo "-shuffle=$s  FALLA"; done`

código de salida `0` (esperado `0`) · 1.2 s de reloj

```
-shuffle=1  pasa
-shuffle=2  FALLA
-shuffle=3  FALLA
-shuffle=4  FALLA
-shuffle=5  pasa
-shuffle=6  FALLA
-shuffle=7  pasa
-shuffle=8  FALLA
-shuffle=9  FALLA
-shuffle=10  FALLA
```

### Una semilla que falla, reproducida

`$ go test -tags trampa -count=1 -v -shuffle=2 -run 'TestMoneda|TestCambia' ./10_flaky | grep -E 'shuffle|^(---|ok|FAIL)|got'`

código de salida `1` (esperado `1`) · 0.2 s de reloj

```
-test.shuffle 2
--- PASS: TestCambiaMoneda (0.00s)
    trampa_test.go:24: got "10,00 USD"
--- FAIL: TestMonedaPorDefecto (0.00s)
FAIL
FAIL	curso-testing/10_flaky	0.003s
FAIL
```

### La carrera de datos, 100 veces sin -race

`$ go test -tags trampa -count=100 -json -run 'TestVisitas$' ./10_flaky | go run ./12_ci/resumen -max-fallos 1 -lentos 0`

código de salida `1` (esperado `1`) · 0.2 s de reloj

```
1 paquetes · 100 tests: 82 pasan, 18 fallan, 0 saltados

FALLAN:
  curso-testing/10_flaky TestVisitas
      trampa_test.go:40: got 46; quiero 50
      trampa_test.go:40: got 49; quiero 50
      trampa_test.go:40: got 46; quiero 50
      trampa_test.go:40: got 49; quiero 50
      ... (14 líneas más)
  ... y 17 fallos más
exit status 1
```

### La carrera de datos, una vez con -race

`$ go test -tags trampa -count=1 -race -run 'TestVisitas$' ./10_flaky`

código de salida `1` (esperado `1`) · 0.3 s de reloj

```
==================
WARNING: DATA RACE
Read at 0x00c000018268 by goroutine 10:
  curso-testing/10_flaky.(*Visitas).Contar.func1()
      /src/10_flaky/flaky.go:37 +0x30

Previous write at 0x00c000018268 by goroutine 21:
  curso-testing/10_flaky.(*Visitas).Contar.func1()
... (57 líneas más)
```

### Las versiones arregladas: -race, -shuffle=on y -count=100

`$ go test -count=100 -race -shuffle=on -json ./10_flaky | go run ./12_ci/resumen -lentos 0`

código de salida `0` (esperado `0`) · 1.3 s de reloj

```
1 paquetes · 400 tests: 400 pasan, 0 fallan, 0 saltados
```

## Lección 11 — Build tags para separar los de integración

### go test ./... (sin etiquetas)

`$ go test -count=1 -json ./... | go run ./12_ci/resumen -lentos 3`

código de salida `0` (esperado `0`) · 0.9 s de reloj

```
13 paquetes · 41 tests: 38 pasan, 0 fallan, 3 saltados

Los 3 más lentos:
    0.05 s  curso-testing/tienda/cobros TestCobrarRespetaElContext
    0.00 s  curso-testing/07_fuzzing TestFormatear
    0.00 s  curso-testing/06_golden TestRecibo
```

### go test -tags integracion ./...

`$ go test -tags integracion -count=1 -json ./... | go run ./12_ci/resumen -lentos 3`

código de salida `0` (esperado `0`) · 3.2 s de reloj

```
14 paquetes · 47 tests: 44 pasan, 0 fallan, 3 saltados

Los 3 más lentos:
    1.67 s  curso-testing/11_etiquetas TestPagadosDePostgres
    0.05 s  curso-testing/tienda/cobros TestCobrarRespetaElContext
    0.05 s  curso-testing/05_testcontainers TestCantidadCeroDeshaceTodo
```

### Renombrar una función sin la etiqueta: todo verde

`$ (copia temporal, PagadosDe renombrada en informe.go e informe_test.go) go vet ./... && go test ./11_etiquetas`

código de salida `0` (esperado `0`) · 1.6 s de reloj

```
ok  	curso-testing/11_etiquetas	0.001s
```

### Y con la etiqueta

`$ (copia temporal, el mismo renombrado) go vet -tags integracion ./11_etiquetas`

código de salida `1` (esperado `1`) · 2.1 s de reloj

```
# curso-testing/11_etiquetas
# [curso-testing/11_etiquetas]
vet: 11_etiquetas/informe_integracion_test.go:41:13: undefined: PagadosDe
```

## Lección 12 — go test -json en CI

### Toda la suite, con integración, -race y -shuffle, por el resumen

`$ go test -tags integracion -race -shuffle=on -count=1 -json ./... | go run ./12_ci/resumen -lentos 5`

código de salida `0` (esperado `0`) · 16.2 s de reloj

```
14 paquetes · 47 tests: 44 pasan, 0 fallan, 3 saltados

Los 5 más lentos:
    1.72 s  curso-testing/11_etiquetas TestPagadosDePostgres
    0.07 s  curso-testing/05_testcontainers TestCantidadCeroDeshaceTodo
    0.06 s  curso-testing/05_testcontainers TestAislados
    0.06 s  curso-testing/05_testcontainers TestActualizar
    0.06 s  curso-testing/05_testcontainers TestGuardarYBuscar
```

### La trampa de la tubería: un test que falla, y el código de salida de la tubería

`$ go test -json ./09_cobertura | cat > /dev/null; echo $?   (sh, sin pipefail)`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
código de la tubería: 0
```

### Lo mismo con set -o pipefail

`$ set -o pipefail; go test -json ./09_cobertura | cat > /dev/null; echo $?`

código de salida `0` (esperado `0`) · 0.2 s de reloj

```
código de la tubería: 1
```
