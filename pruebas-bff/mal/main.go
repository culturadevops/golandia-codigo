package main

import "pruebas/api"

type Servidor struct{}

// A proposito: no implementa GetPantalla. El contrato dice que hace falta.
var _ api.ServerInterface = Servidor{}

func main() {}
