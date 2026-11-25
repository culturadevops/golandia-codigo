```go
package tarea

func LineaTabla(celdas []string) string {
    var resultado string
    for i, celda := range celdas {
        if i > 0 {
            resultado += " | "
        }
        resultado += celda
    }
    return "|" + resultado + "|"
}
```