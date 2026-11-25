package tarea

// El caso del vídeo arreglado: respuesta de qwen2.5-coder:7b (prompt A, seed 1) con dos cambios.
// 1) El búfer tiene sitio para TODAS las respuestas: ninguna goroutine se queda esperando para enviar.
// 2) Fuera el WaitGroup y la goroutine que cerraba el canal: nadie lee después del primero,
//    así que no hace falta cerrarlo (y esa goroutine también se quedaba colgada en wg.Wait).
func ElMasRapido(consultas []func() string) string {
	resultChan := make(chan string, len(consultas))

	for _, consulta := range consultas {
		go func() {
			resultChan <- consulta()
		}()
	}

	return <-resultChan
}
