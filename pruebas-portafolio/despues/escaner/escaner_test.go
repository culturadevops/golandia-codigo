package escaner

import (
	"context"
	"errors"
	"net"
	"slices"
	"testing"
	"time"
)

// abrir levanta n puertos de verdad en 127.0.0.1 y devuelve sus números. Se cierran al terminar la prueba.
func abrir(t *testing.T, n int) []int {
	t.Helper()
	var puertos []int
	for range n {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = l.Close() })
		go func() {
			for {
				c, err := l.Accept()
				if err != nil {
					return
				}
				_ = c.Close()
			}
		}()
		puertos = append(puertos, l.Addr().(*net.TCPAddr).Port)
	}
	slices.Sort(puertos)
	return puertos
}

func TestEncuentraLosPuertosAbiertos(t *testing.T) {
	abiertos := abrir(t, 3)
	desde, hasta := abiertos[0]-5, abiertos[len(abiertos)-1]+5
	res, err := Escanear(context.Background(), "127.0.0.1", desde, hasta, 50, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range abiertos {
		if !slices.Contains(res.Abiertos, p) {
			t.Errorf("no encontró el puerto abierto %d (encontró %v)", p, res.Abiertos)
		}
	}
}

func TestValidaLosParametros(t *testing.T) {
	casos := []struct {
		nombre             string
		desde, hasta, conc int
	}{
		{"concurrencia cero", 1, 10, 0},
		{"rango al revés", 100, 10, 5},
		{"puerto imposible", 1, 70000, 5},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if _, err := Escanear(context.Background(), "127.0.0.1", c.desde, c.hasta, c.conc, time.Second); err == nil {
				t.Error("esperaba un error")
			}
		})
	}
}

func TestSeCancelaConElContexto(t *testing.T) {
	ctx, cancelar := context.WithCancel(context.Background())
	cancelar()
	_, err := Escanear(ctx, "127.0.0.1", 1, 65535, 10, time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("esperaba context.Canceled, llegó %v", err)
	}
}

func TestClasificar(t *testing.T) {
	casos := map[string]string{
		"dial tcp 127.0.0.1:9: connect: connection refused": "rechazado",
		"dial tcp 10.0.0.1:80: i/o timeout":                 "timeout",
		"socket: too many open files":                       "sin-descriptores",
		"algo raro":                                         "otro",
	}
	for msg, quiero := range casos {
		if got := Clasificar(errors.New(msg)); got != quiero {
			t.Errorf("Clasificar(%q) = %q, quería %q", msg, got, quiero)
		}
	}
}
