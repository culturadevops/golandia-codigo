package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// cargarConfig: dentro de un pod usa el ServiceAccount; fuera, el kubeconfig.
func cargarConfig() (*rest.Config, string, error) {
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, "in-cluster", nil
	}
	if !errors.Is(err, rest.ErrNotInCluster) {
		return nil, "", err
	}
	ruta := os.Getenv("KUBECONFIG")
	if ruta == "" {
		ruta = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	cfg, err = clientcmd.BuildConfigFromFlags("", ruta)
	return cfg, "kubeconfig " + ruta, err
}

func mostrarConfig(cfg *rest.Config, origen string) {
	_, err := rest.InClusterConfig()
	logf("InClusterConfig(): %v", err)
	logf("origen: %s", origen)
	logf("host: %s", cfg.Host)
	logf("QPS=%v Burst=%v (0 = valores por defecto de client-go)", cfg.QPS, cfg.Burst)
}

// ── contador de peticiones ──────────────────────────────────────────

var (
	peticiones atomic.Int64
	bytesLeidos atomic.Int64
	muVerbos    sync.Mutex
	porVerbo    = map[string]int{}
)

type contado struct{ rt http.RoundTripper }

func (c contado) RoundTrip(req *http.Request) (*http.Response, error) {
	peticiones.Add(1)
	v := req.Method
	if req.URL.Query().Get("watch") == "true" {
		v = "WATCH"
	}
	if v == "WATCH" && os.Getenv("KGO_URLS") != "" {
		logf("→ %s %s", v, req.URL.RequestURI())
	}
	muVerbos.Lock()
	porVerbo[v]++
	muVerbos.Unlock()
	resp, err := c.rt.RoundTrip(req)
	if resp != nil {
		resp.Body = &cuerpo{resp.Body}
	}
	return resp, err
}

type cuerpo struct{ r interface{ Read([]byte) (int, error); Close() error } }

func (b *cuerpo) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	bytesLeidos.Add(int64(n))
	return n, err
}
func (b *cuerpo) Close() error { return b.r.Close() }

func contar(cfg *rest.Config) {
	cfg.Wrap(func(rt http.RoundTripper) http.RoundTripper { return contado{rt} })
}

func resumenPeticiones() string {
	muVerbos.Lock()
	defer muVerbos.Unlock()
	ks := make([]string, 0, len(porVerbo))
	for k := range porVerbo {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	s := ""
	for _, k := range ks {
		s += " " + k + "=" + strconv.Itoa(porVerbo[k])
	}
	return s
}

func reiniciarContador() {
	peticiones.Store(0)
	bytesLeidos.Store(0)
	muVerbos.Lock()
	porVerbo = map[string]int{}
	muVerbos.Unlock()
}
