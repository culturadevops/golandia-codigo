package cache

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

// Go 1.27: synctest.Sleep duerme y luego espera, en una sola llamada.
func TestCaducaSleep127(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := Nueva(30*time.Second, 10*time.Second)
		defer c.Cerrar()
		c.Poner("sesion", "ana")
		synctest.Sleep(41 * time.Second)
		if n := c.Tam(); n != 0 {
			t.Fatalf("quedan %d entradas", n)
		}
		t.Logf("con synctest.Sleep: quedan %d entradas", c.Tam())
	})
}

// Un timeout de contexto de cinco segundos, sin esperar cinco segundos.
func TestContextoTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		inicio := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		<-ctx.Done()
		t.Logf("error: %v · pasaron %v de reloj falso", ctx.Err(), time.Since(inicio))
	})
}

// Un ticker de un segundo durante un minuto: sesenta ticks.
func TestTickerUnMinuto(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tk := time.NewTicker(time.Second)
		defer tk.Stop()
		fin := time.After(time.Minute)
		ticks := 0
		for {
			select {
			case <-tk.C:
				ticks++
			case <-fin:
				t.Logf("ticks en un minuto: %d", ticks)
				return
			}
		}
	})
}

// La solución al caso de la red: una conexión en memoria (net.Pipe) sí deja avanzar el reloj.
func TestConPipe(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cli, srv := net.Pipe()
		defer cli.Close()
		defer srv.Close()
		go func() {
			time.Sleep(3 * time.Second) // el «servidor» tarda en contestar
			srv.Write([]byte("hola"))
		}()
		inicio := time.Now()
		buf := make([]byte, 4)
		if _, err := io.ReadFull(cli, buf); err != nil {
			t.Fatal(err)
		}
		t.Logf("recibido %q tras %v de reloj falso", buf, time.Since(inicio))
	})
}

// Un mutex tampoco es bloqueo duradero: el reloj no avanza y el test se cuelga.
func TestBloqueadoEnMutex(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var mu sync.Mutex
		mu.Lock()
		go func() { mu.Lock() }() // espera un Unlock que nunca llega
		time.Sleep(time.Second)
	})
}

// Si la función termina y deja una goroutine bloqueada para siempre.
func TestFugaAlTerminar(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ch := make(chan int)
		go func() { <-ch }() // nadie le va a enviar nada
	})
}

// La hora de fuera y la de dentro.
func TestDosRelojes(t *testing.T) {
	t.Logf("fuera de la burbuja: año %d", time.Now().Year())
	synctest.Test(t, func(t *testing.T) {
		t.Logf("dentro de la burbuja: %v", time.Now().UTC())
	})
}

// Un canal creado FUERA de la burbuja: esperar en él no es un bloqueo duradero.
func TestCanalDeFuera(t *testing.T) {
	ch := make(chan int)
	synctest.Test(t, func(t *testing.T) {
		go func() { <-ch }() // el canal no es de la burbuja
		time.Sleep(time.Second)
	})
}
