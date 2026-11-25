```go
package tarea

// Config representa una configuración donde los punteros nil significan "sin valor".
type Config struct {
	Timeout *int
	Nombre  *string
}

// PorDefecto devuelve una Config con Timeout en 30 y Nombre en "api".
func PorDefecto() Config {
	timeout := 30
	nombre := "api"
	return Config{
		Timeout: &timeout,
		Nombre:  &nombre,
	}
}
```