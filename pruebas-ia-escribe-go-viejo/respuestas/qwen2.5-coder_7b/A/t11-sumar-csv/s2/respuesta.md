```go
package tarea

import (
	"strconv"
	"strings"
)

func SumarCSV(s string) (int, error) {
	numbers := strings.Split(s, ",")
	sum := 0
	for _, num := range numbers {
		val, err := strconv.Atoi(num)
		if err != nil {
			return 0, err
		}
		sum += val
	}
	return sum, nil
}
```