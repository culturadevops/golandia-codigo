# ☁️ nubecita — escribe tu propia nube en Go

Un emulador de servicios de AWS, en Go, con **cero dependencias fuera del SDK de pruebas**.
El SDK oficial de AWS para Go y la CLI oficial de AWS hablan con esto **sin enterarse de que no
es Amazon**. Lo único que cambia en tu código es el endpoint.

```bash
docker build -t nubecita . && docker run --rm -p 4566:4566 nubecita
# ...o sin contenedor:
go run .
# nubecita · S3, DynamoDB y SQS en http://127.0.0.1:4566 · sin comprobar firmas · Ctrl+C para parar
```

```bash
export AWS_ACCESS_KEY_ID=prueba AWS_SECRET_ACCESS_KEY=prueba AWS_DEFAULT_REGION=us-east-1
aws --endpoint-url http://127.0.0.1:4566 s3 mb s3://fotos
aws --endpoint-url http://127.0.0.1:4566 s3 cp gato.txt s3://fotos/gato.txt
aws --endpoint-url http://127.0.0.1:4566 s3 ls s3://fotos
```

Y desde tu código en Go, las **dos líneas** que lo cambian todo:

```go
c := s3.New(s3.Options{
    Region:       "us-east-1",
    Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
    BaseEndpoint: aws.String("http://127.0.0.1:4566"), // ve aquí, no a s3.amazonaws.com
    UsePathStyle: true,                                // /bucket/clave, no bucket.host/clave
})
```

> ⚠️ `UsePathStyle: true` no es un capricho. Por defecto el SDK mete el bucket en el **nombre del
> host** (`fotos.s3.amazonaws.com`), y contra `localhost` eso no resuelve. Es el error número uno
> al apuntar el SDK a cualquier emulador.

---

## ¿Por qué escribir esto si ya existe LocalStack?

Tres razones, y la tercera es la de verdad:

1. **LocalStack dejó de ser gratis para uso comercial** el 23 de marzo de 2026: se acabó la
   edición Community, ahora pide cuenta y token, y el plan gratuito excluye el uso comercial.
2. Un binario de 8 MB arranca en milisegundos. Un contenedor de emulación, no.
3. 🥇 **Porque escribiéndolo entiendes S3 de verdad.** Cuando implementas `ListObjectsV2` con
   delimitador te enteras, de golpe y para siempre, de que **S3 no tiene carpetas**. Eso no lo
   aprendes usando LocalStack: lo aprendes escribiéndolo.

⛔ Y lo que **no** es: esto no es para producción y no lo va a ser nunca. No hay firmas, no hay
permisos, no hay disco. Es un doble de pruebas, y un doble de pruebas vale justo por lo que
**no** hace.

---

## Lo que hay hecho — S3 · DynamoDB · SQS · SNS · SigV4 · permisos · Lambda · Logs · persistencia · lanzador · disparadores

**1.946 líneas** de código (sin comentarios ni blancos), **30 operaciones**, comprobación de
firma opcional y una imagen que se descarga en **2,6 MiB**. Los tres servicios viven en **el
mismo puerto y el mismo proceso**.

⚠️ Dos números distintos: la imagen se **descarga** en 2,61 MiB (capas comprimidas) y **ocupa**
9,17 MB en disco. Confundirlos es publicar una cifra falsa.

### Episodio 1 — S3

| operación | ruta | hecho |
|---|---|:--:|
| `ListBuckets` | `GET /` | ✅ |
| `CreateBucket` | `PUT /{bucket}` | ✅ |
| `DeleteBucket` | `DELETE /{bucket}` | ✅ |
| `HeadBucket` | `HEAD /{bucket}` | ✅ |
| `ListObjectsV2` | `GET /{bucket}?list-type=2` | ✅ prefijo · delimitador · max-keys · encoding-type |
| `PutObject` | `PUT /{bucket}/{clave...}` | ✅ ETag = md5 |
| `GetObject` | `GET /{bucket}/{clave...}` | ✅ |
| `HeadObject` | `HEAD /{bucket}/{clave...}` | ✅ |
| `DeleteObject` | `DELETE /{bucket}/{clave...}` | ✅ idempotente |

Y con eso ya funcionan `aws s3 mb`, `cp`, `sync`, `ls --recursive`, `rm --recursive` y `rb`.

### Episodio 2 — DynamoDB

⚠️ **No comparte protocolo con S3, y esa es la lección.** S3 dice qué quieres con el método y la
ruta. DynamoDB manda **siempre** un `POST /` con el cuerpo en JSON y la operación en una
cabecera:

```http
POST / HTTP/1.1
X-Amz-Target: DynamoDB_20120810.PutItem
Content-Type: application/x-amz-json-1.0

{"TableName":"pedidos","Item":{"cliente":{"S":"ana"},"total":{"N":"42.5"}}}
```

Dos servicios de la misma nube, dos diseños. No es incoherencia: S3 es de 2006 y DynamoDB de
2012, cuando Amazon ya había estandarizado el estilo «JSON con la operación en la cabecera».
Los dos caben en el mismo puerto porque `enrutador.go` mira esa cabecera y reparte.

| operación | hecho |
|---|:--:|
| `CreateTable` · `DescribeTable` · `DeleteTable` · `ListTables` | ✅ clave simple o compuesta |
| `PutItem` · `GetItem` · `DeleteItem` | ✅ con `ConditionExpression` (`attribute_exists` / `attribute_not_exists`) y `ReturnValues=ALL_OLD` |
| `Query` | ✅ `KeyConditionExpression` completa: `=` en la partición, y `=` `<` `<=` `>` `>=` `BETWEEN` `begins_with` en el orden |
| `Scan` | ✅ con `Limit` y `Select=COUNT` |

⛔ **No hay**: índices secundarios, `UpdateItem`, transacciones, streams, paginación real ni
`FilterExpression`. No es una lista de tareas pendientes: es la frontera de un doble de pruebas.

### Episodio 3 — SQS

Habla como DynamoDB —`POST /`, JSON, operación en la cabecera— pero con otro prefijo:
`AmazonSQS.SendMessage`. Y con una historia detrás: **SQS es más viejo que S3** y durante años
habló el protocolo «Query» de AWS, con los parámetros en el cuerpo tipo formulario. Amazon lo
migró a JSON. O sea: el protocolo de un servicio **no es para siempre**.

| operación | hecho |
|---|:--:|
| `CreateQueue` · `ListQueues` · `GetQueueUrl` · `DeleteQueue` · `PurgeQueue` | ✅ con `VisibilityTimeout` y `RedrivePolicy` |
| `GetQueueAttributes` | ✅ los dos números «aproximados» |
| `SendMessage` · `SendMessageBatch` | ✅ con el MD5 que el SDK **comprueba** |
| `ReceiveMessage` | ✅ visibilidad, `ApproximateReceiveCount` y recibo nuevo en cada entrega |
| `DeleteMessage` · `DeleteMessageBatch` · `ChangeMessageVisibility` | ✅ |
| cola de mensajes muertos | ✅ a la N-ésima entrega, se aparta |

⛔ **No hay**: colas FIFO, `DelaySeconds`, atributos de mensaje, sondeo largo de verdad
(`WaitTimeSeconds` se acepta pero no bloquea) ni políticas de acceso.

### Episodio 4 — SigV4, la firma

Hasta aquí el doble no autenticaba a nadie, que es lo correcto para un doble de pruebas. Con
`-firmar` comprueba la firma **SigV4** de cada petición:

```bash
docker run --rm -p 4566:4566 nubecita -escuchar 0.0.0.0 -firmar
```

⚠️ **Firmar no es cifrar.** La petición viaja en claro, con su cuerpo legible. Lo que la firma
demuestra son otras dos cosas: **quién** la manda —sin que el secreto viaje nunca por el cable—
y que **nadie la ha tocado** por el camino.

Los cuatro pasos, que son todo SigV4:

```
1. PETICIÓN CANÓNICA   se reescribe la petición en un formato exacto y sin ambigüedades
2. CADENA A FIRMAR     la fecha, el ámbito, y el resumen de esa petición canónica
3. CLAVE DE FIRMA      una cadena de HMAC: secreto -> fecha -> región -> servicio -> aws4
4. FIRMA               HMAC de la cadena a firmar con esa clave
```

Y lo que se aprende implementándolo:

- 🥇 **El paso 1 es el que duele.** Para que salga la misma firma hay que reconstruir la
  petición **exactamente** igual que el cliente: orden de los parámetros, cabeceras firmadas,
  codificación de la ruta. Una barra de más y la firma cambia entera — y el error solo dice
  `SignatureDoesNotMatch`, nunca por qué. Aquí sí **decimos por qué**.
- **La cadena de HMAC acota la clave**: fecha → región → servicio. Una firma de hoy no vale
  mañana, y una de S3 no vale para DynamoDB. Está limitada por diseño.
- 🥇 **La firma cubre el RESUMEN del cuerpo, no el cuerpo.** Si alguien cambia el cuerpo y deja
  la cabecera `x-amz-content-sha256` como estaba, la firma sigue cuadrando. Por eso hay que
  comprobar **además** que el cuerpo hace ese resumen.
- **`UNSIGNED-PAYLOAD` significa literalmente «no firmo el cuerpo».** S3 lo admite sobre TLS,
  porque entonces confías en el canal. Ahí la firma dice quién eres, pero **no** protege el
  contenido.
- **Una firma caduca** (ventana de 15 minutos). Sin eso, quien capture una sola petición firmada
  podría repetirla para siempre.
- **`hmac.Equal` y no `==`.** Comparar cadenas sale antes en cuanto hay una diferencia, y ese
  «antes» se puede medir: es un canal por tiempos.

⛔ **No hay**: firmas en la URL (presigned), STS ni credenciales temporales, permisos ni
políticas. Comprobar quién eres no es lo mismo que decidir qué puedes hacer.

### Episodio 5 — Lambda, con la API del runtime de verdad

🥇 **La biblioteca oficial de Amazon se conecta a nubecita sin cambiar una línea.** No hay
parche, no hay «modo local»: lo único que cambia es una variable de entorno.

```bash
# el servicio
docker run --rm -p 4566:4566 nubecita -escuchar 0.0.0.0

# tu función, compilada con aws-lambda-go, sin tocar nada
docker run -e AWS_LAMBDA_RUNTIME_API=nubecita:4566/f/saludar funcion-saludar:1
```

Y funciona porque Lambda no es una API: son **dos**.

```
1) EL PLANO DE DATOS      lo llama TU código, con el SDK de AWS
   POST /2015-03-31/functions/saludar/invocations

2) LA API DEL RUNTIME     la llama la FUNCIÓN, desde dentro de su contenedor
   GET  /2018-06-01/runtime/invocation/next          <- se queda esperando
   POST /2018-06-01/runtime/invocation/{id}/response
```

La segunda es la que casi nadie ha visto, y es la que explica Lambda entera: **tu función no
recibe peticiones, las va a buscar**. `aws-lambda-go` es un bucle alrededor de ese `GET`.

Lo que se aprende implementándolo:

- 🥇 **El arranque en frío es esperar en un canal.** Quien invoca deja la invocación y se
  bloquea; el runtime la viene a buscar. Si todavía no hay nadie al otro lado, esa espera es
  exactamente el arranque en frío. No es una penalización que te cobran: es que el receptor aún
  no ha llegado.
- 🥇 **Si tu función falla, la llamada contesta `200`.** El fallo viaja en la cabecera
  `X-Amz-Function-Error`, no en el código HTTP — porque para Lambda «se ejecutó y devolvió un
  error» es una invocación correcta. Un cliente escrito a mano que solo mire el código va a
  tratar como éxito **cada fallo de tu función**.
- **Un runtime atiende de una en una.** Medido: 3 invocaciones de 0,8 s con 1 contenedor son
  2,9 s; con 3 contenedores, 1,3 s. Escalar en Lambda es más contenedores, no más hilos.
- **El plazo viaja en una cabecera** (`Lambda-Runtime-Deadline-Ms`), y de ahí sale el
  `ctx.Deadline()` que le llega a tu manejador. Si nadie contesta a tiempo: `Sandbox.Timedout`
  — otra vez con un `200`.
- **La invocación asíncrona (`Event`) contesta `202` y sin resultado.** No es que se pierda: es
  que cuando la API contesta, todavía no se ha ejecutado nada.
- **Las variables de paquete sobreviven entre invocaciones** mientras el contenedor viva. Por
  eso el arranque en frío se paga una vez — y por eso guardar estado en una global es una
  trampa: a veces está y a veces no.

⚠️ **Dos detalles de Go que se ven aquí y en ningún otro sitio:**

- El canal de la respuesta tiene **buffer 1**. Quien invoca puede irse antes de tiempo; si el
  canal fuera sin buffer, el runtime se quedaría bloqueado para siempre entregando algo que ya
  no espera nadie. Una gorrutina perdida por invocación abandonada.
- **No se espera en un canal con el candado cogido.** El sondeo largo saca el canal del mapa
  *antes* de bloquearse; si no, nadie podría ni listar funciones mientras haya un runtime
  esperando.

⚠️ Y un tropiezo real: Lambda va por **ruta**, como S3, así que **no cabe en el mismo `mux`**.
S3 se queda con la raíz entera con comodines, y el `mux` de Go **ni siquiera arranca**: suelta
un pánico al registrar, diciendo que `GET /2015-03-31/functions` atiende más métodos que
`HEAD /{bucket}/{clave...}` —porque un `GET` registrado contesta también a `HEAD`— pero tiene la
ruta más concreta. Solución: su propio `mux`, y entrar **antes** que S3 en el enrutador.

⛔ **No hay**: subir un zip de verdad (el código vive en la imagen del contenedor), versiones,
alias, capas, VPC, disparadores ni registro en CloudWatch.

### Episodio 6 — La persistencia, y qué es estado de verdad

Hasta aquí nubecita no tocaba disco **a propósito**: matabas el proceso y no quedaba nada. Con
el episodio 6 hay dos banderas, y hacen cosas distintas:

```bash
# la máquina de desarrollo: lee al arrancar Y escribe cuando algo cambia
docker run --rm -p 4566:4566 -v "$PWD/datos:/datos" nubecita \
    -escuchar 0.0.0.0 -datos /datos/estado.json

# la prueba reproducible: lee al arrancar y NO escribe NUNCA
docker run --rm -p 4566:4566 -v "$PWD/datos:/datos" nubecita \
    -escuchar 0.0.0.0 -semilla /datos/semilla.json
```

🥇 **La pregunta del episodio: ¿qué es estado y qué no?** Un objeto es estado. Un mensaje es
estado. Pero **el rato que a un mensaje le queda de estar invisible no lo es**: es una
conversación a medias con alguien que ya no existe. Se abandona, no se guarda.

| no se guarda | por qué |
|---|---|
| el **recibo** de un mensaje | era una promesa hecha a alguien que murió con el proceso |
| **cuándo vuelve a verse** un mensaje | el reloj siguió corriendo; el consumidor no |
| el **ETag** y el **tamaño** de un objeto | se calculan de los datos: guardarlos es guardarse la oportunidad de que un día no cuadren |
| el **índice** de claves de una tabla | se reconstruye con el mismo código que lo construye |
| los **canales** de Lambda y las invocaciones en vuelo | un canal no es un dato: es un sitio donde dos gorrutinas se esperan |

Y lo que **sí**, por el mismo razonamiento: la **cuenta de entregas** de un mensaje (si se
reiniciara a cero, un mensaje que falla siempre no llegaría nunca a la cola de muertos) y el
**número de versión** del fichero.

Lo que se aprende implementándolo:

- 🥇 **No hace falta adivinar qué operaciones escriben.** Se avisa después de *cada* petición y
  quien escribe compara el resultado con lo último que guardó; si es igual, no escribe.
  Adivinar es una trampa: `ReceiveMessage` suena a lectura y **modifica** el estado.
- 🥇 **Por eso hay que ordenar al exportar.** El recorrido de un mapa en Go es aleatorio a
  propósito: sin ordenar, dos exportaciones del mismo estado dan JSON distinto, la comparación
  dice «ha cambiado» siempre, y el fichero se reescribe entero en cada petición.
- **Se escribe a un temporal y se renombra.** Escribir encima del fichero bueno tiene un
  instante en el que el fichero existe y está a medias; renombrar dentro del mismo sistema de
  ficheros es atómico. Y hay que `Sync()` **antes** de renombrar, o el nombre nuevo puede llegar
  al disco antes que el contenido.
- **Se guarda cuando algo cambia, no al salir.** Medido con `docker kill` —SIGKILL, sin apagado
  limpio, sin darle tiempo a nada—: el estado sobrevive entero. Si solo se escribiera al salir,
  ahí se habría perdido todo.
- **Un `[]byte` en JSON sale en base64.** El fichero no se lee a simple vista y ocupa un tercio
  más. Mejor saberlo antes de abrirlo y pensar que se corrompió.
- **Un número de versión, y `DisallowUnknownFields`.** Cargar la mitad de un estado es peor que
  no cargarlo: los campos que faltan se quedan en su valor cero **en silencio**, y una cola con
  cero segundos de visibilidad entrega el mismo mensaje en bucle.
- ⚠️ **El detalle de Go que muerde:** un `time.Time` que vuelve de JSON es el **mismo instante**
  —`Equal` dice que sí— pero `==` dice que no, porque el original lleva pegado un reloj monótono
  que el texto no puede llevar. Por eso `reflect.DeepEqual` sobre structs con fechas falla
  después de un ida y vuelta.

⛔ **No hay**: base de datos (es un JSON entero, reescrito de una pieza), registro de
operaciones, recuperación a un punto en el tiempo, ni varios procesos compartiendo el fichero
—el último que escriba gana—.

### Episodio 7 — Que los levante ella

En el episodio 5 tu función se ejecutaba de verdad, pero el contenedor lo levantabas **tú**, con
un `docker run`. Eso está bien para entender el protocolo y es una mentira para usarlo: en
Lambda de verdad tú no levantas nada.

```bash
docker run --rm -p 4566:4566 --network demo --name nubecita \
    -v /var/run/docker.sock:/var/run/docker.sock \
    nubecita -escuchar 0.0.0.0 \
    -docker /var/run/docker.sock -red demo -yo nubecita:4566 -max 3 -ocioso 60s
```

Y a partir de ahí, **solo la CLI de AWS**:

```bash
aws lambda create-function --function-name saludar \
    --package-type Image --code ImageUri=funcion-saludar:1 --timeout 20 --role ...
aws lambda invoke --function-name saludar --payload '{"nombre":"Ada"}' salida.json
```

Los contenedores aparecen y desaparecen solos.

🥇 **Y lo que lo hace posible cabe en una frase: la API de Docker es HTTP.** Un socket de dominio
Unix con peticiones HTTP normales encima. Sin SDK y sin dependencias — se cambia cómo se **abre**
la conexión y se deja igual todo lo demás:

```go
&http.Transport{
    DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
        var d net.Dialer
        return d.DialContext(ctx, "unix", socket)
    },
}
```

**La regla del lanzador, entera:**

```
si no hay NINGUN runtime esperando trabajo  ->  levanta otro
si ya hay uno libre                         ->  no hagas nada
si llevas un rato sin que te invoquen       ->  apagalos
```

De esas tres líneas sale el escalado de Lambda sin proponérselo:

- 🥇 **Una ráfaga levanta contenedores hasta el tope.** Medido: 5 invocaciones a la vez con el
  tope en 3 → 3 contenedores, y las que sobran esperan turno. Eso es lo que significa
  «concurrencia» en la factura de Lambda: contenedores, no hilos.
- 🥇 **El arranque en frío VUELVE.** Una función que nadie llama se queda sin contenedores, y la
  siguiente llamada vuelve a pagarlo. Nadie te guarda una máquina encendida por si acaso.
- **Crear la función no ejecuta nada**: solo apunta qué imagen sabe ejecutarla. Por eso
  `create-function` es instantáneo, aquí y en Amazon.
- **Borrar la función apaga sus contenedores**, o quedarían runtimes pidiéndole trabajo a algo
  que ya no existe.
- **`AutoRemove: false`** a propósito: si el contenedor se borrara solo al morir, un fallo de
  arranque desaparecería antes de que puedas leer por qué.
- **Si Docker dice que no**, la invocación no se cae ahí: sigue su camino y vence como cualquier
  otra. Puede que haya un runtime tuyo, levantado a mano, a punto de llegar.

⚠️ **Y como el protocolo es HTTP, se puede probar sin Docker.** Las once pruebas del lanzador
corren contra un **demonio de Docker de mentira** montado con `httptest` en veinte líneas: se
comprueba qué se le pide —la imagen, la red, la etiqueta y la variable de entorno— sin tocar el
demonio de verdad.

⛔ **AVISO SERIO:** montar `/var/run/docker.sock` dentro de un contenedor le da control del
demonio **entero** de la máquina: puede crear, borrar y montar lo que quiera. Es lo que hace
LocalStack, es lo que hace esto, y es exactamente por eso que un doble de pruebas no se pone en
un servidor compartido.

⛔ **No hay**: descargar la imagen si falta, límites de memoria o CPU por función, concurrencia
reservada o aprovisionada, ni reintentos.

### Episodio 8 — Los disparadores

El episodio que junta todo lo anterior. Hasta aquí había servicios sueltos; aquí, por primera
vez, **subes un fichero y se ejecuta código**.

```bash
# la función reacciona a lo que se suba a originales/*.txt
aws s3api put-bucket-notification-configuration --bucket fotos \
    --notification-configuration file://notificacion.json

# y esto es lo único que se hace
aws s3api put-object --bucket fotos --key originales/pedido.txt --body pedido.txt
```

```console
$ aws s3 ls s3://fotos/ --recursive
2026-08-25 14:37:13         21 originales/pedido.txt
2026-08-25 14:37:13        108 procesados/originales_pedido.txt.json    <- nadie subio esto
```

Y hay **dos formas** de que algo invoque a tu función sin que tú llames, muy distintas entre sí:

```
SQS -> Lambda   un BUCLE que hace ReceiveMessage por ti y te invoca con lo que saca.
                Lo llaman «event source mapping». No es magia: es un for.

S3  -> Lambda   S3 AVISA cuando pasa algo. No hay bucle: hay una llamada.
                Se configura en el BUCKET, no en la funcion.
```

Lo que se aprende implementándolo:

- 🥇 **Tu función NO borra el mensaje.** Lo borra el disparador, y solo si fuiste bien. Por eso
  un consumidor de Lambda no llama nunca a `DeleteMessage`, y por eso un fallo se reintenta sin
  que escribas una línea de reintentos.
- 🥇 **Los fallos parciales.** Si de un lote de diez falla uno y no dices cuál, vuelven **los
  diez**: los nueve buenos se procesan dos veces. Decir cuál es lo que hacen los
  `batchItemFailures`. Medido: de 5 mensajes con uno malo, queda 1.
- ⚠️ **Y la trampa que va con ellos:** si escribes `itemIdentifiers` en vez de `itemIdentifier`,
  no hay ningún error — se borra el lote entero. Una errata ahí no da fallo, da pérdida de
  mensajes. Está comprobado en `TestDisparadorUnaErrataEnElCampoBorraTodo`.
- **El filtro de S3 se evalúa en S3**, no dentro de tu función. Si tu código se ejecuta con un
  fichero que no querías, el filtro está mal escrito.
- **La notificación sale DESPUÉS de contestar.** Cuando S3 te dice «200», tu objeto ya está
  guardado y la notificación todavía no ha salido. Por eso una notificación perdida no deja el
  objeto a medias — y por eso tu `PutObject` no se entera de que la función falló.
- **El nombre del evento viaja sin el `s3:` de delante**: configuras `s3:ObjectCreated:*` y te
  llega `ObjectCreated:Put`. Es un detalle tonto y es de los que hacen que tu `if` no entre.
- ⚠️ **Los nombres del XML no son los del SDK**: en Go el campo se llama `LambdaFunctionArn` y
  en el XML la etiqueta es `CloudFunction`; `QueueArn` viaja como `Queue`. Es historia — cuando
  se diseñó esto, Lambda todavía se llamaba «cloud function». Se descubre mirando el XML que
  manda el SDK de verdad, no leyendo la documentación.
- ⚠️ **Y el mux de Go no mira la cadena de consulta.** S3 mete ahí sus subrecursos
  (`?notification`, `?acl`, `?versioning`), así que la misma ruta y el mismo método son dos
  operaciones distintas y hay que repartirlas a mano dentro del manejador.
- **Una gorrutina que nadie sabe parar es una fuga con buenos modales.** El bucle del disparador
  vive fuera de cualquier petición: se le da su propio contexto, y `Borrar` **espera** a que
  termine antes de volver.

⛔ **No hay**: disparadores de DynamoDB Streams ni de Kinesis, escalado automático del sondeador,
`MaximumBatchingWindow`, ni reintentos de las notificaciones de S3.

### Episodio 9 — SNS

Una cola y un tema parecen lo mismo y no lo son, y la diferencia cabe en una frase:

```
COLA   el mensaje se lo lleva UNO. Quien llega primero.
TEMA   el mensaje se lo llevan TODOS los que estaban apuntados.
```

```bash
aws sns create-topic --name pedidos
aws sns subscribe --topic-arn ... --protocol sqs --notification-endpoint arn:aws:sqs:...:facturacion
aws sns publish --topic-arn ... --message '{"pedido":42}'
```

Una publicación, y el mensaje aparece en **todas** las colas suscritas.

Lo que se aprende implementándolo:

- 🥇 **Un tema NO guarda nada.** Publicar sin suscriptores no pierde el mensaje: es que nunca
  existió. Por eso el patrón de verdad no es «un tema» ni «una cola», sino un **tema que reparte
  a colas**: el tema hace la copia y cada cola la guarda hasta que su consumidor pueda.
- 🥇 **Sin `RawMessageDelivery`, el mensaje llega ENVUELTO.** Tu consumidor lee el cuerpo del
  mensaje de la cola y se encuentra un sobre, no la carta. Lo que publicaste está dentro, en
  `Message`, y además **como cadena**: hay que deserializar dos veces.
- 🥇 **El filtro lo pone el SUSCRIPTOR**, no quien publica. Quien publica no sabe ni cuántos hay
  apuntados. Y filtrar ahí es gratis: lo que no pasa el filtro no llega a tu cola, y lo que no
  llega a tu cola no se factura.
- ⚠️ **Si el mensaje no trae el atributo, el filtro NO deja pasar.** No es «da igual»: es que no
  coincide. Es el motivo número uno de «mi filtro no deja pasar nada».
- **Crear un tema es idempotente**: crearlo dos veces devuelve el mismo ARN.

⚠️ **Y la sorpresa de protocolo:** SNS **no** habla como SQS. Habla el protocolo «Query» de toda
la vida — un formulario en el cuerpo, la operación en un campo `Action`, y la respuesta en texto
etiquetado — que es exactamente el que tenía SQS antes de que Amazon lo migrara a JSON. SNS se
quedó donde estaba. Es el **cuarto** protocolo distinto dentro de la misma nube:

```
S3         metodo y ruta, respuesta en texto etiquetado
DynamoDB   POST a "/", JSON, operacion en una cabecera
SQS        igual que DynamoDB (lo migraron; SNS no)
SNS        formulario en el cuerpo, operacion en un campo
```

Y para repartirlo hace falta un cuarto criterio en el enrutador, otra vez **positivo**: el
**tipo de contenido**. SNS manda un formulario; DynamoDB y SQS mandan JSON. Se distingue sin
tocar el cuerpo — que es lo importante, porque leerlo ahí se lo quitaría a quien viene detrás.

⛔ **No hay**: entrega real a correo, HTTP o SMS (se aceptan y se dice en el registro que no
llega nada), políticas de filtrado con `numeric`/`prefix`/`anything-but`, temas FIFO,
confirmación de suscripción, ni cola de fallos de entrega.

### Episodio 10 — Los permisos

La otra mitad del episodio 4. Aquel terminaba diciendo, con todas las letras: «comprobar
**quién** eres no es decidir **qué** puedes hacer». Esto es lo segundo.

```
AUTENTICAR   ¿eres quien dices?   -> la firma.       Episodio 4.
AUTORIZAR    ¿puedes hacer esto?  -> las politicas.  Este.
```

```bash
docker run --rm -p 4566:4566 -v "$PWD/pol:/pol" nubecita -escuchar 0.0.0.0 \
    -firmar -permisos /pol/permisos.json
```

El fichero lleva las tres cosas juntas, porque eso es un usuario: **quién es, con qué se
identifica, y qué puede hacer**.

```json
{"lector": {"nombre": "lector", "secreto": "...", "politica": {
  "Version": "2012-10-17",
  "Statement": [{"Effect": "Allow", "Action": ["s3:Get*", "s3:ListBucket"],
                 "Resource": ["arn:aws:s3:::fotos", "arn:aws:s3:::fotos/*"]}]}}}
```

🥇 **Las tres reglas de la evaluación**, y en este orden:

```
1. por defecto NO. Sin un Allow que te toque, no pasas.
2. un Allow explicito te deja pasar.
3. un Deny explicito GANA SIEMPRE, aunque haya mil Allow.
```

De la tercera sale el patrón que usa todo el mundo sin saber por qué: un `Deny` general como red
de seguridad, porque **nada** lo puede levantar. Ni un administrador, ni una política nueva, ni
un despiste. Y el `Deny` se busca **entero antes** de mirar ningún `Allow`: si no, el resultado
dependería del orden en que escribiste las sentencias.

Lo que se aprende implementándolo:

- 🥇 **La parte difícil no es decidir si puedes.** Eso son veinte líneas. La parte difícil es
  saber **qué estás pidiendo**: una política habla de `s3:PutObject` sobre
  `arn:aws:s3:::fotos/gato.png`, y por el cable llega `PUT /fotos/gato.png`. Alguien tiene que
  traducir, y la traducción es **distinta en cada servicio**.
- ⚠️ **Y para saber sobre qué recurso pides algo, a veces hay que leer el cuerpo** — DynamoDB y
  SQS llevan ahí el nombre de la tabla y la URL de la cola. Y si lo lees, hay que **volver a
  dejarlo puesto**, o el servicio de detrás se encuentra una petición vacía. Es el mismo tropiezo
  del episodio 4.
- 🥇 **El comodín no sabe de servicios: sabe de letras.** `s3:*` incluye `s3:DeleteBucket`, y
  `s3:Get*` incluye `s3:GetBucketNotification`, que probablemente no era lo que querías. Y no
  son expresiones regulares: solo `*` y `?`; un punto es un punto.
- ⚠️ **El bucket y sus objetos son DOS recursos.** `arn:aws:s3:::fotos/*` no incluye
  `arn:aws:s3:::fotos`, y por eso `s3:ListBucket` se queda fuera si solo pones la forma con
  barra. Es el error más repetido escribiendo políticas de S3.
- **`Action` y `Resource` pueden ser una cadena o una lista**, y hay que aceptar las dos formas.
  El JSON de la biblioteca estándar no lo hace solo: se escribe un `UnmarshalJSON` a mano.
- **El «no» tiene la forma del servicio.** Si le devuelves un JSON a un cliente de S3, el SDK no
  puede convertirlo en un `AccessDenied` tipado y a tu código le llega «respuesta inesperada».
- 🥇 **Y por qué un error dice qué falta y el otro no:**

| | qué contesta | por qué |
|---|---|---|
| firma que no cuadra | `SignatureDoesNotMatch` | el servidor **no sabe** qué está mal: solo sabe que su firma no es la tuya |
| permiso que falta | `... is not authorized to perform: s3:PutObject on resource: ...` | el permiso que falta **está escrito**, así que se puede nombrar |

Por eso uno se depura en dos minutos y el otro en dos horas.

⛔ **No hay**: usuarios, grupos ni roles (una clave y sus políticas); `AssumeRole` ni credenciales
temporales; `Condition`, `NotAction` ni `Principal`; políticas de recurso; ni permisos entre
servicios —un disparador invoca tu función sin pedir permiso, y en AWS de verdad haría falta uno.

⚠️ Y la consecuencia honesta: esto sirve para **entender** cómo se evalúa una política, no para
comprobar si la tuya es segura.

### CloudWatch Logs — sin episodio (todavía)

Esto no es un episodio: es una pieza que se escribió sola, sin vídeo, porque **cerraba un hueco
declarado**. El episodio 7 terminaba diciendo, con estas palabras: *«lo que imprima la función se
queda en su contenedor»*. Se quedaba ahí porque nadie lo sacaba. Esto es el que lo saca.

```bash
aws --endpoint-url http://localhost:4566 logs tail /aws/lambda/saludar
aws --endpoint-url http://localhost:4566 logs filter-log-events \
    --log-group-name /aws/lambda/saludar --filter-pattern ERROR
```

🥇 **Un registro no es una cola**, aunque las dos guarden cosas en orden:

```
una cola      leer SACA el mensaje. Lo lee UNO.    El puntero lo lleva el SERVIDOR.
un registro   leer no borra nada.   Lo leen TODOS. El puntero lo lleva el LECTOR.
```

De esa única diferencia sale todo lo demás: por eso un registro tiene `nextToken` y una cola
tiene `ReceiptHandle`, y por eso dos consumidores de una cola se reparten el trabajo mientras dos
lectores de un registro ven exactamente lo mismo.

Y la jerarquía son tres niveles, que conviene tener claros antes de leer nada:

```
grupo    /aws/lambda/saludar        lo que TU nombras. Aqui vive la retencion.
flujo    2026/08/25/[$LATEST]a1b2   una fuente concreta: un contenedor, una maquina.
apunte   {marca, mensaje}           una linea, con la hora en MILISEGUNDOS.
```

Lo que se aprende implementándolo:

- 🥇 **La salida de un contenedor no viene en texto: viene en MARCOS.** Docker mete `stdout` y
  `stderr` por la misma conexión, y para poder separarlos le pone a cada trozo ocho bytes por
  delante: uno con el flujo, tres a cero, y cuatro con el tamaño en *big endian*. Si lo lees como
  texto te encuentras basura binaria cada pocas líneas y piensas que el log está corrupto. No lo
  está: estás leyendo el sobre.
- ⚠️ **Y un marco no es una línea.** Puede traer media, tres, o una partida por la mitad de una
  palabra. Hay que ir juntando por flujo hasta ver un salto.
- ⚠️ **El mismo endpoint devuelve dos formatos distintos** según cómo se creó el contenedor: con
  `Tty: true` no hay marcos, porque con terminal solo hay un flujo y no hay nada que separar. Y
  no te avisa: hay que mirar los primeros ocho bytes.
- 🥇 **`nextToken` nunca viene vacío.** Ni cuando no queda nada. Un bucle *«mientras haya token,
  pide otra página»* **no termina jamás**: se queda dando vueltas devolviendo cero apuntes. La
  forma correcta de parar es comparar — si el token que te devuelven es el mismo que mandaste, se
  acabó.
- 🥇 **La historia del `sequenceToken`**, que es la mejor de este servicio. Durante nueve años,
  para escribir había que mandar el token de la escritura anterior; si te equivocabas, el error
  te devolvía el token bueno **dentro**. Servía para que dos escritores del mismo flujo no se
  pisaran. En 2023 Amazon dejó de exigirlo. Aquí está implementado entero porque es lo que
  explica por qué existía: un contador de versión, con otro nombre.
- ⚠️ **La ventana de tiempo, que es la explicación de «puse un log y no aparece».** Un apunte de
  hace más de 14 días o de dentro de más de 2 horas se **rechaza**, y la llamada contesta `200`.
  Si el contenedor tiene la hora mal, sus líneas desaparecen sin decir nada.
- ⚠️ **`endTime` es exclusivo.** Pedir «hasta las 12:00:00.000» no trae el apunte de las
  12:00:00.000.
- **La hora la pone Docker, no nosotros** (`timestamps=1`). Es la correcta: si la pusiéramos al
  recibirla, un contenedor rápido y una red lenta te ordenarían mal el registro.
- **Los campos van en minúscula**: `logGroupName`, no `LogGroupName`. DynamoDB manda `TableName`
  y SQS manda `QueueUrl`. Mismo protocolo, misma empresa, misma cabecera, y otra convención de
  nombres. Cada equipo eligió la suya y ya no se puede cambiar sin romper a todo el mundo.
- **El ARN de un grupo acaba en `:*`**, y eso es parte del nombre, no un comodín. Una política
  escrita sin él no encaja.

Y una cosa que se comprueba sola: **el servicio nuevo entra por la misma puerta**. Las firmas del
episodio 4 y los permisos del episodio 10 valen aquí sin tocar nada. Lo único suyo es cómo se
traduce su petición a un permiso.

⛔ **No hay**: métricas ni filtros de métrica; suscripciones; `StartQuery` (o sea, Logs Insights);
exportar a S3; cifrado con KMS; ni Live Tail. La retención **se guarda y no se aplica**: un doble
de pruebas que te borra los registros mientras depuras no ayuda a nadie. Y los registros **no se
persisten** con `-datos`: son de esta corrida.

### Los ficheros

```
main.go        arranque, banderas y apagado limpio
enrutador.go   como UN puerto sirve VARIOS servicios: mirando X-Amz-Target
almacen.go     S3: buckets y objetos, con un RWMutex
s3.go          S3: rutas, XML y errores. HTTP a secas
tablas.go      DynamoDB: tablas, elementos y el ORDEN (que no es alfabetico si la clave es N)
dynamo.go      DynamoDB: JSON, X-Amz-Target y el analizador de KeyConditionExpression
colas.go       SQS: la VISIBILIDAD, el recibo por entrega y la cola de muertos
sqs.go         SQS: JSON con otro prefijo, y el RedrivePolicy con JSON dentro de una cadena
firma.go       SigV4: peticion canonica, cadena de HMAC, ventana de validez y hmac.Equal
funciones.go   Lambda: el ENCUENTRO por canales entre quien invoca y quien ejecuta
lambda.go      Lambda: las DOS APIs — el plano de datos y la API del runtime
ejemplo-funcion/  una funcion de verdad, con `aws-lambda-go`, la biblioteca oficial
persistencia.go   que es estado y que no; guardar agrupando, comparando y de una pieza
lanzador.go       levantar los contenedores hablando con Docker por su socket (HTTP a secas)
disparadores.go   S3 avisa y SQS se sondea: quien invoca tu funcion cuando tu no llamas
temas.go          SNS: la diferencia entre una cola y un tema, y el sobre
sns.go            SNS: el protocolo Query, con formulario en el cuerpo
permisos.go       las tres reglas, el comodin, y leer una politica de verdad
autorizar.go      lo dificil: traducir una peticion HTTP a un permiso
registros.go      CloudWatch Logs: grupo, flujo y apunte. Un registro NO es una cola
logs.go           CloudWatch Logs: el mismo protocolo, con los campos en minuscula
bitacora.go       de donde salen: los MARCOS de ocho bytes de la salida de un contenedor
Dockerfile     dos etapas: golang:1.26 para construir, `scratch` para correr
*_test.go      pruebas contra los SDK de AWS de VERDAD, no contra clientes de mentira
```

### Probarlo

```bash
go test ./... -count=1 -v
go test ./... -count=1 -race     # el candado no lo valida nadie sin esto
```

⚠️ **En este Windows las pruebas se corren en Docker**, no con el Go local:

```bash
docker run --rm -v "$PWD:/app" -w /app golang:1.26 go test ./... -count=1 -race
```

El Application Control de esta máquina bloquea **a ratos** el binario que `go test` deja en el
temporal («Una directiva de Control de aplicaciones bloqueó este archivo»), y una prueba que a
veces corre no sirve para nada. El servidor sí se compila y corre en Windows: tiene nombre fijo
y no lo bloquea. De regalo, en el contenedor hay cgo y por tanto **detector de carreras**, que
en este Windows (con `CGO_ENABLED=0`) no hay.

Las pruebas usan `github.com/aws/aws-sdk-go-v2/service/s3`, el mismo paquete que usarías contra
Amazon. Es la única forma honesta de decir «esto es compatible»: si el SDK oficial lo entiende,
es compatible.

La demostración completa —incluida la CLI oficial— se regenera con:

```bash
python canales/golandia/investigacion/pruebas_nubecita.py
```

y deja la salida **literal** en `canales/golandia/recursos/pruebas-nubecita/SALIDA.md`, dentro
del proyecto de la fábrica de vídeos. El **código** vive aquí, en su propia carpeta a propósito:
es una herramienta que crece por episodios y se copia sola.

---

## Las tres cosas que te vas a llevar del episodio 1

1. **La API de S3 es HTTP a secas.** El bucket es el primer trozo de la ruta, la clave es el
   resto y la respuesta es XML. Ni protocolo binario ni magia de Amazon.
2. **S3 no tiene carpetas.** Lo que parece una carpeta es el parámetro `delimiter`: las claves
   que comparten el trozo hasta la siguiente barra se colapsan en un «prefijo común».
3. **El error importa tanto como el acierto.** Si devuelves un 404 pelado en vez del XML con
   `<Code>NoSuchKey</Code>`, el SDK del que te llama no puede distinguir «no existe» de «se cayó
   la red». Un emulador que falla mal es peor que no tener emulador.

---

## Las tres del episodio 4

1. **Firmar no es cifrar.** La petición viaja en claro. La firma dice quién eres y que nadie la
   tocó; el secreto no viaja nunca.
2. **`SignatureDoesNotMatch` casi siempre es el paso 1.** La petición canónica no cuadra: la
   ruta, el orden de la consulta, o una cabecera firmada que no viene.
3. **La firma cubre el resumen del cuerpo, no el cuerpo.** Hay que comprobar los dos, o cambiar
   el cuerpo dejando la cabecera intacta pasa desapercibido.

---

## Las tres del episodio 3

1. **Recibir NO saca el mensaje de la cola: lo esconde.** Si no lo borras dentro de la ventana de
   visibilidad, vuelve. De ahí que la entrega sea **al menos una vez** y de ahí que tu consumidor
   tenga que ser idempotente: procesar dos veces no puede cobrar dos veces.
2. **El `ReceiptHandle` cambia en cada entrega.** No es el identificador del mensaje. El fallo
   clásico del consumidor lento: borra con un recibo caducado, cree que borró, y el mensaje sigue.
3. **La cola de mensajes muertos no es magia:** es un contador de entregas y un sitio donde
   apartar. Sin ella, un mensaje que siempre falla se reintenta para siempre.

---

## Las tres del episodio 2

1. **La misma nube, protocolos distintos.** Que S3 sea REST-con-XML y DynamoDB sea JSON-RPC no
   es un descuido: es historia. Y explica por qué el SDK es un paquete por servicio.
2. **Los números viajan como cadena** (`{"N":"42.5"}`) para no pasar por el float de ningún
   lenguaje. Un identificador de 20 dígitos vuelve intacto; con JSON normal, no.
3. **`Query` solo sabe de la clave, y `Scan` lee la tabla entera.** Por eso la clave se elige una
   vez y con cuidado, y por eso las fechas van en ISO-8601: el orden de bytes tiene que coincidir
   con el orden real.

---

## Las tres del episodio 5

1. **Tu función no recibe peticiones: las va a buscar.** La API del runtime es un `GET` que se
   queda abierto hasta que hay trabajo. Cuando ves eso, `aws-lambda-go` deja de parecer magia.
2. **El arranque en frío es una espera en un canal**, no un impuesto. Y por eso la segunda
   invocación es rápida: el contenedor sigue ahí, con sus variables de paquete y todo.
3. **Un fallo de tu función llega con un `200`.** Si tu cliente solo mira el código HTTP, no te
   vas a enterar hasta que alguien pregunte por qué faltan datos.

---

## Las tres del episodio 6

1. **No todo lo que está en memoria es estado.** El recibo de un mensaje en vuelo es una
   promesa hecha a alguien que ya no existe: al restaurar, el mensaje vuelve visible.
2. **No adivines qué operaciones escriben: compara.** Y para poder comparar, ordena — el
   recorrido de un mapa en Go es aleatorio y sin ordenar el JSON cambia solo.
3. **Guardar al salir no es guardar.** `docker kill` no avisa. Se guarda cuando algo cambia, a
   un temporal, y se renombra.

---

## Las tres del episodio 7

1. **La API de Docker es HTTP.** Un socket y `net/http`. Cambias cómo se abre la conexión y ya
   estás hablando con el demonio, sin SDK ni dependencias.
2. **La regla del escalado cabe en tres líneas.** «Si no hay nadie libre, levanta otro» produce
   por sí sola el comportamiento que te facturan.
3. **Un protocolo HTTP se puede fingir.** El demonio de mentira de las pruebas son veinte líneas
   de `httptest`, y con él se comprueba todo lo que se le pide a Docker sin tocar Docker.

---

## Las tres del episodio 10

1. **Un `Deny` gana siempre**, y por eso se usa como red de seguridad: nada lo puede levantar.
2. **El comodín sabe de letras, no de servicios.** `s3:Get*` deja pasar cosas que no esperabas.
3. **Lo difícil no es decidir si puedes: es saber qué estás pidiendo.** Y para eso, a veces hay
   que leer el cuerpo de la petición y volver a dejarlo puesto.

---

## Las tres del episodio 9

1. **Un tema no guarda: reparte.** Publicar sin suscriptores no pierde el mensaje; es que nunca
   existió. Por eso el patrón es tema **que reparte a colas**.
2. **Sin `RawMessageDelivery` llega un sobre**, y lo tuyo va dentro como cadena: se deserializa
   dos veces.
3. **El filtro lo pone el suscriptor**, y si el mensaje no trae el atributo, no pasa.

---

## Las tres del episodio 8

1. **Un disparador de cola es un `for`.** Hace `ReceiveMessage` por ti y te invoca con lo que
   saca. Cuando lo ves, deja de parecer infraestructura y pasa a ser código.
2. **Tu función no borra nada.** Lo borra el disparador, y solo si fuiste bien. De ahí sale el
   reintento gratis, y de ahí sale también que tu consumidor tenga que ser idempotente.
3. **Si no dices qué falló, vuelve todo.** Un lote de diez con uno malo se reintenta entero.

---

## Lo que hay, y lo que falta

Cada episodio se puede seguir por separado, y cada uno añade un servicio al mismo binario.

| ep | servicio | lo que se aprende de verdad |
|---:|---|---|
| **1** | **S3** ✅ | HTTP + XML, enrutado de la estándar, que S3 no tiene carpetas |
| **2** | **DynamoDB** ✅ | otro protocolo en la misma nube: JSON con la operación en la cabecera. La clave se elige una vez. Los números viajan como cadena. El orden es de bytes, salvo si la clave es numérica |
| **3** | **SQS** ✅ | la visibilidad: recibir no saca el mensaje, lo esconde. Por eso la entrega es «al menos una vez» y tu consumidor tiene que ser idempotente |
| **4** | **SigV4** ✅ | qué significa de verdad firmar: no es cifrar. La cadena de HMAC, la petición canónica y por qué el error nunca dice el motivo |
| **5** | **Lambda** ✅ | las DOS APIs de Lambda. El arranque en frío es esperar en un canal. Y tu función fallando con un `200` |
| **6** | **persistencia** ✅ | qué es estado y qué es una conversación a medias. Guardar sin adivinar, y por qué hay que ordenar antes |
| **7** | **el lanzador** ✅ | la API de Docker es HTTP. Tres líneas de regla y sale el escalado de Lambda entero |
| **8** | **los disparadores** ✅ | subes un fichero y se ejecuta código. Tu función no borra el mensaje, y los fallos parciales |
| **9** | **SNS** ✅ | cola contra tema. El sobre en el que llega tu mensaje, y el filtro que lo pone el suscriptor |
| **10** | **los permisos** ✅ | la otra mitad del 4. Un Deny gana siempre, el comodín sabe de letras, y por qué este error sí dice qué falta |
| — | **CloudWatch Logs** ✅ | sin vídeo. Un registro no es una cola, la salida de un contenedor viene en marcos, y el `nextToken` que no se acaba nunca |


### Lo que falta, por orden de lo que más aporta

El criterio no es «qué servicios tiene AWS» —son doscientos— sino **cuál enseña algo que no se
puede contar sin escribirlo**, y cuál convierte esta pila de servicios sueltos en una
arquitectura.

#### ✅ Los tres que convertían esto en una nube de verdad, ya están

**Los disparadores** (episodio 8), **SNS** (episodio 9) y **los permisos** (episodio 10). Lo que
queda son piezas más acotadas:

| qué | por qué este y no otro | lo que se aprendería |
|---|---|---|

✅ **CloudWatch Logs ya está** (arriba), aunque sin vídeo: cerraba el hueco que dejó el episodio
7. Queda esto:

#### 🥈 Los que valen mucho y son más acotados

| qué | lo que se aprendería |
|---|---|
| **API Gateway** (HTTP API) | rutas → función. Por qué tu manejador devuelve un JSON con `statusCode` **dentro** del cuerpo, y qué es `isBase64Encoded` |
| **DynamoDB Streams** | media pieza ya está (la tabla existe). El **registro ordenado por partición**, el iterador y el punto de control. Es donde se ve la diferencia entre una cola y un log |
| **Secrets Manager / Parameter Store** | pequeño, y con una lección real: el versionado (`AWSCURRENT` / `AWSPREVIOUS`), la rotación, y por qué el SDK **cachea** y tú no te enteras |

#### 🥉 Grandes, vistosos y caros de hacer

- **Step Functions** — la máquina de estados: reintentos, `Catch`, paralelo. Mucho vídeo, mucho trabajo.
- **Kinesis** — parecido a DynamoDB Streams pero con particiones a mano.

#### ⛔ Lo que NO merece la pena imitar

**EC2, VPC, RDS, ECS.** No se pueden emular de forma honesta en un binario —o son una máquina
virtual, o son mentira— y lo que enseñarían no es un protocolo, es una consola. Un doble de
pruebas vale por lo que **no** hace.
