# Salida ejecutada — vídeo 224 «Escribí Kafka en Go»

> Corrida del 2026-09-08. `kafkita` (Go, biblioteca estándar y nada más) en un contenedor
> `golang:alpine`; el **cliente oficial de Apache Kafka 4.3.1** (`apache/kafka:latest`) en otro;
> misma red de Docker. Todo lo que se afirma aquí lo dice el cliente oficial, no yo.
> Reproducible con `bash demo.sh`.

---

## 1 · 🥇 LO PRIMERO QUE MANDA UN CLIENTE DE KAFKA

Se pone un socket en el 9092 que **no contesta nada**, solo mira lo que llega:

```
  peticion 0: largo=49 api=18 (ApiVersions) version=4
  peticion 1: largo=49 api=18 (ApiVersions) version=4
  peticion 2: largo=49 api=18 (ApiVersions) version=4
```

No es producir. No es consumir. Son **49 bytes preguntando «¿qué sabes hacer y en qué versión?»**.
Y hasta que no le contestas, el cliente reintenta en bucle **sin decir qué le pasa**. Eso explica
por qué un Kafka mal configurado «no responde» en vez de dar un error legible.

## 2 · 🥇 LA NEGOCIACIÓN, QUE ES LA CLAVE DE TODO

```
$ kafka-topics.sh --bootstrap-server kafkita:9092 --list
pruebas

  <- ApiVersions  v4  corr=0   cliente="adminclient-1"
  <- Metadata     v3  corr=1   cliente="adminclient-1"
```

El cliente **pide** ApiVersions en la v4, y a continuación manda Metadata en la **v3**: la más
alta que le dijimos que sabemos. Declarando versiones bajas, el cliente deja de usar el formato
«flexible» (campos etiquetados, enteros de longitud variable) y habla el formato viejo de
longitudes fijas, que se escribe en una tarde.

⚠️ Y la respuesta a ApiVersions lleva **cabecera v0**, aunque la petición venga en v4: es el único
mensaje que se contesta antes de haber negociado nada.

## 3 · Producir, con el productor oficial

```
   guardado en "pruebas": 98 bytes · 3 mensajes · offsets 0..2
```

**Tres mensajes en UN solo lote de 98 bytes.** La unidad del protocolo no es el mensaje: es el
lote. El productor los agrupa y el broker guarda el bloque tal cual, sin abrirlo.

Y el broker **reescribe** los primeros 8 bytes del lote con el offset real, porque el productor lo
manda con base 0: no sabe qué offsets le van a tocar. Eso lo decide quien escribe en el registro.

## 4 · Consumir, con el consumidor oficial

```
hola
que tal
adios
Processed a total of 3 messages
```

## 5 · 🥇 LOS DOS FALLOS DE VERSIÓN, Y SON DISTINTOS

Este es el hallazgo más útil de todo el vídeo. La negociación hacia abajo **tiene un suelo**, y el
suelo lo pone la funcionalidad que el cliente quiera usar. Pero los dos fallos no se parecen en
nada:

**Produce demasiado viejo → grita.**
```
UnsupportedVersionException: The node does not support PRODUCE with version
in range [3,13]. The supported range is [0,2].
```
El productor moderno es **idempotente de serie** desde Kafka 3.0, y eso exige Produce v3 como
mínimo, porque los campos del identificador de productor no existen antes. (También exige
`InitProducerId`: sin contestarlo, aborta con «The node does not support INIT_PRODUCER_ID».)

**Fetch demasiado viejo → se calla.**
```
Fetch(1): 0 to 3 [unusable: node too old]
```
Eso está en el registro del CLIENTE, en nivel de depuración. No lanza ninguna excepción: marca
Fetch como inservible, no manda la petición **nunca**, y el consumidor muere por tiempo agotado
sin explicar por qué. Hubo que subir a **Fetch v11**, la última antes del formato flexible.

> **Cuando un consumidor de Kafka no lee y no dice nada, mira las versiones antes que la red.**

## 6 · 🥇 LA ESPERA LARGA: DE 316.000 PETICIONES A 18

Un broker de verdad **no contesta enseguida** cuando no hay datos: se queda la petición abierta
hasta que llegue algo o se agote el tiempo que pidió el cliente.

```
   sin espera larga:  316.000 peticiones de Fetch en 10 segundos
   con espera larga:       18 peticiones
```

Cuatro órdenes de magnitud. Sin eso, un consumidor parado quema una CPU entera preguntando
«¿hay algo?» sin parar.

## 7 · 🥇 LA TRAMPA DE METADATA: el 90 % de los «no me conecto a Kafka»

El **mismo** servidor, lo único que cambia es el nombre que anuncia en su respuesta de Metadata:

```
java.net.UnknownHostException: no-existe-este-host: Name does not resolve

  <- ApiVersions  v4  corr=0   cliente="console-producer"
  <- Metadata     v3  corr=1   cliente="console-producer"
```

Fíjate en las dos últimas líneas: **en el servidor no hay ni un error**. El cliente se conectó
perfectamente, mandó sus dos peticiones, las dos entraron bien... y a partir de ahí se fue al
nombre anunciado y se perdió. El puerto estuvo abierto todo el rato.

Es exactamente lo que pasa con Kafka en Docker cuando `advertised.listeners` anuncia un nombre que
el cliente no resuelve: **el diagnóstico no está en el puerto, está en lo que anuncias**.

---

## ⛔ Lo que NO hace, dicho entero

- No hay grupos de consumidores (`JoinGroup`, `SyncGroup`, `Heartbeat`): es un protocolo aparte y
  no cabe. Por eso el consumidor se usa con `--partition 0`.
- No hay disco: todo en memoria, se pierde al reiniciar.
- No hay réplicas, ni controlador, ni elección de líder.
- No se abre el lote: no se comprueba la suma de comprobación ni se descomprime.
- Una sola partición por tema.

## Tropiezos reales (van en el guion)

1. El sondeo dio `UnknownHostException` en varios intentos **aunque la sonda sí recibía** las
   peticiones: es una carrera del DNS de Docker al arrancar. Hay que dar margen o el vídeo enseña
   un error que no es el del tema.
2. `Fetch` inservible en silencio (arriba). Costó una sesión entera de depuración hasta subir el
   nivel de registro del cliente con `-Dorg.apache.logging.log4j.level=DEBUG`, que fue lo que lo
   dijo con todas las letras.
