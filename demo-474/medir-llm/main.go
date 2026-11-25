// medir-llm: mide un LLM local en Ollama: arranque en frío, tiempo hasta el primer token (TTFT) y tokens por segundo.
//
//	go run . [-rondas 5] [-cpu] <modelo> [<modelo> ...]
//
// Solo librería estándar. Por cada modelo: lo saca de memoria y ESPERA a que /api/ps lo confirme; una petición en frío y N en
// caliente, con streaming. TTFT = desde la petición hasta el primer trozo con texto; tokens/s = eval_count / eval_duration del
// último trozo. -cpu fuerza num_gpu=0 (todo en la CPU).
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"
)

const base = "http://127.0.0.1:11434"
const prompt = "Explica en tres frases qué es un contenedor de Docker."

type trozo struct {
	Response     string `json:"response"`
	Done         bool   `json:"done"`
	EvalCount    int    `json:"eval_count"`
	EvalDuration int64  `json:"eval_duration"`
	LoadDuration int64  `json:"load_duration"`
}

type medida struct{ ttft, tps, carga float64 }

func post(cuerpo any) (*http.Response, error) {
	b, _ := json.Marshal(cuerpo)
	return http.Post(base+"/api/generate", "application/json", bytes.NewReader(b))
}

func cargados() []string {
	r, err := http.Get(base + "/api/ps")
	if err != nil {
		return nil
	}
	defer r.Body.Close()
	var ps struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	json.NewDecoder(r.Body).Decode(&ps)
	var n []string
	for _, m := range ps.Models {
		n = append(n, m.Name)
	}
	return n
}

// descargar saca el modelo de memoria y espera a que /api/ps confirme que ya no está: si no, el «frío» sale caliente.
func descargar(modelo string) {
	if r, err := post(map[string]any{"model": modelo, "keep_alive": 0}); err == nil {
		r.Body.Close()
	}
	for i := 0; i < 120; i++ {
		if !slices.Contains(cargados(), modelo) {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println(modelo, "sigue cargado: no mido un frío que no es frío")
	os.Exit(1)
}

func una(modelo string, cpu bool) medida {
	opciones := map[string]any{"temperature": 0, "seed": 42, "num_predict": 128}
	if cpu {
		opciones["num_gpu"] = 0
	}
	t0 := time.Now()
	r, err := post(map[string]any{"model": modelo, "prompt": prompt, "stream": true, "options": opciones})
	if err != nil {
		fmt.Println("error con Ollama:", err)
		os.Exit(2)
	}
	defer r.Body.Close()
	var m medida
	sc := bufio.NewScanner(r.Body)
	for sc.Scan() {
		var t trozo
		json.Unmarshal(sc.Bytes(), &t)
		if m.ttft == 0 && t.Response != "" {
			m.ttft = time.Since(t0).Seconds()
		}
		if t.Done {
			m.tps = float64(t.EvalCount) / (float64(t.EvalDuration) / 1e9)
			m.carga = float64(t.LoadDuration) / 1e9
		}
	}
	return m
}

func mediana(v []float64) float64 {
	s := slices.Clone(v)
	slices.Sort(s)
	if len(s)%2 == 1 {
		return s[len(s)/2]
	}
	return (s[len(s)/2-1] + s[len(s)/2]) / 2
}

func main() {
	rondas := flag.Int("rondas", 5, "peticiones en caliente por modelo")
	cpu := flag.Bool("cpu", false, "solo CPU (num_gpu=0)")
	flag.Parse()
	modo := "GPU"
	if *cpu {
		modo = "SOLO CPU"
	}
	fmt.Printf("prompt: %q · máx. 128 tokens · temperatura 0 · %s\n", prompt, modo)
	fmt.Printf("%-22s %11s %10s %13s %9s\n", "modelo", "frío: carga", "frío: TTFT", "TTFT caliente", "tokens/s")
	for _, m := range cargados() {
		descargar(m)
	}
	for _, m := range flag.Args() {
		descargar(m)
		frio := una(m, *cpu)
		var ttft, tps []float64
		for i := 0; i < *rondas; i++ {
			x := una(m, *cpu)
			ttft, tps = append(ttft, x.ttft), append(tps, x.tps)
		}
		fmt.Printf("%-22s %9.2f s %8.2f s %10.0f ms %9.1f\n", m, frio.carga, frio.ttft, mediana(ttft)*1000, mediana(tps))
		descargar(m)
	}
}
