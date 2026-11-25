# Pruebas ejecutadas — nubecita · S3, DynamoDB, SQS y SigV4

> Salida **literal** de esta máquina. La genera `investigacion/pruebas_nubecita.py` y se
> vuelve a generar sola. ⛔ Lo que no esté aquí no puede salir en pantalla.

**Entorno citable en el vídeo:** `go version go1.26.7 linux/amd64` (dentro de `golang:1.26`) · `Docker version 29.6.2, build dfc4efb` · `aws-cli/2.36.5 Python/3.14.6 Windows/11 exe/AMD64` · `aws-sdk-go-v2` para `s3`, `dynamodb` y `sqs`

**El código vive en `<nubecita>`**, en su propia carpeta, fuera del canal: es una herramienta que crece por episodios, no material de un vídeo.

**La imagen entera se descarga en 2.743.004 bytes** (2.62 MiB comprimidos) y ocupa **9.19MB** en disco. La segunda etapa es `scratch`: dentro no hay NADA más que el binario, porque con `CGO_ENABLED=0` el binario de Go es estático.

⚠️ Son dos números distintos y confundirlos es publicar una cifra falsa: lo que te
descargas (capas comprimidas) no es lo que ocupa descomprimida en tu disco.

⚠️ **Todo corre en Docker, y no es una preferencia.** El Application Control de esta
máquina bloquea a ratos los binarios recién compilados («Una directiva de Control de
aplicaciones bloqueó este archivo»): primero el de `go test`, y desde el episodio 3
también el del servidor. Una prueba que a veces corre no sirve. Lo único que se ejecuta
en Windows es la CLI de AWS, que está firmada. De regalo, en el contenedor hay cgo y por
tanto **detector de carreras**, que en este Windows no hay.

---

## 1 · `go vet` y las pruebas contra el SDK de AWS de verdad

Las pruebas NO usan un cliente de mentira: usan `aws-sdk-go-v2/service/s3`,
`.../dynamodb` y `.../sqs`, los mismos que usarías contra Amazon. Es la única forma
honesta de decir «es compatible»: si el SDK oficial lo entiende, lo es.

```console
$ docker run --rm -v <nubecita>:/app -v nubecita-gomod:/go/pkg/mod -w /app -e GOFLAGS=-buildvcs=false golang:1.26 go vet ./...

[codigo de salida: 0]
```

```console
$ docker run --rm -v <nubecita>:/app -v nubecita-gomod:/go/pkg/mod -w /app -e GOFLAGS=-buildvcs=false golang:1.26 go test ./... -count=1 -v
=== RUN   TestTablaCicloCompleto
--- PASS: TestTablaCicloCompleto (0.00s)
=== RUN   TestGuardarYLeerConStructDeGo
--- PASS: TestGuardarYLeerConStructDeGo (0.00s)
=== RUN   TestLoQueNoExisteNoEsUnError
--- PASS: TestLoQueNoExisteNoEsUnError (0.00s)
=== RUN   TestTablaQueNoExisteLlegaComoResourceNotFound
--- PASS: TestTablaQueNoExisteLlegaComoResourceNotFound (0.00s)
=== RUN   TestCrearDosVecesLaMismaTablaLlegaComoResourceInUse
--- PASS: TestCrearDosVecesLaMismaTablaLlegaComoResourceInUse (0.00s)
=== RUN   TestQuerySoloMiraLaClave
--- PASS: TestQuerySoloMiraLaClave (0.00s)
=== RUN   TestNoSePuedeBuscarPorPrefijoEnLaClaveDeParticion
--- PASS: TestNoSePuedeBuscarPorPrefijoEnLaClaveDeParticion (0.00s)
=== RUN   TestQueryAlReves
--- PASS: TestQueryAlReves (0.00s)
=== RUN   TestEscrituraCondicional
--- PASS: TestEscrituraCondicional (0.00s)
=== RUN   TestScanLeeTodo
--- PASS: TestScanLeeTodo (0.00s)
=== RUN   TestBorrarElemento
--- PASS: TestBorrarElemento (0.00s)
=== RUN   TestLosNumerosViajanComoCadena
--- PASS: TestLosNumerosViajanComoCadena (0.00s)
=== RUN   TestLaClaveDeOrdenNumericaSeOrdenaComoNumero
--- PASS: TestLaClaveDeOrdenNumericaSeOrdenaComoNumero (0.00s)
=== RUN   TestLosDosServiciosConvivenEnElMismoPuerto
--- PASS: TestLosDosServiciosConvivenEnElMismoPuerto (0.00s)
=== RUN   TestFirmaDelSDKOficialEncaja
--- PASS: TestFirmaDelSDKOficialEncaja (0.00s)
=== RUN   TestConElSecretoEquivocadoNoPasa
--- PASS: TestConElSecretoEquivocadoNoPasa (0.00s)
=== RUN   TestClaveDeAccesoDesconocida
--- PASS: TestClaveDeAccesoDesconocida (0.00s)
=== RUN   TestSiAlguienToquetealCuerpoLaFirmaSeCae
--- PASS: TestSiAlguienToquetealCuerpoLaFirmaSeCae (0.00s)
=== RUN   TestSiAlguienCambiaLaRutaLaFirmaSeCae
--- PASS: TestSiAlguienCambiaLaRutaLaFirmaSeCae (0.00s)
=== RUN   TestUnaFirmaCaduca
--- PASS: TestUnaFirmaCaduca (0.00s)
=== RUN   TestSinFirmaNiSeMira
--- PASS: TestSinFirmaNiSeMira (0.00s)
=== RUN   TestSinClavesElVerificadorNoEstorba
--- PASS: TestSinClavesElVerificadorNoEstorba (0.00s)
=== RUN   TestLaClaveDeFirmaCoincideConElEjemploDeAWS
--- PASS: TestLaClaveDeFirmaCoincideConElEjemploDeAWS (0.00s)
=== RUN   TestCicloCompleto
--- PASS: TestCicloCompleto (0.00s)
=== RUN   TestClaveQueNoExisteLlegaComoNoSuchKey
--- PASS: TestClaveQueNoExisteLlegaComoNoSuchKey (0.00s)
=== RUN   TestBucketQueNoExisteLlegaComoNoSuchBucket
--- PASS: TestBucketQueNoExisteLlegaComoNoSuchBucket (0.00s)
=== RUN   TestListarConDelimitadorAgrupaComoSiHubieraCarpetas
--- PASS: TestListarConDelimitadorAgrupaComoSiHubieraCarpetas (0.00s)
=== RUN   TestNoSeBorraUnBucketConCosasDentro
--- PASS: TestNoSeBorraUnBucketConCosasDentro (0.00s)
=== RUN   TestBorrarDosVecesNoEsUnError
--- PASS: TestBorrarDosVecesNoEsUnError (0.00s)
=== RUN   TestListarBuckets
--- PASS: TestListarBuckets (0.00s)
=== RUN   TestColaCicloCompleto
--- PASS: TestColaCicloCompleto (0.00s)
=== RUN   TestRecibirNoSacaElMensajeLoEsconde
--- PASS: TestRecibirNoSacaElMensajeLoEsconde (0.00s)
=== RUN   TestElReciboCambiaEnCadaEntrega
--- PASS: TestElReciboCambiaEnCadaEntrega (0.00s)
=== RUN   TestLaColaDeMuertosSeLlevaElQueSiempreFalla
--- PASS: TestLaColaDeMuertosSeLlevaElQueSiempreFalla (0.00s)
=== RUN   TestColaVaciaNoEsUnError
--- PASS: TestColaVaciaNoEsUnError (0.00s)
=== RUN   TestColaQueNoExisteLlegaComoQueueDoesNotExist
--- PASS: TestColaQueNoExisteLlegaComoQueueDoesNotExist (0.00s)
=== RUN   TestLosNumerosAproximadosDeLaCola
--- PASS: TestLosNumerosAproximadosDeLaCola (0.00s)
=== RUN   TestCambiarLaVisibilidad
--- PASS: TestCambiarLaVisibilidad (0.00s)
=== RUN   TestEnviarYBorrarEnLote
--- PASS: TestEnviarYBorrarEnLote (0.00s)
=== RUN   TestLosTresServiciosConvivenEnElMismoPuerto
--- PASS: TestLosTresServiciosConvivenEnElMismoPuerto (0.00s)
PASS
ok  	nubecita	0.032s
[codigo de salida: 0]
```

### Y con el detector de carreras

El almacén, las tablas y las colas se tocan desde varias gorrutinas a la vez: sin esto,
los candados no los valida nadie.

```console
$ docker run --rm -v <nubecita>:/app -v nubecita-gomod:/go/pkg/mod -w /app -e GOFLAGS=-buildvcs=false golang:1.26 go test ./... -count=1 -race
ok  	nubecita	1.095s
[codigo de salida: 0]
```

---

## 2 · S3 con la CLI oficial de AWS, sin parchear

Lo único que cambia respecto a hablar con Amazon es `--endpoint-url`.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 mb s3://fotos
make_bucket: fotos
[codigo de salida: 0]
```

> `mb` es CreateBucket. La CLI oficial, sin parchear, sin saber que no es Amazon.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 cp .\gato.txt s3://fotos/gato.txt
Completed 15 Bytes/15 Bytes (368 Bytes/s) with 1 file(s) remaining
upload: recursos\pruebas-nubecita\_demo\gato.txt to s3://fotos/gato.txt
[codigo de salida: 0]
```

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 sync .\2026 s3://fotos/2026
Completed 5 Bytes/10 Bytes (145 Bytes/s) with 2 file(s) remaining
upload: recursos\pruebas-nubecita\_demo\2026\enero.txt to s3://fotos/2026/enero.txt
Completed 5 Bytes/10 Bytes (145 Bytes/s) with 1 file(s) remaining
Completed 10 Bytes/10 Bytes (197 Bytes/s) with 1 file(s) remaining
upload: recursos\pruebas-nubecita\_demo\2026\febrero.txt to s3://fotos/2026/febrero.txt
[codigo de salida: 0]
```

> `sync` es la prueba de fuego: por dentro son varios ListObjectsV2 y varios PutObject encadenados. Si la compatibilidad fuera de mentira, se cae aquí.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 ls s3://fotos --recursive
2026-08-25 06:40:15          5 2026/enero.txt
2026-08-25 06:40:15          5 2026/febrero.txt
2026-08-25 06:40:14         15 gato.txt
[codigo de salida: 0]
```

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3api list-objects-v2 --bucket fotos --delimiter /
{
    "Contents": [
        {
            "Key": "gato.txt",
            "LastModified": "2026-08-25T11:40:14.805000+00:00",
            "ETag": "\"2dab52e0e8834a6cbeb558a98af86ffa\"",
            "Size": 15,
            "StorageClass": "STANDARD"
        }
    ],
    "CommonPrefixes": [
        {
            "Prefix": "2026/"
        }
    ],
    "RequestCharged": null,
    "Prefix": ""
}
[codigo de salida: 0]
```

> 🥇 Aquí está lo que nadie cuenta: S3 **no tiene carpetas**. Con `--delimiter /` las dos claves que empiezan por `2026/` se colapsan en UN prefijo común. Eso es todo lo que hay detrás de lo que tú ves como una carpeta.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3api head-object --bucket fotos --key gato.txt
{
    "AcceptRanges": "bytes",
    "LastModified": "2026-08-25T11:40:14+00:00",
    "ContentLength": 15,
    "ETag": "\"2dab52e0e8834a6cbeb558a98af86ffa\"",
    "ContentType": "text/plain",
    "Metadata": {}
}
[codigo de salida: 0]
```

> El ETag es el md5 del contenido entre comillas, calculado igual que en S3.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3api get-object --bucket fotos --key no-existe .\x   # ⚠️ falla A PROPÓSITO

aws: [ERROR]: An error occurred (NoSuchKey) when calling the GetObject operation: The specified key does not exist.
[codigo de salida: 254]
```

> El error llega a la CLI con su **código**, no como un 404 pelado. Es lo que convierte un fallo en algo que tu código puede distinguir.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 rb s3://fotos   # ⚠️ falla A PROPÓSITO
remove_bucket failed: s3://fotos An error occurred (BucketNotEmpty) when calling the DeleteBucket operation: The bucket you tried to delete is not empty
[codigo de salida: 1]
```

> S3 se niega a borrar un bucket con cosas dentro. La nubecita también.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 rm s3://fotos --recursive
delete: s3://fotos/2026/enero.txt
delete: s3://fotos/2026/febrero.txt
delete: s3://fotos/gato.txt
[codigo de salida: 0]
```

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 rb s3://fotos
remove_bucket: fotos
[codigo de salida: 0]
```

> Y así se destruye la demo entera: vaciar y borrar. Sin factura, porque no hay factura.

---

## 3 · DynamoDB, en el MISMO puerto y el MISMO proceso

S3 dice qué quieres con el método y la ruta; DynamoDB manda siempre un `POST /` con el
cuerpo en JSON y la operación en una cabecera, `X-Amz-Target: DynamoDB_20120810.PutItem`.
Dos servicios de la misma nube, dos diseños — y los dos caben en el mismo puerto porque
el enrutador mira esa cabecera.

```console
$ aws --endpoint-url http://127.0.0.1:4566 dynamodb create-table --table-name pedidos --attribute-definitions AttributeName=cliente,AttributeType=S AttributeName=fecha,AttributeType=S --key-schema AttributeName=cliente,KeyType=HASH AttributeName=fecha,KeyType=RANGE --billing-mode PAY_PER_REQUEST
{
    "TableDescription": {
        "AttributeDefinitions": [
            {
                "AttributeName": "cliente",
                "AttributeType": "S"
            },
            {
                "AttributeName": "fecha",
                "AttributeType": "S"
            }
        ],
        "TableName": "pedidos",
        "KeySchema": [
            {
                "AttributeName": "cliente",
                "KeyType": "HASH"
            },
            {
                "AttributeName": "fecha",
                "KeyType": "RANGE"
            }
        ],
        "TableStatus": "ACTIVE",
        "CreationDateTime": "2026-08-25T06:40:18.349000-05:00",
        "TableSizeBytes": 0,
        "ItemCount": 0,
        "TableArn": "arn:aws:dynamodb:us-east-1:000000000000:table/pedidos",
        "BillingModeSummary": {
            "BillingMode": "PAY_PER_REQUEST"
        }
    }
}
[codigo de salida: 0]
```

> Clave COMPUESTA: `cliente` particiona y `fecha` ordena. Esto se elige una vez y no se cambia nunca: es lo que decide qué consultas van a ser posibles.

```console
$ aws --endpoint-url http://127.0.0.1:4566 dynamodb put-item --table-name pedidos --item {"cliente":{"S":"ana"},"fecha":{"S":"2026-01-15"},"total":{"N":"42.5"}}

[codigo de salida: 0]
```

> El número viaja como **cadena**. No es un descuido — así no pasa por el float de ningún lenguaje y no pierde precisión.

```console
$ aws --endpoint-url http://127.0.0.1:4566 dynamodb put-item --table-name pedidos --item {"cliente":{"S":"ana"},"fecha":{"S":"2026-02-03"},"total":{"N":"20"}}

[codigo de salida: 0]
```

```console
$ aws --endpoint-url http://127.0.0.1:4566 dynamodb put-item --table-name pedidos --item {"cliente":{"S":"ana"},"fecha":{"S":"2026-03-21"},"total":{"N":"30"}}

[codigo de salida: 0]
```

```console
$ aws --endpoint-url http://127.0.0.1:4566 dynamodb query --table-name pedidos --key-condition-expression cliente = :c AND fecha BETWEEN :a AND :b --expression-attribute-values {":c":{"S":"ana"},":a":{"S":"2026-02-01"},":b":{"S":"2026-12-31"}}
{
    "Items": [
        {
            "cliente": {
                "S": "ana"
            },
            "fecha": {
                "S": "2026-02-03"
            },
            "total": {
                "N": "20"
            }
        },
        {
            "cliente": {
                "S": "ana"
            },
            "fecha": {
                "S": "2026-03-21"
            },
            "total": {
                "N": "30"
            }
        }
    ],
    "Count": 2,
    "ScannedCount": 2,
    "ConsumedCapacity": null
}
[codigo de salida: 0]
```

> 🥇 Un rango sobre la clave de orden. Y funciona porque las fechas van en ISO-8601: el orden alfabético coincide con el cronológico. Con `15/01/2026` esto daría basura, y sin un solo error.

```console
$ aws --endpoint-url http://127.0.0.1:4566 dynamodb query --table-name pedidos --key-condition-expression begins_with(cliente, :p) --expression-attribute-values {":p":{"S":"a"}}   # ⚠️ falla A PROPÓSITO

aws: [ERROR]: An error occurred (ValidationException) when calling the Query operation: la condicion tiene que fijar la clave de particion con =
[codigo de salida: 254]
```

> ⛔ Sobre la clave de **partición** solo hay igualdad: no existe «los clientes que empiecen por a». No es una limitación de nubecita, es la regla de DynamoDB — y por eso el error lo dice.

```console
$ aws --endpoint-url http://127.0.0.1:4566 dynamodb scan --table-name pedidos --select COUNT
{
    "Count": 3,
    "ScannedCount": 3,
    "ConsumedCapacity": null
}
[codigo de salida: 0]
```

> `Scan` lee la tabla **entera**. Aquí son tres elementos; en Amazon es la diferencia entre unos centavos y una factura que duele.

```console
$ aws --endpoint-url http://127.0.0.1:4566 dynamodb delete-table --table-name pedidos
{
    "TableDescription": {
        "AttributeDefinitions": [
            {
                "AttributeName": "cliente",
                "AttributeType": "S"
            },
            {
                "AttributeName": "fecha",
                "AttributeType": "S"
            }
        ],
        "TableName": "pedidos",
        "KeySchema": [
            {
                "AttributeName": "cliente",
                "KeyType": "HASH"
            },
            {
                "AttributeName": "fecha",
                "KeyType": "RANGE"
            }
        ],
        "TableStatus": "ACTIVE",
        "CreationDateTime": "2026-08-25T06:40:18.349000-05:00",
        "TableSizeBytes": 0,
        "ItemCount": 0,
        "TableArn": "arn:aws:dynamodb:us-east-1:000000000000:table/pedidos",
        "BillingModeSummary": {
            "BillingMode": "PAY_PER_REQUEST"
        }
    }
}
[codigo de salida: 0]
```

---

## 4 · SQS, también en el mismo puerto

SQS habla como DynamoDB —`POST /`, JSON, operación en la cabecera— pero con otro prefijo:
`AmazonSQS.SendMessage`. Y eso tiene una historia detrás: SQS es más viejo que S3 y
durante años habló el protocolo «Query» de AWS, con los parámetros en el cuerpo tipo
formulario. Amazon lo migró a JSON. O sea: **el protocolo de un servicio no es para
siempre**.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs create-queue --queue-name muertos
{
    "QueueUrl": "http://127.0.0.1:57322/000000000000/muertos"
}
[codigo de salida: 0]
```

> Primero la cola de muertos, porque la otra la va a nombrar.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs create-queue --queue-name trabajos --attributes {"VisibilityTimeout": "0", "RedrivePolicy": "{\"deadLetterTargetArn\": \"arn:aws:sqs:us-east-1:000000000000:muertos\", \"maxReceiveCount\": \"2\"}"}
{
    "QueueUrl": "http://127.0.0.1:57322/000000000000/trabajos"
}
[codigo de salida: 0]
```

> `RedrivePolicy` lleva **JSON dentro de una cadena** de otro JSON. Sí, de verdad. Y dice: a la tercera entrega, a la cola de muertos.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs send-message --queue-url http://127.0.0.1:4566/000000000000/trabajos --message-body procesa el pedido 7
{
    "MD5OfMessageBody": "e81fffde762e0107aa3a3a83bb98c1c0",
    "MessageId": "trabajos-1"
}
[codigo de salida: 0]
```

> El `MD5OfMessageBody` no es decorativo: el SDK lo **comprueba**. Si lo calculas mal, el cliente da un error de suma y parece que se corrompió la red.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs receive-message --queue-url http://127.0.0.1:4566/000000000000/trabajos --attribute-names All
{
    "Messages": [
        {
            "MessageId": "trabajos-1",
            "ReceiptHandle": "recibo-2-1",
            "MD5OfBody": "e81fffde762e0107aa3a3a83bb98c1c0",
            "Body": "procesa el pedido 7",
            "Attributes": {
                "ApproximateReceiveCount": "1",
                "SentTimestamp": "1787658022142"
            }
        }
    ]
}
[codigo de salida: 0]
```

> 🥇 Recibir NO saca el mensaje de la cola: lo esconde. Y el `ReceiptHandle` es de ESTA entrega — no es el identificador del mensaje, y el de la entrega anterior deja de valer.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs receive-message --queue-url http://127.0.0.1:4566/000000000000/trabajos --attribute-names All
{
    "Messages": [
        {
            "MessageId": "trabajos-1",
            "ReceiptHandle": "recibo-3-2",
            "MD5OfBody": "e81fffde762e0107aa3a3a83bb98c1c0",
            "Body": "procesa el pedido 7",
            "Attributes": {
                "ApproximateReceiveCount": "2",
                "SentTimestamp": "1787658022142"
            }
        }
    ]
}
[codigo de salida: 0]
```

> Segunda entrega del MISMO mensaje, porque nadie lo borró. `ApproximateReceiveCount` va por dos: ese contador es lo que delata a un consumidor que se está atragantando.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs receive-message --queue-url http://127.0.0.1:4566/000000000000/trabajos

[codigo de salida: 0]
```

> Y a la tercera ya no llega: se pasó de `maxReceiveCount` y se apartó. Salida vacía, sin error — porque una cola sin trabajo tampoco es un error.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs receive-message --queue-url http://127.0.0.1:4566/000000000000/muertos
{
    "Messages": [
        {
            "MessageId": "trabajos-1",
            "ReceiptHandle": "recibo-4-1",
            "MD5OfBody": "e81fffde762e0107aa3a3a83bb98c1c0",
            "Body": "procesa el pedido 7",
            "Attributes": {
                "ApproximateReceiveCount": "1",
                "SentTimestamp": "1787658022142"
            }
        }
    ]
}
[codigo de salida: 0]
```

> 🥇 Ahí está, entero, en la cola de muertos. Eso es todo lo que hay detrás del «dead letter queue»: un contador de entregas y un sitio donde apartar.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs get-queue-attributes --queue-url http://127.0.0.1:4566/000000000000/trabajos --attribute-names All
{
    "Attributes": {
        "ApproximateNumberOfMessages": "0",
        "ApproximateNumberOfMessagesNotVisible": "0",
        "QueueArn": "arn:aws:sqs:us-east-1:000000000000:trabajos",
        "VisibilityTimeout": "0"
    }
}
[codigo de salida: 0]
```

> Los dos números que SQS llama «aproximados», y aquí se ve por qué: dependen del instante exacto en que preguntas.

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs delete-queue --queue-url http://127.0.0.1:4566/000000000000/trabajos

[codigo de salida: 0]
```

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs delete-queue --queue-url http://127.0.0.1:4566/000000000000/muertos

[codigo de salida: 0]
```

```console
$ aws --endpoint-url http://127.0.0.1:4566 sqs list-queues
{
    "QueueUrls": []
}
[codigo de salida: 0]
```

> Sin salida: no queda nada. Los tres servicios se destruyen igual de fácil, porque no hay nada que destruir.

---

## 5 · SigV4: la misma imagen, comprobando firmas

Hasta aquí el doble no autenticaba a nadie, que es lo correcto para un doble de pruebas.
Con `-firmar`, la misma imagen comprueba la firma **SigV4** de cada petición. Y aquí lo
importante: quien FIRMA es la CLI oficial de Amazon y quien COMPRUEBA es nuestro código.

⚠️ **Firmar no es cifrar.** La petición viaja en claro. Lo que la firma demuestra es
quién la manda —sin que el secreto viaje nunca— y que nadie la ha tocado por el camino.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 mb s3://fotos
make_bucket: fotos
[codigo de salida: 0]
```

> La CLI oficial firma cada petición sin que se lo pidas, y nuestra comprobación la acepta. Si esto pasa, nuestra implementación de SigV4 entiende exactamente lo que produce Amazon.

```console
$ AWS_SECRET_ACCESS_KEY=el-secreto-malo aws --endpoint-url http://127.0.0.1:4566 s3 ls   # ⚠️ falla A PROPÓSITO

aws: [ERROR]: An error occurred (SignatureDoesNotMatch) when calling the ListBuckets operation: The request signature we calculated does not match the signature you provided. nubecita dice por qué: la firma no coincide (revisa la petición canónica: ruta, consulta o cabeceras)
[codigo de salida: 254]
```

> 🥇 Con otro secreto, la firma sale distinta — y el secreto NO ha viajado por el cable ni una vez. Lo que no coincide es el resultado de una cadena de HMAC.

```console
$ AWS_ACCESS_KEY_ID=no-te-conozco aws --endpoint-url http://127.0.0.1:4566 s3 ls   # ⚠️ falla A PROPÓSITO

aws: [ERROR]: An error occurred (SignatureDoesNotMatch) when calling the ListBuckets operation: The request signature we calculated does not match the signature you provided. nubecita dice por qué: no conozco la clave de acceso no-te-conozco
[codigo de salida: 254]
```

> Y una clave de acceso desconocida se rechaza antes de calcular nada.

```console
$ curl -s -i http://127.0.0.1:4566/
HTTP/1.1 403 Forbidden
Content-Type: application/xml
Date: Tue, 25 Aug 2026 11:40:26 GMT
Content-Length: 243

<?xml version="1.0" encoding="UTF-8"?><Error><Code>SignatureDoesNotMatch</Code><Message>The request signature we calculated does not match the signature you provided. nubecita dice por qué: no viene la cabecera Authorization</Message></Error>
[codigo de salida: 0]
```

> Sin cabecera `Authorization` ni se mira qué pedías: 403 y a otra cosa.

---

## 6 · Lo que vio el servidor

Esto es la ventaja de tener tu propio emulador y no uno de caja: **ves el HTTP crudo que
manda el cliente de Amazon**. Fíjate en `?list-type=2`, en `encoding-type=url` (que la
CLI manda SIEMPRE), en `delimiter=%2F` — y en cómo cambia el registro cuando entran
DynamoDB y SQS: de rutas a `POST / [Operacion]`.

```console
GET /
  -> 200 0 bucket(s)
PUT /fotos
  -> 200 creado
PUT /fotos/gato.txt
  -> 200 15 byte(s) etag="2dab52e0e8834a6cbeb558a98af86ffa"
GET /fotos?list-type=2&prefix=2026%2F&encoding-type=url
  -> 200 0 objeto(s), 0 prefijo(s)
PUT /fotos/2026/enero.txt
  -> 200 5 byte(s) etag="a424b3d67f65e86bb5864579f9ab0d1d"
PUT /fotos/2026/febrero.txt
  -> 200 5 byte(s) etag="d5415f7c0df2ea624c49f1c243402c07"
GET /fotos?list-type=2&prefix=&encoding-type=url
  -> 200 3 objeto(s), 0 prefijo(s)
GET /fotos?list-type=2&delimiter=%2F&encoding-type=url
  -> 200 1 objeto(s), 1 prefijo(s)
HEAD /fotos/gato.txt
  -> 200 15 byte(s)
GET /fotos/no-existe
  -> 404 NoSuchKey
DELETE /fotos
  -> 409 BucketNotEmpty
GET /fotos?list-type=2&prefix=&encoding-type=url
  -> 200 3 objeto(s), 0 prefijo(s)
DELETE /fotos/2026/enero.txt
  -> 204
DELETE /fotos/2026/febrero.txt
  -> 204
DELETE /fotos/gato.txt
  -> 204
DELETE /fotos
  -> 204 borrado
POST / [CreateTable] 298 byte(s)
  -> 200 tabla "pedidos" creada
POST / [PutItem] 113 byte(s)
  -> 200 guardado en "pedidos"
POST / [PutItem] 111 byte(s)
  -> 200 guardado en "pedidos"
POST / [PutItem] 111 byte(s)
  -> 200 guardado en "pedidos"
POST / [Query] 199 byte(s)
  -> 200 Query: 2 de 3
POST / [Query] 127 byte(s)
  -> 400 ValidationException
POST / [Scan] 43 byte(s)
  -> 200 Scan: 3 elemento(s)
POST / [DeleteTable] 24 byte(s)
  -> 200 tabla "pedidos" borrada
POST / [SQS CreateQueue] 24 byte(s)
  -> 200 cola "muertos" creada
POST / [SQS CreateQueue] 189 byte(s)
  -> 200 cola "trabajos" creada
POST / [SQS SendMessage] 98 byte(s)
  -> 200 enviado trabajos-1
POST / [SQS ReceiveMessage] 87 byte(s)
  -> 200 1 mensaje(s) entregados
POST / [SQS ReceiveMessage] 87 byte(s)
  -> 200 1 mensaje(s) entregados
POST / [SQS ReceiveMessage] 60 byte(s)
  -> 200 0 mensaje(s) entregados
POST / [SQS ReceiveMessage] 59 byte(s)
  -> 200 1 mensaje(s) entregados
POST / [SQS GetQueueAttributes] 87 byte(s)
  -> 200 0 visible(s), 0 en vuelo
POST / [SQS DeleteQueue] 60 byte(s)
  -> 200 cola borrada
POST / [SQS DeleteQueue] 59 byte(s)
  -> 200 cola borrada
POST / [SQS ListQueues] 2 byte(s)
  -> 200 0 cola(s)
2026/08/25 11:40:14 nubecita · S3, DynamoDB y SQS en http://0.0.0.0:4566 · sin comprobar firmas · Ctrl+C para parar

# ── y con -firmar, la misma imagen ──
  -> 403 SignatureDoesNotMatch (no viene la cabecera Authorization)
PUT /fotos
  -> 200 creado
  -> 403 SignatureDoesNotMatch (la firma no coincide (revisa la petición canónica: ruta, consulta o cabeceras))
  -> 403 SignatureDoesNotMatch (no conozco la clave de acceso no-te-conozco)
  -> 403 SignatureDoesNotMatch (no viene la cabecera Authorization)
2026/08/25 11:40:25 nubecita · S3, DynamoDB y SQS en http://0.0.0.0:4566 · comprobando firmas SigV4 · Ctrl+C para parar
```

---

## 7 · El código completo

### `almacen.go`

```go
package main

// El almacen. Es lo unico que guarda estado, y por eso es lo unico que lleva candado.
//
// Decision de diseno del episodio 1: TODO en memoria. No hay disco, no hay base de datos.
// Cuando matas el proceso, se acabo. Eso no es una limitacion: es lo que hace que una prueba
// arranque limpia siempre, que es justo para lo que sirve un emulador.
//
// El candado es un RWMutex y no un Mutex porque el trafico real de un emulador de S3 es
// muchisimo mas de lectura que de escritura: listar y descargar. Con RWMutex, N lecturas
// pasan a la vez y solo las escrituras se serializan.

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"
)

// Objeto es un fichero guardado dentro de un bucket.
type Objeto struct {
	Clave      string
	Datos      []byte
	Tipo       string
	ETag       string // el md5 entre comillas, que es lo que S3 devuelve
	Tam        int    // el tamano va APARTE de Datos: en un listado no se devuelve contenido
	Modificado time.Time
}

// Bucket es un contenedor de objetos. En S3 el nombre es unico en toda la cuenta.
type Bucket struct {
	Nombre  string
	Creado  time.Time
	objetos map[string]*Objeto
}

// Almacen guarda todos los buckets de esta instancia.
type Almacen struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
	reloj   func() time.Time // inyectable: las pruebas necesitan un reloj que no se mueva
}

func NuevoAlmacen() *Almacen {
	return &Almacen{buckets: map[string]*Bucket{}, reloj: time.Now}
}

// CrearBucket devuelve false si ya existia. S3 distingue los dos casos con codigos de error
// distintos, asi que el que llama necesita saberlo.
func (a *Almacen) CrearBucket(nombre string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ya := a.buckets[nombre]; ya {
		return false
	}
	a.buckets[nombre] = &Bucket{Nombre: nombre, Creado: a.reloj().UTC(), objetos: map[string]*Objeto{}}
	return true
}

func (a *Almacen) HayBucket(nombre string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, ok := a.buckets[nombre]
	return ok
}

// BorrarBucket: S3 se niega a borrar un bucket que no esta vacio, y ese "no" es
// deliberadamente ruidoso. Aqui se replica.
func (a *Almacen) BorrarBucket(nombre string) (existia, vacio bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	b, ok := a.buckets[nombre]
	if !ok {
		return false, false
	}
	if len(b.objetos) > 0 {
		return true, false
	}
	delete(a.buckets, nombre)
	return true, true
}

// Buckets devuelve la lista ORDENADA por nombre. S3 los devuelve ordenados, y un mapa de Go
// no tiene orden: si no ordenas, la salida cambia en cada llamada y las pruebas parpadean.
func (a *Almacen) Buckets() []Bucket {
	a.mu.RLock()
	defer a.mu.RUnlock()
	fuera := make([]Bucket, 0, len(a.buckets))
	for _, b := range a.buckets {
		fuera = append(fuera, Bucket{Nombre: b.Nombre, Creado: b.Creado})
	}
	sort.Slice(fuera, func(i, j int) bool { return fuera[i].Nombre < fuera[j].Nombre })
	return fuera
}

// Poner guarda un objeto y devuelve su ETag. El ETag de S3 para una subida simple es el md5
// del contenido, entre comillas: no lo inventamos, lo calculamos igual que el de verdad.
func (a *Almacen) Poner(bucket, clave string, datos []byte, tipo string) (string, bool) {
	suma := md5.Sum(datos)
	etag := `"` + hex.EncodeToString(suma[:]) + `"`
	a.mu.Lock()
	defer a.mu.Unlock()
	b, ok := a.buckets[bucket]
	if !ok {
		return "", false
	}
	b.objetos[clave] = &Objeto{
		Clave: clave, Datos: datos, Tipo: tipo, ETag: etag, Tam: len(datos),
		Modificado: a.reloj().UTC(),
	}
	return etag, true
}

// Obtener devuelve una COPIA de los datos. Devolver el slice interno dejaria que quien llama
// lo modificara por debajo del candado: seria una carrera de datos que el -race caza y que en
// produccion se ve como bytes corrompidos sin explicacion.
func (a *Almacen) Obtener(bucket, clave string) (*Objeto, bool, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	b, ok := a.buckets[bucket]
	if !ok {
		return nil, false, false
	}
	o, ok := b.objetos[clave]
	if !ok {
		return nil, true, false
	}
	copia := *o
	copia.Datos = append([]byte(nil), o.Datos...)
	return &copia, true, true
}

func (a *Almacen) Borrar(bucket, clave string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	b, ok := a.buckets[bucket]
	if !ok {
		return false
	}
	delete(b.objetos, clave)
	return true
}

// Listar implementa lo que de verdad hace ListObjectsV2: filtrar por prefijo, agrupar por
// delimitador y cortar en maxClaves.
//
// El delimitador es la parte que nadie espera: S3 NO tiene carpetas. Lo que parece una carpeta
// es esto — cuando pides delimiter="/" , todas las claves que comparten el trozo hasta la
// siguiente barra se colapsan en un unico "prefijo comun". Por eso `aws s3 ls` parece un `ls`.
func (a *Almacen) Listar(bucket, prefijo, delimitador string, maxClaves int) ([]Objeto, []string, bool, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	b, ok := a.buckets[bucket]
	if !ok {
		return nil, nil, false, false
	}
	claves := make([]string, 0, len(b.objetos))
	for k := range b.objetos {
		if strings.HasPrefix(k, prefijo) {
			claves = append(claves, k)
		}
	}
	sort.Strings(claves)

	var objetos []Objeto
	var comunes []string
	vistos := map[string]bool{}
	truncado := false
	for _, k := range claves {
		if delimitador != "" {
			resto := k[len(prefijo):]
			if i := strings.Index(resto, delimitador); i >= 0 {
				grupo := prefijo + resto[:i+len(delimitador)]
				if !vistos[grupo] {
					if len(objetos)+len(comunes) >= maxClaves {
						truncado = true
						break
					}
					vistos[grupo] = true
					comunes = append(comunes, grupo)
				}
				continue
			}
		}
		if len(objetos)+len(comunes) >= maxClaves {
			truncado = true
			break
		}
		o := b.objetos[k]
		objetos = append(objetos, Objeto{
			Clave: o.Clave, ETag: o.ETag, Modificado: o.Modificado, Tipo: o.Tipo,
			Tam:   o.Tam,
			Datos: nil, // el listado no lleva contenido: seria devolver el bucket entero
		})
	}
	return objetos, comunes, truncado, true
}

// ⚠️ Aqui NO hay un metodo Tamano(bucket, clave) a proposito. Tendria que coger el RLock, y
// llamarlo desde dentro de Listar —que ya lo tiene— es la receta clasica del interbloqueo en Go:
// si un escritor esta esperando entre los dos RLock, el segundo se queda parado para siempre.
// Por eso el tamano viaja DENTRO del Objeto y no se vuelve a preguntar.
```

### `ayudas_test.go`

```go
package main

// Ayudas de las pruebas de firma: clientes de HTTP que se meten EN MEDIO y toquetean la
// petición DESPUÉS de que el SDK la haya firmado. Es la forma de representar a alguien que
// intercepta el tráfico, y lo que demuestra para qué sirve firmar.

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func credencialesFijas(acceso, secreto string) aws.CredentialsProvider {
	return credentials.NewStaticCredentialsProvider(acceso, secreto, "")
}

func hexDe(b []byte) string { return hex.EncodeToString(b) }

// clienteQueToquetea cambia el CUERPO después de la firma, dejando las cabeceras intactas.
type clienteQueToquetea struct{}

func (c *clienteQueToquetea) Do(r *http.Request) (*http.Response, error) {
	nuevo := []byte("transferir 10000 euros")
	r.Body = io.NopCloser(bytes.NewReader(nuevo))
	r.ContentLength = int64(len(nuevo))
	return http.DefaultClient.Do(r)
}

// clienteQueCambiaLaRuta manda la petición firmada a OTRO sitio.
type clienteQueCambiaLaRuta struct{}

func (c *clienteQueCambiaLaRuta) Do(r *http.Request) (*http.Response, error) {
	r.URL.Path = "/otro-bucket"
	return http.DefaultClient.Do(r)
}
```

### `colas.go`

```go
package main

// El estado de SQS: colas y mensajes. Y aqui vive el concepto que hace a SQS distinto de
// cualquier lista de la que hayas sacado elementos: la VISIBILIDAD.
//
// Un mensaje no se saca de la cola cuando lo recibes. Se vuelve INVISIBLE durante un rato. Si
// dentro de ese rato lo borras, desaparece. Si no lo borras —porque tu proceso murio, o tardo
// demasiado— vuelve a aparecer y se lo lleva otro.
//
// De ahi salen las dos consecuencias que definen como se escribe un consumidor:
//
//	1. La entrega es AL MENOS UNA VEZ, nunca exactamente una. Tu consumidor tiene que ser
//	   idempotente: procesar dos veces el mismo mensaje no puede cobrar dos veces.
//	2. Un mensaje que falla siempre se reintenta para siempre. Por eso existe la cola de
//	   mensajes muertos: a la N-esima entrega, en vez de darlo otra vez, se aparta.

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Mensaje struct {
	ID        string
	Cuerpo    string
	MD5       string
	Recibido  int       // cuantas veces se ha entregado ya
	VisibleEn time.Time // antes de este instante, no se entrega a nadie
	Recibo    string    // el ReceiptHandle vigente: cambia en CADA entrega
	Enviado   time.Time
}

type Cola struct {
	Nombre      string
	Visibilidad time.Duration
	MaxRecibos  int    // del RedrivePolicy: a partir de aqui, a la cola de muertos
	ColaMuertos string // nombre de la cola de muertos, si hay
	mensajes    []*Mensaje
	siguiente   int // para generar identificadores estables y legibles
}

type Colas struct {
	mu    sync.Mutex // aqui un Mutex normal: recibir MODIFICA (marca invisible), no solo lee
	colas map[string]*Cola
	reloj func() time.Time
	seq   int
}

// ⚠️ Aqui NO se guarda la direccion publica a proposito: la URL de la cola la arma quien atiende
// la peticion, con el Host que trajo el cliente. Guardarla al arrancar da URLs equivocadas en
// cuanto el servidor vive detras de un contenedor o de un proxy.
func NuevasColas() *Colas {
	return &Colas{colas: map[string]*Cola{}, reloj: time.Now}
}

const visibilidadPorDefecto = 30 * time.Second

// ── colas ────────────────────────────────────────────────────────────────

// Crear recibe la visibilidad como PUNTERO, y no es capricho: en SQS `0` es un valor legitimo
// («que reaparezca al instante») y hay que poder distinguirlo de «no me lo has dicho». Con un
// `time.Duration` a secas, `0` y «sin especificar» son el mismo valor y el cero se pierde.
func (cs *Colas) Crear(nombre string, visibilidad *time.Duration, maxRecibos int,
	colaMuertos string) *Cola {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if c, ya := cs.colas[nombre]; ya {
		// SQS es idempotente al crear si los atributos coinciden. Aqui se simplifica: se
		// devuelve la que hay. Lo que NO se hace es fallar en silencio.
		return c
	}
	vis := visibilidadPorDefecto
	if visibilidad != nil {
		vis = *visibilidad
	}
	c := &Cola{Nombre: nombre, Visibilidad: vis, MaxRecibos: maxRecibos,
		ColaMuertos: colaMuertos}
	cs.colas[nombre] = c
	return c
}

func (cs *Colas) Buscar(nombre string) (*Cola, bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	return c, ok
}

func (cs *Colas) Borrar(nombre string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if _, ok := cs.colas[nombre]; !ok {
		return false
	}
	delete(cs.colas, nombre)
	return true
}

func (cs *Colas) Listar(prefijo string) []string {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	fuera := make([]string, 0, len(cs.colas))
	for n := range cs.colas {
		if prefijo == "" || len(n) >= len(prefijo) && n[:len(prefijo)] == prefijo {
			fuera = append(fuera, n)
		}
	}
	sort.Strings(fuera)
	return fuera
}

func (cs *Colas) Vaciar(nombre string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return false
	}
	c.mensajes = nil
	return true
}

// ── mensajes ─────────────────────────────────────────────────────────────

func (cs *Colas) Enviar(nombre, cuerpo string) (*Mensaje, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return nil, errNoExisteCola
	}
	suma := md5.Sum([]byte(cuerpo))
	c.siguiente++
	cs.seq++
	m := &Mensaje{
		ID:      fmt.Sprintf("%s-%d", nombre, c.siguiente),
		Cuerpo:  cuerpo,
		MD5:     hex.EncodeToString(suma[:]),
		Enviado: cs.reloj().UTC(),
		// VisibleEn en cero: visible desde ya. (Aqui iria el retraso de entrega si se
		// implementara DelaySeconds.)
	}
	c.mensajes = append(c.mensajes, m)
	return m, nil
}

// Recibir entrega hasta `cuantos` mensajes visibles y los esconde durante `visibilidad`.
//
// 🥇 Este metodo es el corazon de SQS, y hace TRES cosas que sorprenden:
//   - marca el mensaje como invisible en vez de sacarlo;
//   - genera un RECIBO NUEVO en cada entrega — el recibo no es el identificador del mensaje,
//     y el de la entrega anterior deja de valer;
//   - si el mensaje ya se entrego demasiadas veces, NO lo entrega: lo aparta a la cola de
//     muertos. Esa es toda la magia del "dead letter queue".
func (cs *Colas) Recibir(nombre string, cuantos int, visibilidad *time.Duration) ([]*Mensaje, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return nil, errNoExisteCola
	}
	vis := c.Visibilidad
	if visibilidad != nil {
		vis = *visibilidad
	}
	ahora := cs.reloj().UTC()
	var salen []*Mensaje
	var quedan []*Mensaje
	for _, m := range c.mensajes {
		if len(salen) >= cuantos || m.VisibleEn.After(ahora) {
			quedan = append(quedan, m)
			continue
		}
		if c.MaxRecibos > 0 && m.Recibido+1 > c.MaxRecibos {
			if muertos, hay := cs.colas[c.ColaMuertos]; hay {
				m.VisibleEn = time.Time{}
				m.Recibido = 0
				m.Recibo = ""
				muertos.mensajes = append(muertos.mensajes, m)
			}
			// Si no hay cola de muertos configurada, el mensaje se descarta. Es lo que hace
			// SQS y es mejor que dar vueltas para siempre — pero por eso conviene tenerla.
			continue
		}
		m.Recibido++
		m.VisibleEn = ahora.Add(vis)
		cs.seq++
		m.Recibo = fmt.Sprintf("recibo-%d-%d", cs.seq, m.Recibido)
		salen = append(salen, m)
		quedan = append(quedan, m)
	}
	c.mensajes = quedan
	return salen, nil
}

// Quitar borra por RECIBO, no por identificador. Y devuelve si encontro algo: borrar con un
// recibo caducado no es un error en SQS, pero saberlo sirve para el registro.
func (cs *Colas) Quitar(nombre, recibo string) (bool, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return false, errNoExisteCola
	}
	for i, m := range c.mensajes {
		if m.Recibo != "" && m.Recibo == recibo {
			c.mensajes = append(c.mensajes[:i], c.mensajes[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// CambiarVisibilidad es lo que usa un consumidor que va a tardar mas de lo previsto: pide mas
// tiempo antes de que el mensaje reaparezca. Con 0, lo devuelve a la cola de inmediato.
func (cs *Colas) CambiarVisibilidad(nombre, recibo string, d time.Duration) (bool, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return false, errNoExisteCola
	}
	for _, m := range c.mensajes {
		if m.Recibo != "" && m.Recibo == recibo {
			m.VisibleEn = cs.reloj().UTC().Add(d)
			return true, nil
		}
	}
	return false, nil
}

// Cuentas devuelve lo que SQS llama "aproximado", y aqui se ve por que lo llama asi: son dos
// numeros que dependen del instante exacto en que preguntas.
func (cs *Colas) Cuentas(nombre string) (visibles, enVuelo int, ok bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, hay := cs.colas[nombre]
	if !hay {
		return 0, 0, false
	}
	ahora := cs.reloj().UTC()
	for _, m := range c.mensajes {
		if m.VisibleEn.After(ahora) {
			enVuelo++
		} else {
			visibles++
		}
	}
	return visibles, enVuelo, true
}
```

### `dynamo.go`

```go
package main

// El servicio DynamoDB. Y aqui viene la sorpresa del episodio 2: NO se parece en nada a S3.
//
//	S3         el metodo y la ruta dicen que quieres.  PUT /fotos/gato.png  ->  respuesta en XML
//	DynamoDB   SIEMPRE POST a "/", cuerpo JSON, y la operacion en una cabecera:
//	           X-Amz-Target: DynamoDB_20120810.PutItem
//
// Dos servicios de la misma nube, dos protocolos distintos. No es incoherencia: es historia.
// S3 es de 2006 y nacio como almacen de objetos sobre HTTP; DynamoDB es de 2012, cuando Amazon
// ya habia estandarizado el estilo "JSON con la operacion en la cabecera" para lo demas.
//
// Lo que se implementa aqui, y es todo lo que se usa un dia normal:
//
//	CreateTable · DescribeTable · DeleteTable · ListTables
//	PutItem · GetItem · DeleteItem
//	Query (con KeyConditionExpression completa) · Scan
//
// ⛔ Lo que NO: indices secundarios, UpdateItem, transacciones, streams, paginacion real y
// FilterExpression. No es una lista de tareas: es la frontera de un doble de pruebas.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	errNoExisteTabla = errors.New("la tabla no existe")
	errCondicion     = errors.New("la condicion no se cumple")
)

const prefijoTipo = "com.amazonaws.dynamodb.v20120810#"

type ServicioDynamo struct {
	tablas   *Tablas
	registro func(formato string, args ...any)
}

func NuevoDynamo(t *Tablas, registro func(string, ...any)) *ServicioDynamo {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioDynamo{tablas: t, registro: registro}
}

func (s *ServicioDynamo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	destino := r.Header.Get("X-Amz-Target")
	op := destino
	if i := strings.LastIndex(destino, "."); i >= 0 {
		op = destino[i+1:]
	}
	cuerpo, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	s.registro("POST / [%s] %d byte(s)", op, len(cuerpo))

	switch op {
	case "CreateTable":
		s.crearTabla(w, cuerpo)
	case "DescribeTable":
		s.describirTabla(w, cuerpo)
	case "DeleteTable":
		s.borrarTabla(w, cuerpo)
	case "ListTables":
		s.listarTablas(w)
	case "PutItem":
		s.ponerElemento(w, cuerpo)
	case "GetItem":
		s.obtenerElemento(w, cuerpo)
	case "DeleteItem":
		s.borrarElemento(w, cuerpo)
	case "Query":
		s.consultar(w, cuerpo)
	case "Scan":
		s.recorrer(w, cuerpo)
	default:
		// ⚠️ Fallo RUIDOSO y con el nombre de la operacion dentro. Un doble que contesta 200 a
		// algo que no implementa hace creer que la prueba paso.
		s.fallo(w, "UnknownOperationException",
			fmt.Sprintf("nubecita no implementa %q (episodio 2: S3 + DynamoDB)", op))
	}
}

// ── respuestas y errores ─────────────────────────────────────────────────
// El SDK convierte el campo `__type` en un tipo de error de Go que tu codigo puede preguntar con
// errors.As. Si contestas un 400 sin ese campo, el que te llama recibe "algo salio mal".

func (s *ServicioDynamo) responder(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *ServicioDynamo) fallo(w http.ResponseWriter, tipo, mensaje string) {
	s.registro("  -> 400 %s", tipo)
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"__type": prefijoTipo + tipo, "message": mensaje,
	})
}

func (s *ServicioDynamo) fallaPorError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, errNoExisteTabla):
		s.fallo(w, "ResourceNotFoundException", "Requested resource not found")
	case errors.Is(err, errCondicion):
		s.fallo(w, "ConditionalCheckFailedException", "The conditional request failed")
	default:
		s.fallo(w, "ValidationException", err.Error())
	}
	return true
}

// ── tablas ───────────────────────────────────────────────────────────────

type esquemaClave struct {
	AttributeName string `json:"AttributeName"`
	KeyType       string `json:"KeyType"` // HASH | RANGE
}

type defAtributo struct {
	AttributeName string `json:"AttributeName"`
	AttributeType string `json:"AttributeType"` // S | N | B
}

type peticionCrearTabla struct {
	TableName            string         `json:"TableName"`
	KeySchema            []esquemaClave `json:"KeySchema"`
	AttributeDefinitions []defAtributo  `json:"AttributeDefinitions"`
}

func (s *ServicioDynamo) fichaTabla(t *Tabla) map[string]any {
	esquema := []map[string]string{{"AttributeName": t.PK.Nombre, "KeyType": "HASH"}}
	atrib := []map[string]string{{"AttributeName": t.PK.Nombre, "AttributeType": t.PK.Tipo}}
	if t.SK != nil {
		esquema = append(esquema, map[string]string{"AttributeName": t.SK.Nombre, "KeyType": "RANGE"})
		atrib = append(atrib, map[string]string{"AttributeName": t.SK.Nombre, "AttributeType": t.SK.Tipo})
	}
	return map[string]any{
		"TableName": t.Nombre,
		// ACTIVE desde el primer instante: en Amazon la tabla tarda en crearse y hay que esperar.
		// Aqui no, y eso es una ventaja del doble, no una mentira: se dice en el README.
		"TableStatus":          "ACTIVE",
		"KeySchema":            esquema,
		"AttributeDefinitions": atrib,
		"CreationDateTime":     float64(t.Creada.UnixMilli()) / 1000.0,
		"ItemCount":            s.tablas.Cuenta(t.Nombre),
		"TableSizeBytes":       0,
		"TableArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/" + t.Nombre,
		"BillingModeSummary":   map[string]any{"BillingMode": "PAY_PER_REQUEST"},
	}
}

func (s *ServicioDynamo) crearTabla(w http.ResponseWriter, cuerpo []byte) {
	var p peticionCrearTabla
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	tipos := map[string]string{}
	for _, a := range p.AttributeDefinitions {
		tipos[a.AttributeName] = a.AttributeType
	}
	var pk Clave
	var sk *Clave
	for _, k := range p.KeySchema {
		c := Clave{Nombre: k.AttributeName, Tipo: tipos[k.AttributeName]}
		if c.Tipo == "" {
			s.fallo(w, "ValidationException",
				"el atributo "+k.AttributeName+" esta en KeySchema pero no en AttributeDefinitions")
			return
		}
		if k.KeyType == "HASH" {
			pk = c
		} else {
			copia := c
			sk = &copia
		}
	}
	if pk.Nombre == "" {
		s.fallo(w, "ValidationException", "falta la clave de particion (KeyType HASH)")
		return
	}
	t, nuevo := s.tablas.Crear(p.TableName, pk, sk)
	if !nuevo {
		s.fallo(w, "ResourceInUseException", "Table already exists: "+p.TableName)
		return
	}
	s.registro("  -> 200 tabla %q creada", p.TableName)
	s.responder(w, map[string]any{"TableDescription": s.fichaTabla(t)})
}

func (s *ServicioDynamo) describirTabla(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		TableName string `json:"TableName"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	t, ok := s.tablas.Buscar(p.TableName)
	if !ok {
		s.fallo(w, "ResourceNotFoundException", "Requested resource not found: Table: "+p.TableName)
		return
	}
	s.responder(w, map[string]any{"Table": s.fichaTabla(t)})
}

func (s *ServicioDynamo) borrarTabla(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		TableName string `json:"TableName"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	t, ok := s.tablas.Borrar(p.TableName)
	if !ok {
		s.fallo(w, "ResourceNotFoundException", "Requested resource not found: Table: "+p.TableName)
		return
	}
	s.registro("  -> 200 tabla %q borrada", p.TableName)
	s.responder(w, map[string]any{"TableDescription": s.fichaTabla(t)})
}

func (s *ServicioDynamo) listarTablas(w http.ResponseWriter) {
	nombres := s.tablas.Listar()
	s.registro("  -> 200 %d tabla(s)", len(nombres))
	s.responder(w, map[string]any{"TableNames": nombres})
}

// ── elementos ────────────────────────────────────────────────────────────

type peticionElemento struct {
	TableName                string            `json:"TableName"`
	Item                     Elemento          `json:"Item"`
	Key                      Elemento          `json:"Key"`
	ConditionExpression      string            `json:"ConditionExpression"`
	ExpressionAttributeNames map[string]string `json:"ExpressionAttributeNames"`
	ReturnValues             string            `json:"ReturnValues"`
}

func (s *ServicioDynamo) ponerElemento(w http.ResponseWriter, cuerpo []byte) {
	var p peticionElemento
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	cond, err := condicionSimple(p.ConditionExpression, p.ExpressionAttributeNames)
	if err != nil {
		s.fallo(w, "ValidationException", err.Error())
		return
	}
	anterior, err := s.tablas.Poner(p.TableName, p.Item, cond)
	if s.fallaPorError(w, err) {
		return
	}
	s.registro("  -> 200 guardado en %q", p.TableName)
	res := map[string]any{}
	if p.ReturnValues == "ALL_OLD" && anterior != nil {
		res["Attributes"] = anterior
	}
	s.responder(w, res)
}

func (s *ServicioDynamo) obtenerElemento(w http.ResponseWriter, cuerpo []byte) {
	var p peticionElemento
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	e, err := s.tablas.Obtener(p.TableName, p.Key)
	if s.fallaPorError(w, err) {
		return
	}
	res := map[string]any{}
	if e != nil {
		res["Item"] = e
	}
	// ⚠️ Que no exista NO es un error en DynamoDB: es un 200 sin campo Item. Si aqui devolvieras
	// un error, el codigo del que te llama trataria "no hay" como "se rompio".
	s.registro("  -> 200 %s", map[bool]string{true: "encontrado", false: "no existe"}[e != nil])
	s.responder(w, res)
}

func (s *ServicioDynamo) borrarElemento(w http.ResponseWriter, cuerpo []byte) {
	var p peticionElemento
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	cond, err := condicionSimple(p.ConditionExpression, p.ExpressionAttributeNames)
	if err != nil {
		s.fallo(w, "ValidationException", err.Error())
		return
	}
	anterior, err := s.tablas.Quitar(p.TableName, p.Key, cond)
	if s.fallaPorError(w, err) {
		return
	}
	res := map[string]any{}
	if p.ReturnValues == "ALL_OLD" && anterior != nil {
		res["Attributes"] = anterior
	}
	s.registro("  -> 200 borrado")
	s.responder(w, res)
}

// ── Query y Scan ─────────────────────────────────────────────────────────

type peticionConsulta struct {
	TableName                 string            `json:"TableName"`
	KeyConditionExpression    string            `json:"KeyConditionExpression"`
	ExpressionAttributeNames  map[string]string `json:"ExpressionAttributeNames"`
	ExpressionAttributeValues map[string]AV     `json:"ExpressionAttributeValues"`
	ScanIndexForward          *bool             `json:"ScanIndexForward"`
	Limit                     *int              `json:"Limit"`
	Select                    string            `json:"Select"`
}

// soloContar: con `--select COUNT` DynamoDB devuelve el numero y NO los elementos. Devolverlos
// igualmente parece inofensivo hasta que alguien cuenta un millon de filas y se lleva el millon
// de filas por la red.
func soloContar(sel string) bool { return strings.EqualFold(sel, "COUNT") }

func (s *ServicioDynamo) consultar(w http.ResponseWriter, cuerpo []byte) {
	var p peticionConsulta
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	todos, t, err := s.tablas.Recorrer(p.TableName)
	if s.fallaPorError(w, err) {
		return
	}
	cond, err := analizarClave(p.KeyConditionExpression, p.ExpressionAttributeNames,
		p.ExpressionAttributeValues, t)
	if err != nil {
		s.fallo(w, "ValidationException", err.Error())
		return
	}
	var salen []Elemento
	for _, e := range todos {
		if cond.cumple(e) {
			salen = append(salen, e)
		}
	}
	// Descendente: `ScanIndexForward: false`. Los elementos ya vienen ordenados por clave de
	// orden, asi que basta con darles la vuelta.
	if p.ScanIndexForward != nil && !*p.ScanIndexForward {
		for i, j := 0, len(salen)-1; i < j; i, j = i+1, j-1 {
			salen[i], salen[j] = salen[j], salen[i]
		}
	}
	escaneados := len(salen)
	if p.Limit != nil && *p.Limit >= 0 && *p.Limit < len(salen) {
		salen = salen[:*p.Limit]
	}
	s.registro("  -> 200 Query: %d de %d", len(salen), len(todos))
	res := map[string]any{"Count": len(salen), "ScannedCount": escaneados}
	if !soloContar(p.Select) {
		res["Items"] = vaciaSiNil(salen)
	}
	s.responder(w, res)
}

func (s *ServicioDynamo) recorrer(w http.ResponseWriter, cuerpo []byte) {
	var p peticionConsulta
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	todos, _, err := s.tablas.Recorrer(p.TableName)
	if s.fallaPorError(w, err) {
		return
	}
	escaneados := len(todos)
	if p.Limit != nil && *p.Limit >= 0 && *p.Limit < len(todos) {
		todos = todos[:*p.Limit]
	}
	// ⚠️ Scan LEE LA TABLA ENTERA. Aqui da igual porque son cuatro elementos en memoria; en
	// Amazon eso es la diferencia entre unos centavos y una factura que duele.
	s.registro("  -> 200 Scan: %d elemento(s)", len(todos))
	res := map[string]any{"Count": len(todos), "ScannedCount": escaneados}
	if !soloContar(p.Select) {
		res["Items"] = vaciaSiNil(todos)
	}
	s.responder(w, res)
}

// vaciaSiNil evita que un slice nil salga como `null` en el JSON. El SDK lo aguanta, pero la CLI
// lo enseña como `null` en vez de `[]` y parece que algo se rompio.
func vaciaSiNil(e []Elemento) []Elemento {
	if e == nil {
		return []Elemento{}
	}
	return e
}

// ── las expresiones ──────────────────────────────────────────────────────
// Esta es la parte que de verdad enseña como piensa DynamoDB. `KeyConditionExpression` solo
// admite: igualdad EXACTA en la clave de particion, y opcionalmente UNA comparacion sobre la
// clave de orden. No hay OR, no hay otros atributos. Esa limitacion es el diseño.

type condicionClave struct {
	pkNombre string
	pkValor  AV
	skNombre string
	skOp     string // "" | = | < | <= | > | >= | BETWEEN | begins_with
	skA, skB AV
}

func (c *condicionClave) cumple(e Elemento) bool {
	if x, ok := compara(e[c.pkNombre], c.pkValor); !ok || x != 0 {
		return false
	}
	if c.skOp == "" {
		return true
	}
	v, hay := e[c.skNombre]
	if !hay {
		return false
	}
	if c.skOp == "begins_with" {
		s, ok := v["S"].(string)
		p, ok2 := c.skA["S"].(string)
		return ok && ok2 && strings.HasPrefix(s, p)
	}
	x, ok := compara(v, c.skA)
	if !ok {
		return false
	}
	switch c.skOp {
	case "=":
		return x == 0
	case "<":
		return x < 0
	case "<=":
		return x <= 0
	case ">":
		return x > 0
	case ">=":
		return x >= 0
	case "BETWEEN":
		y, ok2 := compara(v, c.skB)
		return ok2 && x >= 0 && y <= 0
	}
	return false
}

// trocear parte la expresion en piezas. Se hace a mano y no con strings.Split porque los
// simbolos van pegados a los nombres: "id=:v" es tan valido como "id = :v".
func trocear(expr string) []string {
	var piezas []string
	i := 0
	for i < len(expr) {
		c := expr[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n':
			i++
		case c == '(' || c == ')' || c == ',':
			piezas = append(piezas, string(c))
			i++
		case c == '=' || c == '<' || c == '>':
			if i+1 < len(expr) && expr[i+1] == '=' {
				piezas = append(piezas, expr[i:i+2])
				i += 2
			} else {
				piezas = append(piezas, string(c))
				i++
			}
		default:
			j := i
			for j < len(expr) && !strings.ContainsRune(" \t\n(),=<>", rune(expr[j])) {
				j++
			}
			piezas = append(piezas, expr[i:j])
			i = j
		}
	}
	return piezas
}

func analizarClave(expr string, nombres map[string]string, valores map[string]AV,
	t *Tabla) (*condicionClave, error) {
	if strings.TrimSpace(expr) == "" {
		return nil, errors.New("falta KeyConditionExpression")
	}
	nombre := func(s string) string {
		if strings.HasPrefix(s, "#") {
			if v, ok := nombres[s]; ok {
				return v
			}
		}
		return s
	}
	valor := func(s string) (AV, error) {
		if !strings.HasPrefix(s, ":") {
			return nil, fmt.Errorf("se esperaba un marcador de valor y llego %q", s)
		}
		v, ok := valores[s]
		if !ok {
			return nil, fmt.Errorf("el marcador %s no esta en ExpressionAttributeValues", s)
		}
		return v, nil
	}

	piezas := trocear(expr)
	c := &condicionClave{}
	i := 0
	primero := true
	for i < len(piezas) {
		if !primero {
			if !strings.EqualFold(piezas[i], "AND") {
				return nil, fmt.Errorf("solo se admite AND entre condiciones, no %q", piezas[i])
			}
			i++
		}
		primero = false
		if i >= len(piezas) {
			return nil, errors.New("la expresion termina en AND")
		}
		// begins_with(nombre, :valor)
		if strings.EqualFold(piezas[i], "begins_with") {
			if i+5 >= len(piezas) || piezas[i+1] != "(" || piezas[i+3] != "," ||
				piezas[i+5] != ")" {
				return nil, errors.New("begins_with se escribe begins_with(nombre, :valor)")
			}
			v, err := valor(piezas[i+4])
			if err != nil {
				return nil, err
			}
			c.skNombre, c.skOp, c.skA = nombre(piezas[i+2]), "begins_with", v
			i += 6
			continue
		}
		if i+2 >= len(piezas) {
			return nil, fmt.Errorf("condicion incompleta cerca de %q", piezas[i])
		}
		n, op := nombre(piezas[i]), piezas[i+1]
		if strings.EqualFold(op, "BETWEEN") {
			if i+4 >= len(piezas) || !strings.EqualFold(piezas[i+3], "AND") {
				return nil, errors.New("BETWEEN se escribe: nombre BETWEEN :a AND :b")
			}
			a, err := valor(piezas[i+2])
			if err != nil {
				return nil, err
			}
			b, err := valor(piezas[i+4])
			if err != nil {
				return nil, err
			}
			c.skNombre, c.skOp, c.skA, c.skB = n, "BETWEEN", a, b
			i += 5
			continue
		}
		v, err := valor(piezas[i+2])
		if err != nil {
			return nil, err
		}
		if n == t.PK.Nombre {
			if op != "=" {
				// ⛔ Esta es LA regla de DynamoDB, y por eso el error es explicito: sobre la
				// clave de particion solo hay igualdad. No hay "usuarios que empiecen por a".
				return nil, errors.New("sobre la clave de particion solo se admite =")
			}
			c.pkNombre, c.pkValor = n, v
		} else {
			if !strings.Contains("=,<,<=,>,>=", op) {
				return nil, fmt.Errorf("operador no admitido: %q", op)
			}
			c.skNombre, c.skOp, c.skA = n, op, v
		}
		i += 3
	}
	if c.pkNombre == "" {
		return nil, errors.New("la condicion tiene que fijar la clave de particion con =")
	}
	return c, nil
}

// condicionSimple entiende `attribute_exists(x)` y `attribute_not_exists(x)`, que es lo que hace
// falta para la escritura condicional: la unica herramienta de concurrencia que da DynamoDB.
//
//	"escribe esto SOLO si nadie lo escribio antes"  ->  attribute_not_exists(pk)
//
// Devuelve nil cuando no hay condicion: quien llama interpreta nil como "adelante".
func condicionSimple(expr string, nombres map[string]string) (func(Elemento) bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, nil
	}
	piezas := trocear(expr)
	if len(piezas) != 4 || piezas[1] != "(" || piezas[3] != ")" {
		return nil, fmt.Errorf("nubecita solo entiende attribute_exists(x) y "+
			"attribute_not_exists(x) en ConditionExpression, no %q", expr)
	}
	atributo := piezas[2]
	if v, ok := nombres[atributo]; ok {
		atributo = v
	}
	switch strings.ToLower(piezas[0]) {
	case "attribute_not_exists":
		return func(anterior Elemento) bool {
			if anterior == nil {
				return true
			}
			_, hay := anterior[atributo]
			return !hay
		}, nil
	case "attribute_exists":
		return func(anterior Elemento) bool {
			if anterior == nil {
				return false
			}
			_, hay := anterior[atributo]
			return hay
		}, nil
	}
	return nil, fmt.Errorf("funcion no admitida en ConditionExpression: %q", piezas[0])
}
```

### `dynamo_test.go`

```go
package main

// Igual que en S3: estas pruebas NO usan un cliente hecho a medida. Usan
// `github.com/aws/aws-sdk-go-v2/service/dynamodb` y `feature/dynamodb/attributevalue`, que es lo
// que usarias contra Amazon. Si el SDK oficial lo entiende, es compatible.
//
//	go test ./... -count=1 -v

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Pedido es una struct de Go normal. El paquete `attributevalue` la convierte al mapa con tipos
// que entiende DynamoDB, y al reves. Es el puente que evita escribir {"S": ...} a mano.
type Pedido struct {
	Cliente string  `dynamodbav:"cliente"` // clave de particion
	Fecha   string  `dynamodbav:"fecha"`   // clave de orden, en ISO-8601 A PROPOSITO
	Total   float64 `dynamodbav:"total"`
	Estado  string  `dynamodbav:"estado"`
}

func nubeDynamo(t *testing.T) *dynamodb.Client {
	t.Helper()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"DynamoDB_": NuevoDynamo(NuevasTablas(), nil),
	}))
	t.Cleanup(srv.Close)

	return clienteDynamo(srv.URL)
}

// clienteDynamo esta separado porque lo reutiliza la prueba de que los tres servicios conviven
// en el mismo puerto (`sqs_test.go`).
func clienteDynamo(url string) *dynamodb.Client {
	return dynamodb.New(dynamodb.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(url),
	})
}

// tablaPedidos crea la tabla con clave compuesta: particion `cliente`, orden `fecha`.
func tablaPedidos(t *testing.T, c *dynamodb.Client) {
	t.Helper()
	ctx := context.Background()
	_, err := c.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("pedidos"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("cliente"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("fecha"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("cliente"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("fecha"), KeyType: types.KeyTypeRange},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
}

func guardar(t *testing.T, c *dynamodb.Client, p Pedido) {
	t.Helper()
	elem, err := attributevalue.MarshalMap(p)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if _, err := c.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"), Item: elem,
	}); err != nil {
		t.Fatalf("PutItem: %v", err)
	}
}

func TestTablaCicloCompleto(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)

	lista, err := c.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	if len(lista.TableNames) != 1 || lista.TableNames[0] != "pedidos" {
		t.Fatalf("ListTables = %v", lista.TableNames)
	}

	ficha, err := c.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String("pedidos")})
	if err != nil {
		t.Fatalf("DescribeTable: %v", err)
	}
	if ficha.Table.TableStatus != types.TableStatusActive {
		t.Errorf("TableStatus = %v, quiero ACTIVE", ficha.Table.TableStatus)
	}
	if len(ficha.Table.KeySchema) != 2 {
		t.Errorf("KeySchema tiene %d elementos, quiero 2", len(ficha.Table.KeySchema))
	}

	if _, err := c.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String("pedidos"),
	}); err != nil {
		t.Fatalf("DeleteTable: %v", err)
	}
	lista, _ = c.ListTables(ctx, &dynamodb.ListTablesInput{})
	if len(lista.TableNames) != 0 {
		t.Errorf("quedaron tablas: %v", lista.TableNames)
	}
}

func TestGuardarYLeerConStructDeGo(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15", Total: 42.5, Estado: "pagado"})

	res, err := c.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("pedidos"),
		Key: map[string]types.AttributeValue{
			"cliente": &types.AttributeValueMemberS{Value: "ana"},
			"fecha":   &types.AttributeValueMemberS{Value: "2026-01-15"},
		},
	})
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	var vuelve Pedido
	if err := attributevalue.UnmarshalMap(res.Item, &vuelve); err != nil {
		t.Fatalf("UnmarshalMap: %v", err)
	}
	if vuelve.Total != 42.5 || vuelve.Estado != "pagado" {
		t.Errorf("volvio %+v", vuelve)
	}
}

// Que no exista NO es un error en DynamoDB: es un 200 sin Item. Si lo tratas como error, tu
// codigo confunde "no hay" con "se rompio".
func TestLoQueNoExisteNoEsUnError(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)

	res, err := c.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("pedidos"),
		Key: map[string]types.AttributeValue{
			"cliente": &types.AttributeValueMemberS{Value: "nadie"},
			"fecha":   &types.AttributeValueMemberS{Value: "2026-01-01"},
		},
	})
	if err != nil {
		t.Fatalf("GetItem devolvio error y no deberia: %v", err)
	}
	if res.Item != nil {
		t.Errorf("devolvio un elemento que no existe: %v", res.Item)
	}
}

func TestTablaQueNoExisteLlegaComoResourceNotFound(t *testing.T) {
	c := nubeDynamo(t)
	_, err := c.Scan(context.Background(), &dynamodb.ScanInput{TableName: aws.String("fantasma")})
	var noEsta *types.ResourceNotFoundException
	if !errors.As(err, &noEsta) {
		t.Fatalf("el SDK no lo vio como ResourceNotFoundException: %v", err)
	}
}

func TestCrearDosVecesLaMismaTablaLlegaComoResourceInUse(t *testing.T) {
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	_, err := c.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: aws.String("pedidos"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("cliente"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("cliente"), KeyType: types.KeyTypeHash},
		},
	})
	var enUso *types.ResourceInUseException
	if !errors.As(err, &enUso) {
		t.Fatalf("el SDK no lo vio como ResourceInUseException: %v", err)
	}
}

// 🥇 La prueba que enseña como piensa DynamoDB: Query SOLO sabe de la clave.
func TestQuerySoloMiraLaClave(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15", Total: 10, Estado: "pagado"})
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-02-03", Total: 20, Estado: "enviado"})
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-03-21", Total: 30, Estado: "pagado"})
	guardar(t, c, Pedido{Cliente: "beto", Fecha: "2026-02-10", Total: 99, Estado: "pagado"})

	// Todos los de ana: igualdad EXACTA en la clave de particion.
	res, err := c.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String("pedidos"),
		KeyConditionExpression:    aws.String("cliente = :c"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":c": &types.AttributeValueMemberS{Value: "ana"}},
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if res.Count != 3 {
		t.Errorf("Count = %d, quiero 3", res.Count)
	}

	// Un rango sobre la clave de ORDEN. Y funciona porque las fechas van en ISO-8601: el orden
	// alfabetico coincide con el cronologico. Con "15/01/2026" esto daria basura.
	res, err = c.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("pedidos"),
		KeyConditionExpression: aws.String("cliente = :c AND fecha BETWEEN :a AND :b"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":c": &types.AttributeValueMemberS{Value: "ana"},
			":a": &types.AttributeValueMemberS{Value: "2026-02-01"},
			":b": &types.AttributeValueMemberS{Value: "2026-02-28"},
		},
	})
	if err != nil {
		t.Fatalf("Query BETWEEN: %v", err)
	}
	if res.Count != 1 {
		t.Fatalf("BETWEEN dio %d, quiero 1", res.Count)
	}
	var uno Pedido
	_ = attributevalue.UnmarshalMap(res.Items[0], &uno)
	if uno.Fecha != "2026-02-03" {
		t.Errorf("salio %s", uno.Fecha)
	}

	// begins_with sobre la clave de orden: "todo lo de 2026-03".
	res, err = c.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("pedidos"),
		KeyConditionExpression: aws.String("cliente = :c AND begins_with(fecha, :p)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":c": &types.AttributeValueMemberS{Value: "ana"},
			":p": &types.AttributeValueMemberS{Value: "2026-03"},
		},
	})
	if err != nil {
		t.Fatalf("Query begins_with: %v", err)
	}
	if res.Count != 1 {
		t.Errorf("begins_with dio %d, quiero 1", res.Count)
	}
}

// ⛔ Sobre la clave de particion SOLO hay igualdad. No es una limitacion de nubecita: es la
// regla de DynamoDB, y por eso el error tiene que llegar como ValidationException.
func TestNoSePuedeBuscarPorPrefijoEnLaClaveDeParticion(t *testing.T) {
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	_, err := c.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              aws.String("pedidos"),
		KeyConditionExpression: aws.String("begins_with(cliente, :p)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: "a"},
		},
	})
	if err == nil {
		t.Fatal("dejo buscar por prefijo en la clave de particion, y DynamoDB no deja")
	}
	// El mensaje tiene que DECIR por que. Un "ValidationException" a secas no ayuda a nadie.
	if !strings.Contains(err.Error(), "clave de particion") {
		t.Fatalf("el error no explica por que: %v", err)
	}
}

func TestQueryAlReves(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15", Total: 10})
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-03-21", Total: 30})

	res, err := c.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String("pedidos"),
		KeyConditionExpression:    aws.String("cliente = :c"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":c": &types.AttributeValueMemberS{Value: "ana"}},
		ScanIndexForward:          aws.Bool(false),
		Limit:                     aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	var uno Pedido
	_ = attributevalue.UnmarshalMap(res.Items[0], &uno)
	if uno.Fecha != "2026-03-21" {
		t.Errorf("el mas reciente deberia ser 2026-03-21 y salio %s", uno.Fecha)
	}
}

// La escritura condicional: la unica herramienta de concurrencia que da DynamoDB.
func TestEscrituraCondicional(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	elem, _ := attributevalue.MarshalMap(Pedido{Cliente: "ana", Fecha: "2026-01-15", Total: 10})

	// La primera pasa: nadie lo habia escrito.
	if _, err := c.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"), Item: elem,
		ConditionExpression: aws.String("attribute_not_exists(cliente)"),
	}); err != nil {
		t.Fatalf("la primera escritura condicional deberia pasar: %v", err)
	}
	// La segunda no, y tiene que llegar con SU tipo de error.
	_, err := c.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"), Item: elem,
		ConditionExpression: aws.String("attribute_not_exists(cliente)"),
	})
	var fallo *types.ConditionalCheckFailedException
	if !errors.As(err, &fallo) {
		t.Fatalf("el SDK no lo vio como ConditionalCheckFailedException: %v", err)
	}
}

// Scan lee la tabla ENTERA. Aqui son cuatro elementos; en Amazon es la diferencia entre unos
// centavos y una factura que duele.
func TestScanLeeTodo(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15"})
	guardar(t, c, Pedido{Cliente: "beto", Fecha: "2026-02-10"})

	res, err := c.Scan(ctx, &dynamodb.ScanInput{TableName: aws.String("pedidos")})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if res.Count != 2 || res.ScannedCount != 2 {
		t.Errorf("Count=%d ScannedCount=%d, quiero 2 y 2", res.Count, res.ScannedCount)
	}
}

func TestBorrarElemento(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15"})

	clave := map[string]types.AttributeValue{
		"cliente": &types.AttributeValueMemberS{Value: "ana"},
		"fecha":   &types.AttributeValueMemberS{Value: "2026-01-15"},
	}
	if _, err := c.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("pedidos"), Key: clave,
	}); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	// Borrar dos veces no es un error, igual que en S3.
	if _, err := c.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("pedidos"), Key: clave,
	}); err != nil {
		t.Fatalf("el segundo DeleteItem no deberia fallar: %v", err)
	}
	res, _ := c.Scan(ctx, &dynamodb.ScanInput{TableName: aws.String("pedidos")})
	if res.Count != 0 {
		t.Errorf("quedaron %d elementos", res.Count)
	}
}

// 🥇 Los numeros viajan como CADENA por el cable, y no es un descuido: asi no pasan por el float
// de ningun lenguaje y no pierden precision.
func TestLosNumerosViajanComoCadena(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	if _, err := c.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"),
		Item: map[string]types.AttributeValue{
			"cliente": &types.AttributeValueMemberS{Value: "ana"},
			"fecha":   &types.AttributeValueMemberS{Value: "2026-01-15"},
			// 20 digitos: no cabe entero en un float64.
			"exacto": &types.AttributeValueMemberN{Value: "12345678901234567890"},
		},
	}); err != nil {
		t.Fatalf("PutItem: %v", err)
	}
	res, err := c.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("pedidos"),
		Key: map[string]types.AttributeValue{
			"cliente": &types.AttributeValueMemberS{Value: "ana"},
			"fecha":   &types.AttributeValueMemberS{Value: "2026-01-15"},
		},
	})
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	n, ok := res.Item["exacto"].(*types.AttributeValueMemberN)
	if !ok {
		t.Fatalf("el atributo no volvio como N: %T", res.Item["exacto"])
	}
	if n.Value != "12345678901234567890" {
		t.Errorf("volvio %q: perdio precision por el camino", n.Value)
	}
}

// Clave de orden NUMERICA: el 10 va DESPUES del 9. Si ordenaras por la cadena, iria antes.
func TestLaClaveDeOrdenNumericaSeOrdenaComoNumero(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	if _, err := c.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("eventos"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("flujo"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("secuencia"), AttributeType: types.ScalarAttributeTypeN},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("flujo"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("secuencia"), KeyType: types.KeyTypeRange},
		},
	}); err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	for _, n := range []string{"10", "9", "2"} {
		if _, err := c.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("eventos"),
			Item: map[string]types.AttributeValue{
				"flujo":     &types.AttributeValueMemberS{Value: "a"},
				"secuencia": &types.AttributeValueMemberN{Value: n},
			},
		}); err != nil {
			t.Fatalf("PutItem %s: %v", n, err)
		}
	}
	res, err := c.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String("eventos"),
		KeyConditionExpression:    aws.String("flujo = :f"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":f": &types.AttributeValueMemberS{Value: "a"}},
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	var orden []string
	for _, e := range res.Items {
		orden = append(orden, e["secuencia"].(*types.AttributeValueMemberN).Value)
	}
	quiero := []string{"2", "9", "10"}
	for i := range quiero {
		if i >= len(orden) || orden[i] != quiero[i] {
			t.Fatalf("orden = %v, quiero %v (si ordenas por cadena sale 10, 2, 9)", orden, quiero)
		}
	}
}

// S3 y DynamoDB en el MISMO puerto: se reparten mirando la cabecera X-Amz-Target.
func TestLosDosServiciosConvivenEnElMismoPuerto(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"DynamoDB_": NuevoDynamo(NuevasTablas(), nil),
	}))
	defer srv.Close()

	din := dynamodb.New(dynamodb.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(srv.URL),
	})
	if _, err := din.ListTables(ctx, &dynamodb.ListTablesInput{}); err != nil {
		t.Fatalf("DynamoDB en el puerto compartido: %v", err)
	}
	// Y S3, en el mismo sitio, sigue funcionando.
	s3c := clienteS3(srv.URL)
	if _, err := s3c.ListBuckets(ctx, nil); err != nil {
		t.Fatalf("S3 en el puerto compartido: %v", err)
	}
}
```

### `enrutador.go`

```go
package main

// El enrutador de la raiz: como UN puerto sirve VARIOS servicios de AWS.
//
// Aqui hay una sorpresa que solo se ve montando esto. Los servicios de AWS no comparten
// protocolo:
//
//	S3         el metodo y la RUTA dicen que quieres.   PUT /fotos/gato.png   -> respuesta XML
//	DynamoDB   siempre POST a "/", cuerpo JSON, y la operacion en una CABECERA:
//	           X-Amz-Target: DynamoDB_20120810.PutItem
//	SQS        igual que DynamoDB, con otro prefijo:
//	           X-Amz-Target: AmazonSQS.SendMessage
//
// Son diseños distintos dentro de la misma nube, y el motivo es historico: S3 es de 2006 y se
// penso como almacen de objetos sobre HTTP; DynamoDB es de 2012 y usa el estilo "JSON-RPC" que
// Amazon adopto despues. SQS es MAS viejo que S3 (2006 tambien, con el protocolo Query de toda
// la vida), pero Amazon lo migro a JSON, asi que hoy habla como DynamoDB.
//
// Asi que para servir varios en el mismo puerto no hace falta magia: se mira la cabecera.

import (
	"net/http"
	"strings"
)

// Raiz devuelve el manejador que reparte entre servicios.
//
// `porObjetivo` va del prefijo de `X-Amz-Target` al servicio. Se pasa como mapa y no con un
// `switch` a mano para que añadir el servicio siguiente sea una linea en `main`, no una
// modificacion aqui dentro.
//
// ⚠️ El criterio tiene que ser POSITIVO: se pregunta por lo que identifica a cada servicio, no
// por "lo que no es S3". Un criterio por descarte funciona con dos y se rompe callado con el
// tercero — que es exactamente lo que acaba de entrar.
func Raiz(s3 http.Handler, porObjetivo map[string]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		objetivo := r.Header.Get("X-Amz-Target")
		for prefijo, servicio := range porObjetivo {
			if strings.HasPrefix(objetivo, prefijo) {
				servicio.ServeHTTP(w, r)
				return
			}
		}
		s3.ServeHTTP(w, r)
	})
}
```

### `firma.go`

```go
package main

// 🥇 SigV4: qué significa de verdad FIRMAR una petición.
//
// Casi todo el mundo usa esto todos los días sin haberlo mirado nunca por dentro, porque el SDK
// lo hace solo. Y cuando falla, el mensaje es siempre el mismo y nunca dice por qué:
//
//	SignatureDoesNotMatch
//
// Lo primero, y lo que más cuesta creer: **firmar no es cifrar**. La petición viaja tal cual, en
// claro, con su cuerpo legible. Lo que la firma demuestra son otras dos cosas:
//
//	1. QUIÉN la manda, sin que la contraseña viaje nunca por el cable;
//	2. que NADIE la ha tocado por el camino — ni el cuerpo, ni las cabeceras firmadas, ni la ruta.
//
// ## Cómo funciona, en cuatro pasos
//
//	1. PETICIÓN CANÓNICA   se reescribe la petición en un formato exacto y sin ambigüedades
//	2. CADENA A FIRMAR     la fecha, el ámbito, y el resumen de esa petición canónica
//	3. CLAVE DE FIRMA      una cadena de HMAC: secreto -> fecha -> región -> servicio -> aws4
//	4. FIRMA               HMAC de la cadena a firmar con esa clave
//
// El servidor rehace los cuatro pasos con SU copia del secreto y compara. Si sale lo mismo, la
// petición es auténtica. Si no, `SignatureDoesNotMatch` — y el motivo suele estar en el paso 1.
//
// ⚠️ POR QUÉ EL PASO 1 ES EL QUE DUELE: para que salga la misma firma, el servidor tiene que
// reconstruir la petición EXACTAMENTE igual que el cliente. Una barra de más en la ruta, un
// parámetro de consulta en otro orden, una cabecera con un espacio de más… y la firma cambia
// entera. El error no dice «te sobra una barra»: dice que la firma no coincide.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	algoritmo   = "AWS4-HMAC-SHA256"
	formatoAmz  = "20060102T150405Z"
	formatoDia  = "20060102"
	sinFirmar   = "UNSIGNED-PAYLOAD"
	ventanaVida = 15 * time.Minute // lo mismo que usa AWS: una firma vieja no vale
)

// Verificador comprueba la firma de cada petición. Con `claves` vacío no comprueba nada, que es
// el modo de los episodios 1 a 3: un doble de pruebas no necesita autenticar a nadie.
type Verificador struct {
	claves   map[string]string // identificador de acceso -> secreto
	registro func(string, ...any)
	ahora    func() time.Time // inyectable: sin esto no se puede probar que una firma CADUCA
}

func NuevoVerificador(claves map[string]string, registro func(string, ...any)) *Verificador {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &Verificador{claves: claves, registro: registro, ahora: time.Now}
}

// Envolver mete la comprobación delante de lo que ya había. Si no hay claves configuradas, se
// aparta del camino: devuelve el mismo manejador.
func (v *Verificador) Envolver(siguiente http.Handler) http.Handler {
	if len(v.claves) == 0 {
		return siguiente
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ⚠️ Hay que leer el cuerpo para poder resumirlo... y volver a dejarlo puesto, o el
		// servicio de detrás se encuentra un cuerpo vacío. Es el fallo clásico de escribir un
		// middleware que mira el cuerpo.
		cuerpo, err := io.ReadAll(r.Body)
		if err != nil {
			v.rechazar(w, "IncompleteBody", err.Error())
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(cuerpo)))

		if motivo := v.Comprobar(r, cuerpo); motivo != "" {
			v.registro("  -> 403 SignatureDoesNotMatch (%s)", motivo)
			v.rechazar(w, "SignatureDoesNotMatch",
				"The request signature we calculated does not match the signature you provided. "+
					"nubecita dice por qué: "+motivo)
			return
		}
		siguiente.ServeHTTP(w, r)
	})
}

func (v *Verificador) rechazar(w http.ResponseWriter, codigo, mensaje string) {
	// Se contesta en XML porque el que más se usa aquí es S3. Los servicios de estilo JSON
	// aceptan igualmente el 403; lo que importa es el código.
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusForbidden)
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+
		"<Error><Code>"+codigo+"</Code><Message>"+mensaje+"</Message></Error>")
}

// Comprobar devuelve "" si la firma es buena, y si no, EL MOTIVO. Eso es lo único que esta
// implementación hace mejor que la de verdad: decir por qué.
func (v *Verificador) Comprobar(r *http.Request, cuerpo []byte) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "no viene la cabecera Authorization"
	}
	if !strings.HasPrefix(auth, algoritmo+" ") {
		return "el algoritmo no es " + algoritmo
	}
	cred, firmadas, firma := "", "", ""
	for _, trozo := range strings.Split(strings.TrimPrefix(auth, algoritmo+" "), ",") {
		trozo = strings.TrimSpace(trozo)
		switch {
		case strings.HasPrefix(trozo, "Credential="):
			cred = strings.TrimPrefix(trozo, "Credential=")
		case strings.HasPrefix(trozo, "SignedHeaders="):
			firmadas = strings.TrimPrefix(trozo, "SignedHeaders=")
		case strings.HasPrefix(trozo, "Signature="):
			firma = strings.TrimPrefix(trozo, "Signature=")
		}
	}
	if cred == "" || firmadas == "" || firma == "" {
		return "la cabecera Authorization está incompleta"
	}
	partes := strings.Split(cred, "/")
	if len(partes) != 5 {
		return "la credencial no tiene los cinco tramos: clave/fecha/region/servicio/aws4_request"
	}
	acceso, dia, region, servicio := partes[0], partes[1], partes[2], partes[3]
	secreto, hay := v.claves[acceso]
	if !hay {
		return "no conozco la clave de acceso " + acceso
	}

	// La fecha va en una cabecera propia, y hay que usar ESA, no la del reloj del servidor.
	fechaAmz := r.Header.Get("X-Amz-Date")
	if fechaAmz == "" {
		fechaAmz = r.Header.Get("Date")
	}
	t, err := time.Parse(formatoAmz, fechaAmz)
	if err != nil {
		return "la cabecera X-Amz-Date no se entiende: " + fechaAmz
	}
	// 🥇 La ventana de validez: una firma caduca. Sin esto, quien capture UNA petición firmada
	// puede repetirla para siempre. Es lo que impide el ataque de repetición.
	if d := v.ahora().UTC().Sub(t); d > ventanaVida || d < -ventanaVida {
		return fmt.Sprintf("la firma es de hace %s: caducó (la ventana es de %s)",
			d.Round(time.Second), ventanaVida)
	}
	if t.UTC().Format(formatoDia) != dia {
		return "el día de la credencial no coincide con X-Amz-Date"
	}

	canonica, motivo := peticionCanonica(r, cuerpo, firmadas)
	if motivo != "" {
		return motivo
	}
	ambito := dia + "/" + region + "/" + servicio + "/aws4_request"
	aFirmar := strings.Join([]string{
		algoritmo, fechaAmz, ambito, resumen([]byte(canonica)),
	}, "\n")

	mia := hex.EncodeToString(hmacSHA256(claveDeFirma(secreto, dia, region, servicio),
		[]byte(aFirmar)))
	// ⚠️ hmac.Equal y NO `==`. Comparar cadenas con == sale antes en cuanto encuentra una
	// diferencia, y ese "antes" se puede MEDIR: es un canal por tiempos que deja adivinar la
	// firma byte a byte. Aquí no importa mucho; en un servidor de verdad, sí.
	if !hmac.Equal([]byte(mia), []byte(firma)) {
		return "la firma no coincide (revisa la petición canónica: ruta, consulta o cabeceras)"
	}

	// 🥇 Y falta un paso que casi nadie menciona, y sin el cual todo lo anterior no protege el
	// cuerpo: la firma cubre el RESUMEN del cuerpo, no el cuerpo. Si alguien cambia el cuerpo y
	// deja la cabecera del resumen como estaba, la firma SIGUE CUADRANDO. Hay que comprobar
	// además que el cuerpo hace ese resumen.
	if h := r.Header.Get("X-Amz-Content-Sha256"); h != "" && !especial(h) {
		if resumen(cuerpo) != h {
			return "el cuerpo no coincide con x-amz-content-sha256: alguien lo tocó " +
				"después de firmar"
		}
	}
	return ""
}

// especial dice si el valor de x-amz-content-sha256 no es un resumen, sino una palabra clave.
//
// ⚠️ `UNSIGNED-PAYLOAD` significa literalmente «no firmo el cuerpo». S3 lo admite cuando vas por
// TLS, porque entonces confías en el cifrado del canal para que nadie lo toque. Ahí la firma
// autentica QUIÉN eres, pero NO protege el contenido: eso lo hace el TLS.
func especial(h string) bool {
	return h == sinFirmar || strings.HasPrefix(h, "STREAMING-")
}

// peticionCanonica reescribe la petición en el formato exacto que manda SigV4. Es el paso 1, y
// el que se lleva el 90 % de los dolores de cabeza.
func peticionCanonica(r *http.Request, cuerpo []byte, firmadas string) (string, string) {
	ruta := r.URL.EscapedPath()
	if ruta == "" {
		ruta = "/"
	}

	// La consulta va ORDENADA por nombre y con cada trozo codificado. Si el cliente manda
	// `b=2&a=1`, aquí tiene que quedar `a=1&b=2`, o la firma no sale.
	q := r.URL.Query()
	nombres := make([]string, 0, len(q))
	for k := range q {
		nombres = append(nombres, k)
	}
	sort.Strings(nombres)
	var trozos []string
	for _, k := range nombres {
		vs := append([]string(nil), q[k]...)
		sort.Strings(vs)
		for _, val := range vs {
			trozos = append(trozos, url.QueryEscape(k)+"="+url.QueryEscape(val))
		}
	}
	consulta := strings.Join(trozos, "&")

	// Solo las cabeceras que el cliente DECLARÓ firmadas entran aquí. Por eso un proxy puede
	// añadir cabeceras por el camino sin romper la firma: si no están en la lista, no cuentan.
	lista := strings.Split(firmadas, ";")
	sort.Strings(lista)
	var cab strings.Builder
	for _, nombre := range lista {
		valor := r.Header.Get(nombre)
		if nombre == "host" {
			valor = r.Host // el Host NO está en r.Header: Go lo saca aparte
		}
		if valor == "" && nombre != "host" {
			return "", "la cabecera firmada " + nombre + " no viene en la petición"
		}
		cab.WriteString(strings.ToLower(nombre) + ":" + strings.TrimSpace(valor) + "\n")
	}

	// El resumen del cuerpo. En S3 viene en una cabecera —y puede decir literalmente «sin
	// firmar»—; en los servicios de estilo JSON se calcula.
	hash := r.Header.Get("X-Amz-Content-Sha256")
	if hash == "" {
		hash = resumen(cuerpo)
	}

	return strings.Join([]string{
		r.Method, ruta, consulta, cab.String(), firmadas, hash,
	}, "\n"), ""
}

// claveDeFirma es la cadena de HMAC que hace que el secreto NO viaje nunca, ni siquiera
// derivado de una sola vez: cada eslabón acota para qué sirve la clave resultante.
//
//	AWS4+secreto -> fecha -> región -> servicio -> aws4_request
//
// Consecuencia práctica: una firma de hoy no vale mañana, y una de S3 no vale para DynamoDB.
// Está acotada por diseño.
func claveDeFirma(secreto, dia, region, servicio string) []byte {
	k := hmacSHA256([]byte("AWS4"+secreto), []byte(dia))
	k = hmacSHA256(k, []byte(region))
	k = hmacSHA256(k, []byte(servicio))
	return hmacSHA256(k, []byte("aws4_request"))
}

func hmacSHA256(clave, datos []byte) []byte {
	m := hmac.New(sha256.New, clave)
	m.Write(datos)
	return m.Sum(nil)
}

func resumen(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
```

### `firma_test.go`

```go
package main

// Las pruebas de la firma. Y aquí hay una gracia: quien FIRMA es el SDK oficial de AWS, y quien
// COMPRUEBA es nuestro código. O sea que si estas pruebas pasan, nuestra comprobación entiende
// exactamente lo que produce Amazon — que es la única forma honesta de decirlo.
//
//	go test ./... -count=1 -v -run Firma

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithy "github.com/aws/smithy-go"
)

// nubeFirmada levanta la nubecita CON comprobación de firma, y devuelve además el verificador
// para poder moverle el reloj (y así probar que una firma caduca sin esperar quince minutos).
func nubeFirmada(t *testing.T, acceso, secreto string) (*s3.Client, *Verificador, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	v := NuevoVerificador(map[string]string{acceso: secreto}, nil)
	srv := httptest.NewServer(v.Envolver(Raiz(mux, map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(NuevasTablas(), nil),
		"AmazonSQS.": NuevoSQS(NuevasColas(), nil),
	})))
	t.Cleanup(srv.Close)
	return clienteS3(srv.URL), v, srv
}

func codigoDeError(err error) string {
	var api smithy.APIError
	if errors.As(err, &api) {
		return api.ErrorCode()
	}
	return ""
}

// 🥇 La prueba de fondo: el SDK oficial firma, nosotros comprobamos, y encaja.
func TestFirmaDelSDKOficialEncaja(t *testing.T) {
	c, _, _ := nubeFirmada(t, "prueba", "prueba")
	if _, err := c.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String("fotos")}); err != nil {
		t.Fatalf("el SDK firmó y nuestra comprobación la rechazó: %v", err)
	}
}

// Con el secreto equivocado, la firma sale distinta. Y fíjate en que el secreto NUNCA viaja:
// lo que no coincide es el resultado de una cadena de HMAC.
func TestConElSecretoEquivocadoNoPasa(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "el-de-verdad")
	malo := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "otro-secreto"),
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
	})
	_, err := malo.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err == nil {
		t.Fatal("aceptó una firma hecha con otro secreto")
	}
	if c := codigoDeError(err); c != "SignatureDoesNotMatch" {
		t.Fatalf("el error llegó como %q y esperaba SignatureDoesNotMatch: %v", c, err)
	}
}

// Una clave de acceso que no conocemos se rechaza antes de calcular nada.
func TestClaveDeAccesoDesconocida(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "prueba")
	otro := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("no-te-conozco", "prueba"),
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
	})
	_, err := otro.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err == nil || !strings.Contains(err.Error(), "no conozco la clave de acceso") {
		t.Fatalf("esperaba que dijera que no conoce la clave: %v", err)
	}
}

// 🥇 LA PRUEBA QUE EXPLICA PARA QUÉ SIRVE FIRMAR: alguien en el medio cambia el cuerpo DESPUÉS
// de que el SDK lo firmara. El cuerpo viaja en claro y se puede leer — firmar no es cifrar —
// pero en cuanto se toca, la firma deja de cuadrar.
func TestSiAlguienToquetealCuerpoLaFirmaSeCae(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "prueba")
	entrometido := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "prueba"),
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
		HTTPClient:   &clienteQueToquetea{},
	})
	_, err := entrometido.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
		Body: bytes.NewReader([]byte("transferir 10 euros")),
	})
	if err == nil {
		t.Fatal("aceptó una petición cuyo cuerpo cambió después de firmarse")
	}
	if c := codigoDeError(err); c != "SignatureDoesNotMatch" {
		t.Fatalf("el error llegó como %q: %v", c, err)
	}
}

// Y lo mismo con la ruta: cambiar el destino después de firmar tampoco cuela.
func TestSiAlguienCambiaLaRutaLaFirmaSeCae(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "prueba")
	entrometido := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "prueba"),
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
		HTTPClient:   &clienteQueCambiaLaRuta{},
	})
	_, err := entrometido.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String("mio")})
	if err == nil {
		t.Fatal("aceptó una petición cuya ruta cambió después de firmarse")
	}
}

// 🥇 Una firma CADUCA. Sin esto, quien capture una sola petición firmada podría repetirla para
// siempre. Y se prueba sin esperar: el reloj del verificador es inyectable.
func TestUnaFirmaCaduca(t *testing.T) {
	c, v, _ := nubeFirmada(t, "prueba", "prueba")
	// El servidor cree que han pasado dos horas desde que se firmó.
	v.ahora = func() time.Time { return time.Now().UTC().Add(2 * time.Hour) }

	_, err := c.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err == nil {
		t.Fatal("aceptó una firma de hace dos horas")
	}
	if !strings.Contains(err.Error(), "caducó") {
		t.Fatalf("el error no dice que caducó: %v", err)
	}
}

// Sin cabecera de autorización, ni se mira qué pedías.
func TestSinFirmaNiSeMira(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "prueba")
	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("código = %d, quiero 403", res.StatusCode)
	}
}

// Y sin claves configuradas, el verificador se aparta del camino: es el modo de los episodios
// 1 a 3, y tiene que seguir funcionando igual.
func TestSinClavesElVerificadorNoEstorba(t *testing.T) {
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	v := NuevoVerificador(nil, nil)
	srv := httptest.NewServer(v.Envolver(mux))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("código = %d, quiero 200: el verificador estorba cuando no debería",
			res.StatusCode)
	}
}

// La cadena de HMAC, comprobada contra el ejemplo que AWS publica en su documentación de SigV4.
// Es el único número de todo el proyecto que NO sale de esta máquina: sale de Amazon, y por eso
// sirve para saber si nuestra implementación es la de verdad o solo se parece.
func TestLaClaveDeFirmaCoincideConElEjemploDeAWS(t *testing.T) {
	// Ejemplo oficial: secreto wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY,
	// fecha 20150830, región us-east-1, servicio iam.
	k := claveDeFirma("wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "20150830", "us-east-1", "iam")
	quiero := "c4afb1cc5771d871763a393e44b703571b55cc28424d1a5e86da6ed3c154a4b9"
	if got := hexDe(k); got != quiero {
		t.Fatalf("clave de firma = %s\nquiero          = %s", got, quiero)
	}
}
```

### `main.go`

```go
// nubecita — una nube de mentira, escrita en Go, para probar tu codigo sin pagar ni pedir
// permiso a nadie.
//
// Episodio 1: S3.  ·  Episodio 2: DynamoDB.  ·  Episodio 3: SQS.  ·  Episodio 4: SigV4.
// El SDK de AWS para Go y la CLI de AWS hablan con esto sin saber que no es Amazon. Lo unico
// que cambia en tu codigo es el endpoint.
//
//	go run .                       # escucha en 127.0.0.1:4566
//	go run . -puerto 9000 -callado
//
// El puerto 4566 es el de LocalStack a proposito: si ya tenias algo apuntando ahi, no tocas
// nada.
//
// ⛔ Esto NO es para produccion y no lo va a ser nunca. No hay firmas, no hay permisos, no hay
// disco. Es un doble de pruebas, y un doble de pruebas vale justo por lo que NO hace.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	puerto := flag.Int("puerto", 4566, "puerto en el que escuchar")
	// ⛔ Dentro de un contenedor hay que escuchar en 0.0.0.0: con 127.0.0.1 el proceso solo se
	// oye a si mismo y el `-p` de docker publica un puerto al que no contesta nadie. Es el
	// fallo numero uno al meter un servidor en un contenedor, y no da ningun error.
	escuchar := flag.String("escuchar", "127.0.0.1", "interfaz en la que escuchar")
	callado := flag.Bool("callado", false, "no imprimir cada peticion")
	// ⛔ Apagado por defecto: un doble de pruebas no necesita autenticar a nadie, y pedir
	// firmas por defecto solo haria mas dificil empezar. Con -firmar se comprueban de verdad.
	firmar := flag.Bool("firmar", false, "comprobar la firma SigV4 de cada peticion")
	acceso := flag.String("acceso", "prueba", "identificador de acceso admitido")
	secreto := flag.String("secreto", "prueba", "el secreto de ese identificador")
	flag.Parse()

	registro := log.New(os.Stdout, "", 0).Printf
	if *callado {
		registro = func(string, ...any) {}
	}

	// S3 se enruta por METODO y RUTA; DynamoDB y SQS van siempre a POST "/" con la operacion
	// en una cabecera. `Raiz` mira esa cabecera y reparte. Por eso caben en el mismo puerto, y
	// añadir el siguiente servicio es UNA linea en este mapa.
	dir := fmt.Sprintf("%s:%d", *escuchar, *puerto)
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), registro).Rutas(mux)
	porObjetivo := map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(NuevasTablas(), registro),
		"AmazonSQS.": NuevoSQS(NuevasColas(), registro),
	}

	var claves map[string]string
	if *firmar {
		claves = map[string]string{*acceso: *secreto}
	}
	verificador := NuevoVerificador(claves, registro)

	srv := &http.Server{
		Addr: dir,
		// La comprobacion de firma envuelve TODO, antes del reparto por servicio: una peticion
		// mal firmada no debe llegar ni a saber a que servicio iba.
		Handler:           verificador.Envolver(Raiz(mux, porObjetivo)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Apagado limpio. Un emulador se arranca y se mata cien veces al dia desde un script o
	// desde un Makefile: si no cierras bien, la siguiente corrida se encuentra el puerto
	// ocupado y el fallo aparece en OTRA prueba, que es lo peor que puede pasar.
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	go func() {
		firma := "sin comprobar firmas"
		if *firmar {
			firma = "comprobando firmas SigV4"
		}
		log.Printf("nubecita · S3, DynamoDB y SQS en http://%s · %s · Ctrl+C para parar",
			dir, firma)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("no puedo escuchar en %s: %v", dir, err)
		}
	}()

	<-ctx.Done()
	log.Println("cerrando...")
	cierre, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()
	if err := srv.Shutdown(cierre); err != nil {
		log.Printf("cierre forzado: %v", err)
	}
}
```

### `s3.go`

```go
package main

// El servicio S3. Aqui esta TODO lo que hace falta para que el SDK de AWS y la CLI de AWS
// hablen con esto creyendo que es Amazon.
//
// Lo importante del episodio 1, y lo que casi nadie sabe: **la API de S3 no es un protocolo
// raro, es HTTP a secas**. El bucket es el primer trozo de la ruta, la clave del objeto es el
// resto, y las respuestas son XML. Ni gRPC, ni protobuf, ni una libreria de Amazon: net/http.
//
// Y el enrutado es el de la biblioteca estandar, sin framework. Desde Go 1.22 el ServeMux
// entiende el metodo y los comodines en el patron, que es exactamente lo que hace falta:
//
//	PUT /{bucket}            -> crear bucket
//	PUT /{bucket}/{clave...} -> subir objeto      ({clave...} se traga las barras
//	GET /{bucket}            -> listar             de "fotos/2026/gato.png")

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// El espacio de nombres que S3 pone en todas sus respuestas. Si te lo saltas, algunos clientes
// se lo tragan y otros no: el SDK de Go si, la CLI tambien, pero no todos.
const nsS3 = "http://s3.amazonaws.com/doc/2006-03-01/"

// formatoISO es el que usa S3 en los listados: hasta milisegundos y con Z. Con el RFC3339 de
// Go a secas sale con nanosegundos y algunos analizadores se quejan.
const formatoISO = "2006-01-02T15:04:05.000Z"

type ServicioS3 struct {
	almacen *Almacen
	// registro es opcional: cuando esta puesto, cada peticion se imprime. Es LA razon por la
	// que un emulador propio gana a uno de caja: ves el HTTP crudo que manda el SDK.
	registro func(formato string, args ...any)
}

func NuevoS3(a *Almacen, registro func(string, ...any)) *ServicioS3 {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioS3{almacen: a, registro: registro}
}

func (s *ServicioS3) Rutas(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.listarBuckets)
	mux.HandleFunc("PUT /{bucket}", s.crearBucket)
	mux.HandleFunc("DELETE /{bucket}", s.borrarBucket)
	mux.HandleFunc("HEAD /{bucket}", s.cabezaBucket)
	mux.HandleFunc("GET /{bucket}", s.listarObjetos)
	mux.HandleFunc("PUT /{bucket}/{clave...}", s.ponerObjeto)
	mux.HandleFunc("GET /{bucket}/{clave...}", s.obtenerObjeto)
	mux.HandleFunc("HEAD /{bucket}/{clave...}", s.obtenerObjeto)
	mux.HandleFunc("DELETE /{bucket}/{clave...}", s.borrarObjeto)
}

// ── los errores ──────────────────────────────────────────────────────────
// S3 no devuelve un 404 pelado: devuelve un XML con un Code que el SDK convierte en un tipo de
// error de Go (NoSuchBucket, NoSuchKey...). Si aqui contestas un 404 vacio, el SDK del usuario
// recibe "unexpected response" y no puede distinguir "no existe" de "se rompio la red".

type errorXML struct {
	XMLName  xml.Name `xml:"Error"`
	Code     string   `xml:"Code"`
	Message  string   `xml:"Message"`
	Resource string   `xml:"Resource,omitempty"`
}

func (s *ServicioS3) fallo(w http.ResponseWriter, r *http.Request, estado int, codigo, mensaje string) {
	s.registro("  -> %d %s", estado, codigo)
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(estado)
	_, _ = io.WriteString(w, xml.Header)
	_ = xml.NewEncoder(w).Encode(errorXML{Code: codigo, Message: mensaje, Resource: r.URL.Path})
}

func (s *ServicioS3) responderXML(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(estado)
	_, _ = io.WriteString(w, xml.Header)
	_ = xml.NewEncoder(w).Encode(v)
}

// ── buckets ──────────────────────────────────────────────────────────────

type bucketXML struct {
	Nombre string `xml:"Name"`
	Creado string `xml:"CreationDate"`
}

type listaBucketsXML struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	NS      string   `xml:"xmlns,attr"`
	Owner   struct {
		ID          string `xml:"ID"`
		DisplayName string `xml:"DisplayName"`
	} `xml:"Owner"`
	Buckets struct {
		Bucket []bucketXML `xml:"Bucket"`
	} `xml:"Buckets"`
}

func (s *ServicioS3) listarBuckets(w http.ResponseWriter, r *http.Request) {
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	var res listaBucketsXML
	res.NS = nsS3
	res.Owner.ID = "nubecita"
	res.Owner.DisplayName = "nubecita"
	for _, b := range s.almacen.Buckets() {
		res.Buckets.Bucket = append(res.Buckets.Bucket, bucketXML{
			Nombre: b.Nombre, Creado: b.Creado.Format(formatoISO),
		})
	}
	s.registro("  -> 200 %d bucket(s)", len(res.Buckets.Bucket))
	s.responderXML(w, http.StatusOK, res)
}

func (s *ServicioS3) crearBucket(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("bucket")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	// El cuerpo trae el LocationConstraint cuando la region no es us-east-1. Da igual el
	// contenido, pero hay que LEERLO Y CERRARLO o la conexion no se puede reutilizar.
	_, _ = io.Copy(io.Discard, r.Body)
	if !s.almacen.CrearBucket(nombre) {
		s.fallo(w, r, http.StatusConflict, "BucketAlreadyOwnedByYou",
			"Your previous request to create the named bucket succeeded and you already own it.")
		return
	}
	s.registro("  -> 200 creado")
	w.Header().Set("Location", "/"+nombre)
	w.WriteHeader(http.StatusOK)
}

func (s *ServicioS3) borrarBucket(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("bucket")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	existia, vacio := s.almacen.BorrarBucket(nombre)
	if !existia {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	if !vacio {
		s.fallo(w, r, http.StatusConflict, "BucketNotEmpty",
			"The bucket you tried to delete is not empty")
		return
	}
	s.registro("  -> 204 borrado")
	w.WriteHeader(http.StatusNoContent)
}

func (s *ServicioS3) cabezaBucket(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("bucket")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	if !s.almacen.HayBucket(nombre) {
		// ⚠️ HeadBucket NO puede llevar cuerpo: es un HEAD. El codigo es toda la respuesta.
		s.registro("  -> 404")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.registro("  -> 200")
	w.WriteHeader(http.StatusOK)
}

// ── listado de objetos ───────────────────────────────────────────────────

type contenidoXML struct {
	Clave        string `xml:"Key"`
	Modificado   string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Tam          int    `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

type prefijoComunXML struct {
	Prefijo string `xml:"Prefix"`
}

type listaObjetosXML struct {
	XMLName        xml.Name          `xml:"ListBucketResult"`
	NS             string            `xml:"xmlns,attr"`
	Nombre         string            `xml:"Name"`
	Prefijo        string            `xml:"Prefix"`
	Delimitador    string            `xml:"Delimiter,omitempty"`
	MaxClaves      int               `xml:"MaxKeys"`
	NumClaves      int               `xml:"KeyCount"`
	Truncado       bool              `xml:"IsTruncated"`
	TipoCodificado string            `xml:"EncodingType,omitempty"`
	Contenido      []contenidoXML    `xml:"Contents"`
	PrefijosComun  []prefijoComunXML `xml:"CommonPrefixes"`
}

func (s *ServicioS3) listarObjetos(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	q := r.URL.Query()
	s.registro("%s %s", r.Method, r.URL.RequestURI())

	prefijo, delim := q.Get("prefix"), q.Get("delimiter")
	maxClaves := 1000
	if v := q.Get("max-keys"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxClaves = n
		}
	}
	objetos, comunes, truncado, hay := s.almacen.Listar(bucket, prefijo, delim, maxClaves)
	if !hay {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}

	// encoding-type=url: la CLI SIEMPRE lo pide. Si lo ignoras, una clave con un espacio o un
	// acento vuelve mal y el cliente muestra basura. Es un detalle de dos lineas que separa
	// "casi funciona" de "funciona".
	codificar := strings.EqualFold(q.Get("encoding-type"), "url")
	esc := func(v string) string {
		if codificar {
			return url.QueryEscape(v)
		}
		return v
	}

	res := listaObjetosXML{
		NS: nsS3, Nombre: bucket, Prefijo: esc(prefijo), Delimitador: esc(delim),
		MaxClaves: maxClaves, Truncado: truncado,
	}
	if codificar {
		res.TipoCodificado = "url"
	}
	for _, o := range objetos {
		res.Contenido = append(res.Contenido, contenidoXML{
			Clave: esc(o.Clave), Modificado: o.Modificado.Format(formatoISO),
			ETag: o.ETag, Tam: o.Tam, StorageClass: "STANDARD",
		})
	}
	for _, p := range comunes {
		res.PrefijosComun = append(res.PrefijosComun, prefijoComunXML{Prefijo: esc(p)})
	}
	res.NumClaves = len(res.Contenido) + len(res.PrefijosComun)
	s.registro("  -> 200 %d objeto(s), %d prefijo(s)", len(res.Contenido), len(res.PrefijosComun))
	s.responderXML(w, http.StatusOK, res)
}

// ── objetos ──────────────────────────────────────────────────────────────

func (s *ServicioS3) ponerObjeto(w http.ResponseWriter, r *http.Request) {
	bucket, clave := r.PathValue("bucket"), r.PathValue("clave")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	datos, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, r, http.StatusBadRequest, "IncompleteBody", err.Error())
		return
	}
	tipo := r.Header.Get("Content-Type")
	if tipo == "" {
		tipo = "binary/octet-stream"
	}
	etag, hay := s.almacen.Poner(bucket, clave, datos, tipo)
	if !hay {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	s.registro("  -> 200 %d byte(s) etag=%s", len(datos), etag)
	w.Header().Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
}

func (s *ServicioS3) obtenerObjeto(w http.ResponseWriter, r *http.Request) {
	bucket, clave := r.PathValue("bucket"), r.PathValue("clave")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	o, hayBucket, hayObjeto := s.almacen.Obtener(bucket, clave)
	if !hayBucket {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	if !hayObjeto {
		s.fallo(w, r, http.StatusNotFound, "NoSuchKey", "The specified key does not exist.")
		return
	}
	w.Header().Set("Content-Type", o.Tipo)
	w.Header().Set("Content-Length", strconv.Itoa(len(o.Datos)))
	w.Header().Set("ETag", o.ETag)
	w.Header().Set("Last-Modified", o.Modificado.Format(http.TimeFormat))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusOK)
	// En un HEAD, net/http tira el cuerpo el solo. No hace falta ramificar.
	s.registro("  -> 200 %d byte(s)", len(o.Datos))
	_, _ = w.Write(o.Datos)
}

func (s *ServicioS3) borrarObjeto(w http.ResponseWriter, r *http.Request) {
	bucket, clave := r.PathValue("bucket"), r.PathValue("clave")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	if !s.almacen.Borrar(bucket, clave) {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	// ⚠️ S3 devuelve 204 aunque la clave NO existiera. Borrar es idempotente a proposito: si
	// devolvieras 404, un reintento de red fallaria despues de haber borrado bien.
	s.registro("  -> 204")
	w.WriteHeader(http.StatusNoContent)
}
```

### `s3_test.go`

```go
package main

// Estas pruebas NO hablan con nuestro servidor a mano: usan el SDK de AWS para Go de verdad,
// el mismo `github.com/aws/aws-sdk-go-v2/service/s3` que usarias contra Amazon. Es la unica
// forma honesta de decir "esto es compatible": si el SDK lo entiende, es compatible.
//
//	go test ./... -race -v

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// nube levanta una nubecita en un puerto libre y devuelve un cliente de S3 apuntado a ella.
//
// Las DOS lineas que hacen que el SDK hable con nosotros y no con Amazon:
//
//	o.BaseEndpoint = aws.String(srv.URL)   // ve aqui en vez de a s3.amazonaws.com
//	o.UsePathStyle = true                  // /bucket/clave, no bucket.s3.amazonaws.com/clave
//
// Lo segundo no es un capricho: el estilo por defecto mete el bucket en el NOMBRE DEL HOST, y
// contra localhost eso no resuelve. Es el error numero uno al apuntar el SDK a un emulador.
func nube(t *testing.T) *s3.Client {
	t.Helper()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return clienteS3(srv.URL)
}

// clienteS3 esta separado porque lo reutiliza la prueba de que S3 y DynamoDB conviven en el
// mismo puerto (`dynamo_test.go`).
func clienteS3(url string) *s3.Client {
	return s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(url),
		UsePathStyle: true,
	})
}

func TestCicloCompleto(t *testing.T) {
	ctx := context.Background()
	c := nube(t)

	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	cuerpo := []byte("hola nubecita")
	sub, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
		Body: bytes.NewReader(cuerpo), ContentType: aws.String("text/plain"),
	})
	if err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	// El ETag tiene que ser el md5 del contenido, entre comillas, igual que en S3 de verdad.
	suma := md5.Sum(cuerpo)
	quiero := `"` + hex.EncodeToString(suma[:]) + `"`
	if aws.ToString(sub.ETag) != quiero {
		t.Errorf("ETag = %q, quiero %q", aws.ToString(sub.ETag), quiero)
	}

	baja, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
	})
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	defer baja.Body.Close()
	leido, _ := io.ReadAll(baja.Body)
	if !bytes.Equal(leido, cuerpo) {
		t.Errorf("bajo %q, subi %q", leido, cuerpo)
	}
	if aws.ToString(baja.ContentType) != "text/plain" {
		t.Errorf("ContentType = %q", aws.ToString(baja.ContentType))
	}
	if aws.ToInt64(baja.ContentLength) != int64(len(cuerpo)) {
		t.Errorf("ContentLength = %d, quiero %d", aws.ToInt64(baja.ContentLength), len(cuerpo))
	}
}

// El error tiene que llegar al SDK como el TIPO correcto, no como un 404 pelado. Si esto
// falla, el codigo del usuario no puede distinguir "no existe" de "se cayo la red".
func TestClaveQueNoExisteLlegaComoNoSuchKey(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("vacio")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	_, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("vacio"), Key: aws.String("no-esta"),
	})
	if err == nil {
		t.Fatal("esperaba un error y no hubo ninguno")
	}
	var noEsta *types.NoSuchKey
	if !errors.As(err, &noEsta) {
		t.Fatalf("el SDK no lo vio como NoSuchKey: %v", err)
	}
}

func TestBucketQueNoExisteLlegaComoNoSuchBucket(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	_, err := c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String("fantasma")})
	var noEsta *types.NoSuchBucket
	if !errors.As(err, &noEsta) {
		t.Fatalf("el SDK no lo vio como NoSuchBucket: %v", err)
	}
}

// El delimitador es lo que hace que S3 PAREZCA que tiene carpetas. No las tiene.
func TestListarConDelimitadorAgrupaComoSiHubieraCarpetas(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("archivo")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	for _, k := range []string{"2025/enero.txt", "2025/febrero.txt", "2026/enero.txt", "raiz.txt"} {
		if _, err := c.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String("archivo"), Key: aws.String(k), Body: bytes.NewReader([]byte("x")),
		}); err != nil {
			t.Fatalf("PutObject %s: %v", k, err)
		}
	}

	res, err := c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("archivo"), Delimiter: aws.String("/"),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2: %v", err)
	}
	// Cuatro objetos, pero con delimitador se ven DOS "carpetas" y UN fichero suelto.
	if len(res.CommonPrefixes) != 2 {
		t.Errorf("prefijos comunes = %d, quiero 2", len(res.CommonPrefixes))
	}
	if len(res.Contents) != 1 {
		t.Errorf("objetos en la raiz = %d, quiero 1", len(res.Contents))
	}
	if aws.ToInt32(res.KeyCount) != 3 {
		t.Errorf("KeyCount = %d, quiero 3", aws.ToInt32(res.KeyCount))
	}

	// Y sin delimitador, los cuatro planos: no hay carpetas, hay claves con barras dentro.
	todo, err := c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String("archivo")})
	if err != nil {
		t.Fatalf("ListObjectsV2 sin delimitador: %v", err)
	}
	if len(todo.Contents) != 4 {
		t.Errorf("sin delimitador = %d objetos, quiero 4", len(todo.Contents))
	}
}

func TestNoSeBorraUnBucketConCosasDentro(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("lleno")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	if _, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("lleno"), Key: aws.String("algo"), Body: bytes.NewReader([]byte("x")),
	}); err != nil {
		t.Fatalf("PutObject: %v", err)
	}
	if _, err := c.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String("lleno")}); err == nil {
		t.Fatal("borro un bucket con cosas dentro, y S3 no deja")
	}

	// Se vacia y ahora si.
	if _, err := c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String("lleno"), Key: aws.String("algo"),
	}); err != nil {
		t.Fatalf("DeleteObject: %v", err)
	}
	if _, err := c.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String("lleno")}); err != nil {
		t.Fatalf("DeleteBucket: %v", err)
	}
}

// Borrar en S3 es idempotente: borrar algo que ya no esta NO es un error. Si lo fuera, un
// reintento de red despues de un borrado bueno fallaria.
func TestBorrarDosVecesNoEsUnError(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("b")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	for i := range 2 {
		if _, err := c.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String("b"), Key: aws.String("nunca-existio"),
		}); err != nil {
			t.Fatalf("DeleteObject numero %d: %v", i+1, err)
		}
	}
}

func TestListarBuckets(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	for _, n := range []string{"zeta", "alfa"} {
		if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(n)}); err != nil {
			t.Fatalf("CreateBucket %s: %v", n, err)
		}
	}
	res, err := c.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if len(res.Buckets) != 2 {
		t.Fatalf("buckets = %d, quiero 2", len(res.Buckets))
	}
	// Ordenados: si no ordenas, el mapa de Go los devuelve en orden distinto cada vez.
	if aws.ToString(res.Buckets[0].Name) != "alfa" {
		t.Errorf("el primero es %q, quiero alfa", aws.ToString(res.Buckets[0].Name))
	}
}
```

### `sqs.go`

```go
package main

// El servicio SQS. Habla el mismo protocolo que DynamoDB —POST a "/", JSON, y la operacion en
// la cabecera— pero con otro prefijo:
//
//	X-Amz-Target: AmazonSQS.SendMessage
//
// ⚠️ Y aqui hay un dato que sorprende: SQS es MAS viejo que DynamoDB y durante años hablo el
// protocolo "Query" de AWS, con los parametros en el cuerpo tipo formulario. Amazon lo migro a
// JSON, y por eso hoy se parece a DynamoDB y no a S3. O sea: el protocolo de un servicio no es
// para siempre.
//
// Lo que se implementa, que es lo que se usa un dia normal:
//
//	CreateQueue · ListQueues · GetQueueUrl · DeleteQueue · PurgeQueue · GetQueueAttributes
//	SendMessage · SendMessageBatch
//	ReceiveMessage · DeleteMessage · DeleteMessageBatch · ChangeMessageVisibility
//
// ⛔ Lo que NO: colas FIFO, DelaySeconds, atributos de mensaje, sondeo largo de verdad
// (WaitTimeSeconds se acepta pero no bloquea) ni politicas de acceso.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var errNoExisteCola = errors.New("la cola no existe")

const tipoSQS = "com.amazonaws.sqs#"

type ServicioSQS struct {
	colas    *Colas
	registro func(formato string, args ...any)
}

func NuevoSQS(c *Colas, registro func(string, ...any)) *ServicioSQS {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioSQS{colas: c, registro: registro}
}

func (s *ServicioSQS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	destino := r.Header.Get("X-Amz-Target")
	op := destino
	if i := strings.LastIndex(destino, "."); i >= 0 {
		op = destino[i+1:]
	}
	cuerpo, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	s.registro("POST / [SQS %s] %d byte(s)", op, len(cuerpo))

	// ⚠️ La URL de la cola se arma con el Host de LA PETICION, no con una direccion fija de
	// arranque. Dentro de un contenedor el servidor escucha en 0.0.0.0, y devolver esa
	// direccion daria una URL que no le sirve a nadie. Asi, el cliente recibe la misma
	// direccion por la que entro.
	base := "http://" + r.Host

	switch op {
	case "CreateQueue":
		s.crearCola(w, cuerpo, base)
	case "ListQueues":
		s.listarColas(w, cuerpo, base)
	case "GetQueueUrl":
		s.urlDeCola(w, cuerpo, base)
	case "DeleteQueue":
		s.borrarCola(w, cuerpo)
	case "PurgeQueue":
		s.vaciarCola(w, cuerpo)
	case "GetQueueAttributes":
		s.atributos(w, cuerpo)
	case "SendMessage":
		s.enviar(w, cuerpo)
	case "SendMessageBatch":
		s.enviarLote(w, cuerpo)
	case "ReceiveMessage":
		s.recibir(w, cuerpo)
	case "DeleteMessage":
		s.borrarMensaje(w, cuerpo)
	case "DeleteMessageBatch":
		s.borrarLote(w, cuerpo)
	case "ChangeMessageVisibility":
		s.cambiarVisibilidad(w, cuerpo)
	default:
		s.fallo(w, "UnsupportedOperation",
			fmt.Sprintf("nubecita no implementa %q (episodio 3: S3, DynamoDB y SQS)", op))
	}
}

func (s *ServicioSQS) responder(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *ServicioSQS) fallo(w http.ResponseWriter, tipo, mensaje string) {
	s.registro("  -> 400 %s", tipo)
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"__type": tipoSQS + tipo, "message": mensaje,
	})
}

// urlDeLaCola arma la URL con el numero de cuenta de mentira que usa todo el mundo en local.
func urlDeLaCola(base, nombre string) string { return base + "/000000000000/" + nombre }

// nombreDeURL saca el nombre del final de la URL de la cola. En SQS la cola se identifica por su
// URL entera, no por su nombre: el nombre es solo el ultimo tramo.
func nombreDeURL(u string) string {
	if i := strings.LastIndex(u, "/"); i >= 0 {
		return u[i+1:]
	}
	return u
}

// ── colas ────────────────────────────────────────────────────────────────

type peticionCrearCola struct {
	QueueName  string            `json:"QueueName"`
	Attributes map[string]string `json:"Attributes"`
}

// redrive es el `RedrivePolicy`, que en SQS viaja como una CADENA con JSON dentro. Sí: JSON
// dentro de un campo de texto de otro JSON. No es un error de lectura.
type redrive struct {
	MaxReceiveCount     any    `json:"maxReceiveCount"`
	DeadLetterTargetArn string `json:"deadLetterTargetArn"`
}

func (s *ServicioSQS) crearCola(w http.ResponseWriter, cuerpo []byte, base string) {
	var p peticionCrearCola
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	if p.QueueName == "" {
		s.fallo(w, "InvalidParameterValue", "falta QueueName")
		return
	}
	var vis *time.Duration
	if v, hay := p.Attributes["VisibilityTimeout"]; hay {
		if n, err := strconv.Atoi(v); err == nil {
			d := time.Duration(n) * time.Second
			vis = &d
		}
	}
	maxRecibos, colaMuertos := 0, ""
	if v, hay := p.Attributes["RedrivePolicy"]; hay && v != "" {
		var rd redrive
		if err := json.Unmarshal([]byte(v), &rd); err != nil {
			s.fallo(w, "InvalidParameterValue", "RedrivePolicy no es JSON valido: "+err.Error())
			return
		}
		// ⚠️ maxReceiveCount llega unas veces como numero y otras como cadena, segun quien
		// arme la peticion. Aceptar solo uno de los dos es un fallo que solo aparece con el
		// cliente que no probaste.
		switch t := rd.MaxReceiveCount.(type) {
		case float64:
			maxRecibos = int(t)
		case string:
			maxRecibos, _ = strconv.Atoi(t)
		}
		colaMuertos = nombreDeURL(rd.DeadLetterTargetArn)
		if i := strings.LastIndex(colaMuertos, ":"); i >= 0 {
			colaMuertos = colaMuertos[i+1:]
		}
	}
	s.colas.Crear(p.QueueName, vis, maxRecibos, colaMuertos)
	s.registro("  -> 200 cola %q creada", p.QueueName)
	s.responder(w, map[string]any{"QueueUrl": urlDeLaCola(base, p.QueueName)})
}

func (s *ServicioSQS) listarColas(w http.ResponseWriter, cuerpo []byte, base string) {
	var p struct {
		QueueNamePrefix string `json:"QueueNamePrefix"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	nombres := s.colas.Listar(p.QueueNamePrefix)
	urls := make([]string, 0, len(nombres))
	for _, n := range nombres {
		urls = append(urls, urlDeLaCola(base, n))
	}
	s.registro("  -> 200 %d cola(s)", len(urls))
	s.responder(w, map[string]any{"QueueUrls": urls})
}

func (s *ServicioSQS) urlDeCola(w http.ResponseWriter, cuerpo []byte, base string) {
	var p struct {
		QueueName string `json:"QueueName"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	if _, ok := s.colas.Buscar(p.QueueName); !ok {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	s.responder(w, map[string]any{"QueueUrl": urlDeLaCola(base, p.QueueName)})
}

func (s *ServicioSQS) borrarCola(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	if !s.colas.Borrar(nombreDeURL(p.QueueUrl)) {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	s.registro("  -> 200 cola borrada")
	s.responder(w, map[string]any{})
}

func (s *ServicioSQS) vaciarCola(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	if !s.colas.Vaciar(nombreDeURL(p.QueueUrl)) {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	s.responder(w, map[string]any{})
}

func (s *ServicioSQS) atributos(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	nombre := nombreDeURL(p.QueueUrl)
	visibles, enVuelo, ok := s.colas.Cuentas(nombre)
	if !ok {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	c, _ := s.colas.Buscar(nombre)
	attr := map[string]string{
		"ApproximateNumberOfMessages":           strconv.Itoa(visibles),
		"ApproximateNumberOfMessagesNotVisible": strconv.Itoa(enVuelo),
		"VisibilityTimeout":                     strconv.Itoa(int(c.Visibilidad.Seconds())),
		"QueueArn":                              "arn:aws:sqs:us-east-1:000000000000:" + nombre,
	}
	s.registro("  -> 200 %d visible(s), %d en vuelo", visibles, enVuelo)
	s.responder(w, map[string]any{"Attributes": attr})
}

// ── mensajes ─────────────────────────────────────────────────────────────

func (s *ServicioSQS) enviar(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl    string `json:"QueueUrl"`
		MessageBody string `json:"MessageBody"`
	}
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	m, err := s.colas.Enviar(nombreDeURL(p.QueueUrl), p.MessageBody)
	if err != nil {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	s.registro("  -> 200 enviado %s", m.ID)
	// ⚠️ El MD5 no es decorativo: el SDK lo COMPRUEBA. Si lo calculas mal, el cliente da un
	// error de checksum y parece que se corrompio la red.
	s.responder(w, map[string]any{"MessageId": m.ID, "MD5OfMessageBody": m.MD5})
}

func (s *ServicioSQS) enviarLote(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
		Entries  []struct {
			Id          string `json:"Id"`
			MessageBody string `json:"MessageBody"`
		} `json:"Entries"`
	}
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	nombre := nombreDeURL(p.QueueUrl)
	var bien []map[string]any
	var mal []map[string]any
	for _, e := range p.Entries {
		m, err := s.colas.Enviar(nombre, e.MessageBody)
		if err != nil {
			mal = append(mal, map[string]any{"Id": e.Id, "Code": "QueueDoesNotExist",
				"SenderFault": false, "Message": err.Error()})
			continue
		}
		bien = append(bien, map[string]any{"Id": e.Id, "MessageId": m.ID,
			"MD5OfMessageBody": m.MD5})
	}
	s.registro("  -> 200 lote: %d bien, %d mal", len(bien), len(mal))
	res := map[string]any{"Successful": bien}
	if mal != nil {
		res["Failed"] = mal
	}
	s.responder(w, res)
}

func (s *ServicioSQS) recibir(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl            string `json:"QueueUrl"`
		MaxNumberOfMessages int    `json:"MaxNumberOfMessages"`
		VisibilityTimeout   *int   `json:"VisibilityTimeout"`
	}
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	if p.MaxNumberOfMessages <= 0 {
		p.MaxNumberOfMessages = 1
	}
	var vis *time.Duration
	if p.VisibilityTimeout != nil {
		d := time.Duration(*p.VisibilityTimeout) * time.Second
		vis = &d
	}
	msgs, err := s.colas.Recibir(nombreDeURL(p.QueueUrl), p.MaxNumberOfMessages, vis)
	if err != nil {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	salida := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		salida = append(salida, map[string]any{
			"MessageId":     m.ID,
			"ReceiptHandle": m.Recibo,
			"Body":          m.Cuerpo,
			"MD5OfBody":     m.MD5,
			"Attributes": map[string]string{
				"ApproximateReceiveCount": strconv.Itoa(m.Recibido),
				"SentTimestamp":           strconv.FormatInt(m.Enviado.UnixMilli(), 10),
			},
		})
	}
	s.registro("  -> 200 %d mensaje(s) entregados", len(salida))
	// ⚠️ Recibir cero mensajes NO es un error: es una cola vacia. Si devolvieras error, el
	// consumidor no distinguiria "no hay trabajo" de "se rompio la cola".
	res := map[string]any{}
	if len(salida) > 0 {
		res["Messages"] = salida
	}
	s.responder(w, res)
}

func (s *ServicioSQS) borrarMensaje(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl      string `json:"QueueUrl"`
		ReceiptHandle string `json:"ReceiptHandle"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	borrado, err := s.colas.Quitar(nombreDeURL(p.QueueUrl), p.ReceiptHandle)
	if err != nil {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	// Borrar con un recibo caducado NO es un error en SQS. Se registra, porque casi siempre
	// significa que el consumidor tardo mas que la visibilidad y otro ya se lo llevo.
	s.registro("  -> 200 borrado=%v", borrado)
	s.responder(w, map[string]any{})
}

func (s *ServicioSQS) borrarLote(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
		Entries  []struct {
			Id            string `json:"Id"`
			ReceiptHandle string `json:"ReceiptHandle"`
		} `json:"Entries"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	nombre := nombreDeURL(p.QueueUrl)
	var bien []map[string]any
	for _, e := range p.Entries {
		if _, err := s.colas.Quitar(nombre, e.ReceiptHandle); err != nil {
			s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
			return
		}
		bien = append(bien, map[string]any{"Id": e.Id})
	}
	s.responder(w, map[string]any{"Successful": bien})
}

func (s *ServicioSQS) cambiarVisibilidad(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl          string `json:"QueueUrl"`
		ReceiptHandle     string `json:"ReceiptHandle"`
		VisibilityTimeout int    `json:"VisibilityTimeout"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	ok, err := s.colas.CambiarVisibilidad(nombreDeURL(p.QueueUrl), p.ReceiptHandle,
		time.Duration(p.VisibilityTimeout)*time.Second)
	if err != nil {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	if !ok {
		s.fallo(w, "ReceiptHandleIsInvalid", "The receipt handle is not valid.")
		return
	}
	s.registro("  -> 200 visibilidad a %d s", p.VisibilityTimeout)
	s.responder(w, map[string]any{})
}
```

### `sqs_test.go`

```go
package main

// Pruebas de SQS contra `github.com/aws/aws-sdk-go-v2/service/sqs`, el cliente oficial.
//
// ⚠️ Ninguna prueba DUERME. El reloj de las colas es inyectable, así que "que pasen treinta
// segundos" es una asignación. Una prueba que duerme para comprobar un tiempo de espera es
// lenta y, cuando la máquina va cargada, además es inestable.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// relojFalso permite mover el tiempo a mano.
type relojFalso struct{ ahora time.Time }

func (r *relojFalso) leer() time.Time         { return r.ahora }
func (r *relojFalso) avanzar(d time.Duration) { r.ahora = r.ahora.Add(d) }

func nubeSQS(t *testing.T) (*sqs.Client, *Colas, *relojFalso) {
	t.Helper()
	reloj := &relojFalso{ahora: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)}
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	colas := NuevasColas()
	colas.reloj = reloj.leer
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"AmazonSQS.": NuevoSQS(colas, nil),
	}))
	t.Cleanup(srv.Close)

	c := sqs.New(sqs.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(srv.URL),
	})
	return c, colas, reloj
}

func crearCola(t *testing.T, c *sqs.Client, nombre string, attr map[string]string) string {
	t.Helper()
	res, err := c.CreateQueue(context.Background(), &sqs.CreateQueueInput{
		QueueName: aws.String(nombre), Attributes: attr,
	})
	if err != nil {
		t.Fatalf("CreateQueue %s: %v", nombre, err)
	}
	return aws.ToString(res.QueueUrl)
}

func TestColaCicloCompleto(t *testing.T) {
	ctx := context.Background()
	c, _, _ := nubeSQS(t)
	url := crearCola(t, c, "pedidos", nil)

	lista, err := c.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		t.Fatalf("ListQueues: %v", err)
	}
	if len(lista.QueueUrls) != 1 {
		t.Fatalf("ListQueues = %v", lista.QueueUrls)
	}

	env, err := c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("procesa el pedido 7"),
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if aws.ToString(env.MessageId) == "" {
		t.Error("no devolvio MessageId")
	}

	rec, err := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if err != nil {
		t.Fatalf("ReceiveMessage: %v", err)
	}
	if len(rec.Messages) != 1 {
		t.Fatalf("recibi %d mensajes, quiero 1", len(rec.Messages))
	}
	if aws.ToString(rec.Messages[0].Body) != "procesa el pedido 7" {
		t.Errorf("cuerpo = %q", aws.ToString(rec.Messages[0].Body))
	}

	if _, err := c.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl: aws.String(url), ReceiptHandle: rec.Messages[0].ReceiptHandle,
	}); err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
	if _, err := c.DeleteQueue(ctx, &sqs.DeleteQueueInput{QueueUrl: aws.String(url)}); err != nil {
		t.Fatalf("DeleteQueue: %v", err)
	}
}

// 🥇 LA prueba del episodio: recibir NO saca el mensaje de la cola. Lo esconde.
func TestRecibirNoSacaElMensajeLoEsconde(t *testing.T) {
	ctx := context.Background()
	c, _, reloj := nubeSQS(t)
	url := crearCola(t, c, "trabajos", map[string]string{"VisibilityTimeout": "30"})
	if _, err := c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("uno"),
	}); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	// Primera recepción: llega.
	primera, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(primera.Messages) != 1 {
		t.Fatalf("la primera recepcion dio %d mensajes", len(primera.Messages))
	}

	// Inmediatamente después: NO llega. Está invisible, pero sigue en la cola.
	vacia, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(vacia.Messages) != 0 {
		t.Fatalf("lo entrego dos veces seguidas: la visibilidad no funciona")
	}

	// Pasan 31 segundos sin borrarlo: vuelve. Nadie lo perdio.
	reloj.avanzar(31 * time.Second)
	otra, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(otra.Messages) != 1 {
		t.Fatalf("no volvio despues de la visibilidad: %d mensajes", len(otra.Messages))
	}
	// Y el contador de entregas va por dos: es lo que delata un consumidor que no borra.
	if otra.Messages[0].Attributes["ApproximateReceiveCount"] != "2" {
		t.Errorf("ApproximateReceiveCount = %q, quiero 2",
			otra.Messages[0].Attributes["ApproximateReceiveCount"])
	}
}

// El recibo cambia en CADA entrega, y el viejo deja de valer. Es el fallo clásico del
// consumidor lento: borra con un recibo caducado, cree que borró, y el mensaje sigue ahí.
func TestElReciboCambiaEnCadaEntrega(t *testing.T) {
	ctx := context.Background()
	c, _, reloj := nubeSQS(t)
	url := crearCola(t, c, "trabajos", map[string]string{"VisibilityTimeout": "10"})
	_, _ = c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("uno"),
	})

	primera, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	reciboViejo := primera.Messages[0].ReceiptHandle

	reloj.avanzar(11 * time.Second)
	segunda, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if aws.ToString(segunda.Messages[0].ReceiptHandle) == aws.ToString(reciboViejo) {
		t.Fatal("el recibo no cambio entre entregas")
	}

	// Borrar con el recibo viejo: SQS no falla, pero NO borra.
	if _, err := c.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl: aws.String(url), ReceiptHandle: reciboViejo,
	}); err != nil {
		t.Fatalf("borrar con recibo caducado no deberia dar error: %v", err)
	}
	reloj.avanzar(11 * time.Second)
	sigue, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(sigue.Messages) != 1 {
		t.Fatal("el recibo caducado SI borro, y no debia")
	}
}

// 🥇 La cola de mensajes muertos: a la N-ésima entrega, en vez de darlo otra vez, se aparta.
func TestLaColaDeMuertosSeLlevaElQueSiempreFalla(t *testing.T) {
	ctx := context.Background()
	c, _, reloj := nubeSQS(t)
	crearCola(t, c, "muertos", nil)
	url := crearCola(t, c, "trabajos", map[string]string{
		"VisibilityTimeout": "5",
		"RedrivePolicy": `{"deadLetterTargetArn":"arn:aws:sqs:us-east-1:000000000000:muertos",` +
			`"maxReceiveCount":2}`,
	})
	_, _ = c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("el que siempre revienta"),
	})

	// Dos entregas: el consumidor lo coge y no lo borra (porque falla).
	for i := 1; i <= 2; i++ {
		r, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
		if len(r.Messages) != 1 {
			t.Fatalf("entrega %d: recibi %d mensajes", i, len(r.Messages))
		}
		reloj.avanzar(6 * time.Second)
	}
	// La tercera vez ya no se entrega: se aparta.
	tercera, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(tercera.Messages) != 0 {
		t.Fatalf("lo entrego una tercera vez: el redrive no funciona")
	}

	// Y está en la cola de muertos, entero.
	muertosURL, err := c.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String("muertos")})
	if err != nil {
		t.Fatalf("GetQueueUrl: %v", err)
	}
	enMuertos, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: muertosURL.QueueUrl})
	if len(enMuertos.Messages) != 1 {
		t.Fatalf("la cola de muertos tiene %d mensajes, quiero 1", len(enMuertos.Messages))
	}
	if aws.ToString(enMuertos.Messages[0].Body) != "el que siempre revienta" {
		t.Errorf("llego con otro cuerpo: %q", aws.ToString(enMuertos.Messages[0].Body))
	}
}

// Pedir a una cola vacía NO es un error: es que no hay trabajo.
func TestColaVaciaNoEsUnError(t *testing.T) {
	ctx := context.Background()
	c, _, _ := nubeSQS(t)
	url := crearCola(t, c, "vacia", nil)
	res, err := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if err != nil {
		t.Fatalf("recibir de una cola vacia dio error: %v", err)
	}
	if len(res.Messages) != 0 {
		t.Errorf("devolvio %d mensajes de una cola vacia", len(res.Messages))
	}
}

func TestColaQueNoExisteLlegaComoQueueDoesNotExist(t *testing.T) {
	c, _, _ := nubeSQS(t)
	_, err := c.GetQueueUrl(context.Background(), &sqs.GetQueueUrlInput{
		QueueName: aws.String("fantasma"),
	})
	var noEsta *types.QueueDoesNotExist
	if !errors.As(err, &noEsta) {
		t.Fatalf("el SDK no lo vio como QueueDoesNotExist: %v", err)
	}
}

// Los dos números que SQS llama "aproximados", y por qué lo son: dependen del instante.
func TestLosNumerosAproximadosDeLaCola(t *testing.T) {
	ctx := context.Background()
	c, _, reloj := nubeSQS(t)
	url := crearCola(t, c, "cuentas", map[string]string{"VisibilityTimeout": "20"})
	for _, cuerpo := range []string{"a", "b", "c"} {
		_, _ = c.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl: aws.String(url), MessageBody: aws.String(cuerpo),
		})
	}
	_, _ = c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl: aws.String(url), MaxNumberOfMessages: 2,
	})

	res, err := c.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(url),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		t.Fatalf("GetQueueAttributes: %v", err)
	}
	if res.Attributes["ApproximateNumberOfMessages"] != "1" {
		t.Errorf("visibles = %q, quiero 1", res.Attributes["ApproximateNumberOfMessages"])
	}
	if res.Attributes["ApproximateNumberOfMessagesNotVisible"] != "2" {
		t.Errorf("en vuelo = %q, quiero 2",
			res.Attributes["ApproximateNumberOfMessagesNotVisible"])
	}
	// Y en cuanto pasa la visibilidad, los mismos tres vuelven a estar visibles. De ahi el
	// "aproximado": la respuesta depende del instante en que preguntas.
	reloj.avanzar(21 * time.Second)
	res, _ = c.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(url),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if res.Attributes["ApproximateNumberOfMessages"] != "3" {
		t.Errorf("tras la visibilidad, visibles = %q, quiero 3",
			res.Attributes["ApproximateNumberOfMessages"])
	}
}

// Un consumidor que va a tardar más pide prórroga; y con 0 lo devuelve a la cola al instante.
func TestCambiarLaVisibilidad(t *testing.T) {
	ctx := context.Background()
	c, _, _ := nubeSQS(t)
	url := crearCola(t, c, "trabajos", map[string]string{"VisibilityTimeout": "300"})
	_, _ = c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("uno"),
	})
	r, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})

	// "No puedo con esto, que lo coja otro YA."
	if _, err := c.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl: aws.String(url), ReceiptHandle: r.Messages[0].ReceiptHandle,
		VisibilityTimeout: 0,
	}); err != nil {
		t.Fatalf("ChangeMessageVisibility: %v", err)
	}
	otra, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(otra.Messages) != 1 {
		t.Fatal("con visibilidad 0 el mensaje deberia volver de inmediato")
	}
}

func TestEnviarYBorrarEnLote(t *testing.T) {
	ctx := context.Background()
	c, _, _ := nubeSQS(t)
	url := crearCola(t, c, "lotes", nil)
	res, err := c.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(url),
		Entries: []types.SendMessageBatchRequestEntry{
			{Id: aws.String("1"), MessageBody: aws.String("uno")},
			{Id: aws.String("2"), MessageBody: aws.String("dos")},
			{Id: aws.String("3"), MessageBody: aws.String("tres")},
		},
	})
	if err != nil {
		t.Fatalf("SendMessageBatch: %v", err)
	}
	if len(res.Successful) != 3 {
		t.Fatalf("enviados %d de 3", len(res.Successful))
	}

	rec, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl: aws.String(url), MaxNumberOfMessages: 10,
	})
	if len(rec.Messages) != 3 {
		t.Fatalf("recibi %d de 3", len(rec.Messages))
	}
	var borrados []types.DeleteMessageBatchRequestEntry
	for i, m := range rec.Messages {
		borrados = append(borrados, types.DeleteMessageBatchRequestEntry{
			Id: aws.String(string(rune('a' + i))), ReceiptHandle: m.ReceiptHandle,
		})
	}
	if _, err := c.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(url), Entries: borrados,
	}); err != nil {
		t.Fatalf("DeleteMessageBatch: %v", err)
	}
	attr, _ := c.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(url),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if attr.Attributes["ApproximateNumberOfMessages"] != "0" ||
		attr.Attributes["ApproximateNumberOfMessagesNotVisible"] != "0" {
		t.Errorf("la cola no quedo vacia: %v", attr.Attributes)
	}
}

// Los tres servicios, en el MISMO puerto.
func TestLosTresServiciosConvivenEnElMismoPuerto(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(NuevasTablas(), nil),
		"AmazonSQS.": NuevoSQS(NuevasColas(), nil),
	}))
	defer srv.Close()

	if _, err := clienteS3(srv.URL).ListBuckets(ctx, nil); err != nil {
		t.Fatalf("S3: %v", err)
	}
	din := clienteDynamo(srv.URL)
	if _, err := din.ListTables(ctx, nil); err != nil {
		t.Fatalf("DynamoDB: %v", err)
	}
	col := sqs.New(sqs.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(srv.URL),
	})
	if _, err := col.ListQueues(ctx, &sqs.ListQueuesInput{}); err != nil {
		t.Fatalf("SQS: %v", err)
	}
}
```

### `tablas.go`

```go
package main

// El estado de DynamoDB: tablas y elementos. Igual que en S3, es lo unico con candado.
//
// Lo que hay que entender ANTES de escribir una linea, y que cambia como diseñas tus tablas:
//
//	· Un elemento es un MAPA de atributos con TIPO: {"nombre": {"S":"ana"}, "edad": {"N":"31"}}.
//	· Los numeros viajan como CADENA. No es un descuido: asi el numero no pasa por el float de
//	  ningun lenguaje y no pierde precision por el camino.
//	· La CLAVE se elige al crear la tabla y no se cambia nunca. Es de particion, o de particion
//	  mas orden. Todo lo demas del elemento es libre: no hay esquema.
//	· `Query` solo sabe buscar POR LA CLAVE. Si preguntas por otra cosa, es `Scan`, y `Scan` lee
//	  la tabla entera. Por eso la clave se elige una vez y con cuidado.

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"
)

// AV es un AttributeValue tal como viaja por el cable: {"S":"hola"} · {"N":"42"} · {"BOOL":true}
type AV map[string]any

// Elemento es una fila: un mapa de nombre de atributo a valor con tipo.
type Elemento map[string]AV

type Clave struct {
	Nombre string // el nombre del atributo
	Tipo   string // S | N | B
}

type Tabla struct {
	Nombre    string
	PK        Clave
	SK        *Clave // nil si la tabla solo tiene clave de particion
	Creada    time.Time
	elementos map[string]Elemento // clave compuesta serializada -> elemento
}

type Tablas struct {
	mu     sync.RWMutex
	tablas map[string]*Tabla
	reloj  func() time.Time
}

func NuevasTablas() *Tablas {
	return &Tablas{tablas: map[string]*Tabla{}, reloj: time.Now}
}

// ── comparar valores, que es lo que hace que el orden exista ─────────────

// escalar saca el valor comparable de un AV de clave. Solo S, N y B pueden ser clave.
func escalar(v AV) (tipo string, s string, n *big.Float, b []byte, ok bool) {
	if x, hay := v["S"]; hay {
		t, _ := x.(string)
		return "S", t, nil, nil, true
	}
	if x, hay := v["N"]; hay {
		t, _ := x.(string)
		f, _, err := big.ParseFloat(t, 10, 200, big.ToNearestEven)
		if err != nil {
			return "N", t, nil, nil, false
		}
		return "N", t, f, nil, true
	}
	if x, hay := v["B"]; hay {
		t, _ := x.(string)
		raw, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			return "B", t, nil, nil, false
		}
		return "B", t, nil, raw, true
	}
	return "", "", nil, nil, false
}

// compara devuelve -1, 0 o 1. ⚠️ Las cadenas se comparan POR BYTES, igual que DynamoDB. Por eso
// las fechas se guardan en ISO-8601: "2026-02-01" va despues de "2026-01-31" por puro orden de
// bytes, sin que nadie sepa que son fechas.
func compara(a, b AV) (int, bool) {
	ta, sa, na, ba, oka := escalar(a)
	tb, sb, nb, bb, okb := escalar(b)
	if !oka || !okb || ta != tb {
		return 0, false
	}
	switch ta {
	case "S":
		return strings.Compare(sa, sb), true
	case "N":
		return na.Cmp(nb), true
	case "B":
		return bytes.Compare(ba, bb), true
	}
	return 0, false
}

// serie convierte un valor de clave en algo que sirve de clave de mapa, CON su tipo delante:
// sin el tipo, el numero 1 y la cadena "1" colisionarian.
func serie(v AV) (string, bool) {
	t, s, _, _, ok := escalar(v)
	if !ok {
		return "", false
	}
	return t + ":" + s, true
}

func (t *Tabla) claveDe(e Elemento) (string, error) {
	pk, hay := e[t.PK.Nombre]
	if !hay {
		return "", fmt.Errorf("falta la clave de particion %q", t.PK.Nombre)
	}
	sp, ok := serie(pk)
	if !ok {
		return "", fmt.Errorf("la clave de particion %q no es S, N ni B", t.PK.Nombre)
	}
	if t.SK == nil {
		return sp, nil
	}
	sk, hay := e[t.SK.Nombre]
	if !hay {
		return "", fmt.Errorf("falta la clave de orden %q", t.SK.Nombre)
	}
	ss, ok := serie(sk)
	if !ok {
		return "", fmt.Errorf("la clave de orden %q no es S, N ni B", t.SK.Nombre)
	}
	return sp + "\x00" + ss, nil
}

// ── operaciones sobre las tablas ─────────────────────────────────────────

func (ts *Tablas) Crear(nombre string, pk Clave, sk *Clave) (*Tabla, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if _, ya := ts.tablas[nombre]; ya {
		return nil, false
	}
	t := &Tabla{Nombre: nombre, PK: pk, SK: sk, Creada: ts.reloj().UTC(),
		elementos: map[string]Elemento{}}
	ts.tablas[nombre] = t
	return t, true
}

func (ts *Tablas) Buscar(nombre string) (*Tabla, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tablas[nombre]
	return t, ok
}

func (ts *Tablas) Borrar(nombre string) (*Tabla, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	t, ok := ts.tablas[nombre]
	if ok {
		delete(ts.tablas, nombre)
	}
	return t, ok
}

// Listar devuelve los nombres ORDENADOS: un mapa de Go no tiene orden y sin ordenar la salida
// cambia en cada llamada.
func (ts *Tablas) Listar() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	fuera := make([]string, 0, len(ts.tablas))
	for n := range ts.tablas {
		fuera = append(fuera, n)
	}
	sort.Strings(fuera)
	return fuera
}

func (ts *Tablas) Cuenta(nombre string) int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if t, ok := ts.tablas[nombre]; ok {
		return len(t.elementos)
	}
	return 0
}

// ── operaciones sobre los elementos ──────────────────────────────────────

// Poner guarda el elemento y devuelve el que habia antes (nil si no habia). Devolverlo es lo que
// permite implementar ReturnValues=ALL_OLD sin volver a leer.
func (ts *Tablas) Poner(tabla string, e Elemento, condicion func(Elemento) bool) (Elemento, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	t, ok := ts.tablas[tabla]
	if !ok {
		return nil, errNoExisteTabla
	}
	k, err := t.claveDe(e)
	if err != nil {
		return nil, err
	}
	anterior := t.elementos[k]
	if condicion != nil && !condicion(anterior) {
		return nil, errCondicion
	}
	t.elementos[k] = e
	return anterior, nil
}

func (ts *Tablas) Obtener(tabla string, clave Elemento) (Elemento, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tablas[tabla]
	if !ok {
		return nil, errNoExisteTabla
	}
	k, err := t.claveDe(clave)
	if err != nil {
		return nil, err
	}
	return t.elementos[k], nil
}

func (ts *Tablas) Quitar(tabla string, clave Elemento, condicion func(Elemento) bool) (Elemento, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	t, ok := ts.tablas[tabla]
	if !ok {
		return nil, errNoExisteTabla
	}
	k, err := t.claveDe(clave)
	if err != nil {
		return nil, err
	}
	anterior := t.elementos[k]
	if condicion != nil && !condicion(anterior) {
		return nil, errCondicion
	}
	delete(t.elementos, k)
	return anterior, nil
}

// Recorrer entrega los elementos ordenados por clave de particion y, dentro, por clave de orden.
// Es lo que necesitan Scan y Query, y lo que hace que el resultado sea estable entre corridas.
//
// ⚠️ Se ordena COMPARANDO LOS VALORES, no la cadena con la que se indexan. Ordenar por la cadena
// pondria el 10 antes que el 9, porque "10" es menor que "9" byte a byte. En una clave de orden
// numerica eso devuelve la pagina equivocada y no lo avisa nadie.
func (ts *Tablas) Recorrer(tabla string) ([]Elemento, *Tabla, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tablas[tabla]
	if !ok {
		return nil, nil, errNoExisteTabla
	}
	fuera := make([]Elemento, 0, len(t.elementos))
	for _, e := range t.elementos {
		fuera = append(fuera, e)
	}
	sort.SliceStable(fuera, func(i, j int) bool { return t.antes(fuera[i], fuera[j]) })
	return fuera, t, nil
}

// antes dice si el elemento a va antes que el b: primero por particion, luego por orden.
func (t *Tabla) antes(a, b Elemento) bool {
	if c, ok := compara(a[t.PK.Nombre], b[t.PK.Nombre]); ok && c != 0 {
		return c < 0
	}
	if t.SK == nil {
		return false
	}
	c, ok := compara(a[t.SK.Nombre], b[t.SK.Nombre])
	return ok && c < 0
}
```

### `Dockerfile`

```dockerfile
# nubecita en una imagen, en dos etapas.
#
# ⚠️ POR QUE HAY DOCKERFILE desde el episodio 3: el Application Control de la maquina de
# desarrollo empezo a bloquear tambien el binario del servidor recien compilado
# («Una directiva de Control de aplicaciones bloqueo este archivo»). Correr en contenedor quita
# de en medio esa fragilidad y, de paso, es como lo va a arrancar cualquiera:
#
#     docker build -t nubecita .
#     docker run --rm -p 4566:4566 nubecita
#
# La segunda etapa es `scratch`: NADA, ni siquiera un shell. Se puede porque el binario de Go es
# estatico con CGO_ENABLED=0, asi que la imagen entera pesa lo que pesa el binario. Es el mejor
# argumento de Go para contenedores, y aqui se ve sin trucos.

FROM golang:1.26 AS construir
WORKDIR /src
# Primero los modulos: asi la capa de dependencias se reutiliza mientras no cambien.
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
# -tags y -ldflags: sin la tabla de simbolos ni la informacion de depuracion.
# ⛔ El binario tiene que ser ESTATICO (CGO_ENABLED=0) o `scratch` no lo puede ejecutar.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /nubecita .

FROM scratch
COPY --from=construir /nubecita /nubecita
EXPOSE 4566
# 0.0.0.0 y no 127.0.0.1: dentro del contenedor, escuchar solo en la interfaz local dejaria
# fuera al que publica el puerto. Es el fallo numero uno al meter un servidor en un contenedor.
ENTRYPOINT ["/nubecita", "-escuchar", "0.0.0.0"]
```
