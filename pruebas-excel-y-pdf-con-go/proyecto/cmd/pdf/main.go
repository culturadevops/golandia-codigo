// pdf: el mismo informe en PDF con maroto v2: cabecera en cada página, tabla por trimestres,
// total y número de página. Con -detalle N añade N líneas de venta para ver cómo pagina.
package main

import (
	"flag"
	"fmt"
	"strings"

	"ejemplo.com/informe/datos"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/fontrepository"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// euros: 12345.6 → "12.345,60 €" (formato de España y Latinoamérica con punto de miles).
func euros(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	ent, dec := s[:len(s)-3], s[len(s)-2:]
	var b strings.Builder
	for i, r := range ent {
		if i > 0 && (len(ent)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return b.String() + "," + dec + " €"
}

func main() {
	detalle := flag.Int("detalle", 0, "líneas de venta extra")
	salida := flag.String("o", "informe.pdf", "fichero de salida")
	comprimir := flag.Bool("comprimir", false, "comprimir el PDF (maroto no lo hace por defecto)")
	fuente := flag.String("fuente", "", "un .ttf para escribir en UTF-8 (sin él: Helvetica, WinAnsi)")
	otros := flag.Bool("otros", false, "añadir una línea con letras que no están en WinAnsi")
	qr := flag.String("qr", "", "añadir al final un código QR con este enlace")
	flag.Parse()

	b := config.NewBuilder()
	if *fuente != "" {
		fuentes, err := fontrepository.New().
			AddUTF8Font("dejavu", fontstyle.Normal, *fuente).
			AddUTF8Font("dejavu", fontstyle.Bold, *fuente).
			Load()
		if err != nil {
			panic(err)
		}
		b.WithCustomFonts(fuentes).WithDefaultFont(&props.Font{Family: "dejavu"})
	}
	cfg := b.
		WithTitle("Informe de ventas", true).
		WithPageNumber(props.PageNumber{Pattern: "Página {current} de {total}", Place: props.Bottom, Size: 8}).
		WithCompression(*comprimir).
		Build()
	m := maroto.New(cfg)

	azul := &props.Color{Red: 0, Green: 173, Blue: 216}
	blanco := &props.Color{Red: 255, Green: 255, Blue: 255}
	cab := props.Text{Style: fontstyle.Bold, Color: blanco, Align: align.Center, Top: 1.5}
	if err := m.RegisterHeader(
		text.NewRow(12, "Informe de ventas 2026 · año, señal, ñandú, €", props.Text{Size: 14, Style: fontstyle.Bold}),
		row.New(8).Add(
			text.NewCol(2, "Producto", cab), text.NewCol(2, "T1", cab), text.NewCol(2, "T2", cab),
			text.NewCol(2, "T3", cab), text.NewCol(2, "T4", cab), text.NewCol(2, "Total", cab),
		).WithStyle(&props.Cell{BackgroundColor: azul}),
	); err != nil {
		panic(err)
	}

	if *otros {
		m.AddRows(text.NewRow(8, "Sucursales: Kraków, Łódź → Москва", props.Text{Top: 1}))
	}
	der := props.Text{Align: align.Right, Top: 1}
	total := 0.0
	for p, prod := range datos.Productos {
		cols := []any{prod}
		suma := 0.0
		for t := 0; t < 4; t++ {
			tri := 0.0
			for m := t * 3; m < t*3+3; m++ {
				tri += datos.Venta(p, m)
			}
			suma += tri
			cols = append(cols, euros(tri))
		}
		total += suma
		r := row.New(7).Add(text.NewCol(2, prod, props.Text{Top: 1}))
		for _, c := range cols[1:] {
			r.Add(text.NewCol(2, c.(string), der))
		}
		r.Add(text.NewCol(2, euros(suma), props.Text{Align: align.Right, Top: 1, Style: fontstyle.Bold}))
		m.AddRows(r)
	}
	m.AddRows(row.New(8).Add(
		text.NewCol(10, "Total del año", props.Text{Style: fontstyle.Bold, Top: 1.5}),
		text.NewCol(2, euros(total), props.Text{Style: fontstyle.Bold, Align: align.Right, Top: 1.5}),
	))

	for i := 0; i < *detalle; i++ {
		p, mes := i%len(datos.Productos), (i/len(datos.Productos))%12
		m.AddRows(row.New(6).Add(
			text.NewCol(2, fmt.Sprintf("Línea %d", i+1), props.Text{Size: 8}),
			text.NewCol(4, datos.Productos[p], props.Text{Size: 8}),
			text.NewCol(3, datos.Meses[mes], props.Text{Size: 8}),
			text.NewCol(3, euros(datos.Venta(p, mes)/30), props.Text{Size: 8, Align: align.Right}),
		))
	}

	if *qr != "" {
		m.AddRows(row.New(30).Add(
			code.NewQrCol(3, *qr),
			text.NewCol(9, "Versión en línea: "+*qr, props.Text{Top: 12}),
		))
	}

	doc, err := m.Generate()
	if err != nil {
		panic(err)
	}
	if err := doc.Save(*salida); err != nil {
		panic(err)
	}
	fmt.Printf("%s: total del año %s\n", *salida, euros(total))
}
