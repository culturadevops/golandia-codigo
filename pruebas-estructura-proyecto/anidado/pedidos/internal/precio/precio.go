// Package precio está en pedidos/internal: solo pedidos y lo que cuelga de pedidos/ pueden importarlo.
package precio

func ConIVA(neto float64) float64 { return neto * 1.16 }
