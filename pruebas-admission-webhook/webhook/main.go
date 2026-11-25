// admision: un admission webhook de Kubernetes en Go, solo con la biblioteca estándar.
//
//	/validar  rechaza pods con imagen :latest (o sin etiqueta) o sin limits de cpu y memory
//	/mutar    añade limits por defecto a los contenedores que no los traen
//	/healthz  para las sondas
//
// Vídeo 678 de Golandia. Sin client-go ni k8s.io/api: los tipos que hacen falta están abajo.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ── Los tipos del AdmissionReview (admission.k8s.io/v1), solo los campos que se usan ──

type AdmissionReview struct {
	APIVersion string             `json:"apiVersion"`
	Kind       string             `json:"kind"`
	Request    *AdmissionRequest  `json:"request,omitempty"`
	Response   *AdmissionResponse `json:"response,omitempty"`
}

type AdmissionRequest struct {
	UID       string          `json:"uid"`
	Namespace string          `json:"namespace"`
	Operation string          `json:"operation"`
	DryRun    bool            `json:"dryRun"`
	Object    json.RawMessage `json:"object"`
}

type AdmissionResponse struct {
	UID       string  `json:"uid"`
	Allowed   bool    `json:"allowed"`
	Status    *Status `json:"status,omitempty"`
	PatchType string  `json:"patchType,omitempty"`
	Patch     []byte  `json:"patch,omitempty"` // encoding/json lo manda en base64, como pide la API
}

type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Pod struct {
	Spec struct {
		Containers     []Container `json:"containers"`
		InitContainers []Container `json:"initContainers"`
	} `json:"spec"`
}

type Container struct {
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	Resources Resources `json:"resources"`
}

type Resources struct {
	Limits   map[string]string `json:"limits,omitempty"`
	Requests map[string]string `json:"requests,omitempty"`
}

// ── Las reglas ──

// etiqueta devuelve el tag de una imagen; "" si no trae (Kubernetes entiende :latest).
func etiqueta(imagen string) string {
	imagen, _, _ = strings.Cut(imagen, "@") // con digest manda el digest
	ultima := imagen[strings.LastIndex(imagen, "/")+1:]
	_, tag, _ := strings.Cut(ultima, ":") // el ":" del puerto del registro queda antes de la última "/"
	return tag
}

func problemas(c Container) []string {
	var p []string
	if t := etiqueta(c.Image); t == "" || t == "latest" {
		p = append(p, fmt.Sprintf("contenedor %q: la imagen %q usa :latest; fija una versión", c.Name, c.Image))
	}
	for _, r := range []string{"cpu", "memory"} {
		if _, ok := c.Resources.Limits[r]; !ok {
			p = append(p, fmt.Sprintf("contenedor %q: falta limits.%s", c.Name, r))
		}
	}
	return p
}

func validar(req *AdmissionRequest) *AdmissionResponse {
	var pod Pod
	if err := json.Unmarshal(req.Object, &pod); err != nil {
		return &AdmissionResponse{UID: req.UID, Status: &Status{Code: 400, Message: err.Error()}}
	}
	var todos []string
	for _, c := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
		todos = append(todos, problemas(c)...)
	}
	if len(todos) == 0 {
		return &AdmissionResponse{UID: req.UID, Allowed: true}
	}
	return &AdmissionResponse{UID: req.UID, Allowed: false,
		Status: &Status{Code: 403, Message: "admision: " + strings.Join(todos, "; ")}}
}

// mutar añade limits por defecto con un JSON Patch; no toca los contenedores que ya los traen.
func mutar(req *AdmissionRequest) *AdmissionResponse {
	var pod Pod
	if err := json.Unmarshal(req.Object, &pod); err != nil {
		return &AdmissionResponse{UID: req.UID, Status: &Status{Code: 400, Message: err.Error()}}
	}
	var ops []map[string]any
	for i, c := range pod.Spec.Containers {
		if len(c.Resources.Limits) > 0 {
			continue
		}
		r := c.Resources
		r.Limits = map[string]string{"cpu": *cpuDef, "memory": *memDef}
		ops = append(ops, map[string]any{"op": "add", "path": fmt.Sprintf("/spec/containers/%d/resources", i), "value": r})
	}
	resp := &AdmissionResponse{UID: req.UID, Allowed: true}
	if len(ops) > 0 {
		resp.PatchType = "JSONPatch"
		resp.Patch, _ = json.Marshal(ops)
	}
	return resp
}

// ── HTTP ──

var (
	volcar = flag.Bool("volcar", false, "escribe en el log el AdmissionReview de entrada y el de salida")
	cpuDef = flag.String("cpu", "500m", "limits.cpu por defecto de /mutar")
	memDef = flag.String("mem", "256Mi", "limits.memory por defecto de /mutar")
)

func manejador(nombre string, decide func(*AdmissionRequest) *AdmissionResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		cuerpo, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		var rev AdmissionReview
		if err == nil {
			err = json.Unmarshal(cuerpo, &rev)
		}
		if err != nil || rev.Request == nil {
			http.Error(w, "se esperaba un AdmissionReview", http.StatusBadRequest)
			return
		}
		if d := os.Getenv("DORMIR"); d != "" { // para la prueba del webhook colgado
			pausa, _ := time.ParseDuration(d)
			time.Sleep(pausa)
		}
		resp := decide(rev.Request)
		salida, _ := json.Marshal(AdmissionReview{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview", Response: resp})
		w.Header().Set("Content-Type", "application/json")
		w.Write(salida)
		log.Printf("%s ns=%s op=%s allowed=%t en %s", nombre, rev.Request.Namespace, rev.Request.Operation, resp.Allowed, time.Since(t))
		if *volcar {
			log.Printf("ENTRADA %s", cuerpo)
			log.Printf("SALIDA %s", salida)
		}
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/validar", manejador("validar", validar))
	http.HandleFunc("/mutar", manejador("mutar", mutar))
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, "ok") })
	srv := &http.Server{Addr: ":8443", ReadHeaderTimeout: 5 * time.Second}
	log.Printf("admision escuchando en :8443 (TLS)")
	log.Fatal(srv.ListenAndServeTLS("/certs/tls.crt", "/certs/tls.key"))
}
