# Salida ejecutada — vídeo «go test -fuzz contra mi propio parser»

> Corrida del 2026-09-10 en `golang:1.26-bookworm` (go1.26.8), 24 trabajadores.
> Reproducible con `docker run --rm -v "$PWD/lab:/lab" -v "$PWD/demo.sh:/demo.sh" -w /lab golang:1.26-bookworm bash /demo.sh`

## ⛔ ÉTICO

Se fuzzea **código propio**, que es exactamente para lo que está la herramienta.

## El parser, y el fallo plantado

Lee el formato más común que existe: `[tipo:1][longitud:2][contenido:longitud]`. Así hablan medio
protocolo de red y medio formato de fichero.

```go
n := binary.BigEndian.Uint16(datos[1:3])
// falta:  if int(n) > len(datos)-3 { return nil, false }
contenido := datos[3 : 3+n]
```

**Se fía de la longitud que viene DENTRO del mensaje.** Si alguien dice «mido 500» y manda 3
bytes, el corte se sale. Así nacen media docena de vulnerabilidades al año en parsers de verdad.

---

## 1 · Las pruebas normales pasan. Las cuatro.

```
   --- PASS: TestParseaMensajeNormal (0.00s)
   --- PASS: TestParseaMensajeVacio (0.00s)
   --- PASS: TestParseaDemasiadoCorto (0.00s)
   --- PASS: TestIdaYVuelta (0.00s)
   PASS
   ok  	parser	0.002s
```

**El parser está probado.** Hay una prueba del caso normal, una del vacío, una del mensaje
demasiado corto y una de ida y vuelta. Es lo que escribiría cualquiera, y es más de lo que
escribe mucha gente. Y tiene un agujero.

## 2 · 🥇 El fuzz lo encuentra en 0,08 segundos

```
$ go test -fuzz='^FuzzParsea$' -fuzztime=60s

fuzz: elapsed: 0s, gathering baseline coverage: 3/3 completed, now fuzzing with 24 workers
fuzz: minimizing 44-byte failing input file
--- FAIL: FuzzParsea (0.08s)
    panic: runtime error: slice bounds out of range [:12339] with capacity 8
        parser.Parsea(...)
        	/lab/parser.go:32
```

El fuzz son **seis líneas**, y lo único que se le pide al parser es que no reviente. Devolver «no
se puede» es una respuesta correcta.

## 🥇 3 · La entrada que lo tumba: tres caracteres

```
   fuzz: minimizing 44-byte failing input file

   testdata/fuzz/FuzzParsea/c2501043394e49f2:
     go test fuzz v1
     []byte("000")
```

**Tres ceros de texto.** El fuzz encontró una entrada de 44 bytes que fallaba y la fue reduciendo
hasta dejarla en tres caracteres, que es lo mínimo que reproduce el fallo.

La cuenta, que es lo bonito:

```
   "000"  =  0x30 0x30 0x30
             ├──┘  └──┬──┘
           tipo=48   longitud = 0x3030 = 12.336

   el parser intenta cortar hasta el byte 12.339
   y el búfer tiene capacidad 8
```

La entrada más inofensiva que te puedas imaginar —tres ceros— hace que el parser intente leer
doce mil bytes de un búfer de ocho.

## 4 · Y ese fichero ya es una prueba, para siempre

```
$ go test -run FuzzParsea        (sin -fuzz)
--- FAIL: FuzzParsea (0.00s)
    --- FAIL: FuzzParsea/c2501043394e49f2 (0.00s)
panic: runtime error: slice bounds out of range [:12339] with capacity 8
```

Go **guardó la entrada en `testdata/`**, dentro del repositorio, y a partir de ahora `go test` a
secas la ejecuta como una prueba normal. No hace falta volver a fuzzear para que el fallo no
vuelva: ya está capturado.

## 5 · La versión corregida aguanta 139 millones de ejecuciones

```
$ go test -fuzz='^FuzzParseaBien$' -fuzztime=60s

fuzz: elapsed: 1m0s, execs: 138.941.781 (2.341.334/sec), new interesting: 2 (total: 3)
PASS
ok  	parser	60.109s
```

**138.941.781 ejecuciones en un minuto**, a 2,34 millones por segundo, con 24 trabajadores. Ni un
fallo.

⚠️ Y hay que decirlo con cuidado: eso **no demuestra que sea correcta**. Demuestra que en un
minuto no encontró cómo hacerla reventar.

## 6 · ⛔ ¿Y sin corpus de semillas?

Lo repetí quitando las tres semillas del fuzz:

```
   semillas que quedan: 0

   fuzz: elapsed: 0s, execs: 14.003 (78.298/sec), new interesting: 1 (total: 2)
   panic: runtime error: slice bounds out of range [:12339] with capacity 8
```

**Lo encontró igual, en 14.003 ejecuciones.** O sea que para este fallo las semillas no hicieron
falta.

⚠️ Y eso va en contra de lo que yo esperaba enseñar. La razón es que este fallo es muy
superficial: casi cualquier secuencia de tres bytes lo dispara. Las semillas importan cuando el
fallo está detrás de una estructura que hay que acertar —una cabecera válida, una suma de
comprobación, un campo con un valor concreto— porque ahí el azar no llega solo. Con este, sí
llegaba.

---

## ⛔ Mi fallo midiendo esto

La primera corrida no ejecutó el fuzz principal **en absoluto**:

```
testing: will not fuzz, -fuzz matches more than one fuzz test: [FuzzParsea FuzzParseaBien]
FAIL
```

`-fuzz` recibe una expresión regular, y `FuzzParsea` casa también con `FuzzParseaBien`. Se
arregla anclándola: `-fuzz='^FuzzParsea$'`.

Lo grave es lo de siempre: los bloques siguientes se ejecutaron igual, sin binario que mirar, y
dieron una salida perfectamente formateada — `testdata` vacío, y un `ok` tranquilizador en el
paso 4. **Ahora el banco para si detecta ese mensaje.**

## ⛔ Lo que el fuzzing NO encuentra

- **Los fallos de lógica que no provocan pánico.** El parser corregido acepta datos de sobra al
  final del mensaje y los ignora en silencio. En un protocolo de red eso es un agujero conocido
  —el contrabando de peticiones— y el fuzz no lo va a encontrar nunca, porque nadie revienta.
- **Lo que no le enseñas a comprobar.** El fuzz solo sabe que algo va mal si el programa entra en
  pánico o si tú escribes una comprobación. Si quieres que detecte que el resultado es
  *incorrecto*, tienes que decirle qué es correcto.
- **Los fallos detrás de estructura**, si no le das semillas que la tengan.
- **Y no demuestra ausencia de fallos.** 139 millones de ejecuciones sin fallar es una buena
  señal, no una demostración.
