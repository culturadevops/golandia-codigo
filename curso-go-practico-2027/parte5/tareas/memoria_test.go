package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// 50 goroutines crean 20 tareas cada una, a la vez, como 50 peticiones simultáneas.
func TestMemoriaConcurrente(t *testing.T) {
	m := NuevaMemoria()
	ctx := context.Background()
	var wg sync.WaitGroup
	for g := range 50 {
		wg.Go(func() {
			for i := range 20 {
				m.Crear(ctx, fmt.Sprintf("tarea %d-%d", g, i))
			}
		})
	}
	wg.Wait()
	lista, _ := m.Listar(ctx)
	if len(lista) != 1000 {
		t.Fatalf("hay %d tareas, quiero 1000", len(lista))
	}
}
