package cmd

// funcionAuxiliar no la llama nadie en ningún otro lado del proyecto.
func funcionAuxiliar() int {
	return 42
}

func calculoConBugDeAsignacion(base int) int {
	total := base * 2
	total = base * 3 // esta reasignacion pisa la anterior sin haberla leido nunca
	return total
}
