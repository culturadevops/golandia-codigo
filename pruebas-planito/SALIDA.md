# Pruebas ejecutadas — planito, un Terraform de bolsillo en Go

> Salida **literal** de esta máquina. La genera `investigacion/pruebas_planito.py` y se
> vuelve a generar sola. ⛔ Lo que no esté aquí no puede salir en pantalla.

**Entorno citable:** `go version go1.26.7 linux/amd64` (dentro de `golang:1.26`) · `aws-sdk-go-v2` para `s3`, `dynamodb` y `sqs`

**El código vive en `c:\p\planito`** y habla con `nubecita` (`<nubecita>`) por una red de
Docker. La demo entera es gratis y se destruye sola.

**La imagen de planito se descarga en 3.633.559 bytes** (3.47 MiB). Segunda etapa `scratch`, igual que nubecita.

---

## 1 · `go vet` y las pruebas del corazón

Fíjate en lo que **no** hay en estas pruebas: red, contenedores, credenciales ni esperas.
`Calcular` recibe tres estructuras y devuelve una cuarta, porque toda la red vive en
`nube.go`. Por eso la lógica que de verdad se puede equivocar se prueba en milisegundos.

```console
$ docker run --rm -v c:/p/planito:/app -v nubecita-gomod:/go/pkg/mod -w /app -e GOFLAGS=-buildvcs=false golang:1.26 go vet ./...

[codigo de salida: 0]
```

```console
$ docker run --rm -v c:/p/planito:/app -v nubecita-gomod:/go/pkg/mod -w /app -e GOFLAGS=-buildvcs=false golang:1.26 go test ./... -count=1 -v
=== RUN   TestLoQueFaltaSeCrea
--- PASS: TestLoQueFaltaSeCrea (0.00s)
=== RUN   TestAplicarDosVecesNoHaceNada
--- PASS: TestAplicarDosVecesNoHaceNada (0.00s)
=== RUN   TestNoSeTocaoLoQueNoCreamos
--- PASS: TestNoSeTocaoLoQueNoCreamos (0.00s)
=== RUN   TestLoQueSobraSeBorraSoloSiLoCreamos
--- PASS: TestLoQueSobraSeBorraSoloSiLoCreamos (0.00s)
=== RUN   TestCambiarLaClaveDeUnaTablaAvisaDeQueSePierdenLosDatos
--- PASS: TestCambiarLaClaveDeUnaTablaAvisaDeQueSePierdenLosDatos (0.00s)
=== RUN   TestDestruirSoloSeLlevaLoNuestro
--- PASS: TestDestruirSoloSeLlevaLoNuestro (0.00s)
=== RUN   TestLaColaDeMuertosSeCreaAntes
--- PASS: TestLaColaDeMuertosSeCreaAntes (0.00s)
=== RUN   TestValidarCazaLosErroresDelFicheroSinLlamarANadie
=== RUN   TestValidarCazaLosErroresDelFicheroSinLlamarANadie/bucket_repetido
=== RUN   TestValidarCazaLosErroresDelFicheroSinLlamarANadie/tabla_sin_clave
=== RUN   TestValidarCazaLosErroresDelFicheroSinLlamarANadie/muertos_que_no_existe
=== RUN   TestValidarCazaLosErroresDelFicheroSinLlamarANadie/muertos_sin_maxRecibos
=== RUN   TestValidarCazaLosErroresDelFicheroSinLlamarANadie/cola_que_se_apunta_a_sí_misma
--- PASS: TestValidarCazaLosErroresDelFicheroSinLlamarANadie (0.00s)
    --- PASS: TestValidarCazaLosErroresDelFicheroSinLlamarANadie/bucket_repetido (0.00s)
    --- PASS: TestValidarCazaLosErroresDelFicheroSinLlamarANadie/tabla_sin_clave (0.00s)
    --- PASS: TestValidarCazaLosErroresDelFicheroSinLlamarANadie/muertos_que_no_existe (0.00s)
    --- PASS: TestValidarCazaLosErroresDelFicheroSinLlamarANadie/muertos_sin_maxRecibos (0.00s)
    --- PASS: TestValidarCazaLosErroresDelFicheroSinLlamarANadie/cola_que_se_apunta_a_sí_misma (0.00s)
=== RUN   TestElEstadoVaYVuelve
--- PASS: TestElEstadoVaYVuelve (0.00s)
=== RUN   TestSinEstadoTodaviaNoEsUnError
--- PASS: TestSinEstadoTodaviaNoEsUnError (0.00s)
=== RUN   TestUnCampoMalEscritoFalla
--- PASS: TestUnCampoMalEscritoFalla (0.00s)
PASS
ok  	planito	0.004s
[codigo de salida: 0]
```

### Con el detector de carreras

```console
$ docker run --rm -v c:/p/planito:/app -v nubecita-gomod:/go/pkg/mod -w /app -e GOFLAGS=-buildvcs=false golang:1.26 go test ./... -count=1 -race
ok  	planito	1.013s
[codigo de salida: 0]
```

---

## 2 · La herramienta, de principio a fin

El fichero de entrada:

```json
{
  "buckets": ["fotos", "copias"],
  "tablas": [
    {"nombre": "pedidos", "particion": "cliente", "orden": "fecha"}
  ],
  "colas": [
    {"nombre": "trabajos", "muertos": "muertos", "maxRecibos": 3},
    {"nombre": "muertos"}
  ]
}
```

```console
$ planito plan -f infra.json -endpoint http://nubecita:4566
planito · 0 recurso(s) en el estado · hablando con http://nubecita:4566

plan:
  + bucket  copias        · no existe
  + bucket  fotos         · no existe
  + tabla   pedidos       · no existe
  + cola    muertos       · no existe
  + cola    trabajos      · no existe

5 cambio(s). Esto NO ha tocado nada.
[codigo de salida: 0]
```

> 🥇 `plan` no toca nada. Y fíjate en el orden: la cola de muertos se crea **antes** que la cola que la nombra. Es la única dependencia del modelo, y se resuelve ordenando.

```console
$ planito aplicar -f infra.json -endpoint http://nubecita:4566
planito · 0 recurso(s) en el estado · hablando con http://nubecita:4566

plan:
  + bucket  copias        · no existe
  + bucket  fotos         · no existe
  + tabla   pedidos       · no existe
  + cola    muertos       · no existe
  + cola    trabajos      · no existe

5 cambio(s). Vuelve a lanzarlo con -si para hacerlo.
[codigo de salida: 0]
```

> ⚠️ Sin `-si` no hace nada. Una herramienta que toca infraestructura sin confirmación acaba borrando lo que no debía.

```console
$ planito aplicar -f infra.json -si -endpoint http://nubecita:4566
planito · 0 recurso(s) en el estado · hablando con http://nubecita:4566

plan:
  + bucket  copias        · no existe
  + bucket  fotos         · no existe
  + tabla   pedidos       · no existe
  + cola    muertos       · no existe
  + cola    trabajos      · no existe

aplicando:
  crear bucket copias
  crear bucket fotos
  crear tabla pedidos
  crear cola muertos
  crear cola trabajos

listo. Estado en planito.estado.json
[codigo de salida: 0]
```

```console
$ planito plan -f infra.json -endpoint http://nubecita:4566
planito · 5 recurso(s) en el estado · hablando con http://nubecita:4566
sin cambios: lo que hay ya es lo que quieres
[codigo de salida: 0]
```

> 🥇 Idempotencia: la segunda pasada no propone nada. Es la propiedad que define a una herramienta declarativa.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 ls
2026-09-09 11:33:11 copias
2026-09-09 11:33:11 fotos
[codigo de salida: 0]
```

> Y ahí están, vistos por el cliente oficial de Amazon.

```console
$ planito plan -f infra-clave.json -endpoint http://nubecita:4566
planito · 5 recurso(s) en el estado · hablando con http://nubecita:4566

plan:
  ~ tabla   pedidos       · la clave cambia de (cliente, fecha) a (id): hay que recrearla y SE PIERDEN LOS DATOS

1 cambio(s). Esto NO ha tocado nada.
[codigo de salida: 0]
```

> 🥇 La clave de una tabla de DynamoDB no se puede modificar. Cambiarla significa recrearla, o sea **perder los datos** — y la herramienta lo dice con todas las letras en vez de hacerlo callando.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 mb s3://de-otro-equipo
make_bucket: de-otro-equipo
[codigo de salida: 0]
```

> Ahora otra persona crea un bucket **sin usar planito**. Esto es lo que pasa en cualquier cuenta real.

```console
$ planito destruir -si -endpoint http://nubecita:4566
planito · 5 recurso(s) en el estado · hablando con http://nubecita:4566

plan:
  - bucket  copias        · lo creamos y ya no lo quieres
  - bucket  fotos         · lo creamos y ya no lo quieres
  - tabla   pedidos       · la creamos y ya no la quieres
  - cola    trabajos      · la creamos y ya no la quieres
  - cola    muertos       · la creamos y ya no la quieres

aplicando:
  borrar bucket copias
  borrar bucket fotos
  borrar tabla pedidos
  borrar cola trabajos
  borrar cola muertos

listo. Estado en planito.estado.json
[codigo de salida: 0]
```

> 🥇 `destruir` se lleva solo lo que planito creó, y las colas en orden **inverso** al de creación. Lo de la otra persona ni aparece en el plan.

```console
$ aws --endpoint-url http://127.0.0.1:4566 s3 ls
2026-09-09 11:33:14 de-otro-equipo
[codigo de salida: 0]
```

> 🥇 Ahí sigue `de-otro-equipo`. Eso es lo que compra el fichero de estado: sin él, «lo que sobra» habría incluido lo ajeno.

---

## 3 · El fichero de estado que quedó

Es la lista de «esto lo hice yo», y es lo **único** que autoriza a borrar. Se escribe de
forma atómica (temporal + renombrado) y después de CADA paso, no al final: si el proceso
muere a mitad, lo ya creado tiene que estar apuntado o nadie lo borraría nunca.

```json
{
  "version": 1,
  "actualizado": "2026-09-09T16:33:14.363928044Z",
  "buckets": [],
  "tablas": [],
  "colas": []
}
```

---

## 4 · Lo que vio la nube

El registro de `nubecita` durante toda la demo: se ve a planito leyendo el estado real
(los tres `List...`) antes de cada plan, y luego creando y borrando.

```console
GET /
  -> 200 0 bucket(s)
GET /?x-id=ListBuckets
  -> 200 0 bucket(s)
POST / [ListTables] 2 byte(s)
  -> 200 0 tabla(s)
POST / [SQS ListQueues] 2 byte(s)
  -> 200 0 cola(s)
GET /?x-id=ListBuckets
  -> 200 0 bucket(s)
POST / [ListTables] 2 byte(s)
  -> 200 0 tabla(s)
POST / [SQS ListQueues] 2 byte(s)
  -> 200 0 cola(s)
GET /?x-id=ListBuckets
  -> 200 0 bucket(s)
POST / [ListTables] 2 byte(s)
  -> 200 0 tabla(s)
POST / [SQS ListQueues] 2 byte(s)
  -> 200 0 cola(s)
PUT /copias
  -> 200 creado
PUT /fotos
  -> 200 creado
POST / [CreateTable] 277 byte(s)
  -> 200 tabla "pedidos" creada
POST / [SQS CreateQueue] 39 byte(s)
  -> 200 cola "muertos" creada
POST / [SQS CreateQueue] 156 byte(s)
  -> 200 cola "trabajos" creada
GET /?x-id=ListBuckets
  -> 200 2 bucket(s)
POST / [ListTables] 2 byte(s)
  -> 200 1 tabla(s)
POST / [DescribeTable] 23 byte(s)
POST / [SQS ListQueues] 2 byte(s)
  -> 200 2 cola(s)
GET /
  -> 200 2 bucket(s)
GET /?x-id=ListBuckets
  -> 200 2 bucket(s)
POST / [ListTables] 2 byte(s)
  -> 200 1 tabla(s)
POST / [DescribeTable] 23 byte(s)
POST / [SQS ListQueues] 2 byte(s)
  -> 200 2 cola(s)
PUT /de-otro-equipo
  -> 200 creado
GET /?x-id=ListBuckets
  -> 200 3 bucket(s)
POST / [ListTables] 2 byte(s)
  -> 200 1 tabla(s)
POST / [DescribeTable] 23 byte(s)
POST / [SQS ListQueues] 2 byte(s)
  -> 200 2 cola(s)
GET /copias?list-type=2
  -> 200 0 objeto(s), 0 prefijo(s)
DELETE /copias
  -> 204 borrado
GET /fotos?list-type=2
  -> 200 0 objeto(s), 0 prefijo(s)
DELETE /fotos
  -> 204 borrado
POST / [DeleteTable] 23 byte(s)
  -> 200 tabla "pedidos" borrada
POST / [SQS GetQueueUrl] 24 byte(s)
POST / [SQS DeleteQueue] 65 byte(s)
  -> 200 cola borrada
POST / [SQS GetQueueUrl] 23 byte(s)
POST / [SQS DeleteQueue] 64 byte(s)
  -> 200 cola borrada
GET /
  -> 200 1 bucket(s)
2026/09/09 16:33:10 nubecita · S3, DynamoDB, SQS, SNS, Logs y Lambda en http://0.0.0.0:4566 · sin comprobar firmas · sin permisos · sin guardar nada · sin levantar runtimes · Ctrl+C para parar
```

---

## 5 · El código completo

### `deseado.go`

```go
package main

// El fichero DESEADO: qué quieres que exista. Nada más.
//
// Terraform usa HCL, que es un lenguaje propio con su analizador. Aquí se usa JSON, que trae la
// biblioteca estándar, porque el tema de esta herramienta NO es analizar un fichero: es lo que
// se hace DESPUÉS de leerlo. Cambiar JSON por HCL es sustituir esta función; lo demás no se
// entera.
//
//	{
//	  "buckets": ["fotos", "copias"],
//	  "tablas":  [{"nombre": "pedidos", "particion": "cliente", "orden": "fecha"}],
//	  "colas":   [{"nombre": "muertos"},
//	              {"nombre": "trabajos", "muertos": "muertos", "maxRecibos": 3}]
//	}

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Tabla struct {
	Nombre    string `json:"nombre"`
	Particion string `json:"particion"`
	Orden     string `json:"orden,omitempty"` // opcional: clave simple o compuesta
}

type Cola struct {
	Nombre     string `json:"nombre"`
	Muertos    string `json:"muertos,omitempty"`    // nombre de la cola de mensajes muertos
	MaxRecibos int    `json:"maxRecibos,omitempty"` // entregas antes de apartar
}

type Deseado struct {
	Buckets []string `json:"buckets"`
	Tablas  []Tabla  `json:"tablas"`
	Colas   []Cola   `json:"colas"`
}

func LeerDeseado(ruta string) (*Deseado, error) {
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	var d Deseado
	// ⚠️ DisallowUnknownFields: si te equivocas escribiendo un campo, quiero un error, no que
	// se ignore en silencio. Un fichero de infraestructura que se lee "a medias" es la forma
	// más cara de descubrir una errata.
	dec := json.NewDecoder(strings.NewReader(string(datos)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("%s: %w", ruta, err)
	}
	if err := d.Validar(); err != nil {
		return nil, fmt.Errorf("%s: %w", ruta, err)
	}
	d.Ordenar()
	return &d, nil
}

// Validar comprueba lo que se puede comprobar SIN tocar la nube. Todo lo que se pueda cazar
// aquí es un error que no cuesta una llamada de red ni deja medio despliegue hecho.
func (d *Deseado) Validar() error {
	vistos := map[string]bool{}
	for _, b := range d.Buckets {
		if b == "" {
			return fmt.Errorf("hay un bucket sin nombre")
		}
		if vistos["bucket:"+b] {
			return fmt.Errorf("el bucket %q está dos veces", b)
		}
		vistos["bucket:"+b] = true
	}
	for _, t := range d.Tablas {
		if t.Nombre == "" || t.Particion == "" {
			return fmt.Errorf("la tabla %q necesita nombre y clave de partición", t.Nombre)
		}
		if vistos["tabla:"+t.Nombre] {
			return fmt.Errorf("la tabla %q está dos veces", t.Nombre)
		}
		vistos["tabla:"+t.Nombre] = true
	}
	nombresCola := map[string]bool{}
	for _, c := range d.Colas {
		if c.Nombre == "" {
			return fmt.Errorf("hay una cola sin nombre")
		}
		if vistos["cola:"+c.Nombre] {
			return fmt.Errorf("la cola %q está dos veces", c.Nombre)
		}
		vistos["cola:"+c.Nombre] = true
		nombresCola[c.Nombre] = true
	}
	// 🥇 La única dependencia real de este modelo: una cola no puede apuntar a una cola de
	// muertos que no vas a crear. Cazarlo aquí evita crear la mitad y morir en la otra mitad.
	for _, c := range d.Colas {
		if c.Muertos != "" && !nombresCola[c.Muertos] {
			return fmt.Errorf("la cola %q apunta a la cola de muertos %q, que no está declarada",
				c.Nombre, c.Muertos)
		}
		if c.Muertos != "" && c.MaxRecibos <= 0 {
			return fmt.Errorf("la cola %q tiene cola de muertos pero maxRecibos es %d",
				c.Nombre, c.MaxRecibos)
		}
		if c.Muertos == c.Nombre && c.Muertos != "" {
			return fmt.Errorf("la cola %q se apunta a sí misma como cola de muertos", c.Nombre)
		}
	}
	return nil
}

// Ordenar deja el deseado en un orden estable. Sin esto, el plan sale en un orden distinto en
// cada ejecución y comparar dos planes deja de ser posible.
func (d *Deseado) Ordenar() {
	sort.Strings(d.Buckets)
	sort.Slice(d.Tablas, func(i, j int) bool { return d.Tablas[i].Nombre < d.Tablas[j].Nombre })
	// ⚠️ Las colas NO se ordenan alfabéticamente: se ordenan por dependencia. Una cola de
	// muertos tiene que existir antes que la cola que la nombra, o la creación falla.
	sort.SliceStable(d.Colas, func(i, j int) bool {
		a, b := d.Colas[i], d.Colas[j]
		if b.Muertos == a.Nombre {
			return true // a es la cola de muertos de b: a primero
		}
		if a.Muertos == b.Nombre {
			return false
		}
		return a.Nombre < b.Nombre
	})
}
```

### `estado.go`

```go
package main

// El fichero de ESTADO: qué creó esta herramienta.
//
// 🥇 Es la pieza que casi nadie menciona al explicar Terraform, y sin ella todo esto es
// PELIGROSO. Razón, en una frase:
//
//	sin estado, «lo que sobra» incluye lo que creó otra persona.
//
// Si comparas solo lo deseado con lo que hay en la nube, cualquier bucket que exista y no esté
// en tu fichero parece basura tuya. Y no lo es: puede ser de otro equipo. El estado es la lista
// de «esto lo hice yo», y es lo único que autoriza a borrar.
//
// ⚠️ Y de ahí sale el otro dolor famoso: si pierdes el estado, tu herramienta se olvida de lo
// que creó. No lo borra: se olvida. Los recursos quedan huérfanos, pagando, y sin nadie que los
// recoja. Por eso el estado se guarda y se respalda.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Estado struct {
	Version     int       `json:"version"`
	Actualizado time.Time `json:"actualizado"`
	Buckets     []string  `json:"buckets"`
	Tablas      []string  `json:"tablas"`
	Colas       []string  `json:"colas"`
}

func LeerEstado(ruta string) (*Estado, error) {
	datos, err := os.ReadFile(ruta)
	if errors.Is(err, os.ErrNotExist) {
		// No hay estado todavía: es la primera vez. No es un error.
		return &Estado{Version: 1}, nil
	}
	if err != nil {
		return nil, err
	}
	var e Estado
	if err := json.Unmarshal(datos, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// Guardar escribe el estado de forma ATÓMICA: primero a un temporal, luego se renombra.
//
// ⚠️ `os.WriteFile` trunca el fichero antes de escribir. Si el proceso muere a mitad —o se va la
// luz— te quedas con un estado vacío o cortado, que es peor que no tenerlo: la herramienta cree
// que no creó nada y deja todo huérfano. El renombrado sí es atómico.
func (e *Estado) Guardar(ruta string) error {
	e.Version = 1
	e.Actualizado = time.Now().UTC()
	sort.Strings(e.Buckets)
	sort.Strings(e.Tablas)
	sort.Strings(e.Colas)
	datos, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	tmp := ruta + ".tmp"
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(tmp, append(datos, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, ruta)
}

func (e *Estado) Anotar(tipo, nombre string) {
	switch tipo {
	case "bucket":
		e.Buckets = anadir(e.Buckets, nombre)
	case "tabla":
		e.Tablas = anadir(e.Tablas, nombre)
	case "cola":
		e.Colas = anadir(e.Colas, nombre)
	}
}

func (e *Estado) Olvidar(tipo, nombre string) {
	switch tipo {
	case "bucket":
		e.Buckets = quitar(e.Buckets, nombre)
	case "tabla":
		e.Tablas = quitar(e.Tablas, nombre)
	case "cola":
		e.Colas = quitar(e.Colas, nombre)
	}
}

func anadir(xs []string, x string) []string {
	for _, y := range xs {
		if y == x {
			return xs
		}
	}
	return append(xs, x)
}

func quitar(xs []string, x string) []string {
	fuera := xs[:0]
	for _, y := range xs {
		if y != x {
			fuera = append(fuera, y)
		}
	}
	return fuera
}
```

### `main.go`

```go
// planito — tu propio Terraform, en Go, para entender qué hace de verdad una herramienta
// declarativa.
//
//	planito plan     -f infra.json     # qué haría, y por qué. No toca nada
//	planito aplicar  -f infra.json     # lo hace
//	planito destruir                   # borra SOLO lo que planito creó
//
// Por defecto habla con `nubecita` en http://127.0.0.1:4566, así que la demo entera es gratis y
// se destruye al final. Con `-endpoint ""` habla con Amazon de verdad.
//
// ⛔ Esto NO es Terraform y no pretende serlo. No hay lenguaje propio, ni proveedores, ni grafo
// de dependencias, ni bloqueo del estado, ni módulos. Lo que sí hay es la idea entera, que cabe
// en una resta:
//
//	plan = deseado − actual
//
// El resto de Terraform es envoltorio alrededor de eso — envoltorio muy valioso, pero envoltorio.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		uso()
	}
	orden := os.Args[1]

	fs := flag.NewFlagSet(orden, flag.ExitOnError)
	fichero := fs.String("f", "infra.json", "el fichero con lo que quieres que exista")
	estadoRuta := fs.String("estado", "planito.estado.json", "dónde se guarda lo que creamos")
	endpoint := fs.String("endpoint", "http://127.0.0.1:4566",
		"a quién le hablamos; vacío = Amazon de verdad")
	region := fs.String("region", "us-east-1", "la región")
	si := fs.Bool("si", false, "no preguntar antes de aplicar o destruir")
	_ = fs.Parse(os.Args[2:])

	ctx := context.Background()
	nube := NuevaNube(*endpoint, *region)

	estado, err := LeerEstado(*estadoRuta)
	if err != nil {
		log.Fatalf("no puedo leer el estado %s: %v", *estadoRuta, err)
	}
	actual, err := nube.Leer(ctx)
	if err != nil {
		log.Fatalf("no puedo mirar qué hay: %v", err)
	}

	var plan *Plan
	var deseado *Deseado
	switch orden {
	case "plan", "aplicar":
		deseado, err = LeerDeseado(*fichero)
		if err != nil {
			log.Fatalf("%v", err)
		}
		plan = Calcular(deseado, actual, estado)
	case "destruir":
		deseado = &Deseado{}
		plan = Destruir(actual, estado)
	default:
		uso()
	}

	fmt.Printf("planito · %d recurso(s) en el estado · hablando con %s\n",
		len(estado.Buckets)+len(estado.Tablas)+len(estado.Colas), destino(*endpoint))
	if plan.Vacio() {
		fmt.Println(plan.VacioTexto())
		return
	}
	fmt.Println("\nplan:")
	for _, p := range plan.Pasos {
		fmt.Println(p)
	}

	if orden == "plan" {
		fmt.Printf("\n%d cambio(s). Esto NO ha tocado nada.\n", len(plan.Pasos))
		return
	}
	if !*si {
		// ⚠️ Sin `-si`, `aplicar` y `destruir` NO hacen nada. Una herramienta que borra
		// infraestructura sin confirmación acaba borrando lo que no debía, y el día que pasa
		// nadie se acuerda de por qué le pareció buena idea.
		fmt.Printf("\n%d cambio(s). Vuelve a lanzarlo con -si para hacerlo.\n", len(plan.Pasos))
		return
	}

	fmt.Println("\naplicando:")
	if err := Ejecutar(ctx, nube, plan, deseado, estado, *estadoRuta); err != nil {
		// El estado ya se ha guardado con lo que SÍ se hizo: un fallo a mitad no puede dejar
		// recursos creados fuera del estado, porque entonces nadie los borraría nunca.
		log.Fatalf("\n[!] %v\n    el estado guarda lo que sí se hizo: %s", err, *estadoRuta)
	}
	fmt.Printf("\nlisto. Estado en %s\n", *estadoRuta)
}

// Ejecutar hace los pasos EN ORDEN y guarda el estado DESPUÉS DE CADA UNO.
//
// 🥇 Guardar al final sería más cómodo y está mal: si el proceso muere a mitad, los recursos ya
// creados no estarían en el estado, y `destruir` no se los llevaría nunca. Quedan huérfanos,
// pagando, y sin nadie que se acuerde de ellos.
func Ejecutar(ctx context.Context, nube *Nube, plan *Plan, deseado *Deseado,
	estado *Estado, rutaEstado string) error {
	tablas := map[string]Tabla{}
	for _, t := range deseado.Tablas {
		tablas[t.Nombre] = t
	}
	colas := map[string]Cola{}
	for _, c := range deseado.Colas {
		colas[c.Nombre] = c
	}
	avisar := func(f string, a ...any) { fmt.Printf(f+"\n", a...) }

	for _, p := range plan.Pasos {
		fmt.Printf("  %s %s %s\n", p.Accion, p.Tipo, p.Nombre)
		var err error
		switch {
		case p.Accion == Cambiar:
			// Recrear: borrar y volver a crear. Se avisa a gritos porque se pierden los datos.
			avisar("      ⚠️  se recrea: SE PIERDEN LOS DATOS de %s", p.Nombre)
			if err = borrar(ctx, nube, p, avisar); err == nil {
				err = crear(ctx, nube, p, tablas, colas)
			}
		case p.Accion == Crear:
			err = crear(ctx, nube, p, tablas, colas)
		case p.Accion == Borrar:
			err = borrar(ctx, nube, p, avisar)
		}
		if err != nil {
			_ = estado.Guardar(rutaEstado)
			return fmt.Errorf("%s %s %s: %w", p.Accion, p.Tipo, p.Nombre, err)
		}
		if p.Accion == Borrar {
			estado.Olvidar(p.Tipo, p.Nombre)
		} else {
			estado.Anotar(p.Tipo, p.Nombre)
		}
		if err := estado.Guardar(rutaEstado); err != nil {
			return fmt.Errorf("guardando el estado: %w", err)
		}
	}
	return nil
}

func crear(ctx context.Context, n *Nube, p Paso, tablas map[string]Tabla,
	colas map[string]Cola) error {
	switch p.Tipo {
	case "bucket":
		return n.CrearBucket(ctx, p.Nombre)
	case "tabla":
		return n.CrearTabla(ctx, tablas[p.Nombre])
	case "cola":
		return n.CrearCola(ctx, colas[p.Nombre])
	}
	return fmt.Errorf("tipo desconocido: %s", p.Tipo)
}

func borrar(ctx context.Context, n *Nube, p Paso, avisar func(string, ...any)) error {
	switch p.Tipo {
	case "bucket":
		return n.BorrarBucket(ctx, p.Nombre, avisar)
	case "tabla":
		return n.BorrarTabla(ctx, p.Nombre)
	case "cola":
		return n.BorrarCola(ctx, p.Nombre)
	}
	return fmt.Errorf("tipo desconocido: %s", p.Tipo)
}

func destino(endpoint string) string {
	if endpoint == "" {
		return "AWS de verdad"
	}
	return endpoint
}

func uso() {
	fmt.Println(`planito — tu propio Terraform, en Go

  planito plan     -f infra.json     qué haría, y por qué. NO toca nada
  planito aplicar  -f infra.json -si lo hace
  planito destruir -si               borra SOLO lo que planito creó

  -endpoint  a quién le hablamos (por defecto http://127.0.0.1:4566, o sea nubecita)
  -estado    dónde se guarda lo que creamos (por defecto planito.estado.json)`)
	os.Exit(2)
}
```

### `nube.go`

```go
package main

// Lo que hay de verdad en la nube, y cómo se cambia. Esta es la única parte con red.
//
// El resto de la herramienta —leer el fichero, calcular el plan, llevar el estado— no sabe que
// existe AWS. Por eso el plan se prueba entero sin levantar nada, y por eso cambiar de proveedor
// sería reescribir SOLO este fichero.
//
// El cliente apunta a donde tú digas con `-endpoint`, así que la misma herramienta sirve contra
// `nubecita` (gratis, en tu máquina) y contra Amazon de verdad. Es la misma idea que el
// `--endpoint-url` de la CLI oficial.

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyntypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Actual es la foto de lo que existe AHORA. Es un dato tonto a propósito: quien calcula el plan
// no debe poder llamar a la red sin querer.
type Actual struct {
	Buckets []string
	Tablas  []Tabla
	Colas   []string
}

type Nube struct {
	s3     *s3.Client
	dynamo *dynamodb.Client
	sqs    *sqs.Client
}

func NuevaNube(endpoint, region string) *Nube {
	cred := credentials.NewStaticCredentialsProvider("planito", "planito", "")
	var punta *string
	if endpoint != "" {
		punta = aws.String(endpoint)
	}
	return &Nube{
		// ⚠️ UsePathStyle: sin esto el SDK mete el bucket en el NOMBRE DEL HOST y contra tu
		// máquina eso no resuelve. Es el error número uno al apuntar el SDK a un emulador.
		s3: s3.New(s3.Options{Region: region, Credentials: cred, BaseEndpoint: punta,
			UsePathStyle: true}),
		dynamo: dynamodb.New(dynamodb.Options{Region: region, Credentials: cred,
			BaseEndpoint: punta}),
		sqs: sqs.New(sqs.Options{Region: region, Credentials: cred, BaseEndpoint: punta}),
	}
}

// Leer saca la foto de los tres servicios. Si alguno falla, falla entero: un plan calculado con
// una foto a medias propondría crear cosas que ya existen o borrar cosas que sí quieres.
func (n *Nube) Leer(ctx context.Context) (*Actual, error) {
	a := &Actual{}

	buckets, err := n.s3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("listando buckets: %w", err)
	}
	for _, b := range buckets.Buckets {
		a.Buckets = append(a.Buckets, aws.ToString(b.Name))
	}

	tablas, err := n.dynamo.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, fmt.Errorf("listando tablas: %w", err)
	}
	for _, nombre := range tablas.TableNames {
		ficha, err := n.dynamo.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(nombre)})
		if err != nil {
			return nil, fmt.Errorf("describiendo la tabla %s: %w", nombre, err)
		}
		t := Tabla{Nombre: nombre}
		for _, k := range ficha.Table.KeySchema {
			if k.KeyType == dyntypes.KeyTypeHash {
				t.Particion = aws.ToString(k.AttributeName)
			} else {
				t.Orden = aws.ToString(k.AttributeName)
			}
		}
		a.Tablas = append(a.Tablas, t)
	}

	colas, err := n.sqs.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return nil, fmt.Errorf("listando colas: %w", err)
	}
	for _, u := range colas.QueueUrls {
		a.Colas = append(a.Colas, ultimoTramo(u))
	}

	sort.Strings(a.Buckets)
	sort.Strings(a.Colas)
	sort.Slice(a.Tablas, func(i, j int) bool { return a.Tablas[i].Nombre < a.Tablas[j].Nombre })
	return a, nil
}

// ── crear ────────────────────────────────────────────────────────────────

func (n *Nube) CrearBucket(ctx context.Context, nombre string) error {
	_, err := n.s3.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(nombre)})
	return err
}

func (n *Nube) CrearTabla(ctx context.Context, t Tabla) error {
	atrib := []dyntypes.AttributeDefinition{{
		AttributeName: aws.String(t.Particion), AttributeType: dyntypes.ScalarAttributeTypeS}}
	esquema := []dyntypes.KeySchemaElement{{
		AttributeName: aws.String(t.Particion), KeyType: dyntypes.KeyTypeHash}}
	if t.Orden != "" {
		atrib = append(atrib, dyntypes.AttributeDefinition{
			AttributeName: aws.String(t.Orden), AttributeType: dyntypes.ScalarAttributeTypeS})
		esquema = append(esquema, dyntypes.KeySchemaElement{
			AttributeName: aws.String(t.Orden), KeyType: dyntypes.KeyTypeRange})
	}
	_, err := n.dynamo.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:            aws.String(t.Nombre),
		AttributeDefinitions: atrib,
		KeySchema:            esquema,
		BillingMode:          dyntypes.BillingModePayPerRequest,
	})
	return err
}

func (n *Nube) CrearCola(ctx context.Context, c Cola) error {
	attr := map[string]string{}
	if c.Muertos != "" {
		// El RedrivePolicy viaja como JSON DENTRO de una cadena. Se arma con el codificador y
		// no a mano: escribirlo a mano es la forma más rápida de colar una comilla mal.
		pol, err := json.Marshal(map[string]any{
			"deadLetterTargetArn": "arn:aws:sqs:us-east-1:000000000000:" + c.Muertos,
			"maxReceiveCount":     strconv.Itoa(c.MaxRecibos),
		})
		if err != nil {
			return err
		}
		attr["RedrivePolicy"] = string(pol)
	}
	_, err := n.sqs.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(c.Nombre), Attributes: attr})
	return err
}

// ── borrar ───────────────────────────────────────────────────────────────

// BorrarBucket vacía antes de borrar. S3 se niega a borrar un bucket con cosas dentro, y esa
// negativa es correcta: si la herramienta lo saltara, un `destroy` mal escrito se llevaría datos
// por delante sin que nadie lo viera venir. Aquí se vacía A PROPÓSITO y se dice en el registro.
func (n *Nube) BorrarBucket(ctx context.Context, nombre string, avisar func(string, ...any)) error {
	lista, err := n.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(nombre)})
	if err != nil {
		return err
	}
	if len(lista.Contents) > 0 {
		avisar("      vaciando %d objeto(s) de %s antes de borrarlo", len(lista.Contents), nombre)
		for _, o := range lista.Contents {
			if _, err := n.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(nombre), Key: o.Key}); err != nil {
				return err
			}
		}
	}
	_, err = n.s3.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(nombre)})
	return err
}

func (n *Nube) BorrarTabla(ctx context.Context, nombre string) error {
	_, err := n.dynamo.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: aws.String(nombre)})
	return err
}

func (n *Nube) BorrarCola(ctx context.Context, nombre string) error {
	u, err := n.sqs.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String(nombre)})
	if err != nil {
		return err
	}
	_, err = n.sqs.DeleteQueue(ctx, &sqs.DeleteQueueInput{QueueUrl: u.QueueUrl})
	return err
}

func ultimoTramo(u string) string {
	for i := len(u) - 1; i >= 0; i-- {
		if u[i] == '/' {
			return u[i+1:]
		}
	}
	return u
}
```

### `plan.go`

```go
package main

// 🥇 EL CORAZÓN DE LA HERRAMIENTA, y cabe en una frase:
//
//	plan = deseado − actual
//
// Eso es todo lo que hace Terraform, y lo que hace un controlador de Kubernetes, y lo que hace
// cualquier cosa que se llame "declarativa". No hay más magia. Lo demás —el lenguaje, los
// proveedores, el grafo— es envoltorio alrededor de esta resta.
//
// Y una decisión que lo cambia todo: aquí NO hay red. Esta función recibe dos listas y devuelve
// una tercera. Por eso se puede probar entera sin levantar nada, y por eso las pruebas de esta
// herramienta tardan milisegundos.

import (
	"fmt"
	"sort"
	"strings"
)

type Accion string

const (
	Crear   Accion = "crear"
	Borrar  Accion = "borrar"
	Cambiar Accion = "cambiar" // existe pero con otra forma: hay que rehacerlo
)

type Paso struct {
	Accion Accion
	Tipo   string // bucket | tabla | cola
	Nombre string
	Porque string // ⚠️ por qué este paso está en el plan. Un plan sin porqués no se audita.
}

func (p Paso) String() string {
	simbolo := map[Accion]string{Crear: "+", Borrar: "-", Cambiar: "~"}[p.Accion]
	linea := fmt.Sprintf("  %s %-7s %-14s", simbolo, p.Tipo, p.Nombre)
	if p.Porque != "" {
		linea += "· " + p.Porque
	}
	return strings.TrimRight(linea, " ")
}

type Plan struct {
	Pasos []Paso
}

func (p *Plan) VacioTexto() string {
	return "sin cambios: lo que hay ya es lo que quieres"
}

func (p *Plan) Vacio() bool { return len(p.Pasos) == 0 }

// Calcular compara lo deseado con lo actual y devuelve QUÉ hay que hacer.
//
// `gestionado` es lo que ESTA herramienta creó (viene del fichero de estado). Es la pieza que
// más gente se salta y sin la cual todo esto es peligroso: sin ella, «lo que sobra» incluiría
// los recursos que creó otra persona, y el destroy se los llevaría por delante.
func Calcular(deseado *Deseado, actual *Actual, gestionado *Estado) *Plan {
	p := &Plan{}

	// ── buckets ──────────────────────────────────────────────────────────
	hay := conjunto(actual.Buckets)
	quiero := conjunto(deseado.Buckets)
	for _, b := range deseado.Buckets {
		if !hay[b] {
			p.Pasos = append(p.Pasos, Paso{Crear, "bucket", b, "no existe"})
		}
	}
	for _, b := range ordenado(gestionado.Buckets) {
		if !quiero[b] && hay[b] {
			p.Pasos = append(p.Pasos, Paso{Borrar, "bucket", b, "lo creamos y ya no lo quieres"})
		}
	}

	// ── tablas ───────────────────────────────────────────────────────────
	// Aquí aparece `cambiar`, y con él la lección incómoda: la clave de una tabla de DynamoDB
	// NO se puede modificar. Cambiarla en el fichero significa borrar y volver a crear, o sea
	// PERDER LOS DATOS. Por eso el plan lo dice con todas las letras en vez de hacerlo callando.
	tablasActuales := map[string]Tabla{}
	for _, t := range actual.Tablas {
		tablasActuales[t.Nombre] = t
	}
	quieroTabla := map[string]bool{}
	for _, t := range deseado.Tablas {
		quieroTabla[t.Nombre] = true
		vieja, existe := tablasActuales[t.Nombre]
		switch {
		case !existe:
			p.Pasos = append(p.Pasos, Paso{Crear, "tabla", t.Nombre, "no existe"})
		case vieja.Particion != t.Particion || vieja.Orden != t.Orden:
			p.Pasos = append(p.Pasos, Paso{Cambiar, "tabla", t.Nombre,
				fmt.Sprintf("la clave cambia de (%s) a (%s): hay que recrearla y SE PIERDEN LOS DATOS",
					clave(vieja), clave(t))})
		}
	}
	for _, n := range ordenado(gestionado.Tablas) {
		if _, existe := tablasActuales[n]; existe && !quieroTabla[n] {
			p.Pasos = append(p.Pasos, Paso{Borrar, "tabla", n, "la creamos y ya no la quieres"})
		}
	}

	// ── colas ────────────────────────────────────────────────────────────
	colasActuales := conjunto(actual.Colas)
	quieroCola := map[string]bool{}
	for _, c := range deseado.Colas {
		quieroCola[c.Nombre] = true
		if !colasActuales[c.Nombre] {
			p.Pasos = append(p.Pasos, Paso{Crear, "cola", c.Nombre, "no existe"})
		}
	}
	// ⚠️ Al borrar colas hay que ir en orden INVERSO al de creación: la cola que apunta a una
	// cola de muertos se va antes que la de muertos. Como `Ordenar()` ya dejó las dependencias
	// primero, aquí basta con recorrer al revés.
	borrables := ordenado(gestionado.Colas)
	for i := len(borrables) - 1; i >= 0; i-- {
		n := borrables[i]
		if colasActuales[n] && !quieroCola[n] {
			p.Pasos = append(p.Pasos, Paso{Borrar, "cola", n, "la creamos y ya no la quieres"})
		}
	}
	return p
}

// Destruir es el plan de «quítalo todo»: borrar TODO lo que esta herramienta creó, y nada más.
// Es literalmente `Calcular` contra un deseado vacío, y por eso no puede llevarse por delante
// lo que creó otra persona: lo que no está en el estado, no se toca.
func Destruir(actual *Actual, gestionado *Estado) *Plan {
	return Calcular(&Deseado{}, actual, gestionado)
}

func clave(t Tabla) string {
	if t.Orden == "" {
		return t.Particion
	}
	return t.Particion + ", " + t.Orden
}

func conjunto(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func ordenado(xs []string) []string {
	fuera := append([]string(nil), xs...)
	sort.Strings(fuera)
	return fuera
}
```

### `plan_test.go`

```go
package main

// Las pruebas del corazón de la herramienta. Y fíjate en lo que NO hay aquí: red, contenedores,
// credenciales, esperas. `Calcular` recibe tres estructuras y devuelve una cuarta.
//
// Esa es la ventaja de haber sacado la red a `nube.go`: la lógica que de verdad se puede
// equivocar —qué crear, qué borrar y qué NO tocar— se prueba entera en milisegundos.
//
//	go test ./... -count=1 -v

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pasos(p *Plan) []string {
	var fuera []string
	for _, x := range p.Pasos {
		fuera = append(fuera, string(x.Accion)+" "+x.Tipo+" "+x.Nombre)
	}
	return fuera
}

func iguales(t *testing.T, tuvo []string, quiero ...string) {
	t.Helper()
	if len(tuvo) != len(quiero) {
		t.Fatalf("plan = %v, quiero %v", tuvo, quiero)
	}
	for i := range quiero {
		if tuvo[i] != quiero[i] {
			t.Fatalf("plan = %v, quiero %v", tuvo, quiero)
		}
	}
}

func TestLoQueFaltaSeCrea(t *testing.T) {
	d := &Deseado{Buckets: []string{"fotos"},
		Tablas: []Tabla{{Nombre: "pedidos", Particion: "cliente", Orden: "fecha"}},
		Colas:  []Cola{{Nombre: "trabajos"}}}
	p := Calcular(d, &Actual{}, &Estado{})
	iguales(t, pasos(p), "crear bucket fotos", "crear tabla pedidos", "crear cola trabajos")
}

// 🥇 Idempotencia: aplicar dos veces no hace nada la segunda. Es la propiedad que define a una
// herramienta declarativa, y la que más fácil se rompe sin darte cuenta.
func TestAplicarDosVecesNoHaceNada(t *testing.T) {
	d := &Deseado{Buckets: []string{"fotos"},
		Tablas: []Tabla{{Nombre: "pedidos", Particion: "cliente"}},
		Colas:  []Cola{{Nombre: "trabajos"}}}
	actual := &Actual{Buckets: []string{"fotos"},
		Tablas: []Tabla{{Nombre: "pedidos", Particion: "cliente"}},
		Colas:  []string{"trabajos"}}
	estado := &Estado{Buckets: []string{"fotos"}, Tablas: []string{"pedidos"},
		Colas: []string{"trabajos"}}
	p := Calcular(d, actual, estado)
	if !p.Vacio() {
		t.Fatalf("la segunda pasada propone %v, y no debería proponer nada", pasos(p))
	}
}

// 🥇 LA PRUEBA QUE JUSTIFICA EL FICHERO DE ESTADO: hay un bucket que no está en tu fichero, pero
// tampoco lo creaste tú. No se toca. Sin estado, esto se borraría — y sería de otro equipo.
func TestNoSeTocaoLoQueNoCreamos(t *testing.T) {
	d := &Deseado{Buckets: []string{"mio"}}
	actual := &Actual{Buckets: []string{"mio", "de-otro-equipo"}}
	estado := &Estado{Buckets: []string{"mio"}} // «de-otro-equipo» NO está aquí
	p := Calcular(d, actual, estado)
	if !p.Vacio() {
		t.Fatalf("propone %v: se está metiendo con lo que no creó", pasos(p))
	}
}

func TestLoQueSobraSeBorraSoloSiLoCreamos(t *testing.T) {
	d := &Deseado{Buckets: []string{}}
	actual := &Actual{Buckets: []string{"viejo", "de-otro-equipo"}}
	estado := &Estado{Buckets: []string{"viejo"}}
	p := Calcular(d, actual, estado)
	iguales(t, pasos(p), "borrar bucket viejo")
}

// Cambiar la clave de una tabla no es una modificación: es recrearla. Y eso pierde los datos.
// La herramienta lo tiene que DECIR, no hacerlo callando.
func TestCambiarLaClaveDeUnaTablaAvisaDeQueSePierdenLosDatos(t *testing.T) {
	d := &Deseado{Tablas: []Tabla{{Nombre: "pedidos", Particion: "cliente", Orden: "fecha"}}}
	actual := &Actual{Tablas: []Tabla{{Nombre: "pedidos", Particion: "id"}}}
	estado := &Estado{Tablas: []string{"pedidos"}}
	p := Calcular(d, actual, estado)
	iguales(t, pasos(p), "cambiar tabla pedidos")
	if !strings.Contains(p.Pasos[0].Porque, "SE PIERDEN LOS DATOS") {
		t.Fatalf("el plan no avisa de que se pierden los datos: %q", p.Pasos[0].Porque)
	}
}

// Destruir es Calcular contra un deseado vacío. Y por eso tampoco puede tocar lo ajeno.
func TestDestruirSoloSeLlevaLoNuestro(t *testing.T) {
	actual := &Actual{Buckets: []string{"mio", "ajeno"},
		Tablas: []Tabla{{Nombre: "pedidos", Particion: "id"}},
		Colas:  []string{"trabajos", "muertos", "ajena"}}
	estado := &Estado{Buckets: []string{"mio"}, Tablas: []string{"pedidos"},
		Colas: []string{"muertos", "trabajos"}}
	p := Calcular(&Deseado{}, actual, estado)
	// Las colas se borran en orden inverso al alfabético del estado: primero «trabajos»
	// (que apunta a la de muertos) y luego «muertos».
	iguales(t, pasos(p),
		"borrar bucket mio", "borrar tabla pedidos",
		"borrar cola trabajos", "borrar cola muertos")
}

// El orden de creación tiene que respetar la única dependencia del modelo: la cola de muertos
// existe antes que la cola que la nombra.
func TestLaColaDeMuertosSeCreaAntes(t *testing.T) {
	d := &Deseado{Colas: []Cola{
		{Nombre: "trabajos", Muertos: "muertos", MaxRecibos: 3},
		{Nombre: "muertos"},
	}}
	d.Ordenar()
	p := Calcular(d, &Actual{}, &Estado{})
	iguales(t, pasos(p), "crear cola muertos", "crear cola trabajos")
}

// Lo que se puede cazar sin tocar la red, se caza sin tocar la red.
func TestValidarCazaLosErroresDelFicheroSinLlamarANadie(t *testing.T) {
	casos := []struct {
		nombre string
		d      Deseado
		dice   string
	}{
		{"bucket repetido", Deseado{Buckets: []string{"a", "a"}}, "dos veces"},
		{"tabla sin clave", Deseado{Tablas: []Tabla{{Nombre: "t"}}}, "clave de partición"},
		{"muertos que no existe",
			Deseado{Colas: []Cola{{Nombre: "t", Muertos: "no-esta", MaxRecibos: 2}}},
			"que no está declarada"},
		{"muertos sin maxRecibos",
			Deseado{Colas: []Cola{{Nombre: "m"}, {Nombre: "t", Muertos: "m"}}},
			"maxRecibos"},
		{"cola que se apunta a sí misma",
			Deseado{Colas: []Cola{{Nombre: "t", Muertos: "t", MaxRecibos: 2}}},
			"a sí misma"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := c.d.Validar()
			if err == nil {
				t.Fatal("lo dio por bueno")
			}
			if !strings.Contains(err.Error(), c.dice) {
				t.Fatalf("el error no explica el problema: %v", err)
			}
		})
	}
}

// El estado se escribe de forma atómica y se relee igual. Es el fichero que no te puedes
// permitir perder, así que merece su prueba.
func TestElEstadoVaYVuelve(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "sub", "planito.estado.json")
	e := &Estado{}
	e.Anotar("bucket", "fotos")
	e.Anotar("bucket", "fotos") // dos veces: no debe duplicar
	e.Anotar("cola", "trabajos")
	if err := e.Guardar(ruta); err != nil {
		t.Fatalf("Guardar: %v", err)
	}
	if _, err := os.Stat(ruta + ".tmp"); err == nil {
		t.Error("quedó el fichero temporal sin renombrar")
	}
	vuelto, err := LeerEstado(ruta)
	if err != nil {
		t.Fatalf("LeerEstado: %v", err)
	}
	if len(vuelto.Buckets) != 1 || vuelto.Buckets[0] != "fotos" {
		t.Errorf("buckets = %v", vuelto.Buckets)
	}
	vuelto.Olvidar("cola", "trabajos")
	if len(vuelto.Colas) != 0 {
		t.Errorf("colas = %v tras olvidar", vuelto.Colas)
	}
}

// Que no haya estado todavía NO es un error: es la primera vez que corres la herramienta.
func TestSinEstadoTodaviaNoEsUnError(t *testing.T) {
	e, err := LeerEstado(filepath.Join(t.TempDir(), "no-existe.json"))
	if err != nil {
		t.Fatalf("la primera vez no debería fallar: %v", err)
	}
	if len(e.Buckets)+len(e.Tablas)+len(e.Colas) != 0 {
		t.Error("el estado nuevo no viene vacío")
	}
}

// Una errata en un campo tiene que fallar, no ignorarse. Un fichero de infraestructura leído
// «a medias» es la forma más cara de descubrir un typo.
func TestUnCampoMalEscritoFalla(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "infra.json")
	os.WriteFile(ruta, []byte(`{"bucketss": ["fotos"]}`), 0o644)
	if _, err := LeerDeseado(ruta); err == nil {
		t.Fatal("se tragó un campo que no existe")
	}
}
```

### `Dockerfile`

```dockerfile
# planito en una imagen, en dos etapas — igual que nubecita.
#
# ⚠️ Se corre en contenedor por el mismo motivo que nubecita: el Application Control de la
# maquina de desarrollo bloquea a ratos los binarios recien compilados. Y de paso, esto es lo
# que hace cualquiera con una herramienta de infraestructura: meterla en una imagen y correrla
# igual en su portatil y en la integracion continua.
#
#     docker network create nube
#     docker run -d --name nubecita --network nube nubecita:ep3
#     docker run --rm --network nube -v "$PWD:/w" -w /w planito \
#         plan -f infra.json -endpoint http://nubecita:4566

FROM golang:1.26 AS construir
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /planito .

FROM scratch
COPY --from=construir /planito /planito
WORKDIR /w
ENTRYPOINT ["/planito"]
```
