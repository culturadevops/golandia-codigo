# El enlace de datos en Fyne: tocas el dato y la interfaz se entera sola — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. sin enlace · el boton lo hace todo

```console
subir := widget.NewButton("Subir", func() {
    n++
    etiqueta.SetText(fmt.Sprintf("%d objetos", n))
    titulo.SetText(...)                             
    estado.SetText(...)                             
    lista.Refresh()                                 
})
```

## 2. el dato, con enlace

```console
objetos := binding.NewInt()
objetos.Set(2)

n, _ := objetos.Get()
objetos.Set(n + 1)
```

## 3. la etiqueta, atada al dato

```console
texto := binding.IntToStringWithFormat(objetos,
             "el bucket tiene %d objetos")

etiqueta := widget.NewLabelWithData(texto)
```

## 4. el boton, con enlace

```console
subir := widget.NewButton("Subir un objeto", func() {
    n, _ := objetos.Get()
    objetos.Set(n + 1)   // y ya esta
})
```

## 5. una lista atada al dato

```console
claves := binding.NewStringList()

lista := widget.NewListWithData(claves,
    func() fyne.CanvasObject { return widget.NewLabel("") },
    func(i binding.DataItem, o fyne.CanvasObject) {
        o.(*widget.Label).Bind(i.(binding.String))
    })
```

## 6. la tercera via: tu propio estado

```console
func (a *App) setObjetos(n int) {
    a.objetos = n
    a.etiqueta.SetText(fmt.Sprintf("%d objetos", n))
    a.estado.SetText(...)
}
```

## 7. el error que casi nadie mira

```console
n, err := objetos.Get()
//    ^^^ esto casi siempre se escribe como _
```

## 8. el material del curso

```console
Fyne v2.8.1  ·  go1.26.8
capturado en Docker sobre Xvfb, escala 3
corrida: 2026-09-03 · salida-real.txt
```
