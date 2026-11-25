package tarea

func MayorDeTres(a, b, c int) int {
    max := a
    if b > max {
        max = b
    }
    if c > max {
        max = c
    }
    return max
}
