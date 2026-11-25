package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// TaskError es el tipo de error propio para fallos de una tarea.
type TaskError struct {
	ID     int
	Reason string
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("tarea %d: %s", e.ID, e.Reason)
}

// Task representa una unidad de trabajo.
type Task struct {
	ID       int
	Duration time.Duration
	Fail     bool
}

// Counter es un contador seguro para uso concurrente.
type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func process(t Task) error {
	time.Sleep(t.Duration)
	if t.Fail {
		return &TaskError{ID: t.ID, Reason: "fallo simulado"}
	}
	return nil
}

// runAll procesa las tareas en paralelo y devuelve los errores envueltos.
func runAll(tasks []Task) (int, []error) {
	var (
		wg      sync.WaitGroup
		counter Counter
		errMu   sync.Mutex
		errs    []error
	)

	for _, t := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			if err := process(t); err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("procesando tarea %d: %w", t.ID, err))
				errMu.Unlock()
				return
			}
			counter.Inc()
		}(t)
	}

	wg.Wait()
	return counter.Value(), errs
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	tasks := make([]Task, 10)
	for i := range tasks {
		tasks[i] = Task{
			ID:       i + 1,
			Duration: time.Duration(r.Intn(100)) * time.Millisecond,
			Fail:     r.Intn(4) == 0,
		}
	}

	done, errs := runAll(tasks)

	fmt.Printf("Tareas terminadas: %d de %d\n", done, len(tasks))
	if len(errs) == 0 {
		return
	}

	fmt.Printf("Fallos: %d\n", len(errs))
	for _, err := range errs {
		var te *TaskError
		if errors.As(err, &te) {
			fmt.Printf("  - %v (ID=%d, motivo=%q)\n", err, te.ID, te.Reason)
		} else {
			fmt.Printf("  - %v\n", err)
		}
	}

	if joined := errors.Join(errs...); joined != nil {
		fmt.Println("\nError combinado:")
		fmt.Println(joined)
	}
}
