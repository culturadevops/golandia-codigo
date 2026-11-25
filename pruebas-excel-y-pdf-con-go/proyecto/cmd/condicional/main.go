// condicional: sobre el informe.xlsx, pinta de verde las ventas de más de 8.000 € y añade
// una columna «Estado» con una lista desplegable. Lo guarda como informe-plus.xlsx.
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
	verde, _ := f.NewConditionalStyle(&excelize.Style{
		Font: &excelize.Font{Color: "006100"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"C6EFCE"}, Pattern: 1},
	})
	if err := f.SetConditionalFormat("Ventas", "B2:M6", []excelize.ConditionalFormatOptions{
		{Type: "cell", Criteria: ">", Format: &verde, Value: "8000"},
	}); err != nil {
		panic(err)
	}
	f.SetCellValue("Ventas", "O1", "Estado")
	dv := excelize.NewDataValidation(true)
	dv.Sqref = "O2:O6"
	dv.SetDropList([]string{"Pendiente", "Revisado", "Aprobado"})
	if err := f.AddDataValidation("Ventas", dv); err != nil {
		panic(err)
	}
	if err := f.SaveAs("informe-plus.xlsx"); err != nil {
		panic(err)
	}
	fmt.Println("informe-plus.xlsx: verde si > 8000 en B2:M6 · lista en O2:O6")
}
