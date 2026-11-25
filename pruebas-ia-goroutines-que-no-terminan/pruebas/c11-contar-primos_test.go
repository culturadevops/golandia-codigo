package tarea

import "testing"

func escenarioC11() (bool, bool) {
	nums := make([]int, 10000)
	for i := range nums {
		nums[i] = i + 1
	}
	var n int
	ok := termina(func() { n = ContarPrimos(nums, 8) })
	return ok, ok && n == 1229
}

func TestC11(t *testing.T) {
	ok, correcto := escenarioC11()
	informe(t, ok, correcto)
}
