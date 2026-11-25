# Tu app arranca sin iconos en otro ordenador: empotrar con go:embed — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. un tema, entero: cuatro metodos

```console
Color(nombre, variante)  -> que color es este
Font(estilo)             -> que tipografia
Icon(nombre)             -> que icono
Size(nombre)             -> que tamano
```

## 2. cambiar un color y delegar todo lo demas

```console
type temaGo struct{ fyne.Theme }   // el de por defecto, dentro

func (t temaGo) Color(n fyne.ThemeColorName,
        v fyne.ThemeVariant) color.Color {
    if n == theme.ColorNamePrimary {
        return color.NRGBA{0x00, 0xAD, 0xD8, 255}   // el cian de Go
    }
    return t.Theme.Color(n, v)                      // y el resto, delegado
}
```

## 3. empotrar, que es del lenguaje

```console
import _ "embed"

//go:embed icono.png
var iconoPNG []byte    // el fichero entra en el BINARIO
```

## 4. de bytes a recurso de Fyne

```console
icono := fyne.NewStaticResource("icono.png", iconoPNG)

a.SetIcon(icono)          // el de la aplicacion
w.SetIcon(icono)          // el de la ventana
widget.NewIcon(icono)     // y dibujado dentro
```

## 5. los tres numeros · medidos

```console
el icono en disco          235 bytes
la ventana pelada   32.343.016 bytes
con tema e icono    32.377.552 bytes
```

## 6. los que ya vienen

```console
theme.FolderIcon()     theme.DocumentIcon()
theme.DeleteIcon()     theme.SearchIcon()
theme.WarningIcon()    theme.ContentAddIcon()
```

## 7. empotrar una carpeta entera

```console
//go:embed recursos/*
var recursos embed.FS

datos, _ := recursos.ReadFile("recursos/plantilla.json")
lista, _ := recursos.ReadDir("recursos")
```

## 8. el material del curso

```console
Fyne v2.8.1  ·  go1.26.8
icono 235 B · binario +34.536 B
corrida: 2026-09-03 · salida-real.txt
```
