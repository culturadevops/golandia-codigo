# Salida ejecutada — vídeo «Por qué el malware se escribe en Go (y cómo se detecta)»

> Corrida del 2026-09-10 en `golang:1.26-bookworm` + gcc (Debian 12.2.0). Reproducible con
> `docker build -t banco-detectar . && docker run --rm -v "$PWD:/lab" -w /lab banco-detectar bash /lab/demo.sh`

## ⛔ DEFENSIVO, y se dice en el vídeo

Aquí **no se escribe nada ofensivo ni se analiza ninguna muestra real**. Lo que se analiza son mis
propios binarios, compilados en este mismo banco. La pregunta es de **formato de ejecutable**, no
de lo que hace el programa.

---

# A · POR QUÉ lo eligen

## A1 · Compilación cruzada, desde una sola máquina

```
   plataformas compiladas: 10 · fallidas: 0
   tiempo total: 32.162 ms
   una orden por plataforma. Ni un compilador cruzado instalado.

     /tmp/h-darwin-amd64     5.604.208 bytes
     /tmp/h-darwin-arm64     5.263.218
     /tmp/h-freebsd-amd64    5.380.801
     /tmp/h-linux-386        5.400.731
     /tmp/h-linux-amd64      5.435.660
     /tmp/h-linux-arm        5.303.153
     /tmp/h-linux-arm64      5.122.578
     /tmp/h-windows-386      5.554.176
     /tmp/h-windows-amd64    5.563.392
     /tmp/h-windows-arm64    5.111.296
```

**Diez plataformas en treinta y dos segundos, desde un Linux, cambiando dos variables.** Windows
en tres arquitecturas, macOS en dos, FreeBSD. Sin instalar un solo compilador cruzado.

Esa es la primera respuesta a la pregunta del título, y no tiene nada de siniestro: es la misma
razón por la que lo usan Docker y Kubernetes.

## A2 · Qué necesita la máquina de destino

```
   ldd sobre el binario de Go:  not a dynamic executable

   ldd sobre el binario de C:
     linux-vdso.so.1
     libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
     /lib64/ld-linux-x86-64.so.2
```

**Nada.** El de Go no pide una sola librería. El de C necesita `libc` y el enlazador dinámico, y
si la versión no encaja, no arranca.

Para quien despliega una herramienta legítima, eso es comodidad. Para quien reparte algo que no
debería, es lo mismo: un archivo que corre donde caiga.

## A3 · El precio, que también es una señal

```
   Go (hola mundo):  5.435.660 bytes
   C  (hola mundo):      16.016 bytes
```

**339 veces más grande.** Y ese tamaño no es solo un coste: es en sí mismo algo llamativo. Un
ejecutable de cinco megas que dice hacer una cosa pequeña ya es raro.

---

# B · 🥇 CÓMO SE DETECTA

Siete señales independientes. La misma pregunta a tres binarios.

| señal | normal | `-ldflags "-s -w"` | un binario de C |
|---|---|---|---|
| 1. la cadena de versión `go1.x` | **SÍ** · go1.26.8 | **SÍ** · go1.26.8 | no |
| 2. información de compilación legible | **SÍ** | **SÍ** | no |
| 3. la sección `.go.buildinfo` | **SÍ** | **SÍ** | no |
| 4. secciones propias del lenguaje | **SÍ** · 8 | **SÍ** · 8 | no |
| 5. mensajes del runtime | **SÍ** · 4.185 | **SÍ** · 2.373 | no |
| 6. nombres de funciones | **SÍ** | **SÍ** | no |
| 7. nombres de paquetes de la biblioteca | **SÍ** · 348 | **SÍ** · 147 | no |
| | **7 de 7** | **7 de 7** | **0 de 7** |

## 🥇 «Despojar» el binario no baja ni una señal

**Siete de siete en los dos.** Lo único que cambia son las cantidades: las cadenas del runtime
bajan de 4.185 a 2.373, y los nombres de paquetes de 348 a 147. Siguen siendo miles y cientos.

Las ocho secciones que lleva dentro, con nombre propio:

```
   .go.buildid  .go.buildinfo  .go.fipsinfo  .go.module
   .gopclntab   .itablink      .noptrdata    .typelink
```

Ninguna de esas existe en un ejecutable que no sea de Go. No hace falta ni mirar el contenido:
con la lista de secciones ya está identificado.

## 🥇 Y el remate: el binario te dice que intentaron despojarlo

```
$ go version -m binario-despojado
/tmp/go-strip: go1.26.8
	path	ejemplo/prog
	mod	ejemplo/prog	(devel)
	build	-buildmode=exe
	build	-compiler=gc
	build	-ldflags="-s -w"        <- AQUÍ
	build	CGO_ENABLED=1
```

**La información de compilación registra las banderas con las que se compiló, incluidas las que
se usaron para intentar ocultar cosas.** El binario lleva dentro la confesión.

## El contraste con C, dicho honestamente

⚠️ Que el binario de C saque 0 de 7 **no significa que sea más discreto**: significa que esas
siete señales son específicas de Go, y las escogí precisamente por eso. La comparación justa es
otra: *qué tienes que hacer para saber qué lleva dentro*.

```
   ¿dice qué librerías usa?  las enlazadas, sí. Las VERSIONES no
   ¿dice el compilador?      SÍ · GCC: (Debian 12.2.0-14+deb12u1)
   ¿dice sus dependencias?   no hay lista: hay que deducirlas
```

Con el de C **deduces**: miras qué símbolos importa, qué cadenas tiene, cómo se comporta. Con el
de Go **lees una lista**. Esa es la diferencia, y es enorme para quien tiene que analizar algo
deprisa.

Curiosamente, el de C también delata su compilador y hasta la versión de la distribución que lo
construyó. Nadie reparte un binario mudo del todo.

---

## ⛔ Los dos fallos de esta corrida, los dos míos

1. **Compilé desde el directorio equivocado.** `go: go.mod file not found`. Las diez plataformas
   «fallaron», los binarios no existían, y el bloque de detección dio un limpísimo **0 de 7 en
   todo**, incluidos los dos binarios de Go. Un resultado perfectamente formateado y sin ningún
   sentido. Es la segunda vez que me pasa lo mismo esta semana.
2. **Y por eso el banco ahora PARA.** Antes de medir una sola señal, comprueba que los tres
   binarios existen. Si falta uno, se detiene y lo dice, en vez de informar de un cero que
   parece un hallazgo.
3. Extra: al reescribir el guion desde Windows le metí finales de línea de Windows y `bash` no
   lo pudo ni leer (`set: -: invalid option`). Se arregla normalizando el fichero.

## ⛔ Lo que NO dice este banco

- **No analiza ninguna muestra real de malware**, ni falta que hace: lo que se mide es el formato
  del ejecutable, y eso es igual en un binario legítimo y en uno que no lo sea.
- **No dice que Go sea inseguro ni que esté mal elegirlo.** Las razones por las que lo eligen los
  atacantes son exactamente las mismas por las que lo eligen Docker y Kubernetes.
- **No entra en ofuscadores**, que existen, borran parte de esto y tienen su propio precio.
- **No cubre Windows ni macOS como formatos**: todo está medido sobre ELF, el formato de Linux.
  Las secciones se llaman distinto en PE y en Mach-O; las señales equivalentes están, pero no se
  han medido aquí.
