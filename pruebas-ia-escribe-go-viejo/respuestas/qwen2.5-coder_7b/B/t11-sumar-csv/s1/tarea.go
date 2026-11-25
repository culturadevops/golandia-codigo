package tarea

import (
	"errors"
	"strconv"
	"strings"
)

func SumarCSV(s string) (int, error) {
	numbers := strings.Split(s, ",")
	sum := 0
	for _, numStr := range numbers {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, errors.New("invalid number in CSV: " + numStr)
		}
		sum += num
	}
	return sum, nil
}
