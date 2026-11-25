# Tu ventana se congela al llamar a la red: el patrón que lo arregla — lo que sale en pantalla

Comandos y salidas **tal y como salen en las terminales del vídeo**, en el mismo orden.
Se sacaron automáticamente del guion del vídeo; no se escribió nada de memoria.
Las líneas con `$` son lo que tecleas; el resto es la salida que se midió al grabarlo
(tus hashes, tiempos e IDs van a ser otros).

## 1. el bucle que no vuelve, por dentro

```console
para siempre {
    mirar si hay eventos     (raton, teclado, tamano)
    atenderlos               (tus manejadores)
    redibujar la ventana
}
```

## 2. el patron entero

```console
go func() {                          // 1 · la gorrutina
    r, err := c.ListBuckets(ctx, ...) // 2 · la red, FUERA
    fyne.Do(func() {                  // 3 · y SOLO el retoque
        // rellenar la lista, quitar el progreso
    })
}()
```

## 3. el cliente, con las dos lineas del episodio 1

```console
s3.New(s3.Options{
    Region:       "us-east-1",
    Credentials:  credentials.NewStaticCredentialsProvider(...),
    BaseEndpoint: aws.String("http://nubecita:4566"),
    UsePathStyle: true,   // el fallo numero uno
})
```

## 4. el plazo, que ya conoces

```console
ctx, parar := context.WithTimeout(
        context.Background(), 5*time.Second)
defer parar()

r, err := c.ListBuckets(ctx, &s3.ListBucketsInput{})
```

## 5. esconder el progreso, pase lo que pase

```console
fyne.Do(func() {
    progreso.Hide()          // lo primero, pase lo que pase
    if err != nil {
        estado.SetText("error: " + err.Error())
        return
    }
    // ... rellenar la lista
})
```

## 6. el material del curso

```console
ventana sola      32.343.016 B
con el kit de AWS 40.651.320 B
Fyne 2.8.1 · go1.26.8 · corrida 2026-09-03
```
