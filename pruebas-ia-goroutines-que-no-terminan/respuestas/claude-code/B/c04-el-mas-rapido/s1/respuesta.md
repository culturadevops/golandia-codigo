```go
package tarea

// ElMasRapido lanza todas las consultas en paralelo y devuelve el resultado
// de la primera que termine. El canal tiene capacidad para todos los
// resultados, así que las consultas más lentas pueden enviar el suyo y
// terminar sin quedarse bloqueadas aunque nadie lo lea.
// Si no hay consultas, devuelve la cadena vacía.
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