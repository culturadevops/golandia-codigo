package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"tareas/internal/bd"
	"tareas/internal/config"
	"tareas/internal/cuentas"
	"tareas/internal/servidor"
	"tareas/internal/tareas"
	"tareas/migraciones"
)

func servir(ctx context.Context, getenv func(string) string, out io.Writer) error {
	cfg, err := config.Cargar(getenv)
	if err != nil {
		return fmt.Errorf("configuración:\n%w", err)
	}
	log := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: cfg.NivelLog}))

	pool, err := bd.Abrir(ctx, cfg.BDURL.Valor())
	if err != nil {
		return err
	}
	defer pool.Close()

	salud := &servidor.Salud{BD: bd.Sonda{Pool: pool}, Esperada: migraciones.Ultima()}
	h := servidor.Rutas(servidor.Deps{
		Log:     log,
		Cuentas: cuentas.Nuevo(pool, cfg.CostoBcrypt),
		Tareas:  tareas.Postgres{BD: pool},
		Salud:   salud,
		Plazo:   cfg.TiempoPeticion,
	})

	ln, err := net.Listen("tcp", ":"+cfg.Puerto)
	if err != nil {
		return err
	}
	log.Info("escuchando", "dir", ln.Addr().String(), "bd", cfg.BDURL,
		"esquema_esperado", salud.Esperada, "apagado_max", cfg.ApagadoMax.String())
	return servidor.Servir(ctx, servidor.Nuevo(h, log), ln, salud, cfg.Drenaje, cfg.ApagadoMax, log)
}

func migrar(ctx context.Context, getenv func(string) string, out io.Writer) error {
	dsn := getenv("BD_URL_MIGRAR")
	if dsn == "" {
		return errors.New("BD_URL_MIGRAR: obligatorio (el usuario dueño del esquema)")
	}
	return bd.Migrar(ctx, dsn, out)
}

// salud existe porque la imagen final no tiene curl ni shell: el propio binario es la sonda.
func salud(ctx context.Context, getenv func(string) string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://127.0.0.1:"+getenv("PUERTO")+"/salud/listo", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("listo contestó %d", resp.StatusCode)
	}
	return nil
}
