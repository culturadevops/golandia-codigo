// Usa la app como lo haría una persona (escribir, tocar el botón) con el driver de pruebas de Fyne, que dibuja la
// ventana en memoria, sin pantalla, y guarda cada paso como PNG.
package main

import (
	"image/png"
	"os"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func guardar(t *testing.T, v *ventana, nombre string) {
	f, err := os.Create("/out/" + nombre + ".png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, v.w.Canvas().Capture())
}

func TestUsarLaApp(t *testing.T) {
	a := test.NewApp()
	a.Settings().SetTheme(theme.DefaultTheme())
	v := construir(a)
	v.w.Resize(fyne.NewSize(420, 320))
	guardar(t, v, "1-vacia")
	for _, tarea := range []string{"Comprar café", "Grabar el vídeo de Fyne", "Responder comentarios"} {
		test.Type(v.entrada, tarea)
		test.Tap(v.agregar)
	}
	guardar(t, v, "2-con-tareas")
	if got := v.conteo.Text; got != "3 tareas" {
		t.Fatalf("el contador dice %q", got)
	}
	test.Tap(v.agregar) // con la caja vacía no debe agregar nada
	if got := v.conteo.Text; got != "3 tareas" {
		t.Fatalf("agregó una tarea vacía: %q", got)
	}
	t.Logf("contador: %s", v.conteo.Text)
}
