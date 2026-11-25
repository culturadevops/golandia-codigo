package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"curso-cli/comun"
)

// resultadoJSON es el contrato con los scripts: nombres estables y unidades en el nombre.
type resultadoJSON struct {
	URL        string  `json:"url"`
	OK         bool    `json:"ok"`
	Estado     int     `json:"estado"`
	LatenciaMS float64 `json:"latencia_ms"`
	Error      string  `json:"error,omitempty"`
}

func revisar(ctx context.Context, args []string, e Entorno) int {
	fs := flag.NewFlagSet("revisar", flag.ContinueOnError)
	fs.SetOutput(e.Stderr)
	f := definirFlags(fs)
	enJSON := fs.Bool("json", false, "salida JSON para scripts")
	soloCaidos := fs.Bool("solo-caidos", false, "muestra solo los endpoints caídos")
	informe := fs.String("informe", "", "además, guarda el informe en este archivo")
	if err := fs.Parse(args); err != nil {
		return codigoParse(err)
	}
	cfg, err := cargar(fs, f, e.Getenv)
	if err != nil {
		fmt.Fprintln(e.Stderr, "vigia revisar:", err)
		return exitUso
	}

	urls := fs.Args()
	if len(urls) == 1 && urls[0] == "-" {
		if urls, err = leerURLs(e.Stdin); err != nil {
			fmt.Fprintln(e.Stderr, "vigia revisar: leyendo stdin:", err)
			return exitUso
		}
	}
	if len(urls) == 0 {
		fmt.Fprintln(e.Stderr, "vigia revisar: no hay URLs (pásalas como argumentos, o «-» para leerlas de stdin)")
		return exitUso
	}

	// El progreso es para la persona que mira: va a stderr y no ensucia la salida.
	fmt.Fprintf(e.Stderr, "revisando %d endpoints (timeout %s, concurrencia %d)\n",
		len(urls), cfg.Timeout, cfg.Concurrencia)
	res := comun.RevisarTodas(ctx, urls, cfg.Timeout, cfg.Concurrencia)
	if ctx.Err() != nil {
		fmt.Fprintln(e.Stderr, "vigia: interrumpido, no escribo nada a medias")
		return exitInterrumpido
	}

	caidos := 0
	var mostrar []comun.Resultado
	for _, r := range res {
		if !r.OK() {
			caidos++
		}
		if !*soloCaidos || !r.OK() {
			mostrar = append(mostrar, r)
		}
	}

	if *informe != "" { // el informe es un archivo: nunca lleva colores
		if err := escribirAtomico(*informe, formatear(mostrar, *enJSON, false)); err != nil {
			fmt.Fprintln(e.Stderr, "vigia revisar:", err)
			return exitUso
		}
	}
	color := esTerminal(e.Stdout) && e.Getenv("NO_COLOR") == ""
	io.WriteString(e.Stdout, formatear(mostrar, *enJSON, color))

	if caidos > 0 {
		fmt.Fprintf(e.Stderr, "%d de %d caídos\n", caidos, len(res))
		return exitCaidos
	}
	return exitOK
}

func leerURLs(r io.Reader) ([]string, error) {
	var urls []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l != "" && !strings.HasPrefix(l, "#") {
			urls = append(urls, l)
		}
	}
	return urls, sc.Err()
}

func formatear(rs []comun.Resultado, enJSON, color bool) string {
	var b strings.Builder
	if enJSON {
		escribirJSON(&b, rs)
	} else {
		escribirHumano(&b, rs, color)
	}
	return b.String()
}

func escribirJSON(w io.Writer, rs []comun.Resultado) {
	out := make([]resultadoJSON, 0, len(rs)) // nunca nil: sin resultados sale [] y no null
	for _, r := range rs {
		out = append(out, resultadoJSON{
			URL: r.URL, OK: r.OK(), Estado: r.Estado, Error: r.Motivo(),
			LatenciaMS: float64(r.Latencia.Microseconds()) / 1000,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}

func escribirHumano(w io.Writer, rs []comun.Resultado, color bool) {
	for _, r := range rs {
		l := comun.Linea(r)
		if color {
			c := "\x1b[32m" // verde
			if !r.OK() {
				c = "\x1b[31m" // rojo
			}
			l = c + l[:6] + "\x1b[0m" + l[6:]
		}
		fmt.Fprintln(w, l)
	}
}

// esTerminal: ¿hay una persona mirando, o la salida va a un pipe o a un archivo?
// Ojo: os.ModeCharDevice no sirve para esto, porque /dev/null también es un dispositivo
// de caracteres. term.IsTerminal le pregunta al sistema si es una terminal de verdad.
func esTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// escribirAtomico escribe en un temporal del mismo directorio y lo renombra al final.
// Quien lea el archivo ve el informe viejo o el nuevo entero, nunca uno a medias.
func escribirAtomico(ruta, contenido string) error {
	tmp, err := os.CreateTemp(filepath.Dir(ruta), ".vigia-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // si algo falla antes del Rename, no queda basura
	if _, err := tmp.WriteString(contenido); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), ruta)
}
