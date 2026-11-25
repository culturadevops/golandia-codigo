package comun

type Number interface {
	~int | ~int64 | ~float64
}

func SumGeneric[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

func SumInterface(nums []interface{}) float64 {
	var total float64
	for _, n := range nums {
		switch v := n.(type) {
		case int:
			total += float64(v)
		case float64:
			total += v
		}
		// si no es int ni float64 (por ejemplo, un string), se lo salta EN SILENCIO
	}
	return total
}
