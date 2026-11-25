// El banco de pruebas: sitios locales que se caen y vuelven, y un Discord FALSO que solo apunta lo que recibe.
// Nada sale de la máquina: ningún webhook real, ninguna credencial.
//
//	./monitor ingenuo | estados | sin-timeout | rondas | secuencial | entorno
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var inicio time.Time

func t() string { return fmt.Sprintf("%5.1f s", time.Since(inicio).Round(100*time.Millisecond).Seconds()) }

// servidor arranca un sitio de prueba en 127.0.0.1 y devuelve su URL y cómo apagarlo/encenderlo.
type servidor struct {
	addr string
	h    http.Handler
	mu   sync.Mutex
	srv  *http.Server
}

func nuevo(h http.Handler) *servidor {
	s := &servidor{h: h}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	s.addr = ln.Addr().String()
	s.servir(ln)
	return s
}

func (s *servidor) servir(ln net.Listener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.srv = &http.Server{Handler: s.h}
	go s.srv.Serve(ln)
}

func (s *servidor) URL() string { return "http://" + s.addr }

func (s *servidor) Apagar() { s.mu.Lock(); s.srv.Close(); s.mu.Unlock() }

func (s *servidor) Encender() {
	ln, err := net.Listen("tcp", s.addr) // el mismo puerto: el sitio «vuelve»
	if err != nil {
		panic(err)
	}
	s.servir(ln)
}

func ok() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
}

// intermitente: un 500 sí, un 200 no. El clásico sitio que «parpadea».
func intermitente() http.Handler {
	var n atomic.Int64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1)%2 == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte("ok"))
	})
}

// lento: contesta 200, pero a los 3 segundos.
func lento(d time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
			w.Write([]byte("ok"))
		case <-r.Context().Done():
		}
	})
}

// discordFalso imita el endpoint de webhooks de Discord: acepta {"content": "..."} y responde 204.
type discordFalso struct {
	mu       sync.Mutex
	mensajes []string
	silencio bool
}

func (d *discordFalso) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var m struct {
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&m)
	d.mu.Lock()
	d.mensajes = append(d.mensajes, m.Content)
	if !d.silencio {
		fmt.Printf("%s  discord ← %s\n", t(), m.Content)
	}
	d.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

type banco struct {
	web, api, pagos, lento *servidor
	discord                *discordFalso
	webhook                string
}

func montar() *banco {
	b := &banco{web: nuevo(ok()), api: nuevo(ok()), pagos: nuevo(intermitente()), lento: nuevo(lento(3 * time.Second)),
		discord: &discordFalso{}}
	d := nuevo(b.discord)
	b.webhook = d.URL() + "/api/webhooks/demo/falso"
	return b
}

func (b *banco) sitios() []Sitio {
	return []Sitio{{"web", b.web.URL()}, {"api", b.api.URL()}, {"pagos", b.pagos.URL()}, {"lento", b.lento.URL()}}
}

// guion: api se cae en el segundo 1 y vuelve en el 3.
func (b *banco) guion() {
	go func() {
		time.Sleep(1 * time.Second)
		b.api.Apagar()
		fmt.Printf("%s  ── api se apaga\n", t())
		time.Sleep(2 * time.Second)
		b.api.Encender()
		fmt.Printf("%s  ── api vuelve\n", t())
	}()
}

func correr(m *Monitor, sitios []Sitio, dur time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	var wg sync.WaitGroup
	for _, s := range sitios {
		wg.Go(func() { m.Vigilar(ctx, s) })
	}
	wg.Wait()
}

func porSitio(b *banco) map[string]int {
	cuenta := map[string]int{}
	for _, m := range b.discord.mensajes {
		for _, s := range []string{"web", "api", "pagos", "lento"} {
			if strings.Contains(m, " "+s+" ") {
				cuenta[s]++
			}
		}
	}
	return cuenta
}

func resumen(b *banco, dur time.Duration) {
	c := porSitio(b)
	fmt.Printf("  avisos en %.0f s: %d  (web %d · api %d · pagos %d · lento %d)\n", dur.Seconds(), len(b.discord.mensajes), c["web"], c["api"], c["pagos"], c["lento"])
}

func main() {
	modo := "estados"
	if len(os.Args) > 1 {
		modo = os.Args[1]
	}
	inicio = time.Now()
	switch modo {
	case "ingenuo", "estados":
		b := montar()
		b.guion()
		m := &Monitor{Cliente: &http.Client{Timeout: 500 * time.Millisecond}, Intervalo: 200 * time.Millisecond,
			Umbral: 3, SoloCambios: modo == "estados", Alertar: Discord(b.webhook)}
		fmt.Println("  sitios: web (bien) · api (se cae del 1 al 3) · pagos (500 una sí, una no) · lento (tarda 3 s)")
		fmt.Printf("  cada 200 ms · timeout 500 ms · umbral %d · solo cambios: %v\n", m.Umbral, m.SoloCambios)
		correr(m, b.sitios(), 5*time.Second)
		resumen(b, 5*time.Second)

	case "sin-timeout":
		b := montar()
		m := &Monitor{Cliente: &http.Client{}, Intervalo: 200 * time.Millisecond, Umbral: 3, SoloCambios: true,
			Alertar: Discord(b.webhook)}
		m.AlRevisar = func(s Sitio, d time.Duration, err error) {
			if s.Nombre == "lento" {
				fmt.Printf("%s  lento: %v en %.1f s\n", t(), estado(err), d.Round(100*time.Millisecond).Seconds())
			}
		}
		fmt.Println("  http.Client{} sin Timeout · lento tarda 3 s en contestar 200")
		correr(m, []Sitio{{"lento", b.lento.URL()}}, 7*time.Second)
		resumen(b, 7*time.Second)

	case "rondas":
		// ¿Cuántas veces se revisa «web» en 5 s si un sitio lento comparte la ronda?
		b := montar()
		for _, v := range []struct {
			nombre  string
			cliente *http.Client
			rondas  bool
		}{
			{"por rondas, sin timeout   ", &http.Client{}, true},
			{"por rondas, timeout 500 ms", &http.Client{Timeout: 500 * time.Millisecond}, true},
			{"por sitio,  timeout 500 ms", &http.Client{Timeout: 500 * time.Millisecond}, false},
		} {
			var web atomic.Int64
			m := &Monitor{Cliente: v.cliente, Intervalo: 200 * time.Millisecond, Umbral: 3, SoloCambios: true,
				Alertar: func(string) {}}
			m.AlRevisar = func(s Sitio, _ time.Duration, _ error) {
				if s.Nombre == "web" {
					web.Add(1)
				}
			}
			s := []Sitio{{"web", b.web.URL()}, {"lento", b.lento.URL()}}
			if v.rondas {
				porRondas(m, s, 5*time.Second)
			} else {
				correr(m, s, 5*time.Second)
			}
			fmt.Printf("  %s → web revisado %2d veces en 5 s\n", v.nombre, web.Load())
		}

	case "secuencial":
		// 10 sitios que tardan 300 ms cada uno: una ronda en fila contra una ronda con una goroutine por sitio.
		var sitios []Sitio
		for i := range 10 {
			sitios = append(sitios, Sitio{fmt.Sprintf("s%d", i), nuevo(lento(300 * time.Millisecond)).URL()})
		}
		m := &Monitor{Cliente: &http.Client{Timeout: 2 * time.Second}}
		ini := time.Now()
		for _, s := range sitios {
			m.Revisar(context.Background(), s)
		}
		fmt.Printf("  en fila (for):            %.1f s\n", time.Since(ini).Round(100*time.Millisecond).Seconds())
		ini = time.Now()
		var wg sync.WaitGroup
		for _, s := range sitios {
			wg.Go(func() { m.Revisar(context.Background(), s) })
		}
		wg.Wait()
		fmt.Printf("  una goroutine por sitio: %.1f s\n", time.Since(ini).Round(100*time.Millisecond).Seconds())

	case "entorno":
		// Sin la variable, el monitor no arranca: mejor que vigilar sin poder avisar.
		log.SetFlags(0)
		Discord(webhookDeEntorno())

	default:
		fmt.Println("modos: ingenuo | estados | sin-timeout | rondas | secuencial | entorno")
	}
}

// porRondas: la versión que parece natural — revisa todos, espera a que acaben TODOS, duerme, repite.
func porRondas(m *Monitor, sitios []Sitio, dur time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	for ctx.Err() == nil {
		var wg sync.WaitGroup
		for _, s := range sitios {
			wg.Go(func() {
				ini := time.Now()
				err := m.Revisar(ctx, s)
				if ctx.Err() == nil {
					m.AlRevisar(s, time.Since(ini), err)
				}
			})
		}
		wg.Wait()
		select {
		case <-ctx.Done():
		case <-time.After(m.Intervalo):
		}
	}
}

func estado(err error) string {
	if err != nil {
		return "FALLO (" + motivo(err) + ")"
	}
	return "200 OK"
}
