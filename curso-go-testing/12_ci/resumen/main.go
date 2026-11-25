// Lección 12 · go test -json en CI.
//
// resumen lee la salida de «go test -json» por la entrada estándar y escribe lo que importa
// en una CI: cuántos pasaron, cuáles fallaron (con su salida) y los más lentos.
// Sale con código 1 si algo falló, para que el paso de la CI se ponga en rojo aunque vaya
// detrás de una tubería.
//
//	go test -json ./... | go run ./12_ci/resumen
//	go test -json ./... | go run ./12_ci/resumen -lentos 5 -md resumen.md
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"
)

// Evento es una línea de «go test -json» (ver «go doc test2json»).
type Evento struct {
	Time    time.Time
	Action  string // start, run, pause, cont, pass, fail, skip, output, bench, build-output, build-fail
	Package string
	Test    string
	Elapsed float64
	Output  string
}

type prueba struct {
	nombre string
	seg    float64
}

func main() {
	nLentos := flag.Int("lentos", 5, "cuántos tests lentos enseñar")
	maxFallos := flag.Int("max-fallos", 10, "cuántos fallos detallar")
	md := flag.String("md", "", "escribe también un resumen en Markdown (p. ej. $GITHUB_STEP_SUMMARY)")
	flag.Parse()
	r := leer(os.Stdin)
	var b strings.Builder
	r.escribir(&b, *nLentos, *maxFallos)
	fmt.Print(b.String())
	if *md != "" {
		os.WriteFile(*md, []byte("```\n"+b.String()+"```\n"), 0o644)
	}
	if len(r.fallos) > 0 || r.paquetesRotos > 0 {
		os.Exit(1)
	}
}

type resultado struct {
	pasan, fallan, saltan, paquetes, paquetesRotos int
	fallos                                         []string
	salida                                         map[string][]string
	lentos                                         []prueba
}

func leer(in io.Reader) *resultado {
	r := &resultado{salida: map[string][]string{}}
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var e Evento
		if json.Unmarshal(sc.Bytes(), &e) != nil {
			continue // líneas que no son JSON (p. ej. avisos del compilador)
		}
		raiz, _, _ := strings.Cut(e.Test, "/") // la salida de un subtest va con su test padre
		clave := e.Package + " " + raiz
		switch {
		case e.Action == "output" && e.Test != "":
			r.salida[clave] = append(r.salida[clave], e.Output)
		case e.Test == "" && e.Action == "pass":
			r.paquetes++
		case e.Test == "" && e.Action == "fail":
			r.paquetes++
			r.paquetesRotos++
		case e.Action == "pass" && !strings.Contains(e.Test, "/"):
			r.pasan++
			r.lentos = append(r.lentos, prueba{clave, e.Elapsed})
		case e.Action == "fail" && !strings.Contains(e.Test, "/"):
			r.fallan++
			r.fallos = append(r.fallos, clave)
		case e.Action == "skip" && !strings.Contains(e.Test, "/"):
			r.saltan++
		}
	}
	slices.SortFunc(r.lentos, func(a, b prueba) int {
		return int((b.seg - a.seg) * 1000)
	})
	return r
}

func (r *resultado) escribir(w io.Writer, nLentos, maxFallos int) {
	fmt.Fprintf(w, "%d paquetes · %d tests: %d pasan, %d fallan, %d saltados\n",
		r.paquetes, r.pasan+r.fallan+r.saltan, r.pasan, r.fallan, r.saltan)
	if len(r.fallos) > 0 {
		fmt.Fprintln(w, "\nFALLAN:")
		for i, f := range r.fallos {
			if i == maxFallos {
				fmt.Fprintf(w, "  ... y %d fallos más\n", len(r.fallos)-maxFallos)
				break
			}
			fmt.Fprintf(w, "  %s\n", f)
			var lineas []string
			for _, l := range r.salida[f] {
				if t := strings.TrimSpace(l); t != "" && !strings.HasPrefix(t, "=== ") && !strings.HasPrefix(t, "--- ") {
					lineas = append(lineas, t)
				}
			}
			for _, t := range lineas[:min(4, len(lineas))] {
				fmt.Fprintf(w, "      %s\n", t)
			}
			if len(lineas) > 4 {
				fmt.Fprintf(w, "      ... (%d líneas más)\n", len(lineas)-4)
			}
		}
	}
	if nLentos == 0 {
		return
	}
	fmt.Fprintf(w, "\nLos %d más lentos:\n", min(nLentos, len(r.lentos)))
	for _, p := range r.lentos[:min(nLentos, len(r.lentos))] {
		fmt.Fprintf(w, "  %6.2f s  %s\n", p.seg, p.nombre)
	}
}
