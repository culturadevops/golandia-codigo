// Receptor de webhooks de GitHub que despliega solo: git push → pull, build y reinicio de la app.
//
//	receptor -repo DIR [-addr :9000] [-app-addr 127.0.0.1:8080] [-rama refs/heads/main] [-sin-cola]
//	receptor firmar SECRETO < cuerpo   (imprime la cabecera X-Hub-Signature-256, para la demo)
//
// El secreto va en la variable WEBHOOK_SECRET. Golandia, vídeo 563.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// GitHub corta los payloads a 25 MB: más que eso no puede ser suyo.
const maxCuerpo = 25 << 20

var inicio = time.Now()

func logf(f string, a ...any) {
	fmt.Printf("%6d ms  %s\n", time.Since(inicio).Milliseconds(), fmt.Sprintf(f, a...))
}

func firmar(secreto, cuerpo []byte) string {
	m := hmac.New(sha256.New, secreto)
	m.Write(cuerpo)
	return "sha256=" + hex.EncodeToString(m.Sum(nil))
}

type pedido struct {
	n     int
	after string
}

type receptor struct {
	secreto []byte
	rama    string
	repo    string
	appAddr string
	sinCola bool

	mu      sync.Mutex
	vistas  map[string]bool // X-GitHub-Delivery ya procesadas
	n       int
	cola    chan pedido
	app     *exec.Cmd
	ultimo  string // commit desplegado
	muDespl sync.Mutex
}

func (r *receptor) webhook(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "solo POST", http.StatusMethodNotAllowed)
		return
	}
	cuerpo, err := io.ReadAll(io.LimitReader(req.Body, maxCuerpo+1))
	if err != nil || len(cuerpo) > maxCuerpo {
		http.Error(w, "cuerpo demasiado grande", http.StatusRequestEntityTooLarge)
		return
	}
	// 1. La firma, antes que nada: sobre los bytes EXACTOS que llegaron, y en tiempo constante.
	firma := req.Header.Get("X-Hub-Signature-256")
	if !hmac.Equal([]byte(firma), []byte(firmar(r.secreto, cuerpo))) {
		logf("rechazado: firma inválida (%q)", recorta(firma))
		http.Error(w, "firma inválida", http.StatusUnauthorized)
		return
	}
	evento := req.Header.Get("X-GitHub-Event")
	if evento == "ping" {
		fmt.Fprintln(w, "pong")
		return
	}
	if evento != "push" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// 2. Repetidas: solo DESPUÉS de validar la firma, o cualquiera podría «gastar» ids ajenos.
	entrega := req.Header.Get("X-GitHub-Delivery")
	r.mu.Lock()
	if entrega == "" || r.vistas[entrega] {
		r.mu.Unlock()
		logf("rechazado: entrega repetida %s", entrega)
		http.Error(w, "entrega repetida", http.StatusConflict)
		return
	}
	r.vistas[entrega] = true
	r.mu.Unlock()

	var p struct {
		Ref   string `json:"ref"`
		After string `json:"after"`
	}
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	if p.Ref != r.rama {
		logf("ignorado: push a %s", p.Ref)
		fmt.Fprintf(w, "ignorado: %s no es %s\n", p.Ref, r.rama)
		return
	}
	r.mu.Lock()
	r.n++
	ped := pedido{r.n, p.After}
	r.mu.Unlock()
	// 3. Contestar YA y desplegar después: GitHub espera la respuesta en 10 segundos.
	if r.sinCola {
		go r.desplegar(ped)
	} else {
		r.cola <- ped
	}
	logf("aceptado: despliegue #%d (%s)", ped.n, corto(ped.after))
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "despliegue #%d en cola\n", ped.n)
}

func (r *receptor) trabajador() {
	for p := range r.cola {
		r.desplegar(p)
	}
}

func (r *receptor) desplegar(p pedido) {
	t0 := time.Now()
	logf("despliegue #%d: empieza", p.n)
	if out, err := ejecutar(r.repo, "git", "pull", "--ff-only", "-q"); err != nil {
		logf("despliegue #%d FALLÓ en git pull: %s", p.n, ultima(out))
		return
	}
	head, _ := ejecutar(r.repo, "git", "rev-parse", "HEAD")
	head = strings.TrimSpace(head)
	r.muDespl.Lock()
	igual := head == r.ultimo
	r.muDespl.Unlock()
	if igual {
		logf("despliegue #%d: nada que hacer, %s ya está desplegado", p.n, corto(head))
		return
	}
	if out, err := ejecutar(r.repo, "go", "build", "-buildvcs=false", "-o", "app.nuevo", "."); err != nil {
		logf("despliegue #%d FALLÓ en go build: %s", p.n, ultima(out))
		return
	}
	if err := os.Rename(filepath.Join(r.repo, "app.nuevo"), filepath.Join(r.repo, "app")); err != nil {
		logf("despliegue #%d FALLÓ al reemplazar el binario: %v", p.n, err)
		return
	}
	if err := r.reiniciar(); err != nil {
		logf("despliegue #%d FALLÓ al reiniciar: %v", p.n, err)
		return
	}
	r.muDespl.Lock()
	r.ultimo = head
	r.muDespl.Unlock()
	logf("despliegue #%d: listo en %d ms · %s · la app dice: %s", p.n, time.Since(t0).Milliseconds(), corto(head), version(r.appAddr))
}

func (r *receptor) reiniciar() error {
	r.muDespl.Lock()
	defer r.muDespl.Unlock()
	var parada time.Time
	if r.app != nil {
		r.app.Process.Signal(syscall.SIGTERM)
		r.app.Wait()
		parada = time.Now()
	}
	defer func() {
		if !parada.IsZero() {
			logf("    la app no contestó durante %.1f ms (de parar la vieja a que contesta la nueva)",
				float64(time.Since(parada).Microseconds())/1000)
		}
	}()
	c := exec.Command(filepath.Join(r.repo, "app"))
	c.Env = append(os.Environ(), "ADDR="+r.appAddr)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	if err := c.Start(); err != nil {
		return err
	}
	r.app = c
	for i := 0; i < 2000; i++ {
		if version(r.appAddr) != "" {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
	return fmt.Errorf("la app no contesta en %s", r.appAddr)
}

func version(addr string) string {
	cl := http.Client{Timeout: 200 * time.Millisecond}
	res, err := cl.Get("http://" + addr + "/")
	if err != nil {
		return ""
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return strings.TrimSpace(string(b))
}

func ejecutar(dir, prog string, args ...string) (string, error) {
	c := exec.Command(prog, args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return string(out), err
}

func ultima(s string) string {
	l := strings.Split(strings.TrimSpace(s), "\n")
	return l[len(l)-1]
}

func corto(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

func recorta(s string) string {
	if len(s) > 20 {
		return s[:20] + "…"
	}
	return s
}

func main() {
	if len(os.Args) == 3 && os.Args[1] == "firmar" {
		cuerpo, _ := io.ReadAll(os.Stdin)
		fmt.Print(firmar([]byte(os.Args[2]), cuerpo))
		return
	}
	addr := flag.String("addr", ":9000", "dónde escucha el webhook")
	appAddr := flag.String("app-addr", "127.0.0.1:8080", "dónde escucha la app desplegada")
	repo := flag.String("repo", "", "clon del repositorio a desplegar")
	rama := flag.String("rama", "refs/heads/main", "rama que despliega")
	sinCola := flag.Bool("sin-cola", false, "desplegar cada push en su propia goroutine (MAL, para la demo)")
	flag.Parse()
	secreto := os.Getenv("WEBHOOK_SECRET")
	if secreto == "" || *repo == "" {
		log.Fatal("faltan WEBHOOK_SECRET o -repo")
	}
	r := &receptor{secreto: []byte(secreto), rama: *rama, repo: *repo, appAddr: *appAddr, sinCola: *sinCola,
		vistas: map[string]bool{}, cola: make(chan pedido, 16)}
	go r.trabajador()
	// Primer despliegue al arrancar, por la misma cola.
	r.cola <- pedido{0, "arranque"}
	http.HandleFunc("/webhook", r.webhook)
	logf("receptor escuchando en %s (sin-cola=%v)", *addr, *sinCola)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
