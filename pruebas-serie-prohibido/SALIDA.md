# Salida ejecutada — serie «Lo que Go NO te deja»

> Reejecutado el 2026-09-05 en **go1.26.5 windows/amd64** (antes: golang:1.23-alpine, 2026-08-01).
> Las cuatro piezas verificadas REPRODUCEN. Salida literal, pegada de la consola.

## 01_coma_flotante.go
```
--- Go tiene EXACTAMENTE el mismo problema ---
0.1 + 0.2      = 0.3
¿es igual a 0.3? true

10 veces 0,10 € = 0.9999999999999999
¿es 1 euro?      false

19,99 x 3      = 59.96999999999999886313

--- en céntimos (enteros) ---
10 veces 10 céntimos = 100 céntimos = 1 €
¿exacto? true
```

## 02_nil_mapas_slices.go
```
=== 1 · nil que no es nil ===
  err == algo fallo
  err == nil ? false   <- todo el mundo espera true

=== 2 · los mapas se recorren en orden ALEATORIO a proposito ===
  vuelta 1: a b c d e 
  vuelta 2: a b c d e 
  vuelta 3: c d e a b 

=== 3 · dos slices comparten memoria y se pisan ===
  original : [1 2 3 4 5]
  trozo    : [2 3]  (len 2 cap 4 )
  tras append(trozo, 99):
  trozo    : [2 3 99]
  original : [1 2 3 99 5]  <- CAMBIO SOLO, sin tocarlo
```

⚠️ Nota honesta sobre los mapas: el orden es ALEATORIO, no "siempre distinto". Con 5 claves
dos vueltas seguidas pueden coincidir (pasó en esta corrida: vueltas 1 y 2 iguales, la 3
distinta). El vídeo tiene que decirlo así o el titular no se sostiene.

## 03_mapas_orden.go — cuánto de aleatorio es (2026-09-05, go1.26.5)
```
=== tres vueltas seguidas, mapa de 5 claves ===
  vuelta 1: a b c d e
  vuelta 2: a b c d e
  vuelta 3: c d e a b

=== 1000 vueltas sobre el MISMO mapa, sin tocarlo ===
  ordenes distintos vistos : 5  (de 120 posibles)

  cuantas veces empezo por cada clave:
    a : 498 de 1000
    b : 112 de 1000
    c : 136 de 1000
    d : 141 de 1000
    e : 113 de 1000

  vueltas seguidas que salieron IGUALES: 299 de 1000
  -> es ALEATORIO, no 'siempre distinto'. El titular tiene que decir eso.
```

## 04_mapas_tamano.go — ¿baraja o punto de partida? (2026-09-05, go1.26.5)
```
claves   ordenes      posibles       % que son rotacion del 1º
5        5            120            100%
8        8            40320          100%
20       20           muchos         100%
100      100          muchos         100%

Si el % de rotaciones es alto, el recorrido NO se baraja:
es la misma vuelta empezando por otro sitio.
```

### 🥇 Lo que sale de aquí, y cambia el vídeo 201

Go **no baraja** el mapa: aleatoriza **por dónde empiezas**. Medido en go1.26.5 sobre
mapas de 5, 8, 20 y 100 claves insertadas en orden y sin borrados, **el 100 % de las
vueltas son rotaciones de la misma secuencia**, y salen exactamente `n` órdenes distintos
de los `n!` posibles.

⚠️ Condiciones de la medida, que hay que decir en el vídeo: inserción secuencial, sin
borrados y sin crecimiento del mapa a mitad. No se ha medido qué pasa con borrados.

⛔ Por eso el guion NO puede decir «cada vuelta sale distinta»: 343 de 1.000 vueltas
seguidas salieron IGUALES, y una clave de cinco arranca el 51 % de las veces.

## 05_slices_intermitente.go — por qué el pisotón es INTERMITENTE (2026-09-05, go1.26.5)
```
=== el MISMO codigo, dos entradas distintas ===

  sobra capacidad        original=[1 2 3 4 5]  trozo=[2 3]  len=2 cap=4
                         tras append(trozo,99): original=[1 2 3 99 5]  -> PISADO

  capacidad justa        original=[1 2 3]  trozo=[2 3]  len=2 cap=2
                         tras append(trozo,99): original=[1 2 3]  -> intacto

=== y lo mismo, barriendo tamanos ===
  len(original)=3  cap(trozo)=2  ->  intacto
  len(original)=4  cap(trozo)=3  ->  PISADO
  len(original)=5  cap(trozo)=4  ->  PISADO
  len(original)=6  cap(trozo)=5  ->  PISADO
  len(original)=7  cap(trozo)=6  ->  PISADO
  len(original)=8  cap(trozo)=7  ->  PISADO

El fallo depende del TAMANO DE LA ENTRADA. Por eso llega a produccion.
```

### 🥇 Lo que sale de aquí, y es el vídeo 202

El mismo código corrompe o no **según el tamaño de la entrada**: con 3 elementos el
original queda intacto (no sobra capacidad, `append` copia a otro array) y a partir de 4
lo pisa. Un fallo que siempre pasa se encuentra en la primera prueba; uno que depende del
tamaño de la entrada llega a producción.

## 06_bucle_for.go — el bug que Go arregló en 1.22 (2026-09-05)

El MISMO fichero, dos contenedores. Esto es el vídeo 210 entero.

### golang:1.21-alpine
```
go version go1.21.13 linux/amd64
=== 1 · goroutines que capturan la variable del bucle ===
  lo que vieron las 5 goroutines: [5 5 5 5 5]
  esperado (1.22+): [0 1 2 3 4]
  en 1.21 y antes : algo como [5 5 5 5 5] — todas comparten LA MISMA variable

=== 2 · lo mismo sin concurrencia: punteros a la variable del bucle ===
  los tres punteros valen: 3 3 3 
  esperado (1.22+): 0 1 2
  en 1.21 y antes : 3 3 3 — los tres apuntan al MISMO sitio

=== 3 · y las direcciones lo demuestran ===
  puntero 0 -> 0xc0000120c8
  puntero 1 -> 0xc0000120c8
  puntero 2 -> 0xc0000120c8
  si las tres direcciones son iguales, es UNA sola variable compartida.
```

### golang:1.22-alpine
```
go version go1.22.12 linux/amd64
=== 1 · goroutines que capturan la variable del bucle ===
  lo que vieron las 5 goroutines: [0 1 2 3 4]
  esperado (1.22+): [0 1 2 3 4]
  en 1.21 y antes : algo como [5 5 5 5 5] — todas comparten LA MISMA variable

=== 2 · lo mismo sin concurrencia: punteros a la variable del bucle ===
  los tres punteros valen: 0 1 2 
  esperado (1.22+): 0 1 2
  en 1.21 y antes : 3 3 3 — los tres apuntan al MISMO sitio

=== 3 · y las direcciones lo demuestran ===
  puntero 0 -> 0xc000192068
  puntero 1 -> 0xc000192070
  puntero 2 -> 0xc000192078
  si las tres direcciones son iguales, es UNA sola variable compartida.
```

### 🥇 El plano que sostiene el vídeo

Las **direcciones**. En 1.21 los tres punteros dan `0xc00012c060`, la misma: hay UNA
variable compartida por todas las vueltas. En 1.22 dan tres direcciones distintas separadas
por 8 bytes: cada vuelta tiene la suya. No hace falta explicar nada más, se ve.

⚠️ Y el giro honrado para el cierre: si tu código dependía del comportamiento viejo, 1.22
te lo cambia. Un arreglo también rompe cosas.

### 🥈 Y cómo lo arreglaron sin romper la promesa de compatibilidad

Lo decide la línea `go` del `go.mod`, **no la versión del compilador**. Comprobado con el
MISMO toolchain go1.22.12, cambiando solo esa línea:
```
toolchain: go version go1.22.12 linux/amd64
--- go.mod dice: go 1.21 ---
  lo que vieron las 5 goroutines: [5 5 5 5 5]
  los tres punteros valen: 3 3 3 
--- go.mod dice: go 1.22 ---
  lo que vieron las 5 goroutines: [0 1 2 3 4]
  los tres punteros valen: 0 1 2 
```

Por eso pudieron cambiar la semántica de un bucle sin romper diez años de código: los
módulos que no suben su línea `go` siguen compilando con las reglas viejas.

## 07_defer_tags_any_fugas.go — cuatro piezas de golpe (2026-09-05, golang:alpine)

Cubre los vídeos **214** (defer en bucle), **213** (struct tags), **212** (any) y **209** (fugas).

### Lo que dice `go vet`
```
main.go:62:2: struct field tag `json:"nombre_completo` not compatible with reflect.StructTag.Get: bad syntax for struct tag value
main.go:63:2: struct field tag `json: "edad"` not compatible with reflect.StructTag.Get: bad syntax for struct tag value
```

### Y lo que imprime al correr
```
=== 1 · defer dentro de un bucle (video 214) ===
  con defer en el bucle : 200 ficheros abiertos A LA VEZ
  sacando el cuerpo     : 1 fichero abierto a la vez
  -> el defer del bucle NO cierra al acabar la vuelta: espera al final de la funcion

=== 2 · los struct tags son cadenas magicas (video 213) ===
  etiqueta bien escrita : {"nombre_completo":"Ana Lopez","edad":33}
  sintaxis rota         : {"NombreCompleto":"Ana Lopez","Edad":33}   <- go vet SI la caza
  errata en el nombre   : {"nombre_complet":"Ana Lopez","edda":33}   <- go vet NO puede cazarla
  -> ni error, ni panic, ni aviso. Solo un JSON distinto del que esperabas.

=== 3 · any mata el tipado (video 212) ===
  la asercion a int NO vale. Y esto COMPILA sin una sola queja.
  y la version sin comprobar revienta en EJECUCION:
    panic: interface conversion: interface {} is string, not int

=== 4 · goroutines que nunca mueren (video 209) ===
  al empezar            : 1 goroutines
  tras 1.000 llamadas   : 1001 goroutines
  y despues de un GC    : 1001 goroutines
  -> el recolector NO se las lleva: no son basura, estan ESPERANDO.
```

### 🥇 El matiz del vídeo 213, que es lo que lo hace distinto

`go vet` **sí** caza las erratas de SINTAXIS (comilla que falta, espacio de más). Lo que no
puede cazar es la etiqueta perfectamente escrita con el NOMBRE mal: ninguna herramienta sabe
cómo querías llamar tu campo. Sale un JSON válido con la clave equivocada y el consumidor
recibe un campo vacío sin que nadie diga nada.

## 08 / 09 / 10 — variable sin usar, imports circulares e init() (2026-09-05)

Tres paquetes de verdad, porque dos de ellos son **errores de compilación** y hay que
enseñar el mensaje literal. Cubren los vídeos **203**, **207** y **211**.
```
########## 203 · variable sin usar ##########
# command-line-arguments
./m.go:7:2: "os" imported and not used
./m.go:11:2: declared and not used: contador

########## 207 · imports circulares ##########
package ciclo/a
	imports ciclo/b from a.go
	imports ciclo/a from b.go: import cycle not allowed

########## 211 · init() magico ##########
  1. init() del paquete importado (driver)
  2. init() de aaa_primero.go
  3. init() de main.go
  3b. init() de zzz_ultimo.go
  4. y por fin main()

  Nadie llamo a ningun init(). Y no puedes llamarlos tu:
  descomenta la linea de abajo y no compila.
```

### 🥇 Lo que sale de aquí

**203** — el titular es literal: `declared and not used: contador` es un ERROR, no un aviso.
Y el import sin usar también. La válvula de escape (`_ = x`, `_ "paquete"`) está en
`08-variable-sin-usar/valvula.go.txt`.

**207** — `import cycle not allowed`, y el compilador imprime **la cadena entera** de
importaciones. Eso es lo que se enseña en pantalla.

**211** — el orden medido es: paquete importado → `aaa_primero.go` → `main.go` →
`zzz_ultimo.go` → `main()`. O sea **alfabético por nombre de fichero**, comprobado con
contraejemplo (`zzz_ultimo.go` corre después de `main.go`).

⚠️ MATIZ OBLIGATORIO EN EL GUION: que la herramienta `go` los ordene por nombre **no lo
garantiza la especificación** — dice que el orden de presentación al compilador es cosa de
la implementación. Así que el vídeo enseña el orden real Y dice que no te apoyes en él.
Las dos cosas, o el vídeo estaría enseñando a escribir código frágil.

### 10b — los dos límites de init() (2026-09-05)
```
### 1 · varias init() en el mismo fichero:
  init() numero 1 del mismo fichero
  init() numero 2 del mismo fichero
  init() numero 3 del mismo fichero
  main()

### 2 · intentar llamarla:
# initlimites
./main.go:13:2: undefined: init
```

Se pueden declarar **varias** `init()` en el mismo fichero y corren en orden de
declaración. Y `undefined: init` es el mejor plano del vídeo 211: la función existe, se
ejecuta sola, y su nombre **no está en el ámbito**. No la puedes llamar.

## 11-ternario — por qué la función genérica NO es un ternario (2026-09-05)
```
=== la funcion generica que todos proponen ===
      ...evaluando rama VERDADERA
      ...evaluando rama FALSA
  resultado: 1
  -> se evaluaron LAS DOS, aunque solo hacia falta una

=== y con un if de toda la vida ===
      ...evaluando rama VERDADERA
  resultado: 1
  -> solo se evaluo la que hacia falta

=== y por eso esto REVIENTA con la funcion, y no con el if ===
  Si(len(lista) > 0, lista[0], -1)
  panic: runtime error: index out of range [0] with length 0
  -> la rama 'lista[0]' se evaluo AUNQUE la condicion decia que no
```

### 🥇 Lo que sale de aquí, y es el final del vídeo 206

Todos los artículos de «cómo hacer un ternario en Go» proponen una función genérica. Y esa
función **no es un ternario**: un ternario de verdad es PEREZOSO, solo evalúa la rama que
gana. Una función de Go evalúa **siempre los dos argumentos** antes de entrar.

La demo que lo cierra:

```go
Si(len(lista) > 0, lista[0], -1)   // panic: index out of range [0] with length 0
```

La condición dice que la lista está vacía, y aun así `lista[0]` se evalúa. Con un `if` de
toda la vida, no. Así que la «solución» que circula por internet es peor que el problema.

## 12-sin-herencia — el empotrado NO es herencia (2026-09-06, go1.26.5)
```
=== 1 · el empotrado parece herencia, y no lo es ===
  h.Quien()   = la HIJA (ana)
  h.Saludar() = hola, soy la base (ana)

  Saludar() esta en Base y llama a Quien().
  Con herencia de verdad diria 'la HIJA'. Aqui dice 'la base'.
  -> Base no sabe que esta empotrada. Llama a SU propio Quien().

=== 2 · la composicion con interfaz SI hace lo que querias ===
  s.Saludar() = hola, soy la persona (ana)
  -> aqui no hay magia: le pasas QUIEN tiene que responder.

=== 3 · y el empotrado sirve para lo que SI sirve: reutilizar ===
  h.nombre = ana  <- el campo de Base, accesible sin escribir .Base
  h.Base.Quien() = la base (ana)  <- y el original sigue ahi
```

### 🥇 La demostración que decide el vídeo 205

`h.Quien()` dice **«la HIJA»**. Pero `h.Saludar()`, que está en `Base` y llama a `Quien()`,
dice **«la base»**. Con herencia de verdad diría «la HIJA»: eso es el despacho virtual.
En Go no lo hay — `Base` no sabe que está empotrada y llama a SU propio método.

O sea que el empotrado **se parece** a la herencia justo lo suficiente para engañarte, y
se comporta distinto en el único sitio donde la herencia aporta algo.

## 13-panic-recover — panic/recover NO son excepciones (2026-09-06, go1.26.5)

Cubre los vídeos **204** (Go no tiene excepciones) y **208** (panic en una librería).
```
=== 1 · recover() FUERA de un defer no captura nada ===
  recover() llamado directamente -> <nil>  (nil: no ha capturado nada)

=== 2 · una libreria que hace panic se lleva a QUIEN LA USA ===
  llamo a libreriaMaleducada(-1)...
  el que la usa tuvo que poner un recover para sobrevivir:
    recuperado: libreria: no acepto negativos
  -> y fijate: no se puede seguir donde estabas. La funcion entera se abandona.

=== 3 · la version educada: un error normal, y el flujo sigue ===
  error: no acepto negativos, me diste -1
  -> aqui SI sigo donde estaba, decido, y continuo

=== 4 · el limite que lo decide: panic en OTRA goroutine ===
  main tiene un recover puesto. Vamos a ver si sirve de algo.
panic: panic desde otra goroutine

goroutine 6 [running]:
main.main.func3()
	<repo>/pruebas-serie-prohibido/13-panic-recover/main.go:74 +0x25
created by main.main in goroutine 1
	<repo>/pruebas-serie-prohibido/13-panic-recover/main.go:73 +0x315
exit status 2
```

### 🥇 El límite que lo decide (vídeo 204)

Un `panic` en OTRA goroutine se lleva el proceso entero: `exit status 2`, y el `recover` de
`main` **ni se ejecuta**. En cualquier lenguaje con excepciones, un hilo que lanza no tumba
el proceso. Esa es la prueba de que no son excepciones, no una opinión.

Y los otros dos límites: `recover()` fuera de un `defer` devuelve `nil`, y tras recuperar
**no se sigue donde estabas** — la función entera se abandona.

### 🥈 Y para el vídeo 208

La librería maleducada obliga a quien la usa a poner un `recover` para sobrevivir. La
educada devuelve un error y el flujo continúa. Es exactamente lo que dice Effective Go:
*«real library functions should avoid panic»*.
