# Pruebas — «Vengo de TypeScript, ¿qué me va a doler en Go?» (vídeo 178)

> **Todo lo que sale en pantalla se ejecutó el 2026-09-03**, dentro de los contenedores oficiales
> `node:22-alpine` (Node v22.23.2 + TypeScript 5.9.2) y `golang:alpine` (go1.26.5 linux/amd64).
> Salida completa y sin retocar en [`salida-real.txt`](salida-real.txt).
>
> ⚠️ Nada de esto se compila en el host: App Control bloquea los binarios sin firmar en esta
> máquina, así que ambos lados corren en contenedor.

## Cómo se reproduce

```sh
cd canales/golandia/recursos/pruebas-go-desde-ts

MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/pruebas" -w //pruebas \
    node:22-alpine sh -c "npm i -g typescript@5.9.2 >/dev/null 2>&1; sh //pruebas/correr_ts.sh"

MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/pruebas" -w //pruebas \
    golang:alpine sh //pruebas/correr_go.sh
```

`tsc` corre siempre con `--strict --noEmit`: se compara Go contra el TypeScript **bien
configurado**, no contra el permisivo. Es lo justo, y además es el que usa el público objetivo.

---

## 1 · La excepción que el tipo no cuenta

`cobrar()` lanza. Su firma es `(saldo: number, importe: number) => number` y **no dice nada** de
que pueda explotar. Quien la llama no se entera hasta ejecución.

```
== compila? ==
   tsc: SIN ERRORES (rc=0)
== y al ejecutarlo? ==
/pruebas/ts/t1_excepciones.ts:7
    throw new Error("saldo insuficiente");
```

El mismo descuido en Go **no llega a compilar**, porque el fallo es el segundo valor devuelto:

```
go/g1_errores.go:22:14: assignment mismatch: 1 variable but cobrar returns 2 values
```

Y escrito bien, el fallo es una rama, no un salto:

```
no se pudo cobrar: saldo insuficiente
```

👉 **El hallazgo:** no es que Go te obligue a *tratar* el error —`_` sigue existiendo—, es que te
obliga a **verlo escrito**. En TypeScript el camino del fallo es invisible en la firma.

---

## 2 · Uniones de literales: esto se pierde, y se echa de menos

TypeScript conoce los valores válidos y caza la letra cambiada **en compilación**:

```
ts/t2_uniones.ts(9,23): error TS2345: Argument of type '"activa"'
                        is not assignable to parameter of type 'Estado'.
```

Go no tiene equivalente. `type Estado string` **no restringe nada**:

```
el cliente esta activo
el cliente esta activa
el cliente esta cualquier cosa que se me ocurra
```

👉 **El hallazgo:** las tres líneas compilan y ninguna se queja jamás. El apaño idiomático
—constantes más una función que valide— mueve la comprobación a **ejecución**. Es la pérdida más
real del cruce, y conviene decirlo sin adornos.

---

## 3 · El valor cero: en Go una variable sin inicializar **ya vale algo**

TypeScript no tiene valor cero: tiene «todavía no vale nada», y lo comprueba.

```
ts/t3_valor_cero.ts(3,26): error TS2454: Variable 'intentos' is used before being assigned.
```

En Go eso es legal, y vale cero:

```
sin inicializar -> intentos=0 nombre="" lista=[] (nil? true)
```

🥇 **Y aquí está el daño de verdad, que no es la variable suelta sino el JSON:**

```
sin el campo -> descuento=0
con cero     -> descuento=0
son iguales? true   <- y son dos cosas DISTINTAS
```

Un campo **ausente** y un campo **con cero explícito** llegan idénticos a la estructura. «No me
mandaron descuento» y «me mandaron descuento cero» son decisiones de negocio distintas y el
lenguaje las aplasta. La salida es el puntero, que sí distingue:

```
con puntero: sin el campo -> <nil> (nil? true)
con puntero: con cero     -> 0 (nil? false)
```

---

## 4 · `nil` no es `null`, y no se porta igual según el tipo

TypeScript mete `null` en el tipo y te obliga a descartarlo, y además tiene `?.`:

```
ts/t4_null.ts(9,13): error TS18047: 'c' is possibly 'null'.
```

Go no tiene `?.` ni nada parecido. Y lo que muerde no es eso: es que **el mismo `nil` se comporta
de tres maneras distintas** según de qué sea:

```
leer de un mapa nil:         0 (y no revienta)
escribir en un mapa nil:     PANIC: assignment to entry in nil map
append a un slice nil:       [1 2 3] (largo 3)
leer un puntero nil:         PANIC: runtime error: invalid memory address or nil pointer dereference
```

👉 **El hallazgo:** un mapa `nil` se **lee** sin problema y revienta al **escribir**; un slice
`nil` acepta `append` como si estuviera inicializado. La regla «comprueba si es nil» que traes de
TypeScript no basta, porque aquí la respuesta depende del tipo.

---

## Lo que este material NO demuestra

- **No se midió rendimiento de nada.** No hay comparación de velocidad entre Node y Go, y en el
  vídeo no se afirma ninguna.
- No es una comparación de ecosistemas, ni de tiempos de compilación, ni de tamaño de binario:
  es solo el **cruce mental** de un lenguaje al otro.
- El TypeScript de referencia es el de `--strict`. Con la configuración permisiva, tres de las
  cuatro comprobaciones de arriba no saltarían — y eso juega a favor de Go, así que se dice.
