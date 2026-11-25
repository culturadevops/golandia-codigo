// dockerfile-ia: un LLM local (Ollama) escribe el Dockerfile de un proyecto y lo arregla con los errores de docker build.
//
//	go run . [-modelo qwen2.5-coder:7b] [-intentos 4] [-errores cola|relevantes] <carpeta-del-proyecto>
//
// Solo librería estándar. Pasos: lee los archivos del proyecto; pide el Dockerfile; docker build; si falla, le devuelve el
// error al modelo; si construye, arranca el contenedor y comprueba GET /salud. Imprime cada intento.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const puerto = 8080

type mensaje struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func leerProyecto(dir string) string {
	var b strings.Builder
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() == "Dockerfile" {
			return nil
		}
		info, _ := d.Info()
		if info.Size() >= 4000 {
			return nil
		}
		datos, _ := os.ReadFile(p)
		rel, _ := filepath.Rel(dir, p)
		fmt.Fprintf(&b, "--- %s ---\n%s\n", filepath.ToSlash(rel), datos)
		return nil
	})
	return b.String()
}

func preguntar(modelo string, msgs []mensaje) (string, error) {
	cuerpo, _ := json.Marshal(map[string]any{"model": modelo, "messages": msgs, "stream": false,
		"options": map[string]any{"temperature": 0, "seed": 42}})
	resp, err := http.Post("http://127.0.0.1:11434/api/chat", "application/json", bytes.NewReader(cuerpo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r struct {
		Message mensaje `json:"message"`
	}
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Message.Content, err
}

var bloque = regexp.MustCompile("(?s)```(?:dockerfile|Dockerfile)?\\s*\\n(.*?)```")

func extraer(texto string) string {
	if m := bloque.FindStringSubmatch(texto); m != nil {
		texto = m[1]
	}
	return strings.TrimSpace(texto) + "\n"
}

func sh(dir string, args ...string) (string, error) {
	c := exec.Command(args[0], args[1:]...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return string(out), err
}

func ultimas(s string, n int) string {
	l := strings.Split(strings.TrimSpace(s), "\n")
	if len(l) > n {
		l = l[len(l)-n:]
	}
	return strings.Join(l, "\n")
}

var relevante = regexp.MustCompile(`(?i)error|not found|no such|cannot|missing|failed|requires`)
var prefijo = regexp.MustCompile(`^#\d+ [\d.]+ `)

// resumenError: "cola" = solo las últimas 15 líneas (v1); "relevantes" = las líneas que dicen POR QUÉ, de todo el log (v2).
func resumenError(log, modo string) string {
	var lineas []string
	for _, l := range strings.Split(strings.TrimSpace(log), "\n") {
		if !strings.Contains(l, "preface") {
			lineas = append(lineas, l)
		}
	}
	if modo == "cola" {
		return ultimas(strings.Join(lineas, "\n"), 15)
	}
	var cand []string
	for _, l := range lineas {
		if relevante.MatchString(l) {
			cand = append(cand, l)
		}
	}
	if len(lineas) > 5 {
		cand = append(cand, lineas[len(lineas)-5:]...)
	}
	vistas := map[string]bool{}
	var salida []string
	for _, l := range cand {
		c := prefijo.ReplaceAllString(l, "")
		if !vistas[c] {
			vistas[c] = true
			salida = append(salida, c)
		}
	}
	if len(salida) > 25 {
		salida = salida[len(salida)-25:]
	}
	return strings.Join(salida, "\n")
}

func probar(dir, etiqueta, modo string) (bool, string) {
	if out, err := sh(dir, "docker", "build", "--progress=plain", "-t", etiqueta, "."); err != nil {
		return false, "docker build falló:\n" + resumenError(out, modo)
	}
	sh(dir, "docker", "rm", "-f", "ia-demo")
	if out, err := sh(dir, "docker", "run", "-d", "--name", "ia-demo", "-p", fmt.Sprintf("18080:%d", puerto), etiqueta); err != nil {
		return false, "docker run falló:\n" + ultimas(out, 10)
	}
	defer sh(dir, "docker", "rm", "-f", "ia-demo")
	cli := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		if r, err := cli.Get("http://127.0.0.1:18080/salud"); err == nil {
			var b bytes.Buffer
			b.ReadFrom(r.Body)
			r.Body.Close()
			return true, fmt.Sprintf("GET /salud -> %d %s", r.StatusCode, strings.TrimSpace(b.String()))
		}
	}
	logs, _ := sh(dir, "docker", "logs", "ia-demo")
	return false, "el contenedor arrancó pero /salud no respondió. Logs:\n" + ultimas(logs, 10)
}

func main() {
	modelo := flag.String("modelo", "qwen2.5-coder:7b", "modelo de Ollama")
	intentos := flag.Int("intentos", 4, "intentos máximos")
	modo := flag.String("errores", "relevantes", "qué parte del error se le pasa al modelo: cola | relevantes")
	flag.Parse()
	dir, _ := filepath.Abs(flag.Arg(0))
	msgs := []mensaje{
		{"system", "Eres un ingeniero DevOps. Respondes SOLO con un Dockerfile dentro de un bloque ```dockerfile."},
		{"user", fmt.Sprintf("Escribe un Dockerfile de producción para este proyecto. La app debe escuchar en el puerto %d "+
			"y responder en GET /salud.\n\n%s", puerto, leerProyecto(dir))},
	}
	t0 := time.Now()
	for n := 1; n <= *intentos; n++ {
		t := time.Now()
		resp, err := preguntar(*modelo, msgs)
		if err != nil {
			fmt.Println("error con Ollama:", err)
			os.Exit(2)
		}
		df := extraer(resp)
		os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(df), 0o644)
		ok, detalle := probar(dir, fmt.Sprintf("ia-demo:%d", n), *modo)
		fmt.Printf("== intento %d (%.1f s)\n", n, time.Since(t).Seconds())
		for _, l := range strings.Split(strings.TrimRight(df, "\n"), "\n") {
			fmt.Println("   |", l)
		}
		fmt.Println("   -> " + strings.ReplaceAll(detalle, "\n", "\n      "))
		if ok {
			tam, _ := sh(dir, "docker", "image", "inspect", "-f", "{{.Size}}", fmt.Sprintf("ia-demo:%d", n))
			var bytesImg float64
			fmt.Sscan(strings.TrimSpace(tam), &bytesImg)
			fmt.Printf("\nRESULTADO: funciona en el intento %d · %.1f s en total · imagen de %.0f MB · modelo %s\n",
				n, time.Since(t0).Seconds(), bytesImg/1e6, *modelo)
			return
		}
		msgs = append(msgs, mensaje{"assistant", resp},
			mensaje{"user", "Ese Dockerfile no funciona. Este es el error:\n" + detalle + "\nCorrígelo. Responde solo con el Dockerfile completo."})
	}
	fmt.Printf("\nRESULTADO: no funcionó en %d intentos · %.1f s · modelo %s\n", *intentos, time.Since(t0).Seconds(), *modelo)
	os.Exit(1)
}
