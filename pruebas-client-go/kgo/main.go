// kgo: la demo del vídeo «client-go: controla Kubernetes desde Go» (golandia 677).
//
//	kgo config | pods | desplegar | watch-ingenuo | watch-reintento | contar | conflicto | rbac
//
// Solo toca el namespace «demo» (lo crea correr.py) y lee el resto.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
)

const ns = "demo"

var t0 = time.Now()

// logf imprime con el tiempo desde que arrancó el programa.
func logf(format string, a ...any) {
	fmt.Printf("%6.1fs  %s\n", time.Since(t0).Seconds(), fmt.Sprintf(format, a...))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: kgo config|pods|desplegar|watch-ingenuo|watch-reintento|contar|conflicto|rbac")
		os.Exit(2)
	}
	cfg, origen, err := cargarConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}
	contar(cfg) // cuenta cada petición al API server
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx := context.Background()
	switch os.Args[1] {
	case "config":
		mostrarConfig(cfg, origen)
	case "pods":
		err = listarPods(ctx, cs)
	case "desplegar":
		err = desplegar(ctx, cs)
	case "watch-ingenuo":
		err = watchIngenuo(ctx, cs)
	case "watch-reintento":
		err = watchReintento(ctx, cs)
	case "contar":
		err = compararLecturas(ctx, cs)
	case "conflicto":
		err = conflicto(ctx, cs)
	case "rbac":
		rbac(ctx, cs)
	default:
		err = fmt.Errorf("subcomando desconocido %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
