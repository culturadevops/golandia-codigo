// ¿Cuánto dinero pierde float64 de verdad? Medido, no contado.
package dinero

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

// El caso más aburrido y más común: sumar un precio muchas veces.
const (
	precio = "0.10"
	veces  = 100000
)

func TestDesvioAcumulado(t *testing.T) {
	// --- float64 ---
	var f float64
	for i := 0; i < veces; i++ {
		f += 0.10
	}

	// --- decimal ---
	d := decimal.Zero
	p, _ := decimal.NewFromString(precio)
	for i := 0; i < veces; i++ {
		d = d.Add(p)
	}

	esperado, _ := decimal.NewFromString("10000")
	fmt.Printf("  sumando %s, %d veces\n", precio, veces)
	fmt.Printf("  float64 : %.10f\n", f)
	fmt.Printf("  decimal : %s\n", d.String())
	fmt.Printf("  exacto  : %s\n", esperado.String())

	desvio := decimal.NewFromFloat(f).Sub(esperado)
	fmt.Printf("  desvio de float64: %s\n", desvio.String())
	if !d.Equal(esperado) {
		t.Fatalf("decimal deberia dar exacto, dio %s", d.String())
	}
}

// ⚠️ El fallo que más se comete: NewFromFloat. El dato ya viene roto.
func TestNewFromFloatYaVieneRoto(t *testing.T) {
	malo := decimal.NewFromFloat(0.1 + 0.2)
	bueno, _ := decimal.NewFromString("0.3")
	fmt.Printf("  NewFromFloat(0.1+0.2) = %s\n", malo.String())
	fmt.Printf("  NewFromString(\"0.3\")  = %s\n", bueno.String())
	fmt.Printf("  ¿son iguales? %v\n", malo.Equal(bueno))
}

// La división no termina: hay que decir cuántos decimales quieres.
func TestDivisionNoTermina(t *testing.T) {
	uno := decimal.NewFromInt(1)
	tres := decimal.NewFromInt(3)
	fmt.Printf("  1/3 con Div (por defecto): %s\n", uno.Div(tres).String())
	fmt.Printf("  1/3 con DivRound a 2:      %s\n", uno.DivRound(tres, 2).String())
	fmt.Printf("  DivisionPrecision global:  %d\n", decimal.DivisionPrecision)
}

// El reparto que no cuadra: 10 euros entre 3 personas.
func TestRepartoQueNoCuadra(t *testing.T) {
	total, _ := decimal.NewFromString("10.00")
	tres := decimal.NewFromInt(3)
	parte := total.DivRound(tres, 2)
	suma := parte.Mul(tres)
	fmt.Printf("  10.00 entre 3 = %s cada uno\n", parte.String())
	fmt.Printf("  y 3 x %s = %s  (faltan %s)\n", parte.String(), suma.String(), total.Sub(suma).String())
}

// ── El precio, medido ────────────────────────────────────────────────────
func BenchmarkSumaFloat64(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f += 0.10
	}
	_ = f
}

func BenchmarkSumaDecimal(b *testing.B) {
	d := decimal.Zero
	p, _ := decimal.NewFromString("0.10")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d = d.Add(p)
	}
	_ = d
}

func BenchmarkSumaEnteroCentimos(b *testing.B) {
	var c int64
	for i := 0; i < b.N; i++ {
		c += 10
	}
	_ = c
}

// ⚠️ Mi encargo decía "NUNCA uses NewFromFloat". El test de arriba NO lo reprodujo:
// con 0.1+0.2 dio exactamente 0.3. Así que busco si falla en algún sitio, en vez de
// repetir el consejo de internet sin comprobarlo.
func TestNewFromFloatCuandoSiDuele(t *testing.T) {
	casos := []float64{0.1 + 0.2, 1.005, 2.675, 0.615, 1e17 + 1, 0.1 * 3}
	for _, f := range casos {
		d := decimal.NewFromFloat(f)
		fmt.Printf("  float64 %.20g -> decimal %s\n", f, d.String())
	}
}
