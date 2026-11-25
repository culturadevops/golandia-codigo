```go
package tarea

type Config struct {
	Timeout *int
	Nombre  *string
}

func PorDefecto() Config {
	timeout := 30
	return Config{
		Timeout: &timeout,
		Nombre:  ptr("api"),
	}
}

func ptr[T any](v T) *T {
	return &v
}
```