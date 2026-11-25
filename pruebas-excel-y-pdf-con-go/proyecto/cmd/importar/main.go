// importar: leer el informe.xlsx fila a fila desde Go, como si lo hubiera hecho otra persona.
package main

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func main() {
	for _, crudo := range []bool{false, true} {
		f, err := excelize.OpenFile("informe.xlsx", excelize.Options{RawCellValue: crudo})
		if err != nil {
			panic(err)
		}
		filas, _ := f.Rows("Ventas")
		filas.Next() // la cabecera
		filas.Next()
		cols, _ := filas.Columns()
		_, err = strconv.ParseFloat(cols[1], 64)
		fmt.Printf("RawCellValue=%-5v %s, enero: %q · ParseFloat: %v\n", crudo, cols[0], cols[1], err)
		filas.Close()
		f.Close()
	}
}
