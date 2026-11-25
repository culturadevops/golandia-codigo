// Lección 11: apagado ordenado, medido. 20 peticiones de 2 s en vuelo y un SIGTERM.
//
//	go run ./lecciones/11_apagado            (en Linux o en Docker: usa SIGTERM)
//
// El programa se lanza a sí mismo como servidor en tres versiones y hace de cliente.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"tareas/internal/servidor"
)

const enVuelo = 20

func lento(w http.ResponseWriter, _ *http.Request) {
	time.Sleep(2 * time.Second)
	_, _ = io.WriteString(w, "ok")
}

func main() {
	modo := flag.String("servidor", "", "sin-senal, ingenuo o plantilla")
	flag.Parse()
	if *modo != "" {
		servir(*modo)
		return
	}
	for _, m := range []string{"sin-senal", "ingenuo", "plantilla"} {
		medir(m)
	}
}

func servir(modo string) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	fmt.Println(ln.Addr()) // el cliente lee la dirección de aquí
	h := http.HandlerFunc(lento)
	switch modo {
	case "sin-senal": // ni NotifyContext ni Shutdown: SIGTERM mata el proceso
		_ = http.Serve(ln, h)

	case "ingenuo": // Shutdown en una goroutine; main acaba cuando Serve vuelve
		srv := &http.Server{Handler: h}
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGTERM)
			<-c
			_ = srv.Shutdown(context.Background())
		}()
		if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		// Serve vuelve EN CUANTO empieza Shutdown, no cuando acaba: main termina aquí.

	case "plantilla":
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
		defer stop()
		log := slog.New(slog.NewTextHandler(io.Discard, nil))
		_ = servidor.Servir(ctx, servidor.Nuevo(h, log), ln, &servidor.Salud{}, 0, 8*time.Second, log)
	}
}

func medir(modo string) {
	yo, _ := os.Executable()
	cmd := exec.Command(yo, "-servidor", modo)
	out, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	dir, _ := bufio.NewReader(out).ReadString('\n')
	url := "http://" + strings.TrimSpace(dir) + "/"

	var mu sync.Mutex
	ok, fallos := 0, map[string]int{}
	var wg sync.WaitGroup
	for range enVuelo {
		wg.Go(func() {
			resp, err := http.Get(url)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fallos[motivo(err)]++
				return
			}
			resp.Body.Close()
			ok++
		})
	}
	time.Sleep(300 * time.Millisecond) // las 20 ya están dentro del handler
	t0 := time.Now()
	_ = cmd.Process.Signal(syscall.SIGTERM)
	_ = cmd.Wait()
	salio := time.Since(t0)
	wg.Wait()
	fmt.Printf("  %-10s -> %2d de %d OK  · el proceso sale a los %4d ms", modo, ok, enVuelo,
		salio.Round(10*time.Millisecond).Milliseconds())
	for m, n := range fallos {
		fmt.Printf("  · %d con «%s»", n, m)
	}
	fmt.Println()
}

func motivo(err error) string {
	s := err.Error()
	for _, m := range []string{"EOF", "connection reset by peer", "connection refused"} {
		if strings.Contains(s, m) {
			return m
		}
	}
	return s
}
