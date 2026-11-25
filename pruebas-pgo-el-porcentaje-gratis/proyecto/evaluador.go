package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"time"
)

// Un evaluador de expresiones: cada nodo del árbol es una interfaz, así que cada Eval es una
// llamada dinámica. Es el tipo de código caliente donde PGO puede ayudar.
type Nodo interface{ Eval(x []float64) float64 }

type Num float64
type Var int
type Suma struct{ A, B Nodo }
type Mul struct{ A, B Nodo }
type Func struct{ A Nodo }

func (n Num) Eval(_ []float64) float64  { return float64(n) }
func (v Var) Eval(x []float64) float64  { return x[int(v)%len(x)] }
func (s Suma) Eval(x []float64) float64 { return s.A.Eval(x) + s.B.Eval(x) }
func (m Mul) Eval(x []float64) float64  { return m.A.Eval(x) * m.B.Eval(x) }
func (f Func) Eval(x []float64) float64 { return suavizar(f.A.Eval(x)) }

// suavizar es lo bastante grande para que el compilador NO la inlinee por su cuenta.
func suavizar(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	if v > 1e6 {
		v = 1e6 + math.Log1p(v-1e6)
	} else if v < -1e6 {
		v = -1e6 - math.Log1p(-v-1e6)
	}
	r := v
	for i := 0; i < 3; i++ {
		r = r*0.5 + v*0.5
	}
	return r
}

// arbol construye siempre el mismo árbol (determinista) de la profundidad pedida.
func arbol(p, semilla int) Nodo {
	if p == 0 {
		if semilla%3 == 0 {
			return Num(float64(semilla%7) + 0.5)
		}
		return Var(semilla)
	}
	switch semilla % 4 {
	case 0, 1:
		return Suma{arbol(p-1, semilla*3+1), arbol(p-1, semilla*5+2)}
	case 2:
		return Mul{arbol(p-1, semilla*7+3), Num(0.99)}
	default:
		return Func{arbol(p-1, semilla*11+4)}
	}
}

func trabajo(vueltas int) float64 {
	t := arbol(12, 1)
	x := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	total := 0.0
	for i := 0; i < vueltas; i++ {
		x[i%len(x)] += 1e-9
		total += t.Eval(x)
	}
	return total
}

func main() {
	perfil := flag.String("perfil", "", "guarda un perfil de CPU en este fichero")
	flag.Parse()
	if *perfil != "" {
		f, err := os.Create(*perfil)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	ini := time.Now()
	r := trabajo(2_000_000)
	fmt.Printf("resultado %.3f en %v\n", r, time.Since(ini).Round(time.Millisecond))
}
