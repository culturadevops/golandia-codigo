# Dos ventanas idénticas y una está mal: la regla de hilos de Fyne — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. el que escribe casi todo el mundo

```console
go func() {
    for i := 1; i <= 5; i++ {
        time.Sleep(300 * time.Millisecond)
        etiqueta.SetText(fmt.Sprintf("subido %d de 5", i))
    }
}()
```

## 2. el correcto · la unica diferencia

```console
go func() {
    for i := 1; i <= 5; i++ {
        time.Sleep(300 * time.Millisecond)   // el TRABAJO, fuera
        n := i
        fyne.Do(func() {                     // y SOLO el retoque, dentro
            etiqueta.SetText(fmt.Sprintf("subido %d de 5", n))
        })
    }
}()
```

## 3. lo que imprime el programa MALO · salida real

```console
*** This application has not been migrated
    to the fyne.Do threading model ***
*** The next major Fyne release will remove
    this safety! ***
*** Read more at docs.fyne.io/started/goroutines ***
```

## 4. y sigue · salida real

```console
*** Error in Fyne call thread, this should have
    been called in fyne.Do[AndWait] ***
*** This error needs to be addressed as it could
    cause concurrent errors soon ***
*** The next major Fyne release will disable
    this safety by default! ***
```

## 5. y te dice la linea · salida real

```console
  From: /app/etapas/e04-mal/main.go:27
```

## 6. thread.go, del propio modulo · desde la 2.6

```console
func DoAndWait(fn func()) {
    CurrentApp().Driver().DoFromGoroutine(fn, true)
}

func Do(fn func()) {
    CurrentApp().Driver().DoFromGoroutine(fn, false)
}
```

## 7. por donde buscar

```console
busca:   go func()
y dentro: SetText · Refresh · Enable · Disable
          Show · Hide · Close · Set(  (enlaces)
```

## 8. la forma fiable

```console
$ go run -race .
WARNING: DATA RACE
  Write at 0x... by goroutine 8
  Previous read at 0x... by main goroutine
```

## 9. el material del curso

```console
Fyne v2.8.1  ·  go1.26.8
las dos capturas: e04-mal.png y e04-bien.png
corrida: 2026-09-03 · salida-real.txt
```
