package tarea

func AlgunNegativo(nums []int) bool {
    for _, num := range nums {
        if num < 0 {
            return true
        }
    }
    return false
}
