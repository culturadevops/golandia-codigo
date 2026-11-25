# SALIDA — curso «Observabilidad en Go: slog, Prometheus y OpenTelemetry»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-10-06 19:27:13Z`
- go version go1.26.8 linux/amd64
- `github.com/prometheus/client_golang v1.24.1`
- `go.opentelemetry.io/otel v1.47.0`
- `go.opentelemetry.io/otel/sdk v1.47.0`
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0`
- Prometheus: `prom/prometheus:v3.1.0`
- Docker `29.6.2`

## Leccion 01 — slog: atributos, la clave que falta y el secreto

`go run ./01_slog` · codigo de salida `0`

```
== 1.1 · Un texto armado con Sprintf contra atributos ==
  {"level":"INFO","msg":"pedido 42 pagado por ana"}
  {"level":"INFO","msg":"pedido pagado","pedido":42,"usuario":"ana"}
  el segundo se puede filtrar por usuario="ana"; el primero, solo con una expresión regular

== 1.2 · La trampa: se olvida una clave ==
  {"level":"INFO","msg":"pago rechazado","pedido":42,"!BADKEY":"tarjeta caducada"}
  compila, se ejecuta y no da error: el valor huérfano sale como !BADKEY
  pero go vet sí lo ve:
    malo.go:7:2: call to slog.Info missing a final value
    (go vet, código de salida 1)

== 1.3 · La otra trampa: registrar el struct entero ==
  {"level":"INFO","msg":"inicio de sesión","usuario":{"Nombre":"ana","Email":"ana@ejemplo.com","Clave":"hunter2"}}
  ¿sale la clave?: true

== 1.4 · Lo correcto: que el tipo decida qué se registra (slog.LogValuer) ==
  {"level":"INFO","msg":"inicio de sesión","usuario":{"nombre":"ana","clave":"[oculta]"}}
  ¿sale la clave?: false
```

## Leccion 02 — Niveles: el Debug que no imprime y cuesta

`go run ./02_niveles` · codigo de salida `0`

```
== 2.1 · La trampa: 10.000 Debug con el nivel en Info ==
  Debug con c.JSON():           96 ms · carrito serializado 10000 veces · líneas escritas: 0
  no se escribió nada, y el carrito se serializó igual: los argumentos se evalúan antes de llamar

== 2.2 · Lo correcto (1): preguntar antes con Enabled ==
  if log.Enabled(Debug):         0 ms · carrito serializado     0 veces · líneas escritas: 0

== 2.3 · Lo correcto (2): un LogValuer, que se evalúa solo si se escribe ==
  Debug con LogValuer:           0 ms · carrito serializado     0 veces · líneas escritas: 0
```

## Leccion 03 — La métrica que no existe hasta el primer error

`go run ./03_sin_iniciar` · codigo de salida `0`

```
== 3.1 · 120 pagos bien, ninguno mal: lo que publica /metrics ==
  tienda_pagos_total{resultado="ok"} 120
  la serie resultado="error" no existe: CounterVec la crea en el primer WithLabelValues

== 3.2 · Lo que ve la alerta, en el Prometheus de verdad ==
  tasa de errores = sum(rate(tienda_pagos_total{resultado="error"}[1m])) / sum(rate(tienda_pagos_total[1m]))
    -> (vacío: ninguna serie)
  no es 0: es «sin datos». El panel dice «No data» y una alerta «> 0.05» no tiene con qué compararse

== 3.3 · La trampa: llegan los 5 primeros errores ==
  tienda_pagos_total{resultado="error"} 5
  tienda_pagos_total{resultado="ok"} 120
  increase(errores[1m]) -> {resultado="error"} 0
  la serie nace valiendo 5 y para Prometheus ese es su punto de partida: esos 5 no se ven

== 3.4 · Lo correcto: las series creadas al arrancar, a 0 ==
  increase(errores[1m]) -> {resultado="error"} 5.107
  increase extrapola a los bordes de la ventana: sale cerca de 5, no exacto
```

## Leccion 04 — Cardinalidad: un id en una etiqueta

`go run ./04_cardinalidad` · codigo de salida `0`

```
== 4.1 · 1.000 peticiones a /pedidos/{id}, con mil ids distintos ==
  con la ruta tal cual: tienda_peticiones_ruta_cruda_total{ruta="/pedidos/1"} 1 · tienda_peticiones_ruta_cruda_total{ruta="/pedidos/10"} 1 · ...
  con la plantilla:     tienda_peticiones_total{ruta="/pedidos/{id}"} 1000

== 4.2 · Las series que Prometheus tiene que guardar ==
  count(tienda_peticiones_ruta_cruda_total) -> 1000
  count(tienda_peticiones_total)            -> 1
  ningún error: cada pedido nuevo es una serie más, para siempre

== 4.3 · Y lo que pesa /metrics en cada raspado ==
  tienda_peticiones_ruta_cruda_total    58025 bytes
  tienda_peticiones_total                 162 bytes
```

## Leccion 05 — Histogramas: el p99 de los buckets por defecto

`go run ./05_histogramas` · codigo de salida `0`

```
== 5.1 · Los buckets por defecto (prometheus.DefBuckets), en segundos ==
  0,005 · 0,01 · 0,025 · 0,05 · 0,1 · 0,25 · 0,5 · 1 · 2,5 · 5 · 10 · +Inf

== 5.2 · La trampa por arriba: 900 pagos de 200 ms y 100 de 15 s ==
  p99 de verdad: 15 s · histogram_quantile(0.99) con DefBuckets: 10 s
  todo lo que pasa de 10 s cae en +Inf, y Prometheus responde el último borde: 10

== 5.3 · La trampa por abajo: 1.000 respuestas de caché de 0,8 ms ==
  p50 de verdad: 0,0008 s · histogram_quantile(0.5) con DefBuckets: 0.0025 s
  todo cae en el primer bucket (0 a 5 ms) y Prometheus interpola por la mitad

== 5.4 · Lo correcto: buckets para TU rango (ExponentialBuckets(0.0005, 2, 16)) ==
  pago, p99:  de verdad 15 s · Prometheus 15.56 s
  caché, p50: de verdad 0,0008 s · Prometheus 0.00075 s
  un histograma no guarda valores, guarda cuántos cayeron en cada bucket: el resultado es una estimación
```

## Leccion 06 — Trazas: el contexto que se pierde

`go run ./06_trazas` · codigo de salida `0`

```
== 6.1 · Dentro del programa: el ctx que se pasa ==
  traza 4ce95ddd · span 4ff425bd · cobrar                       hijo de 7c08fbca
  traza 4ce95ddd · span 7c08fbca · POST /pagar                  (raíz)
  -> trazas distintas: 1

== 6.2 · La trampa: un context.Background() a mitad de camino ==
  traza c91f7215 · span c634a455 · cobrar                       (raíz)
  traza 0c40ef67 · span 2cfce6a7 · POST /pagar                  (raíz)
  -> trazas distintas: 2
  en el visor salen dos peticiones distintas: el cobro ya no cuelga del pago

== 6.3 · Entre servicios: con http.DefaultClient ==
  traceparent que llega al inventario: (ninguna)
  traza 537a8245 · span ac639971 · GET                          (raíz)
  traza a73df4b0 · span d99844d4 · POST /pagar                  (raíz)
  -> trazas distintas: 2

== 6.4 · La trampa: con otelhttp.NewTransport, pero sin propagador ==
  traceparent que llega al inventario: (ninguna)
  traza 7f232dc4 · span 3f516b91 · GET                          (raíz)
  traza 8390939c · span 042c7143 · HTTP GET                     hijo de 0cc98cda
  traza 8390939c · span 0cc98cda · POST /pagar                  (raíz)
  -> trazas distintas: 2
  el propagador global por defecto no hace nada: otelhttp no tiene con qué escribir la cabecera

== 6.5 · Lo correcto: otel.SetTextMapPropagator(propagation.TraceContext{}) ==
  traceparent que llega al inventario: 00-6c6275ff06c226fa822b4ccc10bd485b-6659a5ccf0409e28-01
  traza 6c6275ff · span 9e636fae · GET                          hijo de 6659a5cc
  traza 6c6275ff · span 6659a5cc · HTTP GET                     hijo de 61b9093b
  traza 6c6275ff · span 61b9093b · POST /pagar                  (raíz)
  -> trazas distintas: 1
```

## Leccion 07 — Los spans que nunca llegan

`go run ./07_spans_perdidos` · codigo de salida `0`

```
== 7.1 · La trampa: el programa termina sin tp.Shutdown() ==
  spans creados: 100 · spans que llegaron al exportador: 0
  estaban en la cola del lote, esperando su turno, y el proceso se fue con ellos

== 7.2 · Lo correcto: defer tp.Shutdown(ctx) ==
  spans creados: 100 · spans que llegaron al exportador: 100 · Shutdown tardó 0 ms

== 7.3 · La otra trampa: un span sin End() (1 de cada 10) ==
  spans creados: 100 · spans que llegaron al exportador: 90 · Shutdown tardó 0 ms
  un span sin End no se exporta nunca: ni con Shutdown. Por eso siempre defer span.End()
```

## Leccion 08 — Todo junto: logs, métricas y trazas que se encuentran

`go run ./08_todo_junto` · codigo de salida `0`

```
== 8.1 · Una petición: GET /pedidos/42 ==
  el registro:
    {"level":"INFO","msg":"pedido consultado","pedido":"42","trace_id":"e4eb935248d465d6b637af8a3b6bd1e5","span_id":"8991a340570f8261"}
  la traza:
    e4eb935248d465d6b637af8a3b6bd1e5 · consultar inventario
    e4eb935248d465d6b637af8a3b6bd1e5 · GET /pedidos/{id}
  ¿el trace_id del registro es el de la traza?: true

== 8.2 · 50 pedidos distintos más: las series de la métrica ==
  tienda_http_duracion_seconds{codigo="200",ruta="GET /pedidos/{id}"} · 51 peticiones
  tienda_http_duracion_seconds{codigo="404",ruta=""} · 1 petición
  51 pedidos, una serie: la etiqueta es la plantilla que ya conoce el router (r.Pattern, Go 1.23)
```
