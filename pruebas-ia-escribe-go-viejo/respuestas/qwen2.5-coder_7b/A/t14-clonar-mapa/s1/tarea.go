package tarea

func Clonar(m map[string]int) map[string]int {
    clon := make(map[string]int)
    for clave, valor := range m {
        clon[clave] = valor
    }
    return clon
}
