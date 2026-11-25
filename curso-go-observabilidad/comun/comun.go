// Package comun tiene lo que repiten las lecciones: títulos, servir /metrics y preguntarle a
// un Prometheus DE VERDAD (las lecciones 3, 4 y 5 corren junto a uno, ver ejecutar_todo.py).
package comun

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Titulo(t string) { fmt.Printf("\n== %s ==\n", t) }

func Ms(d time.Duration) string {
	return fmt.Sprintf("%d ms", d.Round(time.Millisecond).Milliseconds())
}

// Prometheus es la dirección del Prometheus que raspa esta lección (lo pone ejecutar_todo.py).
var Prometheus = os.Getenv("PROMETHEUS")

// Servir publica el registro en :2112/metrics, que es donde lo raspa Prometheus cada segundo.
func Servir(reg *prometheus.Registry) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":2112", mux)
}

// Muestra son las líneas de /metrics que empiezan por `nombre`, tal cual las ve Prometheus.
func Muestra(reg *prometheus.Registry, nombre string) []string {
	familias, _ := reg.Gather()
	var out []string
	for _, f := range familias {
		if f.GetName() != nombre {
			continue
		}
		for _, m := range f.GetMetric() {
			etiquetas := []string{}
			for _, e := range m.GetLabel() {
				etiquetas = append(etiquetas, fmt.Sprintf("%s=%q", e.GetName(), e.GetValue()))
			}
			v := m.GetCounter().GetValue()
			out = append(out, fmt.Sprintf("%s{%s} %g", nombre, strings.Join(etiquetas, ","), v))
		}
	}
	if len(out) == 0 {
		out = append(out, fmt.Sprintf("(ninguna línea de %s)", nombre))
	}
	return out
}

type respuesta struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  [2]any            `json:"value"`
		} `json:"result"`
	} `json:"data"`
	Error string `json:"error"`
}

// Consulta lanza una PromQL contra el Prometheus real y la devuelve en una línea por serie.
func Consulta(q string) []string { return consulta(q, true) }

// consulta, con `redondear` a 4 cifras para leerla en pantalla; sin redondear para comparar.
func consulta(q string, redondear bool) []string {
	r, err := http.Get(Prometheus + "/api/v1/query?query=" + url.QueryEscape(q))
	if err != nil {
		return []string{"error: " + err.Error()}
	}
	defer r.Body.Close()
	var res respuesta
	json.NewDecoder(r.Body).Decode(&res)
	if res.Status != "success" {
		return []string{"error: " + res.Error}
	}
	if len(res.Data.Result) == 0 {
		return []string{"(vacío: ninguna serie)"}
	}
	var out []string
	for _, s := range res.Data.Result {
		claves := make([]string, 0, len(s.Metric))
		for k := range s.Metric {
			if k != "job" && k != "instance" {
				claves = append(claves, k)
			}
		}
		sort.Strings(claves)
		et := []string{}
		for _, k := range claves {
			et = append(et, fmt.Sprintf("%s=%q", k, s.Metric[k]))
		}
		valor := fmt.Sprint(s.Value[1])
		if f, err := strconv.ParseFloat(valor, 64); err == nil && redondear {
			valor = strconv.FormatFloat(f, 'g', 4, 64)
		}
		pre := ""
		if len(et) > 0 {
			pre = "{" + strings.Join(et, ",") + "} "
		}
		out = append(out, pre+valor)
	}
	sort.Strings(out)
	return out
}

// EsperarRaspado espera a que Prometheus raspe la aplicación DESPUÉS de este momento: así lo que
// se consulta es lo que la aplicación tiene ahora, no lo de hace un segundo.
func EsperarRaspado() {
	desde := float64(time.Now().UnixMilli()) / 1000
	for i := 0; i < 100; i++ {
		v := consulta(`timestamp(up{job="app"})`, false)
		if f, err := strconv.ParseFloat(v[0], 64); err == nil && f > desde+0.1 {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("  [aviso] Prometheus no raspó en 20 s")
}
