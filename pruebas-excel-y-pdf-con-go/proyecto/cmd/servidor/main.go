// servidor: los dos informes, generados al vuelo y servidos por HTTP, sin tocar el disco.
package main

import (
	"fmt"
	"net/http"

	"ejemplo.com/informe/datos"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/xuri/excelize/v2"
)

func main() {
	http.HandleFunc("/informe.xlsx", func(w http.ResponseWriter, r *http.Request) {
		f := excelize.NewFile()
		defer f.Close()
		for p, prod := range datos.Productos {
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", p+1), prod)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", p+1), datos.Venta(p, 0))
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="informe.xlsx"`)
		f.Write(w) // directo a la respuesta
	})
	http.HandleFunc("/informe.pdf", func(w http.ResponseWriter, r *http.Request) {
		m := maroto.New(config.NewBuilder().WithCompression(true).Build())
		for _, prod := range datos.Productos {
			m.AddRows(text.NewRow(8, prod))
		}
		doc, err := m.Generate()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `inline; filename="informe.pdf"`)
		w.Write(doc.GetBytes())
	})
	http.ListenAndServe("127.0.0.1:8080", nil)
}
