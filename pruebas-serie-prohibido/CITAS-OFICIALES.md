# 📚 Citas oficiales — serie 🚫 «Lo que Go NO te deja»

> Recogidas el **2026-09-05** de las fuentes oficiales de Go. Tres piezas de la serie no se podían
> producir sin ellas, porque su tesis es «el equipo de Go explicó por qué» y sin la cita literal
> eso es palabra del narrador, no un hecho.
>
> ⛔ Regla: se citan **en inglés y literales**, y en el vídeo se traducen en voz diciendo que es
> una traducción. Nunca se pone una traducción entre comillas como si fuera la cita.

---

## Vídeo 206 · `go-sin-ternario`

**Fuente:** [Go FAQ — «Why does Go not have the `?:` operator?»](https://go.dev/doc/faq#Does_Go_have_a_ternary_form)

> There is no ternary testing operation in Go. You may use the following to achieve the same
> result:
>
> ```go
> if expr {
>     n = trueVal
> } else {
>     n = falseVal
> }
> ```
>
> The reason `?:` is absent from Go is that the language's designers had seen the operation used
> too often to create impenetrably complex expressions. The `if-else` form, although longer, is
> unquestionably clearer. A language needs only one conditional control flow construct.

### ⛔ CORRECCIÓN AL TÍTULO REGISTRADO

El vídeo se registró como **«Go no tiene operador ternario, y Rob Pike explicó por qué»**, y eso
**no lo sostiene la fuente**: la FAQ dice *«the language's designers»*, en plural y sin nombres.
Atribuírselo a una persona concreta sería inventar la mitad de la cita.

👉 El título tiene que decir **«y el equipo de Go explicó por qué»**. Corregido antes de producir.

### La frase que sostiene el vídeo

*«…had seen the operation used too often to create impenetrably complex expressions.»*
No es «es feo»: es que lo habían visto usarse para escribir expresiones impenetrables. Y la
última frase es la tesis de diseño entera: **una sola construcción de control condicional**.

---

## Vídeo 204 · `go-sin-excepciones`

**Fuente:** [Go FAQ — «Why does Go not have exceptions?»](https://go.dev/doc/faq#exceptions)

> We believe that coupling exceptions to a control structure, as in the `try-catch-finally` idiom,
> results in convoluted code. It also tends to encourage programmers to label too many ordinary
> errors, such as failing to open a file, as exceptional.
>
> Go takes a different approach. For plain error handling, Go's multi-value returns make it easy to
> report an error without overloading the return value. A canonical error type, coupled with Go's
> other features, makes error handling pleasant but quite different from that in other languages.
>
> Go also has a couple of built-in functions to signal and recover from truly exceptional
> conditions. The recovery mechanism is executed only as part of a function's state being torn down
> after an error, which is sufficient to handle catastrophe but requires no extra control structures
> and, when used well, can result in clean error-handling code.

### La frase que sostiene el vídeo

*«…encourage programmers to label too many ordinary errors, such as failing to open a file, as
exceptional.»* Ahí está el vídeo: **no poder abrir un fichero no es excepcional**, es lo normal.
El argumento no es que las excepciones sean malas, es que se usan para lo que no es excepcional.

Y el tercer párrafo desactiva el malentendido de siempre: `panic`/`recover` **sí** existen, para
lo de verdad excepcional. No son excepciones porque no van atadas a una estructura de control.

---

## Vídeo 208 · `go-panic-en-libreria`

**Fuente:** [Effective Go — sección «Panic»](https://go.dev/doc/effective_go#panic)

> This is only an example but real library functions should avoid `panic`. If the problem can be
> masked or worked around, it's always better to let things continue to run rather than taking down
> the whole program. One possible counterexample is during initialization: if the library truly
> cannot set itself up, it might be reasonable to panic, so to speak.

### La frase que sostiene el vídeo

*«real library functions should avoid panic»* — y fíjate en el verbo: **should**, no *must*. Es
exactamente lo que dice el título de la pieza: está prohibido por convención, no por el
compilador.

### 🔗 Y el regalo: la excepción enlaza con el vídeo 211

*«One possible counterexample is during initialization…»* — o sea que la propia guía oficial
admite un caso: la inicialización del paquete, que es justo `init()`. Enlace natural con la pieza
13 de la serie.
