```go
package tarea

func LineaTabla(celdas []string) string {
    resultado := "|"
    for _, celda := range celdas {
        resultado += " " + celda + " |"
    }
    return resultado
}
```