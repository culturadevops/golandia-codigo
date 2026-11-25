package tarea

func LineaTabla(celdas []string) string {
    resultado := "|"
    for i, celda := range celdas {
        resultado += " " + celda + " |"
        if i < len(celdas)-1 {
            resultado += " "
        }
    }
    return resultado
}
