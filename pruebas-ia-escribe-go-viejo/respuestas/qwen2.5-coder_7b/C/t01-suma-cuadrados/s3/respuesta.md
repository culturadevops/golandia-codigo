```go
package tarea

import "sync"

func SumaCuadrados(n int) int {
    var sum int
    var wg sync.WaitGroup
    for i := range n {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            sum += i * i
        }(i)
    }
    wg.Wait()
    return sum
}
```