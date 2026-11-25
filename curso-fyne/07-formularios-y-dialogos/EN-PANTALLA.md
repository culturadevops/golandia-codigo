# Un botón que se apaga solo, y un diálogo que no bloquea (curso 7/10) — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. un validador, entero

```console
nombre.Validator = func(s string) error {
    if !valido.MatchString(s) {
        return errors.New("minusculas, numeros y guiones; de 3 a 63")
    }
    return nil
}
```

## 2. las reglas de verdad, no de juguete

```console
var valido = regexp.MustCompile(
    `^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

// minusculas, numeros y guiones
// de 3 a 63, y sin empezar ni acabar por guion
```

## 3. donde va lo que pasa despues

```console
dialog.ShowConfirm("Borrar", "¿Seguro?",
    func(si bool) {          // <- AQUI va lo de despues
        if !si { return }
        borrar()
    }, w)

// esta linea se ejecuta YA, sin esperar
```

## 4. el intro que envia, sin inventar nada

```console
nombre.OnSubmitted = func(s string) {
    if nombre.Validate() == nil {
        crear()
    }
}
```

## 5. el material del curso

```console
Fyne v2.8.1  ·  go1.26.8
nombre invalido a proposito: Fotos_2026
corrida: 2026-09-03 · salida-real.txt
```
