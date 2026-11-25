package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrEntradaInvalida es un error centinela que se puede comprobar con errors.Is.
var ErrEntradaInvalida = errors.New("entrada inválida")

// Tarea representa una unidad de trabajo.
type Tarea struct {
	ID    int
	Valor int
}

// TareaError es el tipo de error propio que describe el fallo de una tarea.
type TareaError struct {
	TareaID int
	Motivo  string
	Err     error
}

func (e *TareaError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("tarea %d: %s: %v", e.TareaID, e.Motivo, e.Err)
	}
	return fmt.Sprintf("tarea %d: %s", e.TareaID, e.Motivo)
}

// Unwrap permite que errors.Is y errors.As recorran la cadena de errores.
func (e *TareaError) Unwrap() error {
	return e.Err
}

// Contador es un contador seguro para uso concurrente.
type Contador struct {
	mu sync.Mutex
	n  int
}

func (c *Contador) Incrementar() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *Contador) Valor() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// procesar simula trabajo y falla con valores no válidos.
func procesar(t Tarea) (int, error) {
	time.Sleep(time.Duration(t.ID%5) * 10 * time.Millisecond)

	if t.Valor < 0 {
		return 0, fmt.Errorf("procesar: %w", &TareaError{
			TareaID: t.ID,
			Motivo:  "valor negativo",
			Err:     ErrEntradaInvalida,
		})
	}
	if t.Valor == 0 {
		return 0, fmt.Errorf("procesar: %w", &TareaError{
			TareaID: t.ID,
			Motivo:  "división por cero",
		})
	}
	return 100 / t.Valor, nil
}

// procesarTodas ejecuta cada tarea en su propia goroutine y devuelve
// el número de tareas terminadas con éxito y los errores producidos.
func procesarTodas(tareas []Tarea) (int, []error) {
	var (
		wg       sync.WaitGroup
		contador Contador
		muErrs   sync.Mutex
		errs     []error
	)

	for _, t := range tareas {
		wg.Add(1)
		go func(t Tarea) {
			defer wg.Done()

			res, err := procesar(t)
			if err != nil {
				muErrs.Lock()
				errs = append(errs, fmt.Errorf("procesarTodas: %w", err))
				muErrs.Unlock()
				return
			}

			fmt.Printf("tarea %d completada: resultado=%d\n", t.ID, res)
			contador.Incrementar()
		}(t)
	}

	wg.Wait()
	return contador.Valor(), errs
}

func main() {
	tareas := []Tarea{
		{ID: 1, Valor: 10},
		{ID: 2, Valor: -3},
		{ID: 3, Valor: 5},
		{ID: 4, Valor: 0},
		{ID: 5, Valor: 20},
		{ID: 6, Valor: -1},
		{ID: 7, Valor: 4},
	}

	terminadas, errs := procesarTodas(tareas)

	fmt.Printf("\nTareas terminadas: %d de %d\n", terminadas, len(tareas))

	if len(errs) == 0 {
		return
	}

	fmt.Printf("Fallos: %d\n", len(errs))
	for _, err := range errs {
		fmt.Println(" -", err)

		var te *TareaError
		if errors.As(err, &te) {
			fmt.Printf("   tipo TareaError -> id=%d, motivo=%q\n", te.TareaID, te.Motivo)
		}
		if errors.Is(err, ErrEntradaInvalida) {
			fmt.Println("   causa raíz: ErrEntradaInvalida")
		}
	}

	if combinado := errors.Join(errs...); combinado != nil {
		fmt.Printf("\nError combinado:\n%v\n", combinado)
	}
}
