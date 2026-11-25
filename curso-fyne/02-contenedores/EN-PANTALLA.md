# Le pusiste un tamaño a tu botón y no hizo caso: los contenedores de Fyne — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. caja vertical · cada hijo, su altura minima

```console
container.NewVBox(
    widget.NewLabel("Buckets de tu nube local"),
    entrada,
    widget.NewButton("Crear bucket", crear),
)
```

## 2. rejilla · todos del mismo tamano

```console
container.NewGridWithColumns(2,
    widget.NewLabel("fotos"),    widget.NewLabel("2 objetos"),
    widget.NewLabel("copias"),   widget.NewLabel("0 objetos"),
    widget.NewLabel("informes"), widget.NewLabel("17 objetos"),
)
```

## 3. borde · el que estructura una aplicacion

```console
container.NewBorder(
    arriba,      // la barra de titulo
    abajo,       // la barra de estado
    izquierda,   // la columna de servicios
    nil,         // nada a la derecha
    centro,      // y esto se come TODO lo que sobra
)
```

## 4. pestanas · un contenedor mas

```console
container.NewAppTabs(
    container.NewTabItem("Buckets",   panelS3),
    container.NewTabItem("Colas",     panelColas),
    container.NewTabItem("Funciones", panelLambda),
)
```

## 5. darle nombre a las zonas

```console
func (a *App) barraSuperior() fyne.CanvasObject { ... }
func (a *App) panelLateral() fyne.CanvasObject { ... }
func (a *App) contenido()    fyne.CanvasObject { ... }

container.NewBorder(a.barraSuperior(), a.estado(),
                    a.panelLateral(), nil, a.contenido())
```

## 6. una regla de colocacion, entera

```console
type Layout interface {
    MinSize(objetos []fyne.CanvasObject) fyne.Size
    Layout(objetos []fyne.CanvasObject, tam fyne.Size)
}
```

## 7. el material del curso

```console
Fyne v2.8.1  ·  go1.26.8
capturado en Docker sobre Xvfb, escala 3
corrida: 2026-09-03 · salida-real.txt
```
