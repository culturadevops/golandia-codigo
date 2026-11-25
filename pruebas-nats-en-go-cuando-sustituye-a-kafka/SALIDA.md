# SALIDA — NATS frente a Kafka (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-10 18:42 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## Las imágenes

`$ docker images`

```
nats:2.15  26MB
apache/kafka:4.1.1  667MB
```

## Arrancar NATS (con JetStream)

`$ docker run -d --name fab-nats nats:2.15 -js`

```
listo en 0.1 s (desde que el contenedor existe)
```

`$ docker logs fab-nats | grep -E 'Version|JetStream|ready'`

```
2026/10/10 18:41:40.644280 [INF]   Version:  2.15.0
2026/10/10 18:41:40.644938 [INF] Starting JetStream
2026/10/10 18:41:40.648010 [INF] Server is ready
```

## Arrancar Kafka (KRaft, un nodo)

`$ docker run -d --name fab-kafka apache/kafka:4.1.1`

```
listo en 2.0 s (desde que el contenedor existe)
```

## Memoria en reposo, 20 s después

`$ docker stats --no-stream fab-nats fab-kafka`

```
fab-nats   6.516MiB
fab-kafka  291.8MiB
```

## 1 · Publicar y suscribirse (NATS básico)

`$ go run ./pubsub`

```
45 líneas
recibido: pedidos.nuevo → pedido 2
recibido: pedidos.pagado → pedido 2
el «pedido 1» no llegó: se publicó antes de que hubiera suscriptor
```

## 2 · JetStream: guardar y leer después

`$ go run ./jetstream`

```
publicado pedido 1 · secuencia 1
publicado pedido 2 · secuencia 2
publicado pedido 3 · secuencia 3
en el stream: 3 mensajes, 153 bytes
consumidor facturacion: pedido 1 (secuencia 1)
consumidor facturacion: pedido 2 (secuencia 2)
consumidor facturacion: pedido 3 (secuencia 3)
```

`$ go run ./jetstream   # otra vez: el consumidor ya leyó los 3 primeros`

```
en el stream: 6 mensajes, 306 bytes
consumidor facturacion: pedido 1 (secuencia 4)
consumidor facturacion: pedido 2 (secuencia 5)
consumidor facturacion: pedido 3 (secuencia 6)
```

## 3 · Petición y respuesta

`$ go run ./peticion`

```
respuesta: precio de CAFÉ: 12,50 (en 336µs)
sin nadie que responda: nats: no responders available for request
```

## 3b · Un grupo de cola: repartir el trabajo

`$ go run ./cola`

```
trabajador-a: 98 tareas
trabajador-b: 94 tareas
trabajador-c: 108 tareas
total: 300 de 300, cada una a un solo trabajador
```

## 3c · El almacén clave-valor de JetStream

`$ go run ./kv`

```
app.timeout = 30s (revisión 3)
  historial: revisión 1 → 5s
  historial: revisión 2 → 10s
  historial: revisión 3 → 30s
```

## 4 · nats bench: publicar 1 millón de mensajes de 128 bytes

`$ nats bench pub bench.prueba --msgs 1000000 --size 128`

```
NATS Core NATS publisher stats: 6,462,853 msgs/sec ~ 789 MiB/sec ~ min: 0.04us ~ avg: 0.11us ~ max: 208.78us ~ P50: 0.05us ~ P90: 0.06us ~ P99: 0.69us ~ P99.9: 13.71us
18:42:11 Starting Core NATS publisher benchmark [clients=1, msg-size=128 B, msgs=1,000,000, multi-subject=false, multi-subject-max=100,000, multi-subject-randomize=false, sleep=0s, subject=bench.prueba]
18:42:11 [1] Starting Core NATS publisher, publishing 1,000,000 messages
```

## 4b · nats bench con JetStream: lo mismo, pero guardado en disco

`$ nats bench js pub async bench.js --create --storage file --msgs 1000000 --size 128`

```
NATS JetStream asynchronous publisher stats: 330,190 msgs/sec ~ 40 MiB/sec ~ min: 1,021.64us ~ avg: 1,512.74us ~ max: 5,852.47us ~ P50: 1,417.50us ~ P90: 2,011.08us ~ P99: 2,814.26us ~ P99.9: 4,959.73us
```

## 5 · Kafka: un tema con 3 particiones

`$ kafka-topics.sh --create --topic pedidos --partitions 3`

```
Created topic pedidos.
```

`$ kafka-topics.sh --describe --topic pedidos`

```
Topic: pedidos	TopicId: jmN3fvmVTOGnKjNsVlr9eg	PartitionCount: 3	ReplicationFactor: 1	Configs: min.insync.replicas=1
Topic: pedidos	Partition: 0	Leader: 1	Replicas: 1	Isr: 1	Elr: 	LastKnownElr:
Topic: pedidos	Partition: 1	Leader: 1	Replicas: 1	Isr: 1	Elr: 	LastKnownElr:
Topic: pedidos	Partition: 2	Leader: 1	Replicas: 1	Isr: 1	Elr: 	LastKnownElr:
```

## 5b · Kafka desde Go (franz-go): la misma clave, la misma partición

`$ go run ./kafka`

```
46 líneas
producido cliente-ana  pedido 1 → partición 2, offset 0
producido cliente-luis pedido 2 → partición 0, offset 0
producido cliente-ana  pedido 3 → partición 2, offset 1
producido cliente-luis pedido 4 → partición 0, offset 1
producido cliente-ana  pedido 5 → partición 2, offset 2
producido cliente-eva  pedido 6 → partición 1, offset 0
leído     partición 2: cliente-ana pedido 1
leído     partición 2: cliente-ana pedido 3
leído     partición 2: cliente-ana pedido 5
leído     partición 0: cliente-luis pedido 2
leído     partición 0: cliente-luis pedido 4
leído     partición 1: cliente-eva pedido 6
```

## 6 · kafka-producer-perf-test: 1 millón de mensajes de 128 bytes

`$ kafka-producer-perf-test.sh --topic bench --num-records 1000000 --record-size 128 --throughput -1`

```
1000000 records sent, 464468.2 records/sec (56.70 MB/sec), 388.46 ms avg latency, 530.00 ms max latency, 382 ms 50th, 499 ms 95th, 521 ms 99th, 529 ms 99.9th.
```
