# -*- coding: utf-8 -*-
"""pruebas_bff.py — mide el patrón BFF en Go de verdad, para el vídeo #115.

No se cita ni una cifra de los tres artículos: se construyen las dos versiones y se miden.
Salida literal en `recursos/pruebas-bff/SALIDA.md`.

    python canales/golandia/investigacion/pruebas_bff.py
"""
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
import probar
from probar import episodio

# Este generador no vive bajo pruebas-entrevista-go: tiene su propia carpeta.
probar.RAIZ = Path(r"<repo>")

# ── El banco de pruebas: tres microservicios de mentira con latencia realista ──────────
BANCO = r'''package banco

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
'''

# ── 1. Sin BFF: el cliente llama a los tres y se arma la pantalla ─────────────────────
SIN = r'''package main

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
'''

# ── 2. Con BFF: una llamada, y el BFF pide los tres EN PARALELO ───────────────────────
CON = r'''package main

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
'''

# ── 3. El «cuándo NO»: con UN solo servicio, el BFF es un salto de más ────────────────
CUANDO_NO = r'''package main

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
'''

# ── 4. El contrato: oapi-codegen genera de verdad, y el compilador lo vigila ──────────
OPENAPI = r'''openapi: 3.0.0
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
'''

CFG = r'''package: api
output: api/api.gen.go
generate:
  models: true
  chi-server: true
'''

# El handler que SI cumple el contrato generado.
HANDLER_OK = r'''package main

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
'''

# El handler que NO lo cumple: falta un metodo. Tiene que fallar la compilacion.
HANDLER_MAL = r'''package main

import "pruebas/api"

type Servidor struct{}

// A proposito: no implementa GetPantalla. El contrato dice que hace falta.
var _ api.ServerInterface = Servidor{}

func main() {}
'''

LENTO = r'''package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"pruebas/banco"
)

// El coste REAL del BFF no es el salto de red: es que ACOPLA la pantalla entera
// al servicio mas lento. Aqui avisos tarda 2 segundos.
func main() {
	u := banco.Nuevo(60*time.Millisecond, banco.UsuarioDemo())
	m := banco.Nuevo(90*time.Millisecond, banco.MovimientosDemo(20))
	a := banco.Nuevo(2000*time.Millisecond, banco.AvisosDemo(6))
	defer u.Close()
	defer m.Close()
	defer a.Close()

	pedir := func(ctx context.Context, url string, destino any) error {
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		return json.NewDecoder(res.Body).Decode(destino)
	}

	// BFF INGENUO: espera a los tres. Un servicio lento tumba la pantalla entera.
	ingenuo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var usr banco.Usuario
		var movs []banco.Movimiento
		var avs []banco.Aviso
		wg.Add(3)
		go func() { defer wg.Done(); pedir(r.Context(), u.URL, &usr) }()
		go func() { defer wg.Done(); pedir(r.Context(), m.URL, &movs) }()
		go func() { defer wg.Done(); pedir(r.Context(), a.URL, &avs) }()
		wg.Wait()
		json.NewEncoder(w).Encode(map[string]any{"nombre": usr.Nombre, "avisos": len(avs)})
	}))
	defer ingenuo.Close()

	// BFF CON PLAZO POR DEPENDENCIA: lo que no llega a tiempo se devuelve como ausente.
	// La pantalla se pinta igual y el usuario ve un hueco, no una rueda infinita.
	sensato := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var usr banco.Usuario
		var movs []banco.Movimiento
		var avs []banco.Aviso
		avisosOK := true
		ctx, cancel := context.WithTimeout(r.Context(), 250*time.Millisecond)
		defer cancel()
		wg.Add(3)
		go func() { defer wg.Done(); pedir(ctx, u.URL, &usr) }()
		go func() { defer wg.Done(); pedir(ctx, m.URL, &movs) }()
		go func() {
			defer wg.Done()
			if err := pedir(ctx, a.URL, &avs); err != nil {
				avisosOK = false
			}
		}()
		wg.Wait()
		json.NewEncoder(w).Encode(map[string]any{
			"nombre": usr.Nombre, "movimientos": len(movs),
			"avisos_disponibles": avisosOK,
		})
	}))
	defer sensato.Close()

	medir := func(nombre, url string) {
		t := time.Now()
		res, err := http.Get(url)
		if err != nil {
			fmt.Println("  error:", err)
			return
		}
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()
		fmt.Printf("  %-26s %-7v  %s", nombre, time.Since(t).Round(time.Millisecond), string(b))
	}

	fmt.Println("UN SERVICIO LENTO (avisos tarda 2s) - aqui SI duele")
	medir("BFF ingenuo:", ingenuo.URL)
	medir("BFF con plazo (250ms):", sensato.URL)
	fmt.Println()
	fmt.Println("El BFF ingenuo acopla la pantalla entera al servicio mas lento.")
	fmt.Println("Sin BFF, el movil habria pintado lo demas y dejado ese hueco cargando.")
}
'''

episodio(
    "pruebas-bff",
    "Vídeo 115 — ¿Necesitas un BFF en Go? (medido, no supuesto)",
    [
        {"nombre": "1. Sin BFF: el cliente llama a los tres servicios",
         "pregunta": "Cuanto cuesta pintar UNA pantalla si el movil habla con los microservicios?",
         "archivos": {"banco/banco.go": BANCO, "sin/main.go": SIN},
         "correr": [(["go", "run", "./sin"], False)]},
        {"nombre": "2. Con BFF: una llamada, y el BFF agrega en paralelo",
         "pregunta": "Y con una capa BFF delante?",
         "archivos": {"conbff/main.go": CON},
         "correr": [(["go", "run", "./conbff"], False)],
         "nota": "Las dos cosas que hace el BFF, y que hay que decir juntas: AGREGA (tres "
                 "llamadas en una) y TRANSFORMA (devuelve solo lo que la pantalla usa)."},
        {"nombre": "3. El «cuando NO»: con UN solo backend, el BFF es un salto de mas",
         "pregunta": "Y si detras solo hay un servicio?",
         "archivos": {"solo/main.go": CUANDO_NO},
         "correr": [(["go", "run", "./solo"], False)],
         "nota": "⚠️ ESTA PRUEBA DESMINTIO LA HIPOTESIS. Se esperaba que el salto extra pesara, "
                 "y en la practica son ~200 microsegundos: nada. O sea que el argumento contra el "
                 "BFF NO es la latencia. Donde duele de verdad es en la prueba 6."},
        {"nombre": "6. Donde SI duele: un servicio lento acopla la pantalla entera",
         "pregunta": "Que pasa cuando una de las dependencias va lenta?",
         "archivos": {"lento/main.go": LENTO},
         "correr": [(["go", "run", "./lento"], False)],
         "nota": "El coste real del BFF: pasa de N llamadas independientes a UNA que depende de "
                 "todas. Sin plazo por dependencia, el servicio mas lento manda sobre la pantalla "
                 "entera. Con plazo, se pinta lo que llego y el hueco se dice."},
        {"nombre": "4. El contrato primero: oapi-codegen genera tipos y servidor",
         "pregunta": "Como se hace contract-first de verdad?",
         "archivos": {"api/openapi.yaml": OPENAPI, "api/cfg.yaml": CFG},
         "correr": [(["go", "run", "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest",
                      "-config", "api/cfg.yaml", "api/openapi.yaml"], False),
                    (["sh", "-c", "head -30 api/api.gen.go"], False)]},
        {"nombre": "5. El contrato lo vigila el COMPILADOR, no una revision",
         "pregunta": "Que pasa si el handler no cumple el contrato?",
         "archivos": {"bien/main.go": HANDLER_OK, "mal/main.go": HANDLER_MAL},
         "correr": [(["go", "get", "github.com/go-chi/chi/v5"], False),
                    (["go", "build", "-o", "/tmp/bien", "./bien"], False),
                    (["go", "build", "-o", "/tmp/mal", "./mal"], True)],
         "nota": "Ese error del compilador es el argumento entero del contract-first: el "
                 "contrato deja de ser un documento y pasa a ser algo que no se puede incumplir."},
    ],
    gomod="module pruebas\n\ngo 1.24\n")
