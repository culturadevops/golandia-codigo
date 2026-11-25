```go
package tarea

// ElMasRapido lanza todas las consultas en paralelo y devuelve el resultado
// de la primera que termine. El canal tiene capacidad para todas las
// respuestas, así las goroutines más lentas no se quedan bloqueadas.
func ElMasRapido(consultas []func() string) string {
	if len(consultas) == 0 {
		return ""
	}

	resultados := make(chan string, len(consultas))
	for _, consulta := range consultas {
		go func(c func() string) {
			resultados <- c()
		}(consulta)
	}

	return <-resultados
}
```