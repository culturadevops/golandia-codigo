package tarea

import "fmt"

func Clonar(m map[string]int) map[string]int {
    n := make(map[string]int)
    for k, v := range m {
        n[k] = v
    }
    return n
}
