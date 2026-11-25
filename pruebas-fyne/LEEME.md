# Pruebas — Curso MIEMBROS «Fyne a fondo» (idea #83)

> **Todo se ejecutó el 2026-09-03**: Fyne **2.8.1** con **go1.26.8**, compilado y **capturado**
> dentro de Docker sobre un servidor gráfico de mentira (Xvfb) y `import` de ImageMagick.
>
> 🥇 **`FYNE_SCALE=3` no es opcional para grabar.** La primera tanda se capturó a escala 1: la
> ventana salía de ~430 px de ancho, el motor de vídeo la estiraba a 1920 y **se veía borrosa**.
> Se detectó abriendo un fotograma del vídeo ya renderizado (canon: una imagen no vale hasta que
> la abres). Con escala 3 la captura nace grande y entra nítida — y el Xvfb tiene que ser
> 2560x1600 para que quepa.
>
> ⚠️ Y un tropiezo de Windows: `sed -i` sobre `correr.sh` dejó el fichero con retornos de carro,
> y `dash` dentro del contenedor lo rechazó con `set: Illegal option -`. El mensaje no dice nada
> de CRLF.
> Las capturas de [`capturas/`](capturas/) son **ventanas de verdad**, no maquetas.
> Salida literal sin retocar en [`salida-real.txt`](salida-real.txt).

## Cómo se reproduce

```sh
docker volume create gomodcache
docker build -t fynelab .                 # ver el Dockerfile de esta carpeta

MSYS_NO_PATHCONV=1 docker run --rm \
    -v gomodcache:/go/pkg/mod -v "$(pwd -W):/app" -w //app \
    fynelab sh -c "cd /app && go mod tidy; sh correr.sh"
```

⚠️ **`MSYS_NO_PATHCONV=1` no es opcional en Windows** (mismo motivo que en el curso de la nube
local: Git Bash reescribe las rutas dentro de los argumentos).

---

## 🥇 EP1 · Fyne 2.8 ya exige cabeceras de Wayland, y el error no lo dice

El primer intento de compilar murió así, con las dependencias «de siempre» (`libgl1-mesa-dev`
y `xorg-dev`) instaladas:

```
# github.com/go-gl/glfw/v3.4/glfw
./glfw/src/wl_platform.h:27:10: fatal error: wayland-client-core.h: No such file or directory
compilation terminated.
```

👉 Faltaban `libwayland-dev`, `wayland-protocols` y `libxkbcommon-dev`. El mensaje habla de un
fichero de cabecera de C: **no dice que te falte un paquete del sistema**, y menos cuál. Es el
primer muro de cualquiera que empiece con Fyne en 2026, y va en el episodio 1.

## 🥇 El tamaño del binario — el precio de «un binario, cero dependencias»

Medido en esta corrida, sin optimizar y con la ventana más simple posible:

| etapa | binario |
|---|---:|
| e01 · una ventana con una etiqueta | **32.343.016 B** |
| e02 · contenedores | 32.365.456 B |
| e03 · binding | 32.385.736 B |
| e05 · lista de 50.000 | 32.423.312 B |
| e07 · formulario con validación | 32.575.040 B |

👉 **Treinta y dos megas para una ventana con una etiqueta.** Y fíjate en lo que de verdad
enseña la tabla: entre la ventana más tonta y el formulario con validación hay **232 KB de
diferencia**. El peso no es tu código: es Fyne entero, que se lleva su motor de dibujado dentro.
Es el precio de no depender de nada, y hay que decirlo con el número delante.

⛔ **Restricción editorial heredada** (decidida en la entrega 2 y publicada): este canal **NO
promete apps «bonitas»**, porque el propio vídeo admite que Fyne no se ve nativo. No introducir
«bonitas» ni «beautiful» en ninguna variante.

## 🥇🥇 EP4 · La regla de los hilos, y que Fyne te avisa él solo

Es la columna vertebral del curso. Dos programas idénticos salvo por **dónde** se toca la
interfaz. El «mal» actualiza una etiqueta directamente desde una gorrutina:

```
*** This application has not been migrated to the fyne.Do threading model ***
*** The next major Fyne release will remove this safety! ***
*** Read more at https://docs.fyne.io/started/goroutines ***
*** Error in Fyne call thread, this should have been called in fyne.Do[AndWait] ***
*** This error needs to be addressed as it could cause concurrent errors soon ***
*** The next major Fyne release will disable this safety by default! ***
  From: /app/etapas/e04-mal/main.go:27
```

Y el «bien», que mete solo el retoque dentro de `fyne.Do`, **no imprime absolutamente nada**.

👉 **Tres cosas que hacen a este episodio el centro del curso:**
1. **La ventana se ve IGUAL en los dos casos.** Compara `capturas/e04-mal.png` con
   `capturas/e04-bien.png`: son la misma imagen. El fallo no se ve, se lee en el registro.
2. **Es la propia biblioteca la que anuncia el reloj**, no una interpretación nuestra: *«the
   next major Fyne release will remove this safety»*. Fuente primaria, impresa por el programa.
3. **Señala la línea exacta** (`main.go:27`), que es más de lo que hace casi ninguna biblioteca.

La API es de la versión 2.6 y está en `thread.go` del propio módulo:

```go
// DoAndWait is used to execute a specified function in the main Fyne runtime context.
func DoAndWait(fn func()) { CurrentApp().Driver().DoFromGoroutine(fn, true) }

// Do is used to execute ... without waiting.
func Do(fn func()) { CurrentApp().Driver().DoFromGoroutine(fn, false) }
```

## 🥇 EP5 · `widget.List` no es una caja vertical: 50.000 claves y **9 filas**

La captura lo dice sola (`capturas/e05-lista.png`):

```
50000 claves · filas construidas: 9
```

El programa cuenta cuántas veces llama Fyne a la función que **crea** una fila. Con cincuenta
mil elementos en la lista, la creó **nueve veces**: las que caben en pantalla y poco más.

👉 Por eso `widget.List` pide **tres** funciones y no una lista de widgets: cuántos hay, cómo se
fabrica una fila vacía, y cómo se rellena la fila número N. Quien mete cincuenta mil etiquetas
en un `VBox` construye cincuenta mil widgets, y la ventana no abre.

## EP7 · La validación apaga el botón sola

Con un nombre inválido a propósito (`Fotos_2026`, que lleva mayúsculas y guion bajo), el campo
se marca en rojo con el motivo y **el botón «Crear» queda deshabilitado** sin escribir una línea
para deshabilitarlo. Se ve en `capturas/e07-formulario.png`.

---

## Lo que este material NO demuestra

- **No se midió rendimiento de nada**: ni fotogramas por segundo, ni memoria, ni tiempo de
  arranque. Lo medido es tamaño de binario y número de filas construidas.
- **Las capturas son de Linux sobre Xvfb.** El mismo código dibuja igual en Windows y en Mac
  —esa es la propuesta de Fyne— pero aquí no se ha comprobado en esos dos sistemas.
- **No se ha probado en móvil**, aunque Fyne lo soporta.
- Los 32 MB son **sin optimizar**: no se aplicó `-ldflags="-s -w"` ni empaquetado. El episodio
  de empaquetado tendrá que dar el número de después, y decir que son dos números distintos.
