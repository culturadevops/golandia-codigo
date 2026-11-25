# Salida ejecutada — vídeo «Tu binario de Go delata tu código fuente»

> Corrida del 2026-09-10 en `golang:1.26-bookworm` (go1.26.8). Reproducible con
> `docker run --rm -v "$PWD/app:/casa/mi-proyecto-privado/app" -v "$PWD/demo.sh:/demo.sh" -w /casa/mi-proyecto-privado/app golang:1.26-bookworm bash /demo.sh`

**El programa** es una herramienta corriente: lee un YAML, usa `cobra`, y tiene un paquete
`interno/` con una constante que solo existe en el código privado. Se compila dentro de
`/casa/mi-proyecto-privado/app` para poder ver si esa ruta acaba dentro del binario.

---

## 🥇 LA TABLA: las cuatro variantes, la misma pregunta a cada una

| | normal | `-trimpath` | `-ldflags "-s -w"` | las dos |
|---|---|---|---|---|
| tamaño (bytes) | 4.211.650 | 4.206.520 | **2.883.849** | **2.879.753** |
| lista de dependencias con versión **y hash** | **SÍ, 3** | **SÍ, 3** | **SÍ, 3** | **SÍ, 3** |
| nombre del módulo interno | **SÍ** | **SÍ** | **SÍ** | **SÍ** |
| rutas absolutas de mi máquina | SÍ | **no** | SÍ | **no** |
| la constante privada | **SÍ** | **SÍ** | **SÍ** | **SÍ** |
| el nombre de mi función | **SÍ** | **SÍ** | **SÍ** | **SÍ** |
| mis nombres de fichero | **SÍ** | **SÍ** | **SÍ** | **SÍ** |
| se ve que es Go | **SÍ** | **SÍ** | **SÍ** | **SÍ** |

## 🥇 EL HALLAZGO

**De las siete cosas privadas, la combinación más agresiva quita exactamente UNA: la ruta
absoluta de mi máquina.** Todo lo demás sobrevive a las cuatro variantes.

Y lo más importante: **`-ldflags "-s -w"`, que es lo que todo el mundo llama «despojar el
binario», no quita ni una sola de las siete.** Lo que compra es tamaño: 1.327.801 bytes, un
31,5 %. No compra privacidad.

## La salida completa del más despojado de los cuatro

```
$ go version -m binario-despojado
/tmp/ambos: go1.26.8
	path	miempresa.interno/herramienta
	mod	miempresa.interno/herramienta	(devel)
	dep	github.com/spf13/cobra	v1.10.2	h1:DMTTonx5m65Ic0GOoRY2c16WCbHxOOw6xxezuLaBpcU=
	dep	github.com/spf13/pflag	v1.0.9	h1:9exaQaMOCwffKiiiYk6/BndUBv+iRViNW+4lEMi0PvY=
	dep	gopkg.in/yaml.v3	v3.0.1	h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
	build	-buildmode=exe
	build	-compiler=gc
	build	-trimpath=true
	build	DefaultGODEBUG=containermaxprocs=0,...
	build	CGO_ENABLED=1
	build	GOARCH=amd64
	build	GOOS=linux
	build	GOAMD64=v1
```

Ahí hay: **el nombre del módulo privado** (`miempresa.interno/herramienta`), **cada dependencia
con su versión exacta y su hash**, y hasta **las banderas con las que se compiló**. En el binario
que más he intentado despojar.

⚠️ Que las versiones sean exactas no es un detalle menor: es la lista de la compra para quien
busque una vulnerabilidad conocida. No hay que adivinar qué versión de qué librería llevas.

## La ruta, en el binario normal

```
   /casa/mi-proyecto-privado/app/interno/secreto.go
   /casa/mi-proyecto-privado/app/main.go
```

## ⛔ ¿Se puede quitar la lista de dependencias?

```
   probando -buildvcs=false:
     dependencias que quedan: 3
   probando -ldflags "-s -w -buildid=":
     dependencias que quedan: 3
```

**No con las banderas normales.** La información de compilación vive en su propia sección del
ejecutable, que no es la tabla de símbolos, y por eso `-s` y `-w` no la tocan.

## Lo que una traza de pánico enseña en cada caso

```
   sin -trimpath:   /tmp/p/main.go:3
   con -trimpath:   miempresa.interno/panico/main.go:3
```

⚠️ **`-trimpath` no borra el nombre de tus ficheros: cambia cómo se escribe la ruta.** Deja de
salir dónde estaba tu proyecto en tu disco, y pasa a salir relativa al módulo. `secreto.go` sigue
estando dentro, y sigue saliendo en cualquier traza.

---

## ⛔ MI FALLO MIDIENDO ESTO

La primera versión del banco contaba «símbolos que empiezan por el nombre del módulo». Dio esto:

```
   normal:     1 símbolo del proyecto
   -trimpath:  3 símbolos del proyecto
```

Y me lo creí un momento: parecía que `-trimpath` **añadía** información, que es absurdo. No
añade nada. Lo que pasa es que mi grep medía cosas distintas en cada lado: sin `-trimpath` las
rutas de fichero son absolutas (`/casa/...`) y no empiezan por el nombre del módulo, así que no
las contaba; con `-trimpath` sí empiezan por él y sí las contaba.

Estaba comparando dos preguntas diferentes y llamándolas la misma. La tabla de arriba pregunta
las dos cosas por separado —el nombre de la función y el nombre del fichero— y se la hace igual
a los cuatro binarios.

## ⛔ Lo que NO dice este banco

- **No dice que Go sea peor que otros lenguajes en esto.** Un `.pyc`, un `.jar` o un paquete de
  Node enseñan tanto o más. Lo que pasa es que con Go la gente cree que el binario es opaco.
- **No es un problema de seguridad por sí mismo.** Es información, y la información tiene dueño.
  El riesgo depende de a quién le das el binario.
- **No se ha probado ofuscación de verdad** (garble y parecidos), que es otra conversación y
  tiene sus propios precios.
- Todo es de un binario de Linux compilado en un contenedor. Los números concretos cambian con
  el sistema y la versión del lenguaje; el reparto de qué se ve y qué no, no.
