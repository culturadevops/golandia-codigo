```go
package tarea

import (
	"errors"
	"strconv"
	"strings"
	"sync"
)

func SumarCSV(s string) (int, error) {
	numbers := strings.Split(s, ",")
	var sum int
	var wg sync.WaitGroup

	for _, numStr := range numbers {
		wg.Add(1)
		go func(numStr string) {
			defer wg.Done()
			num, err := strconv.Atoi(numStr)
			if err != nil {
				return
			}
			sum += num
		}(numStr)
	}

	wg.Wait()

	return sum, nil
}
```