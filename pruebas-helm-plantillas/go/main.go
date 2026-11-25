// Renderiza mini/templates/deployment.yaml con text/template, como lo hace Helm.
//
//	go run . solo-go   ../mini   → text/template sin funciones extra
//	go run . sprig     ../mini   → + las funciones de Sprig
//	go run . helm      ../mini   → + toYaml, la que agrega Helm (sin env ni expandenv)
//	go run . cuenta              → cuántas funciones hay en cada capa
//	go run . guiones             → {{ frente a {{- en text/template puro
//	go run . vacios              → qué cuenta como vacío para default (Sprig)
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"sigs.k8s.io/yaml"
)

// toYAML es la de Helm (pkg/engine/funcs.go): yaml.Marshal y sin el salto final.
func toYAML(v any) string {
	b, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(b), "\n")
}

// funciones arma el FuncMap de cada capa.
func funciones(capa string) template.FuncMap {
	switch capa {
	case "solo-go":
		return template.FuncMap{}
	case "sprig":
		return sprig.TxtFuncMap()
	}
	f := sprig.TxtFuncMap()
	delete(f, "env") // Helm las quita: un chart no lee el entorno de quien lo renderiza
	delete(f, "expandenv")
	f["toYaml"] = toYAML
	return f
}

// datos es el punto (.) que recibe la plantilla: un mapa, igual que en Helm.
func datos(chart string) map[string]any {
	crudo, err := os.ReadFile(filepath.Join(chart, "values.yaml"))
	if err != nil {
		panic(err)
	}
	var values map[string]any
	if err := yaml.Unmarshal(crudo, &values); err != nil {
		panic(err)
	}
	return map[string]any{
		"Values":  values,
		"Release": map[string]any{"Name": "demo", "Namespace": "default"},
	}
}

func renderizar(capa, chart string) {
	ruta := filepath.Join(chart, "templates", "deployment.yaml")
	src, _ := os.ReadFile(ruta)
	t, err := template.New("deployment.yaml").Funcs(funciones(capa)).Parse(string(src))
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := t.Execute(os.Stdout, datos(chart)); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func cuenta() {
	go_ := []string{"and", "call", "html", "index", "slice", "js", "len", "not", "or", "print",
		"printf", "println", "urlquery", "eq", "ge", "gt", "le", "lt", "ne"}
	fmt.Println("text/template trae:", len(go_), "funciones")
	fmt.Println("Sprig v3.3.0 agrega:", len(sprig.TxtFuncMap()))
	_, env := sprig.TxtFuncMap()["env"]
	fmt.Println("¿Sprig tiene env?", env, "· ¿y toYaml?", tiene(sprig.TxtFuncMap(), "toYaml"),
		"· ¿required?", tiene(sprig.TxtFuncMap(), "required"), "· ¿include?", tiene(sprig.TxtFuncMap(), "include"),
		"· ¿tpl?", tiene(sprig.TxtFuncMap(), "tpl"))
	// Los nombres de «extra» en funcMap() de helm.sh/helm/v4@v4.3.1/pkg/engine/funcs.go
	helm := []string{"toToml", "mustToToml", "fromToml", "toYaml", "mustToYaml", "toYamlPretty", "fromYaml",
		"fromYamlArray", "toJson", "mustToJson", "fromJson", "fromJsonArray", "mustToDuration", "durationSeconds",
		"durationMilliseconds", "durationMicroseconds", "durationNanoseconds", "durationMinutes", "durationHours",
		"durationDays", "durationWeeks", "durationRoundTo", "durationTruncateTo", "include", "tpl", "required", "lookup"}
	f := funciones("helm")
	nuevas := 0
	for _, n := range helm {
		if !tiene(sprig.TxtFuncMap(), n) {
			nuevas++
		}
		f[n] = nil
	}
	fmt.Println("Helm v4.3.1 quita env y expandenv, declara", len(helm), "propias (", nuevas, "nuevas,",
		len(helm)-nuevas, "pisan a Sprig) → total", len(f)+len(go_))
}

func tiene(f template.FuncMap, n string) bool { _, ok := f[n]; return ok }

func guiones() {
	for _, src := range []string{
		"a:\n  {{ if true }}\n  b: 1\n  {{ end }}\n",
		"a:\n  {{- if true }}\n  b: 1\n  {{- end }}\n",
		"a:\n  {{- if true -}}\n  b: 1\n  {{- end }}\n",
	} {
		t := template.Must(template.New("g").Parse(src))
		fmt.Printf("--- plantilla:\n%s--- salida (· = espacio):\n", src)
		var sb strings.Builder
		t.Execute(&sb, nil)
		fmt.Println(strings.ReplaceAll(sb.String(), " ", "·"))
	}
}

func vacios() {
	t := template.Must(template.New("v").Funcs(sprig.TxtFuncMap()).Parse(
		`{{ range $k, $v := . }}{{ printf "%-8s" $k }} → {{ default "DEFAULT" $v }}{{ "\n" }}{{ end }}`))
	t.Execute(os.Stdout, map[string]any{"a_false": false, "b_cero": 0, "c_vacio": "", "d_lista": []any{},
		"e_nil": nil, "f_texto": "hola", "g_true": true, "h_uno": 1})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("uso: go run . solo-go|sprig|helm <chart> | cuenta | guiones | vacios")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "cuenta":
		cuenta()
	case "guiones":
		guiones()
	case "vacios":
		vacios()
	default:
		renderizar(os.Args[1], os.Args[2])
	}
}
