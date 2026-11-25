```go
package tarea

import "math"

func Acotar(x, lo, hi int) int {
    return int(math.Min(math.Max(float64(x), float64(lo)), float64(hi)))
}
```