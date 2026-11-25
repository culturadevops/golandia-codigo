```go
package tarea

import "math"

func Acotar(x, lo, hi int) int {
    return int(math.Max(float64(lo), math.Min(float64(x), float64(hi))))
}
```