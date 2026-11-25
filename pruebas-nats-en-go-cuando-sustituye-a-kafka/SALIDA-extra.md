# SALIDA — NATS frente a Kafka — reentrega y releer (ejecutado)

Generado por `python correr_extra.py`. **No se edita a mano.**

- fecha: 2026-10-11 02:46 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## 7 · JetStream: el mensaje que nadie confirma, y el stream con límite

`$ go run ./reentrega`

```
0.0 s · recibido "cobrar pedido 7" · entrega número 1
        (el trabajador se cae: no confirma)
2.0 s · recibido "cobrar pedido 7" · entrega número 2
        confirmado
pendientes de confirmar: 0

publicadas 5 · en el stream: 3 (de la secuencia 3 a la 5)
```

## 7b · NATS básico: el suscriptor que no da abasto

`$ go run ./lento`

```
aviso del cliente: nats: slow consumer, messages dropped
publicados 20000 en 5ms
procesados: 1004 · descartados por el cliente: 18996
```

## 7c · JetStream: el mismo mensaje dos veces, con Nats-Msg-Id

`$ go run ./dedup`

```
intento 1 · secuencia 1 · duplicado: false
intento 2 · secuencia 1 · duplicado: true
en el stream: 1 mensaje
```

## 7d · El consumidor que se queda atrás: Kafka y JetStream

`$ go run ./atras`

```
kafka:     producidos 1000 · el grupo lento leyó y confirmó 200
jetstream: publicados 1000 · el consumidor lento leyó y confirmó 200 · pendientes: 800
```

`$ kafka-consumer-groups.sh --describe --group lento`

```
GROUP TOPIC PARTITION CURRENT-OFFSET LOG-END-OFFSET LAG CONSUMER-ID HOST CLIENT-ID
lento lecturas 0 200 1000 800 - - -
```

## 8 · Kafka: leer no borra

`$ go run ./relee`

```
producidas 3 facturas
grupo facturacion, primera vez:
  facturacion leyó factura 1 (partición 2, offset 0)
  facturacion leyó factura 2 (partición 2, offset 1)
  facturacion leyó factura 3 (partición 2, offset 2)
  facturacion total: 3
grupo facturacion, otra vez:
  facturacion total: 0
grupo auditoria, nuevo:
  auditoria   leyó factura 1 (partición 2, offset 0)
  auditoria   leyó factura 2 (partición 2, offset 1)
  auditoria   leyó factura 3 (partición 2, offset 2)
  auditoria   total: 3
```
