package tarea

// ProcesarTodos llama a f con cada item en paralelo, una goroutine por item.
// Devuelve el primer error en cuanto ocurre, sin esperar a las demás
// goroutines, o nil cuando todas terminan sin error.
//
// El canal de resultados tiene capacidad len(items), de modo que ninguna
// goroutine queda bloqueada al enviar su resultado aunque la función ya
// haya retornado; todas terminan en cuanto f retorna.
func ProcesarTodos(items []string, f func(string) error) error {
	if len(items) == 0 {
		return nil
	}

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
