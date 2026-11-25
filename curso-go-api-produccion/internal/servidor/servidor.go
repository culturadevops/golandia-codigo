// Package servidor arma las rutas, los middlewares y el ciclo de vida del http.Server.
package servidor

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"tareas/internal/cuentas"
	"tareas/internal/tareas"
	"tareas/internal/web"
)

// Deps es todo lo que el servidor recibe hecho. Nada de variables globales.
type Deps struct {
	Log     *slog.Logger
	Cuentas *cuentas.Servicio
	Tareas  tareas.Almacen
	Salud   *Salud
	Plazo   time.Duration // por petición
}

// Rutas devuelve el handler completo: mux + middlewares en su orden.
func Rutas(d Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /salud/vivo", d.Salud.vivo)
	mux.HandleFunc("GET /salud/listo", d.Salud.listo)
	d.Cuentas.Rutas(mux)
	tareas.Rutas(mux, d.Tareas, d.Cuentas.Requerir)

	return web.Cadena(mux,
		web.ConID(d.Log),   // 1.º: todo lo de dentro ya tiene request_id
		web.Registro,       // 2.º: registra también los 500 del pánico
		web.Recuperar,      // 3.º: el pánico no llega a net/http
		web.ErroresDelMux,  // 404 y 405 del mux en JSON
		web.Plazo(d.Plazo), // y el plazo, lo más cerca posible del handler
	)
}

// Nuevo crea el http.Server con plazos. Sin ReadHeaderTimeout, un cliente que manda las
// cabeceras byte a byte tiene una conexión abierta para siempre (Slowloris).
func Nuevo(h http.Handler, log *slog.Logger) *http.Server {
	return &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}
}

// Servir atiende en ln hasta que ctx se cancela (SIGTERM) y entonces apaga en orden:
// listo=503, espera el drenaje, deja de aceptar y espera a las peticiones en vuelo hasta max.
// Devuelve cuando la última petición ha terminado, no antes.
func Servir(ctx context.Context, srv *http.Server, ln net.Listener, s *Salud, drenaje, max time.Duration, log *slog.Logger) error {
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return err // no llegó a apagarse: el puerto, por ejemplo
	case <-ctx.Done():
	}

	t0 := time.Now()
	log.Info("apagando", "drenaje", drenaje.String(), "max", max.String())
	s.Apagando()
	time.Sleep(drenaje)

	shctx, cancel := context.WithTimeout(context.Background(), max)
	defer cancel()
	err := srv.Shutdown(shctx) // deja de aceptar y espera a las que están en vuelo
	if err != nil {
		log.Warn("quedaban peticiones al vencer el plazo; se cortan", "err", err)
		_ = srv.Close()
	}
	if e := <-errc; !errors.Is(e, http.ErrServerClosed) {
		err = errors.Join(err, e)
	}
	log.Info("apagado", "ms", time.Since(t0).Milliseconds())
	return err
}
