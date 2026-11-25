package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"
)

// Config es la configuración efectiva, y Origen dice de dónde salió cada valor.
type Config struct {
	Timeout      time.Duration
	Concurrencia int
	Origen       map[string]string
}

// archivoConfig es la forma de vigia.json. Los punteros distinguen «no está» de «vale cero».
type archivoConfig struct {
	Timeout      *string `json:"timeout"`
	Concurrencia *int    `json:"concurrencia"`
}

type flagsConfig struct {
	timeout      time.Duration
	concurrencia int
	archivo      string
}

// definirFlags registra en fs los flags de configuración. Los valores por defecto que se
// ven en -h son los de verdad, pero NO se usan para decidir: eso lo hace cargar().
func definirFlags(fs *flag.FlagSet) *flagsConfig {
	f := &flagsConfig{}
	fs.DurationVar(&f.timeout, "timeout", 10*time.Second, "tiempo máximo por endpoint")
	fs.IntVar(&f.concurrencia, "concurrencia", 4, "endpoints revisados a la vez")
	fs.StringVar(&f.archivo, "config", "", "archivo de configuración (por defecto ./vigia.json)")
	return f
}

// cargar aplica las capas de abajo arriba: defecto, archivo, entorno, flags.
// Se llama DESPUÉS de fs.Parse.
func cargar(fs *flag.FlagSet, f *flagsConfig, getenv func(string) string) (Config, error) {
	c := Config{Timeout: 10 * time.Second, Concurrencia: 4,
		Origen: map[string]string{"timeout": "defecto", "concurrencia": "defecto"}}

	// 1) archivo: el de -config, si no el de VIGIA_CONFIG, si no ./vigia.json (si existe)
	ruta, explicito := f.archivo, f.archivo != ""
	if ruta == "" {
		ruta, explicito = getenv("VIGIA_CONFIG"), getenv("VIGIA_CONFIG") != ""
	}
	if ruta == "" {
		ruta = "vigia.json"
	}
	if err := leerArchivo(&c, ruta, explicito); err != nil {
		return c, err
	}

	// 2) entorno
	if v := getenv("VIGIA_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return c, fmt.Errorf("VIGIA_TIMEOUT=%q: %w", v, err)
		}
		c.Timeout, c.Origen["timeout"] = d, "entorno VIGIA_TIMEOUT"
	}
	if v := getenv("VIGIA_CONCURRENCIA"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("VIGIA_CONCURRENCIA=%q: %w", v, err)
		}
		c.Concurrencia, c.Origen["concurrencia"] = n, "entorno VIGIA_CONCURRENCIA"
	}

	// 3) flags: SOLO los que se escribieron. fs.Visit recorre los puestos a mano;
	//    fs.VisitAll recorrería también los que tienen su valor por defecto.
	fs.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "timeout":
			c.Timeout, c.Origen["timeout"] = f.timeout, "flag -timeout"
		case "concurrencia":
			c.Concurrencia, c.Origen["concurrencia"] = f.concurrencia, "flag -concurrencia"
		}
	})

	if c.Timeout <= 0 {
		return c, fmt.Errorf("timeout %s: tiene que ser mayor que cero (%s)", c.Timeout, c.Origen["timeout"])
	}
	if c.Concurrencia < 1 {
		return c, fmt.Errorf("concurrencia %d: tiene que ser 1 o más (%s)", c.Concurrencia, c.Origen["concurrencia"])
	}
	return c, nil
}

func leerArchivo(c *Config, ruta string, explicito bool) error {
	datos, err := os.ReadFile(ruta)
	if errors.Is(err, fs.ErrNotExist) && !explicito {
		return nil // el archivo por defecto es opcional; el que pides con -config, no
	}
	if err != nil {
		return err
	}
	var a archivoConfig
	if err := json.Unmarshal(datos, &a); err != nil {
		return fmt.Errorf("%s: %w", ruta, err)
	}
	if a.Timeout != nil {
		d, err := time.ParseDuration(*a.Timeout)
		if err != nil {
			return fmt.Errorf("%s: timeout %q: %w", ruta, *a.Timeout, err)
		}
		c.Timeout, c.Origen["timeout"] = d, "archivo "+ruta
	}
	if a.Concurrencia != nil {
		c.Concurrencia, c.Origen["concurrencia"] = *a.Concurrencia, "archivo "+ruta
	}
	return nil
}

func mostrarConfig(args []string, e Entorno) int {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(e.Stderr)
	f := definirFlags(fs)
	if err := fs.Parse(args); err != nil {
		return codigoParse(err)
	}
	c, err := cargar(fs, f, e.Getenv)
	if err != nil {
		fmt.Fprintln(e.Stderr, "vigia config:", err)
		return exitUso
	}
	fmt.Fprintf(e.Stdout, "timeout       %-6s (%s)\n", c.Timeout, c.Origen["timeout"])
	fmt.Fprintf(e.Stdout, "concurrencia  %-6d (%s)\n", c.Concurrencia, c.Origen["concurrencia"])
	return exitOK
}

// codigoParse: -h no es un error (código 0); un flag mal, sí (código 2).
func codigoParse(err error) int {
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}
	return exitUso
}
