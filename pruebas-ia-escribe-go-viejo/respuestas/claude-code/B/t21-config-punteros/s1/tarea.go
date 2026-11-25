package tarea

// Config representa una configuración donde los punteros nil significan "sin valor".
type Config struct {
	Timeout *int
	Nombre  *string
}

// PorDefecto devuelve una Config con los valores por defecto.
func PorDefecto() Config {
	return Config{
		Timeout: new(30),
		Nombre:  new("api"),
	}
}
