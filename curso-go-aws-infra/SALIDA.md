# SALIDA — curso «Infraestructura de AWS desde Go: SDK y CDK»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-09-28 11:38:14Z`
- go version go1.26.5 windows/amd64
- emulador: moto `5.2.3.dev` (imagen `motoserver/moto:latest`), en `http://localhost:5000`
- SDK: `aws-sdk-go-v2/service/ec2 v1.336.1` · CDK: CLI `2.1136.0 (build 536ad69)`, `aws-cdk-go/awscdk/v2 v2.271.0`, node `v24.18.0`
- sistema: `Windows 10`

## Leccion 01 — En que cuenta y region estoy

`go run ./01_donde_estoy` · codigo de salida `0`

```
== 1.1 · Lo primero: quien soy ==
  cuenta:   123456789012
  quien:    arn:aws:sts::123456789012:user/moto
  region:   us-east-1
  endpoint: http://localhost:5000   <- el emulador; sin esta variable, AWS de verdad

== 1.2 · Tres servidores en us-east-1 ==
  us-east-1  -> 3 instancias   (err = <nil>)
  us-west-2  -> 0 instancias   (err = <nil>)
  eu-west-1  -> 0 instancias   (err = <nil>)
  Cero no es «no hay»: es «no hay AQUI».

== 1.3 · La salvaguarda: si la cuenta no es la esperada, no toco nada ==
  cuenta 123456789012 -> ok, sigo
  cuenta 987654321098 -> ⛔ esperaba 123456789012: no hago nada
```

### Y sin el emulador

`AWS_ENDPOINT_URL= go run ./01_donde_estoy` · codigo de salida `1`

```
2026/09/28 06:38:18 ⛔ AWS_ENDPOINT_URL="" no es un emulador local. Este curso para y borra instancias: no lo ejecuto contra una cuenta de verdad.
exit status 1
```

## Leccion 02 — Inventario: reservas, instancias y paginas

`go run ./02_paginacion` · codigo de salida `0`

```
  lanzadas: 1 llamada con 3 instancias + 9 llamadas con 1 = 12 instancias

== 2.1 · Contar reservas no es contar instancias ==
  len(out.Reservations)          -> 10   <- lo que cuenta mucha gente
  instancias dentro de ellas     -> 12

== 2.2 · Una sola pagina (MaxResults=5, el minimo que admite la API) ==
  instancias: 7 de 12   err = <nil>   NextToken vacio? false
  Ni un error. Solo un token que hay que seguir.

== 2.3 · Con el paginador del SDK ==
  instancias: 12 de 12
```

## Leccion 03 — Filtros y etiquetas

`go run ./03_filtros` · codigo de salida `0`

```
  6 instancias: 3 prod, 2 dev, y 1 que alguien etiqueto a mano (Env, sin dueno)

== 3.1 · El mismo filtro, escrito de tres maneras ==
  tag:env = prod  -> 3
  tag:Env = prod  -> 0
  tag:env = Prod  -> 0

== 3.2 · Cuantas son de dev? ==
  tag:env = dev   -> 2    (la de Env no sale)

== 3.3 · Varios filtros = Y; varios valores = O ==
  env=dev,prod                   -> 5
  env=prod Y dueno=pagos         -> 3

== 3.4 · «Sin dueno» no es un filtro: se mira en Go ==
  i-5be724f50153024ee  Name=prueba-juan  -> sin etiqueta dueno
```

## Leccion 04 — Crear sin huerfanos ni duplicados

`go run ./04_crear` · codigo de salida `0`

```
== 4.1 · Crear y DESPUES etiquetar; el primer intento «da timeout» ==
  intento 1 -> timeout esperando la respuesta
  intento 2 -> <nil>
  en la cuenta:
    i-44e1c641b02400355  Name=-
    i-452cc7e7b11668df4  Name=web-nueva

== 4.2 · Buscar por nombre antes de crear, etiquetas EN la creacion ==
  intento 1 -> creada i-6251440dcd22901d4
  intento 2 -> ya existe i-6251440dcd22901d4: no creo nada
  en la cuenta:
    i-6251440dcd22901d4  Name=web-nueva

== 4.3 · Recien creada no es lista ==
  estado al volver de RunInstances: pending
  tras NewInstanceRunningWaiter:    running   (err = <nil>, 0.0 s en el emulador)
```

## Leccion 05 — Grupos de seguridad y errores de la API

`go run ./05_grupos_seguridad` · codigo de salida `0`

```
== 5.1 · Abrir 443 al mundo en web, dos veces ==
  primera vez -> <nil>
  segunda vez -> operation error EC2: AuthorizeSecurityGroupIngress, https response error StatusCode: 400, RequestID: fENRfzvh1QM4iYu6I9VnkYAtyrV4d18MsKhbHOX02wK9oA6geTXx, api error InvalidPermission.Duplicate: The specified rule already exists

== 5.2 · El codigo del error, con errors.As y smithy.APIError ==
  comun.Codigo(err) -> InvalidPermission.Duplicate
  abrir() idempotente, segunda vez -> <nil>

== 5.3 · Auditoria: puertos de administracion abiertos al mundo ==
  ⛔ bastion  sg-e86f35d2c2fd4f99d   puerto 22    (ssh) abierto a 0.0.0.0/0
  ⛔ db       sg-739f008d36e83bc06   puerto 3389  (rdp) abierto a 0.0.0.0/0

== 5.4 · Cerrar ==
  revocar 22 en bastion -> <nil>
```

## Leccion 06 — Una red desde cero, y el orden para borrarla

`go run ./06_vpc` · codigo de salida `0`

```
== 6.1 · Crear, en orden ==
  1. CreateVpc 10.20.0.0/16                     -> <nil>
  2. CreateSubnet 10.20.1.0/24                  -> <nil>
  3. CreateInternetGateway                      -> <nil>
  4. AttachInternetGateway                      -> <nil>
  5. CreateRouteTable                           -> <nil>
  6. CreateRoute 0.0.0.0/0 -> internet gateway  -> <nil>
  7. AssociateRouteTable con la subred          -> <nil>

== 6.2 · Borrar empezando por la VPC ==
  DeleteVpc -> DependencyViolation
   operation error EC2: DeleteVpc, https response error StatusCode: 400, RequestID: ttxQuDSR8QIHgD65hNYuIXLyn1hlSNHENL8UllGkavBNwXZCcd2Y, api error DependencyViolation: The vpc 'vpc-336f0144ad7729fe3' has dependencies and cannot be deleted.

== 6.3 · Borrar al reves de como se creo ==
  7. DisassociateRouteTable                     -> <nil>
  6. DeleteRouteTable (con su ruta)             -> <nil>
  4. DetachInternetGateway                      -> <nil>
  3. DeleteInternetGateway                      -> <nil>
  2. DeleteSubnet                               -> <nil>
  1. DeleteVpc                                  -> <nil>

  Seis recursos para una red minima. Esta cuenta es la que hace el CDK por ti (leccion 9).
```

## Leccion 07 — DryRun: el error que significa adelante

`go run ./07_dry_run` · codigo de salida `0`

```
== 7.1 · StopInstances con DryRun = true ==
  err -> operation error EC2: StopInstances, https response error StatusCode: 412, RequestID: zeRu35Q4WxyxEZQ9edDeN56rDLjAmWayOFQVMzJjEoRS9Mn0Iovd, api error DryRunOperation: Request would have succeeded, but DryRun flag is set.
  codigo -> DryRunOperation

== 7.2 · Leerlo bien ==
  podria parar 2 instancias? true   (err = <nil>)
  i-2ab9bb224f3594c36 sigue running
  i-115fd0055323618f9 sigue running

== 7.3 · Lo que DryRun NO comprueba ==
  Que las instancias sean las que tu querias. Eso lo decide tu filtro (leccion 8).
```

## Leccion 08 — El filtro vacio que selecciona TODO

`go run ./08_filtro_vacio` · codigo de salida `0`

```
== 8.1 · parar(ENTORNO) con ENTORNO=dev ==
    parando i-f15e417cd1247574b  Name=pruebas  env=dev
    parando i-56b51416c805114b4  Name=pruebas  env=dev
    parando i-252d410ea144948b9  Name=pruebas  env=dev
    prod  running=3 stopped=0
    dev   running=0 stopped=3

== 8.2 · El mismo script, y la variable llega vacia ==
    parando i-dee1f8b50846f636e  Name=pagos    env=prod
    parando i-4fc068279d32c051a  Name=pagos    env=prod
    parando i-729733ac485079959  Name=pagos    env=prod
    parando i-ac42f7aec6604dc43  Name=pruebas  env=dev
    parando i-676054fba72b8bac2  Name=pruebas  env=dev
    parando i-ace8836cb65f0c581  Name=pruebas  env=dev
    prod  running=0 stopped=3
    dev   running=0 stopped=3

== 8.3 · Con tres protecciones ==
  entorno vacio -> sin entorno no paro nada: diria que TODAS
  entorno dev, sin -aplicar:
    PLAN (sin -aplicar) i-f210fdc580456103e  Name=pruebas
    PLAN (sin -aplicar) i-bc02a60701ae62266  Name=pruebas
    PLAN (sin -aplicar) i-866ae337fe72458da  Name=pruebas
    prod  running=3 stopped=0
    dev   running=3 stopped=0
```

## Leccion 09 — CDK en Go: lo que escribe una linea

`go run ./09_cdk_synth` · codigo de salida `0`

```
== 9.1 · awsec2.NewVpc(pila, "Vpc", nil) ==
  $ cdk synth -q
    Successfully synthesized to <repo>\curso-go-aws-infra\cdk\cdk.out
    Supply a stack id (Red, Datos, Web) to display its template.
  23 recursos de AWS:
     4  AWS::EC2::Route
     4  AWS::EC2::RouteTable
     4  AWS::EC2::Subnet
     4  AWS::EC2::SubnetRouteTableAssociation
     2  AWS::EC2::EIP
     2  AWS::EC2::NatGateway
     1  AWS::EC2::InternetGateway
     1  AWS::EC2::VPC
     1  AWS::EC2::VPCGatewayAttachment
  NAT gateways: 2  ->  2 x 0.045 USD/h x 730 h = 65.70 USD al mes, sin trafico

== 9.2 · La misma VPC, sin NAT: subredes publicas y aisladas ==
  $ cdk synth -q -c nat=0
    Successfully synthesized to <repo>\curso-go-aws-infra\cdk\cdk.out
    Supply a stack id (Red, Datos, Web) to display its template.
  17 recursos de AWS:
     4  AWS::EC2::RouteTable
     4  AWS::EC2::Subnet
     4  AWS::EC2::SubnetRouteTableAssociation
     2  AWS::EC2::Route
     1  AWS::EC2::InternetGateway
     1  AWS::EC2::VPC
     1  AWS::EC2::VPCGatewayAttachment
  NAT gateways: 0  ->  0 x 0.045 USD/h x 730 h = 0.00 USD al mes, sin trafico
```

## Leccion 10 — cdk diff: cambiar un nombre no es renombrar

`go run ./10_cdk_diff` · codigo de salida `0`

```
== 10.1 · Desplegar la pila Datos, con el bucket «Registros» ==
  $ cdk deploy Datos --method=direct --require-approval never
    Datos: deploying... [1/1]
    Datos: creating stack...
    Datos |   0 | 6:38:44 a. m. | CREATE_IN_PROGRESS      | AWS::CloudFormation::Stack | Datos User Initiated
    Datos |   1 | 6:38:44 a. m. | CREATE_COMPLETE         | AWS::CloudFormation::Stack | Datos 
    ✅  Datos
  buckets que existen: [datos-registros009e72cf-7recxjg0ar0d]

== 10.2 · En el codigo, el id pasa a ser «Logs». Antes de desplegar: cdk diff ==
  $ cdk diff Datos -c id=Logs --method=template
    Stack Datos
    Resources
    [-] AWS::S3::Bucket Registros Registros009E72CF orphan
    [+] AWS::S3::Bucket Logs Logs6819BB44

== 10.3 · Por que «orphan»: lo que dice la plantilla del bucket ==
  Logs6819BB44  DeletionPolicy=Retain  UpdateReplacePolicy=Retain
```

## Leccion 11 — cdk destroy no borra todo: RemovalPolicy

`go run ./11_cdk_destroy` · codigo de salida `0`

```
== 11.1 · Por defecto: RETAIN ==
  $ cdk deploy Datos --method=direct --require-approval never
    Datos: deploying... [1/1]
    Datos: creating stack...
    Datos |   1 | 6:38:55 a. m. | CREATE_COMPLETE         | AWS::CloudFormation::Stack | Datos 
    Datos |   1 | 6:38:55 a. m. | CREATE_IN_PROGRESS      | AWS::CloudFormation::Stack | Datos User Initiated
    ✅  Datos
  plantilla: Registros009E72CF  DeletionPolicy=Retain
  $ cdk destroy Datos -f
    Datos: destroying... [1/1]
    Datos |   1 | 6:38:59 a. m. | DELETE_COMPLETE         | AWS::CloudFormation::Stack | Datos 
    Datos |   1 | 6:38:59 a. m. | DELETE_IN_PROGRESS      | AWS::CloudFormation::Stack | Datos User Initiated
    ✅  Datos: destroyed
  buckets que siguen existiendo: [datos-registros009e72cf-j1pvo3ygtnro]

== 11.2 · Con RemovalPolicy_DESTROY (-c borrar=1) ==
  $ cdk deploy Datos -c borrar=1 --method=direct --require-approval never
    Datos: deploying... [1/1]
    Datos: creating stack...
    Datos |   1 | 6:39:05 a. m. | CREATE_COMPLETE         | AWS::CloudFormation::Stack | Datos 
    Datos |   1 | 6:39:05 a. m. | CREATE_IN_PROGRESS      | AWS::CloudFormation::Stack | Datos User Initiated
    ✅  Datos
  plantilla: Registros009E72CF  DeletionPolicy=Delete
  $ cdk destroy Datos -c borrar=1 -f
    Datos: destroying... [1/1]
    Datos |   1 | 6:39:10 a. m. | DELETE_COMPLETE         | AWS::CloudFormation::Stack | Datos 
    Datos |   1 | 6:39:10 a. m. | DELETE_IN_PROGRESS      | AWS::CloudFormation::Stack | Datos User Initiated
    ✅  Datos: destroyed
  buckets que siguen existiendo: []
  (con datos dentro, S3 se niega a borrar el bucket: haria falta AutoDeleteObjects)
```

## Leccion 12 — La deriva que cdk diff no ve

`go run ./12_deriva` · codigo de salida `0`

```
== 12.1 · Desplegar la pila Web: un grupo con solo el 443 ==
  $ cdk deploy Web --method=direct --require-approval never
    Web: deploying... [1/1]
    Web: creating stack...
    Web |   0 | 6:39:15 a. m. | CREATE_IN_PROGRESS      | AWS::CloudFormation::Stack            | Web User Initiated
    Web |   1 | 6:39:15 a. m. | CREATE_COMPLETE         | AWS::CloudFormation::Stack            | Web 
    ✅  Web

== 12.2 · Alguien abre el 22 «un momento», a mano ==
  AuthorizeSecurityGroupIngress 22 0.0.0.0/0 -> <nil>

== 12.3 · cdk diff ==
  $ cdk diff Web --method=template
    Stack Web
    There were no differences

== 12.4 · Preguntar a la realidad: plantilla contra cuenta ==
  declaradas: [tcp/443 desde 0.0.0.0/0]
  reales:     [tcp/22 desde 0.0.0.0/0 tcp/443 desde 0.0.0.0/0]
  ⛔ DERIVA en sg-e5d68a2a43063d5ef: tcp/22 desde 0.0.0.0/0 no esta en el codigo
```

## Leccion 13 — El proyecto: inventario

### El escenario

`go run ./13_proyecto/escenario` · codigo de salida `0`

```
escenario listo
```

### Con hallazgos

`inventario -cuenta 123456789012` · codigo de salida `1`

```
Cuenta 123456789012
  us-east-1  4 instancias
  us-west-2  2 instancias
  us-east-1  sin-dueno         i-0aba9220bd4af99d2    Name=prueba-juan
  us-east-1  abierto-al-mundo  sg-60ae4d690a4be4f29   bastion: 22 (ssh) desde 0.0.0.0/0
  us-east-1  volumen-suelto    vol-7ddb7b168ba7a8c81  100 GiB, sin instancia
  us-east-1  ip-sin-usar       eipalloc-6b462855a3ffcd721 127.109.77.113 reservada y sin asociar
  us-east-1  nat-gateway       nat-1aea764a49374ab3f  se cobra por hora, pase trafico o no
level=WARN msg="hay hallazgos" total=5
```

### Lo mismo, en JSON

`inventario -cuenta 123456789012 -json` · codigo de salida `1`

```
{
  "cuenta": "123456789012",
  "instancias_por_region": {
    "us-east-1": 4,
    "us-west-2": 2
  },
  "hallazgos": [
    {
      "region": "us-east-1",
      "tipo": "sin-dueno",
      "recurso": "i-0aba9220bd4af99d2",
      "detalle": "Name=prueba-juan"
    },
    {
      "region": "us-east-1",
      "tipo": "abierto-al-mundo",
      "recurso": "sg-60ae4d690a4be4f29",
      "detalle": "bastion: 22 (ssh) desde 0.0.0.0/0"
    },
    {
      "region": "us-east-1",
      "tipo": "volumen-suelto",
      "recurso": "vol-7ddb7b168ba7a8c81",
      "detalle": "100 GiB, sin instancia"
    },
    {
      "region": "us-east-1",
      "tipo": "ip-sin-usar",
      "recurso": "eipalloc-6b462855a3ffcd721",
      "detalle": "127.109.77.113 reservada y sin asociar"
    },
    {
      "region": "us-east-1",
      "tipo": "nat-gateway",
      "recurso": "nat-1aea764a49374ab3f",
      "detalle": "se cobra por hora, pase trafico o no"
    }
  ]
}
level=WARN msg="hay hallazgos" total=5
```

### Una region en orden

`inventario -regiones us-west-2` · codigo de salida `0`

```
Cuenta 123456789012
  us-west-2  2 instancias
  sin hallazgos
```

### La cuenta no es la esperada

`inventario -cuenta 987654321098` · codigo de salida `2`

```
level=ERROR msg="cuenta inesperada" esperaba=987654321098 es=123456789012
```
