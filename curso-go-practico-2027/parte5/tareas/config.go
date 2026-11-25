package main

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

// Config sale de variables de entorno: la misma imagen sirve en tu máquina y en el servidor.
type Config struct {
	Puerto        int           // TAREAS_PUERTO, 8080
	URLBD         string        // DATABASE_URL, obligatoria
	TimeoutBD     time.Duration // TAREAS_TIMEOUT_BD, 2s
	MaxConexiones int32         // TAREAS_MAX_CONEXIONES, 4
	NivelLog      slog.Level    // TAREAS_LOG, info
}

// LeerConfig recibe getenv (os.Getenv en main) para poder probarla sin tocar el entorno.
// Devuelve TODOS los errores juntos, no solo el primero.
func LeerConfig(getenv func(string) string) (Config, error) {
	c := Config{Puerto: 8080, TimeoutBD: 2 * time.Second, MaxConexiones: 4, NivelLog: slog.LevelInfo}
	var errs []error

	if v := getenv("TAREAS_PUERTO"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 || p > 65535 {
			errs = append(errs, fmt.Errorf("TAREAS_PUERTO=%q: tiene que ser un número de 1 a 65535", v))
		}
		c.Puerto = p
	}
	c.URLBD = getenv("DATABASE_URL")
	if c.URLBD == "" {
		errs = append(errs, errors.New("DATABASE_URL es obligatoria"))
	}
	if v := getenv("TAREAS_TIMEOUT_BD"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("TAREAS_TIMEOUT_BD=%q: usa una duración como 2s o 500ms", v))
		}
		c.TimeoutBD = d
	}
	if v := getenv("TAREAS_MAX_CONEXIONES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n < 1 {
			errs = append(errs, fmt.Errorf("TAREAS_MAX_CONEXIONES=%q: tiene que ser un número mayor que 0", v))
		}
		c.MaxConexiones = int32(n)
	}
	if v := getenv("TAREAS_LOG"); v != "" {
		if err := c.NivelLog.UnmarshalText([]byte(v)); err != nil {
			errs = append(errs, fmt.Errorf("TAREAS_LOG=%q: usa debug, info, warn o error", v))
		}
	}
	return c, errors.Join(errs...)
}
