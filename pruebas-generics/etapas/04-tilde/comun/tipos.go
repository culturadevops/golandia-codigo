package comun

// ConTilde: acepta float64 Y cualquier tipo cuyo tipo subyacente sea float64.
type ConTilde interface {
	~float64
}

// SinTilde: acepta SOLO float64, exactamente, ningún tipo nombrado encima.
type SinTilde interface {
	float64
}

// Celsius: un tipo nombrado, con tipo subyacente float64.
type Celsius float64

func SumConTilde[T ConTilde](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

func SumSinTilde[T SinTilde](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}
