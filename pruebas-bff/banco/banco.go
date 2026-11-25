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
