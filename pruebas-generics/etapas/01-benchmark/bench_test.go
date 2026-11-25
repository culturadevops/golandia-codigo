package bench

import "testing"

const n = 10000

func datosInt() []int {
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i + 100000 // fuera del rango 0-255 que el runtime cachea sin asignar memoria
	}
	return nums
}

func datosInterface() []interface{} {
	nums := make([]interface{}, n)
	for i := range nums {
		nums[i] = i + 100000
	}
	return nums
}

func BenchmarkSumInt(b *testing.B) {
	nums := datosInt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SumInt(nums)
	}
}

func BenchmarkSumGeneric(b *testing.B) {
	nums := datosInt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SumGeneric(nums)
	}
}

func BenchmarkSumInterface(b *testing.B) {
	nums := datosInterface()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SumInterface(nums)
	}
}

// sink es una variable de paquete a proposito: si el valor se guardara en una
// variable local, el compilador prueba que no escapa y elimina la asignacion entera,
// dejando el benchmark en cero sin que nada lo avise.
var sink interface{}

func doblar(n int) int { return n * 2 }

func BenchmarkDoubleInts(b *testing.B) {
	nums := datosInt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DoubleInts(nums)
	}
}

func BenchmarkMapGeneric(b *testing.B) {
	nums := datosInt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapGeneric(nums, doblar)
	}
}

func BenchmarkBoxingSolo(b *testing.B) {
	// aisla el costo de meter un int en una interface{}, sin sumar nada
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = i + 100000
	}
}
