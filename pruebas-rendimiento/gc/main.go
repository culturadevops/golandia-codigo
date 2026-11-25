package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

type Registro struct {
	ID     int
	Nombre string
	Datos  [64]byte
}

// ⚠️ El sumidero global es OBLIGATORIO, no adorno. Sin el, el escape analysis ve que el
// puntero no sale de la funcion, lo deja en la PILA y no hay ni una asignacion en el monton:
// la primera version de esta prueba daba 0 ciclos de GC y 0 MB, o sea que no medía nada.
var sumidero *Registro

// Version que ASIGNA MUCHO: un puntero nuevo por vuelta, y que se escapa de verdad.
func asignaMucho(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		r := &Registro{ID: i, Nombre: "linea"}
		sumidero = r // fuerza el monton: sin esto el compilador lo borra
		total += r.ID
	}
	return total
}

// Version que ASIGNA POCO: reutiliza el mismo valor, sin puntero.
func asignaPoco(n int) int {
	total := 0
	var r Registro
	for i := 0; i < n; i++ {
		r.ID = i
		r.Nombre = "linea"
		total += r.ID
	}
	return total
}

func medir(nombre string, f func(int) int, n int) {
	runtime.GC()
	var antes, despues runtime.MemStats
	runtime.ReadMemStats(&antes)
	t := time.Now()
	f(n)
	d := time.Since(t)
	runtime.ReadMemStats(&despues)
	fmt.Printf("  %-34s %8v  ciclos de GC: %-4d  asignado: %5d MB\n",
		nombre, d.Round(time.Millisecond),
		despues.NumGC-antes.NumGC,
		(despues.TotalAlloc-antes.TotalAlloc)/1024/1024)
}

func main() {
	const N = 5000000
	modo := ""
	if len(os.Args) > 1 {
		modo = os.Args[1]
	}
	switch modo {
	case "gogc":
		fmt.Println("MISMO codigo (el que asigna mucho), tocando solo el recolector")
		for _, g := range []int{100, 400, 1600} {
			debug.SetGCPercent(g)
			medir(fmt.Sprintf("GOGC=%d", g), asignaMucho, N)
		}
	default:
		debug.SetGCPercent(100) // el valor por defecto
		fmt.Println("MISMO recolector (GOGC=100), cambiando el codigo")
		medir("asigna mucho (puntero por vuelta)", asignaMucho, N)
		medir("asigna poco (reutiliza el valor)", asignaPoco, N)
	}
}
