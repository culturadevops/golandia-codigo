// Package config lee la configuración del entorno y la valida ENTERA al arrancar.
//
// Regla de la plantilla: si falta algo o algo está mal, el proceso no arranca y dice
// todos los problemas a la vez, no el primero.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"
)

// Config es todo lo que el servicio necesita saber del entorno.
type Config struct {
	Puerto         string        // PUERTO, obligatorio: 1-65535
	BDURL          DSN           // BD_URL, obligatorio: el usuario de la app, sin permisos de DDL
	TiempoPeticion time.Duration // TIEMPO_PETICION: plazo de cada petición (10s)
	ApagadoMax     time.Duration // APAGADO_MAX: lo que espera a las peticiones en vuelo (8s)
	Drenaje        time.Duration // DRENAJE: listo=503 durante esto antes de cerrar (0s)
	NivelLog       slog.Level    // NIVEL_LOG: debug, info, warn, error (info)
	CostoBcrypt    int           // COSTO_BCRYPT: 10-14 (12)
}

// Cargar lee las variables con getenv (os.Getenv en producción, un map en los tests).
func Cargar(getenv func(string) string) (Config, error) {
	c := Config{}
	var errs []error
	mal := func(v, format string, a ...any) {
		errs = append(errs, fmt.Errorf("%s: "+format, append([]any{v}, a...)...))
	}

	c.Puerto = getenv("PUERTO")
	if n, err := strconv.Atoi(c.Puerto); c.Puerto == "" {
		mal("PUERTO", "obligatorio")
	} else if err != nil || n < 1 || n > 65535 {
		mal("PUERTO", "%q no es un puerto (1-65535)", c.Puerto)
	}

	c.BDURL = DSN(getenv("BD_URL"))
	if err := c.BDURL.validar(); err != nil {
		mal("BD_URL", "%v", err)
	}

	c.TiempoPeticion = duracion(getenv, "TIEMPO_PETICION", 10*time.Second, mal)
	c.ApagadoMax = duracion(getenv, "APAGADO_MAX", 8*time.Second, mal)
	c.Drenaje = duracion(getenv, "DRENAJE", 0, mal)

	if err := c.NivelLog.UnmarshalText([]byte(valor(getenv, "NIVEL_LOG", "info"))); err != nil {
		mal("NIVEL_LOG", "%q no es debug, info, warn ni error", getenv("NIVEL_LOG"))
	}

	costo, err := strconv.Atoi(valor(getenv, "COSTO_BCRYPT", "12"))
	if err != nil || costo < 10 || costo > 14 {
		mal("COSTO_BCRYPT", "%q fuera de 10-14", getenv("COSTO_BCRYPT"))
	}
	c.CostoBcrypt = costo

	return c, errors.Join(errs...)
}

func valor(getenv func(string) string, k, defecto string) string {
	if v := getenv(k); v != "" {
		return v
	}
	return defecto
}

func duracion(getenv func(string) string, k string, defecto time.Duration, mal func(string, string, ...any)) time.Duration {
	v := getenv(k)
	if v == "" {
		return defecto
	}
	d, err := time.ParseDuration(v)
	if err != nil || d < 0 {
		mal(k, "%q no es una duración (ejemplo: 30s, 500ms)", v)
		return defecto
	}
	return d
}

// DSN es una cadena de conexión que nunca se imprime con la contraseña.
type DSN string

func (d DSN) validar() error {
	if d == "" {
		return errors.New("obligatorio")
	}
	u, err := url.Parse(string(d))
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
		return errors.New("no es una URL postgres://usuario:clave@host:puerto/base")
	}
	return nil
}

// String tapa la contraseña: así sale en fmt, en %v y en los logs.
func (d DSN) String() string {
	u, err := url.Parse(string(d))
	if err != nil {
		return "(DSN ilegible)"
	}
	return u.Redacted()
}

// LogValue hace lo mismo para slog.
func (d DSN) LogValue() slog.Value { return slog.StringValue(d.String()) }

// Valor devuelve la cadena real, para conectarse. Es la única salida sin tapar.
func (d DSN) Valor() string { return string(d) }
