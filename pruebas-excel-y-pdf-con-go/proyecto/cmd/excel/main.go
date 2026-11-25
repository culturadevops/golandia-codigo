// excel: el informe de ventas en un .xlsx, con estilos, fórmulas, filtro, panel fijo y gráfico.
package main

import (
	"fmt"

	"ejemplo.com/informe/datos"
	"github.com/xuri/excelize/v2"
)

func main() {
	f := excelize.NewFile()
	defer f.Close()
	h := "Ventas"
	f.SetSheetName("Sheet1", h)

	// Cabecera: Producto, los 12 meses y Total.
	f.SetCellValue(h, "A1", "Producto")
	for m, mes := range datos.Meses {
		c, _ := excelize.CoordinatesToCellName(m+2, 1)
		f.SetCellValue(h, c, mes)
	}
	f.SetCellValue(h, "N1", "Total")

	// Una fila por producto; el total es una FÓRMULA, no un número calculado en Go.
	for p, prod := range datos.Productos {
		fila := p + 2
		f.SetCellValue(h, fmt.Sprintf("A%d", fila), prod)
		for m := range datos.Meses {
			c, _ := excelize.CoordinatesToCellName(m+2, fila)
			f.SetCellValue(h, c, datos.Venta(p, m))
		}
		f.SetCellFormula(h, fmt.Sprintf("N%d", fila), fmt.Sprintf("SUM(B%d:M%d)", fila, fila))
	}
	ult := len(datos.Productos) + 1
	f.SetCellValue(h, fmt.Sprintf("A%d", ult+1), "Total")
	for col := 2; col <= 14; col++ {
		letra, _ := excelize.ColumnNumberToName(col)
		f.SetCellFormula(h, fmt.Sprintf("%s%d", letra, ult+1), fmt.Sprintf("SUM(%s2:%s%d)", letra, letra, ult))
	}

	cab, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"00ADD8"}, Pattern: 1},
	})
	euros := "#,##0.00 €"
	dinero, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &euros})
	total, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &euros, Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(h, "A1", "N1", cab)
	f.SetCellStyle(h, "B2", fmt.Sprintf("N%d", ult), dinero)
	f.SetCellStyle(h, fmt.Sprintf("A%d", ult+1), fmt.Sprintf("N%d", ult+1), total)
	f.SetColWidth(h, "A", "A", 14)
	f.SetColWidth(h, "B", "N", 12)

	f.SetPanes(h, &excelize.Panes{Freeze: true, XSplit: 1, YSplit: 1, TopLeftCell: "B2", ActivePane: "bottomRight"})
	f.AutoFilter(h, fmt.Sprintf("A1:N%d", ult), nil)

	// Un gráfico de columnas, una serie por producto, en su propia hoja.
	var series []excelize.ChartSeries
	for p := range datos.Productos {
		fila := p + 2
		series = append(series, excelize.ChartSeries{
			Name:       fmt.Sprintf("%s!$A$%d", h, fila),
			Categories: fmt.Sprintf("%s!$B$1:$M$1", h),
			Values:     fmt.Sprintf("%s!$B$%d:$M$%d", h, fila, fila),
		})
	}
	f.NewSheet("Gráfico")
	if err := f.AddChart("Gráfico", "A1", &excelize.Chart{
		Type: excelize.Col, Series: series,
		Title:     excelize.ChartTitle{Paragraph: []excelize.RichTextRun{{Text: "Ventas por mes"}}},
		Dimension: excelize.ChartDimension{Width: 900, Height: 420},
	}); err != nil {
		panic(err)
	}
	if err := f.SaveAs("informe.xlsx"); err != nil {
		panic(err)
	}
	fmt.Println("informe.xlsx: 2 hojas, 5 productos × 12 meses, 18 fórmulas, 1 gráfico")
}
