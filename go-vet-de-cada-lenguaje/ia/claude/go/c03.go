package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// TaskError es un error propio que describe el fallo de una tarea.
type TaskError struct {
	TaskID int
	Reason string
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("tarea %d falló: %s", e.TaskID, e.Reason)
}

// Task representa una unidad de trabajo.
type Task struct {
	ID   int
	Name string
}

// Counter es un contador seguro para uso concurrente.
type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// process simula el trabajo de una tarea; algunas fallan.
func process(t Task) error {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	if t.ID%3 == 0 {
		return &TaskError{TaskID: t.ID, Reason: "valor no permitido"}
	}
	return nil
}

// runAll procesa las tareas en paralelo y devuelve los fallos envueltos.
func runAll(tasks []Task) (int, []error) {
	var (
		wg      sync.WaitGroup
		done    Counter
		errMu   sync.Mutex
		failure []error
	)

	for _, t := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			if err := process(t); err != nil {
				wrapped := fmt.Errorf("procesando %q: %w", t.Name, err)
				errMu.Lock()
				failure = append(failure, wrapped)
				errMu.Unlock()
				return
			}
			done.Inc()
		}(t)
	}

	wg.Wait()
	return done.Value(), failure
}

func main() {
	tasks := make([]Task, 10)
	for i := range tasks {
		tasks[i] = Task{ID: i + 1, Name: fmt.Sprintf("tarea-%d", i+1)}
	}

	completed, errs := runAll(tasks)

	fmt.Printf("Tareas terminadas: %d de %d\n", completed, len(tasks))
	fmt.Printf("Fallos: %d\n", len(errs))

	for _, err := range errs {
		fmt.Println(" -", err)

		var te *TaskError
		if errors.As(err, &te) {
			fmt.Printf("   (TaskError detectado: id=%d, motivo=%s)\n", te.TaskID, te.Reason)
		}
	}
}
