# Con cgo se acabó cruzar gratis: el precio real de repartir tu app en Go — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. quitar la informacion de depuracion

```console
$ go build -ldflags="-s -w" -o app .
```

## 2. y comprimido, que es lo que se descarga

```console
$ gzip -9 -c app > app.gz
11.456.996 bytes
```

## 3. empaquetar, con la herramienta de Fyne

```console
$ go install fyne.io/tools/cmd/fyne@latest
$ fyne package -os windows -icon icono.png
$ fyne package -os darwin  -icon icono.png
$ fyne package -os linux   -icon icono.png
```

## 4. apagar el cgo · salida real

```console
$ CGO_ENABLED=0 go build .
package fynelab/etapas/e01-ventana
    imports fyne.io/fyne/v2/app
    imports fyne.io/fyne/v2/internal/driver/glfw
    imports fyne.io/fyne/v2/internal/driver/common
```

## 5. cruzar a Windows con cgo · salida real

```console
$ CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build .
gcc: error: unrecognized command-line option
     '-mthreads'; did you mean '-pthread'?
```

## 6. el control · el mismo ordenador, sin Fyne

```console
$ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build .
hola.exe   2.461.184 bytes
[ok] sin instalar absolutamente nada
```

## 7. 1 · cada sistema, en su sistema

```console
trabajo: construir
  matriz:
    sistema: [ubuntu, windows, macos]
  pasos:
    - go build -ldflags="-s -w"
```

## 8. el material del curso

```console
tal cual        32.343.016 B
con -s -w       23.994.000 B
comprimido      11.456.996 B
corrida: 2026-09-03 · salida-real.txt
```
