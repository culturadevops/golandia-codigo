```go
package tarea

type Config struct {
    Timeout *int
    Nombre  *string
}

func PorDefecto() Config {
    timeout := 30
    nombre := "api"
    return Config{
        Timeout: &timeout,
        Nombre:  &nombre,
    }
}
```