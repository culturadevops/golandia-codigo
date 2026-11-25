package tarea

// Ayuda común de las pruebas del vídeo 705 (escrita el 2026-10-29, antes de generar respuestas).
// fugas() usa el perfil goroutineleak de Go 1.27: en 1.27 ya no hace falta GOEXPERIMENT
// (go.dev/doc/go1.27: «The goroutineleakprofile GOEXPERIMENT setting is now deleted»).

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"testing"
	"time"
)

var reTotal = regexp.MustCompile(`goroutineleak profile: total (\d+)`)

// fugas espera a que acaben las funciones lentas del escenario y cuenta las goroutines
// que el runtime da por fugadas: bloqueadas en algo que ya nadie puede desbloquear.
func fugas(t *testing.T) int {
	time.Sleep(400 * time.Millisecond)
	runtime.GC()
	var b bytes.Buffer
	if err := pprof.Lookup("goroutineleak").WriteTo(&b, 1); err != nil {
		t.Fatal(err)
	}
	n := 0
	if m := reTotal.FindStringSubmatch(b.String()); m != nil {
		n, _ = strconv.Atoi(m[1])
	}
	if n > 0 {
		fmt.Printf("PILAS\n%s\nFIN-PILAS\n", b.String())
	}
	return n
}

// termina ejecuta f con un límite de 3 s y dice si volvió.
func termina(f func()) bool {
	hecho := make(chan struct{})
	go func() { f(); close(hecho) }()
	select {
	case <-hecho:
		return true
	case <-time.After(3 * time.Second):
		return false
	}
}

func informe(t *testing.T, ok, correcto bool) {
	n := fugas(t)
	fmt.Printf("RESULTADO termina=%v correcto=%v fugas=%d\n", ok, correcto, n)
}

// fuente envía 1, 2, 3… hasta que ctx se cancela (respeta ctx: nunca fuga ella).
func fuente(ctx interface{ Done() <-chan struct{} }) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := 1; ; i++ {
			select {
			case c <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return c
}

func finito(xs ...int) <-chan int {
	c := make(chan int, len(xs))
	for _, x := range xs {
		c <- x
	}
	close(c)
	return c
}
