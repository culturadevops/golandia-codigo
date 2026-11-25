# SALIDA — curso «Go para quien ya sabe Python»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-09-22 16:39:12Z`
- go version go1.26.5 windows/amd64
- Python: `3.11.9`
- sistema: `Windows 10`

## Leccion 01 — De script a binario

`go run ./01_binario` · codigo de salida `0`

```
== 1.1 · Esto ya esta compilado ==
  go         go1.26.5
  sistema    windows/amd64
  binario    C:\Users\tu-usuario\AppData\Local\go-build\24\24e36be41877e16ce6f5b369677976d161a23f86cb49bde993bd8f0edeb33aff-d\01_binario.exe

  En Python, `python script.py` necesita un interprete instalado,
  la version correcta y las dependencias. En Go, `go build` deja UN
  fichero que se copia a la maquina y se ejecuta. Nada mas.

== 1.2 · Compilar para otra maquina sin salir de la tuya ==
  GOOS=linux GOARCH=amd64 go build -o revisar .
  El tamano y el arranque medidos estan en la leccion 12.
```

Y lo mismo en Python:

```
== en Python ==
  interprete  C:\Users\tu-usuario\AppData\Local\Programs\Python\Python311\python.exe
  version     3.11.9
  y el script solo corre donde haya ESTE interprete instalado
```

## Leccion 02 — Tipos y el valor cero

`go run ./02_valor_cero` · codigo de salida `0`

```
== 2.1 · Toda variable nace con un valor ==
  int     -> 0
  string  -> ""
  bool    -> false
  slice   -> []  (es nil: true)
  No hay None para un int. Si no pones valor, vale su CERO.

== 2.2 · LA TRAMPA: el mapa no da KeyError, da el valor cero ==
  puertos["cache"]  -> 0
  En Python esto era un KeyError. Aqui devuelve 0, y 0 parece
  un puerto valido. No hay error: hay un dato falso.

== 2.3 · La forma de preguntar si ESTA: la coma ok ==
  cache no esta en el mapa   <- ok == false
  web -> 443, ok=true
```

Y lo mismo en Python:

```
== en Python ==
  puertos['cache'] -> KeyError: 'cache'
  puertos.get('cache') -> None
```

## Leccion 03 — Slices que comparten memoria

`go run ./03_slices` · codigo de salida `0`

```
== 3.1 · Un trozo de un slice es una VENTANA, no una copia ==
  hosts    = [CAMBIADO web-02 web-03 web-04]
  primeros = [CAMBIADO web-02]
  Cambiaste el trozo y cambio el original. En Python, hosts[:2]
  es una COPIA: ahi no pasaria.

== 3.2 · LA TRAMPA: append a veces pisa al vecino ==
  base = [a b X]   <- el append de x piso base[2]
  len(base)=3  cap(base)=10
  Habia capacidad libre, asi que append escribio encima en vez de
  crear un array nuevo. Si no hubiera habido capacidad, no pisa.
  El mismo codigo, dos comportamientos, segun la capacidad.

== 3.3 · Copiar de verdad ==
  orig=[web-01 web-02]  copia=[otro web-02]
```

Y lo mismo en Python:

```
== en Python ==
  hosts    = ['web-01', 'web-02', 'web-03', 'web-04']
  primeros = ['CAMBIADO', 'web-02']
  el corte hosts[:2] es una COPIA
```

## Leccion 04 — Mapas sin orden

`go run ./04_mapas` · codigo de salida `0`

```
== 4.1 · Las cinco primeras vueltas al MISMO mapa ==
  vuelta 1: web-03 db-01 cache-01 web-01 web-02
  vuelta 2: web-02 web-03 db-01 cache-01 web-01
  vuelta 3: cache-01 web-01 web-02 web-03 db-01
  vuelta 4: web-01 web-02 web-03 db-01 cache-01
  vuelta 5: web-02 web-03 db-01 cache-01 web-01
  A veces salen iguales varias veces seguidas. Esa es la trampa:
  PARECE estable, y no lo es.

== 4.2 · Mil vueltas: cuantos ordenes distintos salen ==
  en 1000 vueltas salieron 5 ordenes distintos
  Go baraja el orden A PROPOSITO para que nadie dependa de el.
  Desde Python 3.7, un dict mantiene el orden de insercion.
  Por eso un test que 'siempre pasa' en tu maquina falla un dia en CI.

== 4.3 · Si quieres orden, lo pides ==
  ordenadas: cache-01 db-01 web-01 web-02 web-03
```

Y lo mismo en Python:

```
== en Python ==
  vuelta 1: web-01 web-02 web-03 db-01 cache-01
  vuelta 2: web-01 web-02 web-03 db-01 cache-01
  vuelta 3: web-01 web-02 web-03 db-01 cache-01
  siempre el mismo orden: el de insercion
```

## Leccion 05 — Cadenas: bytes y letras

`go run ./05_cadenas` · codigo de salida `0`

```
== 5.1 · La misma palabra que en el curso de Python ==
  len("configuración")                  -> 14   <- BYTES
  utf8.RuneCountInString   -> 13   <- letras
  En Python len() devuelve las letras. En Go, 14: la o con tilde son 2 bytes.

== 5.2 · LA TRAMPA: indexar da un byte, no una letra ==
  s[11] -> 195 ("Ã")   <- el primer byte de la o con tilde: medio caracter

== 5.3 · range recorre letras (runas), no bytes ==
  posicion de byte 0 -> 'a'
  posicion de byte 1 -> 'ñ'
  posicion de byte 3 -> 'b'
  Fijate: salta del 1 al 3. La ene ocupa dos bytes.
```

Y lo mismo en Python:

```
== en Python ==
  len('configuración') -> 13   <- letras
  len(s.encode('utf-8')) -> 14   <- bytes
```

## Leccion 06 — Errores como valores

`go run ./06_errores` · codigo de salida `0`

```
== 6.1 · El error viene en el return ==
  leerPuerto("443")  -> 443
  leerPuerto("cuatro") -> error: leyendo "cuatro": puerto invalido

== 6.2 · Envolver y preguntar (el `from e` de Python) ==
  errors.Is(err, ErrPuertoInvalido) -> true
  %w envuelve el error original: se puede preguntar por el.

== 6.3 · LA TRAMPA: el error que tiras a la basura ==
  f, _ := os.Open(...)  ->  f = <nil>
  El _ descarta el error SIN avisar. Tu programa sigue con un
  fichero nil y revienta mas tarde, lejos de la causa.
  En Python la excepcion te paraba AQUI. En Go, solo si la miras.
```

Y lo mismo en Python:

```
== en Python ==
  int('cuatro') -> ValueError: invalid literal for int() with base 10: 'cuatro'
  la excepcion te para en el sitio, aunque no la hayas pedido
```

## Leccion 07 — El nil que no es nil

`go run ./07_nil` · codigo de salida `0`

```
== 7.1 · Una validacion que va BIEN ==
  err == nil  ->  false
  err         ->  <nil>
  La validacion fue bien, y aun asi err NO es nil.

== 7.2 · Por que ==
  Una interfaz (error lo es) guarda DOS cosas: el tipo y el valor.
  aqui guarda: tipo *main.MiError, valor nil
  Solo es nil si las DOS son nil. El tipo no lo es.
  En Python, None es None y ya: esto no tiene equivalente.

== 7.3 · El arreglo: devolver nil a secas ==
  validarBien(true) == nil  ->  true
  errors.As(err, &me)       ->  true (puerto invalido)
```

Y lo mismo en Python:

```
== en Python ==
  err is None -> True   (None es None, sin tipo escondido)
```

## Leccion 08 — Structs y el receptor

`go run ./08_structs` · codigo de salida `0`

```
== 8.1 · Un struct es lo mas parecido a tu clase de Python ==
  {Nombre:api Replicas:1}

== 8.2 · LA TRAMPA: el metodo que no modifica ==
  tras s.EscalarMal(5)  -> Replicas = 1
  Compila, no da error, y no cambio nada: el metodo recibio una copia.
  En Python `self` es siempre el objeto de verdad. Aqui, depende del *.
  tras s.Escalar(5)     -> Replicas = 5
  La unica diferencia es el * en (s *Servidor).

== 8.3 · Asignar un struct SI copia (al reves que una lista de Python) ==
  a.Replicas=2  b.Replicas=9
  En Python, b = a con un objeto seria otro nombre para lo mismo.
```

Y lo mismo en Python:

```
== en Python ==
  tras s.escalar(5) -> replicas = 5   (self es el objeto de verdad)
  a.replicas=9  b.replicas=9   (b = a NO copia)
```

## Leccion 09 — defer

`go run ./09_defer` · codigo de salida `0`

```
== 9.1 · defer corre al salir de la FUNCION ==
  empieza 1
  termina 1
  defer de 1 (al SALIR de la funcion)

== 9.2 · Varios defer: al reves de como se escriben ==
  (fin del bucle)
  defer 3
  defer 2
  defer 1

== 9.3 · LA TRAMPA: defer dentro de un bucle no cierra en cada vuelta ==
  al acabar el bucle siguen abiertos: 5 ficheros
  Con 5 da igual. Con 50.000 lineas de log, te quedas sin descriptores.
  En Python, `with open(...)` cierra al salir del BLOQUE. Aqui, de la funcion.
```

Y lo mismo en Python:

```
== en Python ==
  vuelta 0: cerrado al salir del with -> True
  vuelta 1: cerrado al salir del with -> True
  vuelta 2: cerrado al salir del with -> True
```

## Leccion 10 — Goroutines que no esperas

`go run ./10_goroutines` · codigo de salida `0`

```
== 10.1 · LA TRAMPA 1: main no espera a nadie ==
  goroutines terminadas al llegar aqui: 0 de 3
  Si main acaba, el programa acaba, y lo que estaba a medias se pierde.

== 10.2 · WaitGroup: esperar de verdad ==
  con wg.Wait(): 3 de 3

== 10.3 · LA TRAMPA 2: dos goroutines tocando el mismo contador ==
  intento 1: esperaba 100000, salio 86603
  intento 2: esperaba 100000, salio 83967
  intento 3: esperaba 100000, salio 87714
  Ni un error. Solo un numero que no cuadra, y distinto cada vez.
  `go run -race` lo detecta. Y un sync.Mutex lo arregla (10.2).

  Nota: el fallo clasico de la variable del for capturada en la
  goroutine YA NO existe desde Go 1.22. Si ves `v := v` en codigo
  viejo, es por eso.
```

## Leccion 11 — Treinta chequeos en paralelo

`go run ./11_paralelo` · codigo de salida `0`

```
== 30 servicios, cada uno tarda 200ms en contestar ==

  uno detras de otro          6015 ms
  todos a la vez               200 ms
  como mucho 5 a la vez       1203 ms

  todos a la vez es 30 veces mas rapido que en fila
  El limite existe porque 'todos a la vez' contra 3.000 servicios
  es un ataque, no un chequeo. Un canal con hueco hace de semaforo.
```

Y lo mismo en Python:

```
== en Python ==
  uno detras de otro          6017 ms
  hilos, todos a la vez        207 ms
  Para ESPERAR (red, disco), los hilos de Python tambien sirven.
```

## Leccion 12 — El proyecto, en Go

`go run ./12_proyecto` · codigo de salida `1`

```
Revisados 4 servicios (umbral 80%)
  CAIDO      web-02
  SATURADO   web-01  cpu 91%
  SATURADO   cache-01  cpu 88%
time=2026-09-22T11:39:27.981-05:00 level=WARN msg="hay problemas" caidos=1 saturados=2
exit status 1
```

```
== Medida: el mismo programa, en Go y en Python ==
  binario de Go                    3.2 MB
  binario de Go (-ldflags=-s -w)   2.2 MB
  script de Python                 3.7 KB  + un interprete instalado
  arranque de Go     (mediana de 20)    18.7 ms
  arranque de Python (mediana de 20)    39.2 ms
  -> Go arranca 2.1 veces antes
```
