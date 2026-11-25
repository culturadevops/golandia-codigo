// leer: abre el informe.xlsx desde Go y lee el total, que es una fórmula.
package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("informe.xlsx")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	formula, _ := f.GetCellFormula("Ventas", "N7")
	valor, _ := f.GetCellValue("Ventas", "N7")
	calc, err := f.CalcCellValue("Ventas", "N7")
	fmt.Printf("N7 fórmula:        =%s\n", formula)
	fmt.Printf("N7 GetCellValue:   %q\n", valor)
	fmt.Printf("N7 CalcCellValue:  %q (err: %v)\n", calc, err)
	filas, _ := f.GetRows("Ventas")
	fmt.Printf("filas en Ventas:   %d\n", len(filas))
}
