package tarea

// ProcesarTodos llama a f con cada item en paralelo, una goroutine por item.
// Devuelve el primer error en cuanto ocurre, sin esperar a las demás
// goroutines, o nil cuando todas terminan sin error.
func ProcesarTodos(items []string, f func(string) error) error {
	// El canal tiene un hueco por item, así ninguna goroutine se queda
	// bloqueada (ni se filtra) si ya devolvimos por un error anterior.
	errs := make(chan error, len(items))

	for _, item := range items {
		go func(it string) {
			errs <- f(it)
		}(item)
	}

	for range items {
		if err := <-errs; err != nil {
			return err
		}
	}
	return nil
}
