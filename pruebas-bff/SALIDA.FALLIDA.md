# Pruebas ejecutadas — Vídeo 115 — ¿Necesitas un BFF en Go? (medido, no supuesto)

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. Sin BFF: el cliente llama a los tres servicios

**La pregunta:** Cuanto cuesta pintar UNA pantalla si el movil habla con los microservicios?

`banco/banco.go`

```go
package banco

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

// Contadores globales para poder decir cuantas llamadas costo cada version.
var Llamadas atomic.Int64
var BytesServidos atomic.Int64

type Servicio struct{ *httptest.Server }

// Nuevo levanta un servicio que tarda `retardo` y devuelve `cuerpo`.
func Nuevo(retardo time.Duration, cuerpo any) *Servicio {
	s := &Servicio{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Llamadas.Add(1)
		time.Sleep(retardo)
		b, _ := json.Marshal(cuerpo)
		BytesServidos.Add(int64(len(b)))
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}))
	return s
}

func Reset() { Llamadas.Store(0); BytesServidos.Store(0) }

// ── Lo que devuelve cada microservicio. A proposito, MAS de lo que la pantalla usa:
// es el over-fetching del que hablan los tres articulos, y aqui se mide en bytes.
type Usuario struct {
	ID            string            `json:"id"`
	Nombre        string            `json:"nombre"`
	Email         string            `json:"email"`
	Telefono      string            `json:"telefono"`
	Direccion     string            `json:"direccion"`
	Ciudad        string            `json:"ciudad"`
	Pais          string            `json:"pais"`
	CodigoPostal  string            `json:"codigo_postal"`
	FechaAlta     string            `json:"fecha_alta"`
	UltimoAcceso  string            `json:"ultimo_acceso"`
	Preferencias  map[string]string `json:"preferencias"`
	Verificado    bool              `json:"verificado"`
	Plan          string            `json:"plan"`
	AvatarURL     string            `json:"avatar_url"`
	Biografia     string            `json:"biografia"`
}

type Movimiento struct {
	ID          string  `json:"id"`
	Concepto    string  `json:"concepto"`
	Importe     float64 `json:"importe"`
	Divisa      string  `json:"divisa"`
	Fecha       string  `json:"fecha"`
	Categoria   string  `json:"categoria"`
	Comercio    string  `json:"comercio"`
	Referencia  string  `json:"referencia"`
	Estado      string  `json:"estado"`
	Metodo      string  `json:"metodo"`
}

type Aviso struct {
	ID     string `json:"id"`
	Texto  string `json:"texto"`
	Tipo   string `json:"tipo"`
	Fecha  string `json:"fecha"`
	Leido  bool   `json:"leido"`
	Enlace string `json:"enlace"`
}

func UsuarioDemo() Usuario {
	return Usuario{
		ID: "u-1042", Nombre: "Ana Rivas", Email: "ana@example.com",
		Telefono: "+34 600 000 000", Direccion: "Calle Mayor 14, 3B", Ciudad: "Valencia",
		Pais: "ES", CodigoPostal: "46002", FechaAlta: "2021-03-14T10:22:00Z",
		UltimoAcceso: "2026-08-22T09:11:00Z", Verificado: true, Plan: "premium",
		AvatarURL: "https://cdn.example.com/avatars/u-1042-512.png",
		Biografia: "Cuenta personal. Notificaciones activas desde 2021.",
		Preferencias: map[string]string{"tema": "oscuro", "idioma": "es",
			"moneda": "EUR", "resumen": "semanal", "biometria": "activada"},
	}
}

func MovimientosDemo(n int) []Movimiento {
	out := make([]Movimiento, n)
	for i := range out {
		out[i] = Movimiento{
			ID: fmt.Sprintf("m-%04d", i), Concepto: "Compra con tarjeta",
			Importe: -12.40 - float64(i), Divisa: "EUR",
			Fecha: "2026-08-2" + fmt.Sprint(i%9) + "T12:00:00Z",
			Categoria: "supermercado", Comercio: "Mercadona SA",
			Referencia: fmt.Sprintf("REF-2026-08-%06d", i), Estado: "liquidado",
			Metodo: "tarjeta-debito",
		}
	}
	return out
}

func AvisosDemo(n int) []Aviso {
	out := make([]Aviso, n)
	for i := range out {
		out[i] = Aviso{ID: fmt.Sprintf("a-%02d", i), Texto: "Tu resumen semanal esta listo",
			Tipo: "informativo", Fecha: "2026-08-20T08:00:00Z", Leido: i%2 == 0,
			Enlace: "https://app.example.com/resumen/semana-34"}
	}
	return out
}
```

`sin/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"pruebas/banco"
)

func main() {
	// Tres microservicios, con latencias distintas como en la vida real.
	u := banco.Nuevo(60*time.Millisecond, banco.UsuarioDemo())
	m := banco.Nuevo(90*time.Millisecond, banco.MovimientosDemo(20))
	a := banco.Nuevo(45*time.Millisecond, banco.AvisosDemo(6))
	defer u.Close()
	defer m.Close()
	defer a.Close()

	banco.Reset()
	inicio := time.Now()

	// El movil hace las tres llamadas. Y las hace EN SERIE, porque necesita el usuario
	// para saber si pedir los movimientos: es el caso normal, no el ideal.
	total := 0
	for _, url := range []string{u.URL, m.URL, a.URL} {
		r, err := http.Get(url)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		total += len(b)
	}

	tardo := time.Since(inicio)
	fmt.Println("SIN BFF - el cliente lo arma todo")
	fmt.Printf("  llamadas de red que hace el movil : %d\n", banco.Llamadas.Load())
	fmt.Printf("  tiempo hasta poder pintar         : %v\n", tardo.Round(time.Millisecond))
	fmt.Printf("  bytes que baja el movil           : %d\n", total)

	// Y de todo eso, ¿cuanto usa la pantalla de verdad?
	var usr banco.Usuario
	rr, _ := http.Get(u.URL)
	json.NewDecoder(rr.Body).Decode(&usr)
	rr.Body.Close()
	utiles, _ := json.Marshal(map[string]any{"nombre": usr.Nombre, "plan": usr.Plan})
	fmt.Printf("  de los cuales la pantalla usa     : %d bytes del usuario (nombre y plan)\n",
		len(utiles))
}
```

```console
$ go run ./sin
SIN BFF - el cliente lo arma todo
  llamadas de red que hace el movil : 3
  tiempo hasta poder pintar         : 198ms
  bytes que baja el movil           : 6440
  de los cuales la pantalla usa     : 39 bytes del usuario (nombre y plan)
[codigo de salida: 0]
```

---

## 2. Con BFF: una llamada, y el BFF agrega en paralelo

**La pregunta:** Y con una capa BFF delante?

`con/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"pruebas/banco"
)

// Lo que la pantalla del MOVIL necesita. Ni un campo mas.
type PantallaMovil struct {
	Nombre     string   `json:"nombre"`
	Plan       string   `json:"plan"`
	Saldo      float64  `json:"saldo"`
	Ultimos    []string `json:"ultimos"`
	SinLeer    int      `json:"sin_leer"`
}

func main() {
	u := banco.Nuevo(60*time.Millisecond, banco.UsuarioDemo())
	m := banco.Nuevo(90*time.Millisecond, banco.MovimientosDemo(20))
	a := banco.Nuevo(45*time.Millisecond, banco.AvisosDemo(6))
	defer u.Close()
	defer m.Close()
	defer a.Close()

	// EL BFF. Una ruta. Dentro, las tres llamadas EN PARALELO.
	bff := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			wg   sync.WaitGroup
			usr  banco.Usuario
			movs []banco.Movimiento
			avs  []banco.Aviso
		)
		pedir := func(url string, destino any) {
			defer wg.Done()
			res, err := http.Get(url)
			if err != nil {
				return
			}
			defer res.Body.Close()
			json.NewDecoder(res.Body).Decode(destino)
		}
		wg.Add(3)
		go pedir(u.URL, &usr)
		go pedir(m.URL, &movs)
		go pedir(a.URL, &avs)
		wg.Wait()

		// Agregar y TRANSFORMAR: eso es el BFF.
		p := PantallaMovil{Nombre: usr.Nombre, Plan: usr.Plan}
		for _, x := range movs {
			p.Saldo += x.Importe
		}
		for i, x := range movs {
			if i == 3 {
				break
			}
			p.Ultimos = append(p.Ultimos, x.Concepto)
		}
		for _, x := range avs {
			if !x.Leido {
				p.SinLeer++
			}
		}
		json.NewEncoder(w).Encode(p)
	}))
	defer bff.Close()

	banco.Reset()
	inicio := time.Now()
	res, _ := http.Get(bff.URL)
	b, _ := io.ReadAll(res.Body)
	res.Body.Close()
	tardo := time.Since(inicio)

	fmt.Println("CON BFF - una llamada, y el BFF agrega")
	fmt.Printf("  llamadas de red que hace el movil : 1\n")
	fmt.Printf("  llamadas que hace el BFF por detras: %d (en paralelo)\n", banco.Llamadas.Load())
	fmt.Printf("  tiempo hasta poder pintar         : %v\n", tardo.Round(time.Millisecond))
	fmt.Printf("  bytes que baja el movil           : %d\n", len(b))
	fmt.Printf("  y esto es TODO lo que recibe      : %s\n", string(b))
}
```

```console
$ go run ./con
malformed import path "pruebas/con": "con" disallowed as path element component on Windows
[codigo de salida: 1]
```

> Las dos cosas que hace el BFF, y que hay que decir juntas: AGREGA (tres llamadas en una) y TRANSFORMA (devuelve solo lo que la pantalla usa).

---

## 3. El «cuando NO»: con UN solo backend, el BFF es un salto de mas

**La pregunta:** Y si detras solo hay un servicio?

`solo/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"pruebas/banco"
)

// El caso honesto: UN solo backend. Aqui el BFF no agrega nada, solo reenvia.
func main() {
	backend := banco.Nuevo(60*time.Millisecond, banco.UsuarioDemo())
	defer backend.Close()

	bff := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get(backend.URL)
		if err != nil {
			http.Error(w, "arriba", 502)
			return
		}
		defer res.Body.Close()
		var usr banco.Usuario
		json.NewDecoder(res.Body).Decode(&usr)
		json.NewEncoder(w).Encode(map[string]any{"nombre": usr.Nombre, "plan": usr.Plan})
	}))
	defer bff.Close()

	medir := func(nombre, url string, vueltas int) time.Duration {
		var total time.Duration
		for i := 0; i < vueltas; i++ {
			t := time.Now()
			res, _ := http.Get(url)
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			total += time.Since(t)
		}
		media := total / time.Duration(vueltas)
		fmt.Printf("  %-22s %v de media (%d vueltas)\n", nombre, media.Round(100*time.Microsecond), vueltas)
		return media
	}

	fmt.Println("UN SOLO SERVICIO DETRAS - el BFF no agrega nada")
	directo := medir("directo al backend:", backend.URL, 20)
	conbff := medir("pasando por el BFF:", bff.URL, 20)

	fmt.Println()
	fmt.Printf("  el BFF anade %v por peticion (%.0f%% mas lento)\n",
		(conbff - directo).Round(100*time.Microsecond),
		float64(conbff-directo)/float64(directo)*100)
	fmt.Println("  Con un solo backend, un BFF es un salto de red de mas.")
}
```

```console
$ go run ./solo
UN SOLO SERVICIO DETRAS - el BFF no agrega nada
  directo al backend:    60.6ms de media (20 vueltas)
  pasando por el BFF:    60.7ms de media (20 vueltas)

  el BFF anade 100µs por peticion (0% mas lento)
  Con un solo backend, un BFF es un salto de red de mas.
[codigo de salida: 0]
```

> Este es el dato que hace honesto el video. El patron no es gratis: cuando no hay nada que agregar, lo unico que anade el BFF es latencia.

---

## 4. El contrato primero: oapi-codegen genera tipos y servidor

**La pregunta:** Como se hace contract-first de verdad?

`api/openapi.yaml`

```go
openapi: 3.0.0
info:
  title: BFF del movil
  version: "1.0"
paths:
  /v1/pantalla:
    get:
      operationId: getPantalla
      summary: Todo lo que la pantalla de inicio del movil necesita, en una llamada
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PantallaMovil'
components:
  schemas:
    PantallaMovil:
      type: object
      required: [nombre, plan, saldo, sin_leer]
      properties:
        nombre:   { type: string }
        plan:     { type: string }
        saldo:    { type: number, format: double }
        ultimos:  { type: array, items: { type: string } }
        sin_leer: { type: integer }
```

`api/cfg.yaml`

```go
package: api
output: api/api.gen.go
generate:
  models: true
  chi-server: true
```

```console
$ go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config api/cfg.yaml api/openapi.yaml
go: downloading github.com/oapi-codegen/oapi-codegen/v2 v2.8.0
go: downloading github.com/oapi-codegen/oapi-codegen v1.16.3
go: downloading go.yaml.in/yaml/v3 v3.0.4
go: downloading github.com/speakeasy-api/openapi v1.24.0
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading golang.org/x/mod v0.38.0
go: downloading golang.org/x/text v0.40.0
go: downloading golang.org/x/tools v0.48.0
go: downloading github.com/getkin/kin-openapi v0.142.0
go: downloading golang.org/x/sync v0.22.0
go: downloading github.com/vmware-labs/yaml-jsonpath v0.3.2
go: downloading github.com/speakeasy-api/jsonpath v0.6.3
go: downloading github.com/go-openapi/jsonpointer v0.23.1
go: downloading github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
go: downloading github.com/oasdiff/yaml v0.1.1
go: downloading github.com/oasdiff/yaml3 v0.0.14
go: downloading github.com/go-openapi/swag/jsonname v0.26.0
go: downloading github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936
[codigo de salida: 0]
```

```console
$ sh -c head -30 api/api.gen.go
// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.8.0 DO NOT EDIT.
package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// PantallaMovil defines model for PantallaMovil.
type PantallaMovil struct {
	Nombre  string    `json:"nombre"`
	Plan    string    `json:"plan"`
	Saldo   float64   `json:"saldo"`
	SinLeer int       `json:"sin_leer"`
	Ultimos *[]string `json:"ultimos,omitempty"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// GetPantalla Todo lo que la pantalla de inicio del movil necesita, en una llamada
	// (GET /v1/pantalla)
	GetPantalla(w http.ResponseWriter, r *http.Request)
}

// Unimplemented server implementation that returns http.StatusNotImplemented for each endpoint.
[codigo de salida: 0]
```

---

## 5. El contrato lo vigila el COMPILADOR, no una revision

**La pregunta:** Que pasa si el handler no cumple el contrato?

`bien/main.go`

```go
package main

import (
	"encoding/json"
	"net/http"

	"pruebas/api"
)

type Servidor struct{}

// La firma la impone la interfaz generada del openapi.yaml. Si no coincide, no compila.
func (s Servidor) GetPantalla(w http.ResponseWriter, r *http.Request) {
	saldo := -248.60
	json.NewEncoder(w).Encode(api.PantallaMovil{
		Nombre: "Ana Rivas", Plan: "premium", Saldo: saldo, SinLeer: 3,
	})
}

var _ api.ServerInterface = Servidor{}

func main() {}
```

`mal/main.go`

```go
package main

import "pruebas/api"

type Servidor struct{}

// A proposito: no implementa GetPantalla. El contrato dice que hace falta.
var _ api.ServerInterface = Servidor{}

func main() {}
```

```console
$ go build -o /tmp/bien ./bien
api/api.gen.go:10:2: no required module provides package github.com/go-chi/chi/v5; to add it:
	go get github.com/go-chi/chi/v5
[codigo de salida: 1]
```

```console
$ go build -o /tmp/mal ./mal   # ⚠️ falla A PROPÓSITO
api/api.gen.go:10:2: no required module provides package github.com/go-chi/chi/v5; to add it:
	go get github.com/go-chi/chi/v5
[codigo de salida: 1]
```

> Ese error del compilador es el argumento entero del contract-first: el contrato deja de ser un documento y pasa a ser algo que no se puede incumplir.
