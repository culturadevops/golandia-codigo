// orden: con StreamWriter, ¿se puede volver atrás a una fila ya escrita?
package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func main() {
	f := excelize.NewFile()
	defer f.Close()
	sw, _ := f.NewStreamWriter("Sheet1")
	fmt.Println("fila 2:", sw.SetRow("A2", []any{"segunda"}))
	fmt.Println("fila 1:", sw.SetRow("A1", []any{"primera"}))
}
