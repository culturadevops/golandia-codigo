package tarea

import (
	"errors"
	"strconv"
	"strings"
)

func SumarCSV(s string) (int, error) {
	sum := 0
	numbers := strings.Split(s, ",")
	for _, numStr := range numbers {
		num, err := strconv.Atoi(strings.TrimSpace(numStr))
		if err != nil {
			return 0, err
		}
		sum += num
	}
	return sum, nil
}
