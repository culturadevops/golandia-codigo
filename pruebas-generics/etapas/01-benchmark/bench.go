package bench

// Number: el tipo generico. La lista queda cerrada a lo que de verdad se puede sumar.
type Number interface {
	~int | ~int64 | ~float64
}

// SumGeneric: una sola funcion, cualquier tipo de la lista.
func SumGeneric[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// SumInt: la version escrita a mano, solo para int. Es el piso de referencia.
func SumInt(nums []int) int {
	var total int
	for _, n := range nums {
		total += n
	}
	return total
}

// MapGeneric: transforma cada elemento con la funcion que le pases. Sirve para
// cualquier par de tipos de entrada y salida.
func MapGeneric[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

// DoubleInts: la version escrita a mano, solo para duplicar enteros.
func DoubleInts(s []int) []int {
	r := make([]int, len(s))
	for i, v := range s {
		r[i] = v * 2
	}
	return r
}

// SumInterface: la version pre-generics, con interface{} y un type switch.
func SumInterface(nums []interface{}) float64 {
	var total float64
	for _, n := range nums {
		switch v := n.(type) {
		case int:
			total += float64(v)
		case float64:
			total += v
		}
	}
	return total
}
