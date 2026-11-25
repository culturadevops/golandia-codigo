// grande: N filas × 10 columnas, con SetCellValue o con StreamWriter. Dice tiempo y memoria.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/xuri/excelize/v2"
)

func main() {
	modo := flag.String("modo", "normal", "normal o stream")
	n := flag.Int("filas", 100000, "filas")
	flag.Parse()
	t0 := time.Now()
	f := excelize.NewFile()
	defer f.Close()
	if *modo == "stream" {
		sw, err := f.NewStreamWriter("Sheet1")
		if err != nil {
			panic(err)
		}
		fila := make([]any, 10)
		for i := 1; i <= *n; i++ {
			for c := range fila {
				fila[c] = i*10 + c
			}
			celda, _ := excelize.CoordinatesToCellName(1, i)
			if err := sw.SetRow(celda, fila); err != nil {
				panic(err)
			}
		}
		if err := sw.Flush(); err != nil {
			panic(err)
		}
	} else {
		for i := 1; i <= *n; i++ {
			for c := 0; c < 10; c++ {
				celda, _ := excelize.CoordinatesToCellName(c+1, i)
				f.SetCellValue("Sheet1", celda, i*10+c)
			}
		}
	}
	nombre := "grande-" + *modo + ".xlsx"
	if err := f.SaveAs(nombre); err != nil {
		panic(err)
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	st, _ := os.Stat(nombre)
	fmt.Printf("%-6s %d filas: %v · fichero %.1f MB · memoria pedida en total %d MB\n",
		*modo, *n, time.Since(t0).Round(10*time.Millisecond), float64(st.Size())/1e6, ms.TotalAlloc>>20)
}
