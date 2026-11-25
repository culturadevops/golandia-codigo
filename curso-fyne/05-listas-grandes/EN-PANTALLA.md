# 50.000 objetos en una lista de Fyne, y solo construye nueve filas — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. las tres funciones de una lista

```console
widget.NewList(
    func() int { return len(claves) },              // cuantos hay
    func() fyne.CanvasObject {                      // fabrica una vacia
        return widget.NewLabel("plantilla")
    },
    func(i widget.ListItemID, o fyne.CanvasObject) {// rellena la fila i
        o.(*widget.Label).SetText(claves[i])
    })
```

## 2. el contador, dentro de la funcion que FABRICA

```console
func() fyne.CanvasObject {   // se llama UNA vez por fila VISIBLE
    construidas++
    return widget.NewLabel("plantilla")
}
```

## 3. la trampa, en codigo

```console
func(i widget.ListItemID, o fyne.CanvasObject) {
    fila := o.(*Fila)
    fila.texto.SetText(claves[i])
    if esCarpeta(claves[i]) {
        fila.icono.Show()
    }                      // <- y si NO lo es, se queda el de antes
}
```

## 4. una fila compuesta

```console
func() fyne.CanvasObject {
    return container.NewBorder(nil, nil, nil,
        widget.NewButton("Borrar", nil),
        container.NewHBox(
            widget.NewLabel("clave"),
            widget.NewLabel("tamano"),
        ))
}
```

## 5. como se llega a los hijos · fragil

```console
func(i widget.ListItemID, o fyne.CanvasObject) {
    caja := o.(*fyne.Container)
    fila := caja.Objects[0].(*fyne.Container)
    fila.Objects[0].(*widget.Label).SetText(...)
}
```

## 6. con un tipo propio · legible y a prueba

```console
type Fila struct {
    widget.BaseWidget
    clave, tamano *widget.Label
    borrar        *widget.Button
}

fila := o.(*Fila)
fila.clave.SetText(claves[i])
```

## 7. el material del curso

```console
Fyne v2.8.1  ·  go1.26.8
50000 claves · filas construidas: 9
corrida: 2026-09-03 · salida-real.txt
```
