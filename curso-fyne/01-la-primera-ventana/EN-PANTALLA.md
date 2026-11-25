# Tu primera ventana en Go no compila, y el error no menciona ni Go ni Fyne — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. el primer go build · salida real

```console
$ go build ./etapas/e01-ventana
# github.com/go-gl/glfw/v3.4/glfw
./glfw/src/wl_platform.h:27:10: fatal error:
wayland-client-core.h: No such file or directory
compilation terminated.
```

## 2. lo que faltaba de verdad

```console
lo que dicen casi todos los tutoriales:
    libgl1-mesa-dev   xorg-dev

y ademas, desde Fyne 2.8:
    libwayland-dev
    wayland-protocols
    libxkbcommon-dev
```

## 3. la primera ventana, entera

```console
a := app.New()
w := a.NewWindow("nubecita")
w.SetContent(widget.NewLabel("Tu nube local, con ventana"))
w.Resize(fyne.NewSize(420, 140))
w.ShowAndRun()
```

## 4. los binarios de las etapas del curso · medidos

```console
e01 · una ventana con una etiqueta   32.343.016 B
e02 · contenedores                   32.365.456 B
e03 · binding                        32.385.736 B
e05 · lista de 50.000                32.423.312 B
e07 · formulario con validacion      32.575.040 B
```

## 5. lo que imprime la propia biblioteca · salida real

```console
*** This application has not been migrated to the
    fyne.Do threading model ***
*** The next major Fyne release will remove this
    safety! ***
```

## 6. el material del curso

```console
Fyne v2.8.1  ·  go1.26.8
compilado y capturado en Docker sobre Xvfb
corrida: 2026-09-03 · salida-real.txt
```
