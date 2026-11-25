package pedidos

import "github.com/golandia/anidado/pedidos/internal/precio"

func Total(neto float64) float64 { return precio.ConIVA(neto) }
