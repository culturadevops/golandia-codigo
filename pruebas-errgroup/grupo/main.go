// Demos de errgroup para el vídeo. Tres descargas simuladas: «precios» falla a los 100 ms, «stock» y «fotos» tardan 300 ms.
// Los tiempos se redondean a 10 ms: lo que importa es la diferencia, no el milisegundo.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

type tarea struct {
	nombre string
	dura   time.Duration
	falla  bool
}

var tareas = []tarea{{"precios", 100 * time.Millisecond, true}, {"stock", 300 * time.Millisecond, false}, {"fotos", 300 * time.Millisecond, false}}

var inicio time.Time

func ms() int64 { return time.Since(inicio).Round(10*time.Millisecond).Milliseconds() }

// hacer espera lo que dura la tarea o hasta que cancelen el contexto.
func hacer(ctx context.Context, t tarea) error {
	select {
	case <-time.After(t.dura):
		if t.falla {
			fmt.Printf("  %4d ms  %-7s falla\n", ms(), t.nombre)
			return fmt.Errorf("%s: la API respondió 503", t.nombre)
		}
		fmt.Printf("  %4d ms  %-7s terminó\n", ms(), t.nombre)
		return nil
	case <-ctx.Done():
		fmt.Printf("  %4d ms  %-7s cancelada (%v)\n", ms(), t.nombre, context.Cause(ctx))
		return ctx.Err()
	}
}

func main() {
	inicio = time.Now()
	switch os.Args[1] {
	case "waitgroup":
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errs []error
		for _, t := range tareas {
			wg.Go(func() {
				if err := hacer(context.Background(), t); err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			})
		}
		wg.Wait()
		fmt.Printf("  %4d ms  total · error: %v\n", ms(), errors.Join(errs...))
	case "errgroup":
		g, ctx := errgroup.WithContext(context.Background())
		for _, t := range tareas {
			g.Go(func() error { return hacer(ctx, t) })
		}
		err := g.Wait()
		fmt.Printf("  %4d ms  total · error: %v\n", ms(), err)
	case "sin-contexto": // errgroup.Group a secas: junta el error, pero no cancela nada
		var g errgroup.Group
		for _, t := range tareas {
			g.Go(func() error { return hacer(context.Background(), t) })
		}
		err := g.Wait()
		fmt.Printf("  %4d ms  total · error: %v\n", ms(), err)
	case "dos-errores":
		g, ctx := errgroup.WithContext(context.Background())
		g.Go(func() error { return hacer(ctx, tarea{"precios", 100 * time.Millisecond, true}) })
		g.Go(func() error { return hacer(context.Background(), tarea{"stock", 150 * time.Millisecond, true}) })
		err := g.Wait()
		fmt.Printf("  %4d ms  Wait devolvió: %v\n", ms(), err)
	case "http", "http-sin-ctx": // las tres llamadas de verdad, por HTTP, a un servidor local
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dura := 300 * time.Millisecond
			if r.URL.Path == "/precios" {
				dura = 100 * time.Millisecond
			}
			select {
			case <-time.After(dura):
				if r.URL.Path == "/precios" {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
			case <-r.Context().Done():
				fmt.Printf("  %4d ms  servidor: el cliente cortó %s\n", ms(), r.URL.Path)
			}
		}))
		defer srv.Close()
		g, ctx := errgroup.WithContext(context.Background())
		for _, ruta := range []string{"/precios", "/stock", "/fotos"} {
			g.Go(func() error {
				var req *http.Request
				if os.Args[1] == "http" {
					req, _ = http.NewRequestWithContext(ctx, "GET", srv.URL+ruta, nil)
				} else {
					req, _ = http.NewRequest("GET", srv.URL+ruta, nil) // sin el ctx del grupo
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Printf("  %4d ms  cliente: %s cancelada\n", ms(), ruta)
					return err
				}
				resp.Body.Close()
				fmt.Printf("  %4d ms  cliente: %s respondió %d\n", ms(), ruta, resp.StatusCode)
				if resp.StatusCode != 200 {
					return fmt.Errorf("%s: la API respondió %d", ruta, resp.StatusCode)
				}
				return nil
			})
		}
		err := g.Wait()
		time.Sleep(50 * time.Millisecond) // deja que el servidor imprima lo suyo
		fmt.Printf("  %4d ms  total (Wait) · error: %v\n", ms()-50, err)
	case "pipeline", "pipeline-sordo", "pipeline-atascado": // un productor y 3 trabajadores; el item 100 falla
		// sordo: los trabajadores no miran ctx · atascado: el productor no mira ctx
		g, ctx := errgroup.WithContext(context.Background())
		items := make(chan int)
		var procesados atomic.Int32
		g.Go(func() error { // productor: 1.000 items
			defer close(items)
			for i := 1; i <= 1000; i++ {
				if os.Args[1] != "pipeline-atascado" {
					select {
					case items <- i:
					case <-ctx.Done():
						fmt.Printf("  productor: paro en el item %d (%v)\n", i, context.Cause(ctx))
						return ctx.Err()
					}
				} else {
					items <- i // sin mirar ctx
				}
			}
			return nil
		})
		for w := 0; w < 3; w++ {
			g.Go(func() error {
				for it := range items {
					if os.Args[1] != "pipeline-sordo" && ctx.Err() != nil {
						return ctx.Err()
					}
					time.Sleep(time.Millisecond)
					if it == 100 {
						return fmt.Errorf("item %d: dato corrupto", it)
					}
					procesados.Add(1)
				}
				return nil
			})
		}
		listo := make(chan error)
		go func() { listo <- g.Wait() }()
		select {
		case err := <-listo:
			fmt.Printf("  Wait: %v · procesados: %d de 1000\n", err, procesados.Load())
		case <-time.After(2 * time.Second):
			fmt.Printf("  a los 2 s Wait sigue bloqueado · procesados: %d de 1000 · el productor quedó atascado\n", procesados.Load())
		}
	case "limite":
		var g errgroup.Group
		g.SetLimit(3)
		var activas, maximo atomic.Int32
		for i := 0; i < 10; i++ {
			g.Go(func() error {
				n := activas.Add(1)
				for {
					m := maximo.Load()
					if n <= m || maximo.CompareAndSwap(m, n) {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
				activas.Add(-1)
				return nil
			})
		}
		g.Wait()
		fmt.Printf("  10 tareas de 50 ms con SetLimit(3): máximo a la vez = %d · total = %d ms\n", maximo.Load(), ms())
		var sin errgroup.Group
		activas.Store(0)
		maximo.Store(0)
		inicio = time.Now()
		for i := 0; i < 10; i++ {
			sin.Go(func() error {
				n := activas.Add(1)
				for {
					m := maximo.Load()
					if n <= m || maximo.CompareAndSwap(m, n) {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
				activas.Add(-1)
				return nil
			})
		}
		sin.Wait()
		fmt.Printf("  10 tareas de 50 ms sin límite:     máximo a la vez = %d · total = %d ms\n", maximo.Load(), ms())
	}
}
