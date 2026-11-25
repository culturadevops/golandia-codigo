# SALIDA — curso «gRPC en Go: protobuf, streaming y las trampas que no dan error»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-10-06 19:03:54Z`
- go version go1.26.8 linux/amd64
- `google.golang.org/grpc v1.84.0`
- `google.golang.org/protobuf v1.36.12`
- `libprotoc 3.21.12`
- Docker `29.6.2`

## Leccion 01 — protobuf: el número es el contrato

`go run ./01_protobuf` · codigo de salida `0`

```
== 1.1 · Por el cable no viaja el nombre del campo, viaja su número ==
  el cliente viejo manda: sku="TAZA-01" stock=5
  en el cable: 11 bytes · 0a 07 54 41 5a 41 2d 30 31 10 05

== 1.2 · La trampa: reutilizar el número de un campo borrado ==
  el servidor nuevo lee: sku="TAZA-01" precio_centavos=5 · error: <nil>
  -> la taza ahora cuesta 5 centavos. protoc no avisó y Unmarshal tampoco.

== 1.3 · Lo correcto: reserved, y un número nuevo ==
  el mismo mensaje viejo, leído con reserved 2: sku="TAZA-01" · precio presente: false · error: <nil>
  el campo 2 no se pierde: queda guardado como desconocido (2 bytes)
  y si alguien intenta volver a usar el 2, protoc se niega:
    x.proto: Field "descuento" uses reserved number 2.
    x.proto: Suggested field numbers for Producto: 3
    (protoc, código de salida 1)

== 1.4 · La otra trampa: en proto3, «no vino» y «vale 0» son lo mismo ==
  stock=0: 9 bytes · stock sin poner: 9 bytes · ¿iguales?: true
  con optional: precio=0 -> 11 bytes, presente: true · sin precio -> 9 bytes, presente: false
```

## Leccion 02 — El cliente: NewClient no conecta

`go run ./02_cliente` · codigo de salida `0`

```
== 2.1 · La trampa: NewClient contra un puerto donde no escucha nadie ==
  grpc.NewClient("127.0.0.1:<puerto vacío>") -> error: <nil> · estado: IDLE
  no ha intentado conectar: NewClient solo prepara. «Conectado» no significa nada todavía.

== 2.2 · La primera llamada es la que se entera, y falla al momento ==
  Consultar -> código Unavailable en 0 ms, con 5 s de plazo
  mensaje: connection error: desc = "transport: Error while dialing: dial tcp 127.0.0.1:41837: connect: connection refused"
  estado de la conexión ahora: TRANSIENT_FAILURE

== 2.3 · El servidor arranca 1 s tarde: sin WaitForReady, falla ==
  Consultar -> código Unavailable en 1 ms

== 2.4 · Con grpc.WaitForReady(true), la llamada espera a que haya servidor ==
  Consultar -> código OK en 1001 ms · Taza de Golandia, stock 12
  el servidor estaba listo a los 1000 ms; espera de más por los reintentos del cliente (backoff): 1 ms
  útil al arrancar; y siempre con plazo, o espera para siempre.
```

## Leccion 03 — Plazos: sin deadline se espera para siempre

`go run ./03_plazos` · codigo de salida `0`

```
== 3.1 · La trampa: sin plazo, contra un servidor colgado ==
  a los 3003 ms: sin respuesta y sin error. gRPC no pone plazo por defecto.
  el servidor vio la llamada sin plazo: true

== 3.2 · Con plazo: context.WithTimeout de 300 ms ==
  Consultar -> código DeadlineExceeded en 301 ms
  el plazo viaja con la llamada: al servidor le llegó con 299 ms por delante
  y como el servidor mira el ctx, paró en el paso 3 de 10

== 3.3 · La otra trampa: el servidor que no mira el ctx ==
  el cliente se rinde: código DeadlineExceeded en 301 ms
  el servidor siguió trabajando para nadie: 10 pasos de 10
```

## Leccion 04 — Errores: lo que cruza el cable, y el pánico

`go run ./04_errores` · codigo de salida `0`

```
== 4.1 · La trampa: devolver un error de Go con fmt.Errorf ==
  el cliente recibe: código Unknown · "sku PLATO-99: no existe"
  errors.Is(err, ErrNoExiste): false · el %w se queda en el servidor

== 4.2 · Lo correcto: status.Errorf con su código ==
  el cliente recibe: código NotFound · "sku PLATO-99: no existe"
  y decide por el código: status.Code(err) == codes.NotFound -> true

== 4.3 · Devolver nil, nil ==
  el cliente recibe: código OK · producto nil: false · sku="" stock=0
  -> sin error y con un producto vacío: el cliente cree que existe y tiene 0 en stock

== 4.4 · La trampa gorda: un pánico en un handler ==
  Consultar(TAZA-01) -> OK
  Consultar(PLATO-99) -> Unavailable
  Consultar(TAZA-01) -> Unavailable
  el proceso del servidor terminó: código de salida 2
  su última palabra: panic: assignment to entry in nil map
  -> un sku raro tumbó el servidor, y con él a todos sus clientes

== 4.5 · Lo correcto: un interceptor que recupera ==
  Consultar(TAZA-01) -> OK
  Consultar(PLATO-99) -> Internal
  Consultar(TAZA-01) -> OK
  el proceso del servidor sigue vivo
  en su log: pánico en /inventario.v1.Inventario/Consultar: assignment to entry in nil map
```

## Leccion 05 — Streaming: control de flujo y cancelar

`go run ./05_streaming` · codigo de salida `0`

```
== 5.1 · Un flujo del servidor: 20 eventos pequeños, leídos hasta io.EOF ==
  recibidos: 20 · el final del flujo es io.EOF, no un error

== 5.2 · El cliente pide 1.000 eventos de 16 KB y no lee ==
  tras 1 s, el servidor lleva enviados 12 de 1.000 (unos 192 KB) y su Send está bloqueado
  tras 2 s, sigue en 12: es el control de flujo de HTTP/2, no un error
  el cliente lee 50 -> el servidor avanza a 62

== 5.3 · La trampa: el cliente deja de leer y no cancela ==
  3 s después: handlers vivos en el servidor = 1, parado en 62
  ese handler (con su gorrutina y sus buffers) sigue ahí mientras la conexión viva: una fuga por cada cliente así

== 5.4 · Lo correcto: cancelar el contexto (defer cancel()) ==
  cancel() -> el Send del servidor devuelve Canceled en 0 ms · handlers vivos = 0
```

## Leccion 06 — El límite de 4 MB

`go run ./06_tamano` · codigo de salida `0`

```
== 6.1 · La misma llamada, con catálogos cada vez más grandes ==
      100 productos · 6 KB · OK
   10.000 productos · 614 KB · OK
   50.000 productos · 3,0 MB · OK
  100.000 productos · 6,0 MB · ResourceExhausted

== 6.2 · El error, entero, y quién lo ve ==
  cliente: grpc: received message larger than max (6296000 vs. 4194304)
  servidor: «el servidor lo envió sin error». Su límite de envío por defecto es math.MaxInt32 (2 GB): no se entera.

== 6.3 · Subir el límite (grpc.MaxCallRecvMsgSize): mueve el muro, no lo quita ==
  100.000 productos con 16 MB de límite -> OK
  400.000 productos con el mismo límite -> ResourceExhausted

== 6.4 · Lo correcto: un flujo, a trozos de 5.000 ==
  400.000 productos en 80 trozos · el trozo más grande: 307 KB · límite intacto (4 MB)
```

## Leccion 07 — El balanceo: la conexión que se pega

`go run ./07_balanceo` · codigo de salida `0`

```
== 7.1 · Dos réplicas con el mismo nombre: el DNS de Docker da las dos ==
  inventario -> 192.168.16.2, 192.168.16.3

== 7.2 · La trampa: el cliente por defecto (pick_first) ==
  100 llamadas -> replica-b: 100

== 7.3 · Lo correcto: round_robin en la configuración del servicio ==
  100 llamadas -> replica-a: 50 · replica-b: 50

== 7.4 · La segunda trampa: escalas a 3 réplicas y la nueva no recibe nada ==
  con 2 réplicas, 100 llamadas -> replica-a: 50 · replica-b: 50
  ya hay 3 réplicas; una llamada cada 200 ms, contadas de 15 en 15 s:
    de  0 a 15 s -> replica-a: 38 · replica-b: 37
    de 15 a 30 s -> replica-a: 37 · replica-b: 38
    de 30 a 45 s -> replica-a: 38 · replica-b: 37

== 7.5 · Lo correcto: MaxConnectionAge en el servidor (aquí, 3 s) ==
  con 2 réplicas, 100 llamadas -> replica-a: 50 · replica-b: 50
  ya hay 3 réplicas; una llamada cada 200 ms, contadas de 15 en 15 s:
    de  0 a 15 s -> replica-a: 38 · replica-b: 37
    de 15 a 30 s -> replica-a: 37 · replica-b: 35 · replica-c: 3
    de 30 a 45 s -> replica-a: 25 · replica-b: 27 · replica-c: 23
```

## Leccion 08 — Apagar sin cortar: GracefulStop, health e interceptores

`go run ./08_apagar` · codigo de salida `0`

```
== 8.1 · La trampa: srv.Stop() con 5 llamadas a medias ==
  las 5 en vuelo: 5 Unavailable · apagar tardó 0 ms · una llamada nueva después: Unavailable

== 8.2 · Lo correcto: srv.GracefulStop() ==
  las 5 en vuelo: 5 OK · apagar tardó 401 ms · una llamada nueva después: Unavailable
  termina lo que tiene a medias y no acepta nada nuevo. Con un tope: si tarda, Stop().

== 8.3 · El health check: avisar ANTES de apagar ==
  arrancado: SERVING
  tras sano.Shutdown(): NOT_SERVING · el balanceador deja de mandar, y entonces GracefulStop

== 8.4 · Un interceptor: el registro de cada llamada, sin tocar un solo handler ==
  Consultar · OK · 20 ms
  Consultar · Canceled · 50 ms
  Listar · Unimplemented · 0 ms
```

## Leccion 09 — ¿gRPC o REST?

`go run ./09_rest_o_grpc` · codigo de salida `0`

```
== 9.1 · El mismo catálogo de 100 productos, en el cable ==
  protobuf: 6296 bytes · JSON: 10975 bytes · JSON ocupa 1,7 veces más

== 9.2 · 2.000 llamadas seguidas, una a una, en la misma máquina ==
  gRPC: mediana 88 µs · REST con JSON: mediana 234 µs
  en local y sin red de por medio: una diferencia que en internet se come la latencia

== 9.3 · curl no habla gRPC ==
  curl http://<servidor>/inventario.v1.Inventario/Listar -> curl: (1) Received HTTP/0.9 when not allowed

== 9.4 · grpcurl sí, pero sin reflexión no sabe qué hay ==
  grpcurl list -> Failed to list services: server does not support the reflection API
  con reflection.Register(s):
    grpc.reflection.v1.ServerReflection
    grpc.reflection.v1alpha.ServerReflection
    inventario.v1.Inventario
  grpcurl -d '{"productos": 1}' ... Listar:
    {
      "productos": [
        {
          "sku": "SKU-0000000",
          "nombre": "Producto 0000000 de la tienda de Golandia",
          "precioCentavos": "1000"
        }
      ]
    }
```
