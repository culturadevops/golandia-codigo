package tarea

import (
	"context"
	"sync/atomic"
)

// ProcesarConWorkers aplica f a cada elemento de trabajos usando un pool de
// workers goroutines y devuelve los resultados en el mismo orden que trabajos.
// Si ctx se cancela antes de terminar, devuelve ctx.Err() sin esperar a que se
// procesen los trabajos pendientes. Los workers dejan de tomar trabajos nuevos
// en cuanto detectan la cancelación y nunca se bloquean al publicar resultados.
func ProcesarConWorkers(ctx context.Context, trabajos []int, workers int, f func(int) int) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	n := len(trabajos)
	if n == 0 {
		return []int{}, nil
	}
	if workers < 1 {
		workers = 1
	}
	if workers > n {
		workers = n
	}

	resultados := make([]int, n)
	done := make(chan struct{})

	var siguiente atomic.Int64
	var activos atomic.Int64
	activos.Store(int64(workers))

	for w := 0; w < workers; w++ {
		go func() {
			defer func() {
				// El último worker en salir avisa que todo terminó.
				if activos.Add(-1) == 0 {
					close(done)
				}
			}()
			for {
				if ctx.Err() != nil {
					return
				}
				i := int(siguiente.Add(1) - 1)
				if i >= n {
					return
				}
				// Cada índice lo escribe un único worker, sin conflictos.
				resultados[i] = f(trabajos[i])
			}
		}()
	}

	select {
	case <-done:
		// Si el contexto se canceló, algún trabajo pudo quedar sin procesar.
		if siguiente.Load() < int64(n) || ctx.Err() != nil && siguiente.Load() <= int64(n) && !todosProcesados(&siguiente, n, workers) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		return resultados, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// todosProcesados indica si todos los índices fueron tomados por algún worker.
// Cada worker que termina normalmente incrementa el contador una vez más allá
// de n, por lo que si todos salieron por agotamiento el contador vale n+workers.
func todosProcesados(siguiente *atomic.Int64, n, workers int) bool {
	return siguiente.Load() >= int64(n+workers)
}
