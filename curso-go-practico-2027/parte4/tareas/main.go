package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := correr(); err != nil {
		nuevoLogger(0).Error("la API no arrancó", "err", err)
		os.Exit(1)
	}
}

// correr devuelve un error en vez de llamar a log.Fatal: así los defer sí se ejecutan.
func correr() error {
	cfg, err := LeerConfig(os.Getenv)
	if err != nil {
		return fmt.Errorf("configuración: %w", err)
	}
	log := nuevoLogger(cfg.NivelLog)

	// ctx se cancela con Ctrl+C (SIGINT) o con SIGTERM, que es lo que manda docker stop.
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	pool, err := ConectarPostgres(ctx, cfg.URLBD, cfg.MaxConexiones)
	if err != nil {
		return err
	}
	defer pool.Close()
	aplicadas, err := Migrar(ctx, pool)
	if err != nil {
		return err
	}
	log.Info("migraciones", "aplicadas", aplicadas)

	api := &API{almacen: &Postgres{pool: pool, timeout: cfg.TimeoutBD}, log: log}
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Puerto),
		Handler:           registrar(log, api.Rutas()),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errServidor := make(chan error, 1)
	go func() {
		log.Info("escuchando", "puerto", cfg.Puerto, "timeout_bd", cfg.TimeoutBD.String())
		errServidor <- srv.ListenAndServe()
	}()

	select {
	case err := <-errServidor:
		return err // no pudo abrir el puerto, por ejemplo
	case <-ctx.Done():
	}

	// Apagado ordenado: deja de aceptar conexiones y espera a las peticiones en curso.
	log.Info("señal recibida: apagando")
	ctxApagado, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	if err := srv.Shutdown(ctxApagado); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("apagado: %w", err)
	}
	log.Info("apagado limpio")
	return nil
}
