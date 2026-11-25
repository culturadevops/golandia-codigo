package main

type Ruta struct{ Pasos []int }

func main() {
	// Un slice NO es comparable, asi que no puede ser clave.
	m := map[[]int]string{}
	_ = m
	// Y un struct que contiene un slice, tampoco.
	m2 := map[Ruta]string{}
	_ = m2
}
