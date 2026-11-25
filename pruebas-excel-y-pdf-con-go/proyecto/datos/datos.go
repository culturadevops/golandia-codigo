// Package datos: las ventas del informe. Deterministas, para que el banco salga igual cada vez.
package datos

var Meses = []string{"Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic"}

var Productos = []string{"Teclado", "Ratón", "Monitor", "Portátil", "Auriculares"}

// Venta de un producto en un mes, en euros con céntimos.
func Venta(p, m int) float64 {
	base := []float64{1200, 450, 3100, 8700, 640}[p]
	return float64(int((base*(1+0.07*float64(m))+float64((p*37+m*53)%300))*100)) / 100
}
