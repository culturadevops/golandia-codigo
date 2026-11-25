package cache

import (
	"net"
	"testing"
	"testing/synctest"
	"time"
)

// El test honesto: espera de verdad a que pase el TTL.
func TestCaducaEsperando(t *testing.T) {
	c := Nueva(30*time.Second, 10*time.Second)
	defer c.Cerrar()
	c.Poner("sesion", "ana")
	time.Sleep(31 * time.Second)
	time.Sleep(10 * time.Second) // y una vuelta del limpiador
	if n := c.Tam(); n != 0 {
		t.Fatalf("quedan %d entradas", n)
	}
}

// El mismo test dentro de una burbuja: el reloj es falso.
func TestCaducaSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := Nueva(30*time.Second, 10*time.Second)
		defer c.Cerrar()
		t.Logf("hora dentro de la burbuja: %v", time.Now().UTC())
		c.Poner("sesion", "ana")
		time.Sleep(31 * time.Second)
		time.Sleep(10 * time.Second)
		synctest.Wait()
		if n := c.Tam(); n != 0 {
			t.Fatalf("quedan %d entradas", n)
		}
		t.Logf("hora al terminar: %v", time.Now().UTC())
	})
}

// Lo que NO es un bloqueo duradero: E/S de red. El reloj no avanza mientras alguien lee de un socket.
func TestBloqueadoEnRed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()
		go func() { ln.Accept() }() // espera una conexión que nunca llega
		time.Sleep(time.Second)      // nunca termina: la goroutine de arriba no está "durablemente" bloqueada
	})
}

// Un deadlock de verdad dentro de la burbuja: synctest lo detecta y falla.
func TestDeadlock(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ch := make(chan int)
		<-ch
	})
}
