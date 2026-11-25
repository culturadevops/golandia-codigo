// Un vigilante que consulta un servicio cada cierto tiempo. Es el patrón más
// común de fuga de goroutines en Go, y el que se ve en cualquier servicio real.
package fuga

import (
	"context"
	"sync"
	"time"
)

// Vigilante con FUGA: `Arrancar` lanza una goroutine que nunca para, porque
// nadie le dio forma de parar. Ni canal de cierre, ni contexto.
type VigilanteConFuga struct {
	mu     sync.Mutex
	vistas int
}

func (v *VigilanteConFuga) Arrancar(cada time.Duration) {
	go func() {
		t := time.NewTicker(cada)
		for range t.C { // ⚠️ este bucle no termina JAMÁS
			v.mu.Lock()
			v.vistas++
			v.mu.Unlock()
		}
	}()
}

func (v *VigilanteConFuga) Vistas() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.vistas
}

// La misma idea, arreglada: recibe un contexto y para cuando se cancela.
type VigilanteSano struct {
	mu     sync.Mutex
	vistas int
}

func (v *VigilanteSano) Arrancar(ctx context.Context, cada time.Duration) {
	go func() {
		t := time.NewTicker(cada)
		defer t.Stop() // y el ticker también se libera
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				v.mu.Lock()
				v.vistas++
				v.mu.Unlock()
			}
		}
	}()
}

func (v *VigilanteSano) Vistas() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.vistas
}
