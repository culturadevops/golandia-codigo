// e06 · El cliente hablando con nubecita DE VERDAD, con el kit oficial de AWS para Go.
//
// El patron entero del episodio, que es la union del 3 y el 4:
//   1. lanzar una gorrutina
//   2. hacer la llamada de red ALLI (fuera)
//   3. y meter SOLO el retoque de la interfaz en fyne.Do
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func cliente(punto string) *s3.Client {
	return s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(punto),
		UsePathStyle: true, // el fallo numero uno: sin esto el bucket va en el HOST
	})
}

func main() {
	c := cliente("http://nubecita:4566")

	a := app.New()
	w := a.NewWindow("nubecita · buckets")

	estado := widget.NewLabel("conectando...")
	progreso := widget.NewProgressBarInfinite()
	var nombres []string

	lista := widget.NewList(
		func() int { return len(nombres) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(nombres[i])
		},
	)

	refrescar := func() {
		go func() { // 1 · la gorrutina
			ctx, parar := context.WithTimeout(context.Background(), 5*time.Second)
			defer parar()
			r, err := c.ListBuckets(ctx, &s3.ListBucketsInput{}) // 2 · la red, FUERA
			fyne.Do(func() { // 3 · y SOLO el retoque, dentro
				progreso.Hide()
				if err != nil {
					estado.SetText("error: " + err.Error())
					return
				}
				nombres = nombres[:0]
				for _, b := range r.Buckets {
					nombres = append(nombres, aws.ToString(b.Name))
				}
				estado.SetText(fmt.Sprintf("%d buckets · 4566 · conectado", len(nombres)))
				lista.Refresh()
			})
		}()
	}

	boton := widget.NewButton("Refrescar", func() {
		progreso.Show()
		estado.SetText("pidiendo...")
		refrescar()
	})

	arriba := container.NewVBox(estado, progreso, boton)
	w.SetContent(container.NewBorder(arriba, nil, nil, nil, lista))
	w.Resize(fyne.NewSize(640, 360))
	refrescar()
	w.ShowAndRun()
}
