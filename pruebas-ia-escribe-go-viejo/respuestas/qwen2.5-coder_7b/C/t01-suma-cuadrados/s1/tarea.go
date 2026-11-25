package tarea

import "sync"

func SumaCuadrados(n int) int {
    var sum atomic.Int64
    var wg sync.WaitGroup

    for i := 0; i < n; i++ {
        wg.Add(1)
        i := i
        go func() {
            defer wg.Done()
            sum.Add(int64(i * i))
        }()
    }

    wg.Wait()
    return int(sum.Load())
}
