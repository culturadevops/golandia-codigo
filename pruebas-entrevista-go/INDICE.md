# 🧪 Pruebas de la serie «Prepárate la entrevista de Go»

> **Todo lo de aquí está EJECUTADO.** Ninguna cifra de esta serie sale de un blog.
> Entorno único y citable en los ocho vídeos: contenedor **`golang:1.26`** · `go1.26.7 linux/amd64`
> · CPU `Intel(R) Core(TM) Ultra 7 270K Plus`, 24 hilos. Ejecutado el **2026-08-21**.

## Cómo se regenera

```powershell
python canales/golandia/investigacion/pruebas_ep01.py   # …ep02 … ep08
```

Cada script escribe su carpeta `NN-tema/` con los `.go` y un **`SALIDA.md`** que lleva la salida
**literal** de cada comando, su código de salida y qué comandos fallan **a propósito**.

⚠️ **Una corrida con fallos NO pisa la buena.** Escribe en `SALIDA.FALLIDA.md` y deja `SALIDA.md`
donde estaba. Se añadió el 2026-08-21 después de que una caída del demonio de Docker a mitad de
corrida sobrescribiera el `SALIDA.md` ya verificado del episodio 1.

⚠️ **Por qué viven aquí y no en el `_trabajo/` del vídeo:** `_trabajo/` es desechable y `fab.py
limpiar` se lo lleva a la papelera al publicar. Esto es el respaldo de lo que se afirma en pantalla
y tiene que sobrevivir al vídeo.

⚠️ **Un solo `episodio()` por script.** Cada llamada reescribe el `SALIDA.md` entero, así que añadir
una prueba es añadir un elemento a la lista de la llamada que ya existe, nunca una segunda llamada.
Costó dos `SALIDA.md` truncados el 2026-08-21.

---

## El titular verificado de cada episodio

| Ep | Carpeta | El dato que sostiene el vídeo (salida real) |
|---|---|---|
| 1 | `01-slices-y-arrays` | `original[1:3]` tiene **cap 4**, así que el append **pisa**: `original: [1 2 3 99 5]`. Con `base[3:5]` (cap 2) reasigna y no pisa. La capacidad **dobla hasta 512 y ahí deja de doblar**: 848 · 1280 · 1792 · 2560. Y `nil` contra vacío: los dos con len 0 y cap 0, pero solo uno es `nil`. |
| 2 | `02-mapas-y-el-valor-cero` | Leer de un mapa nil da el cero; escribir da `panic: assignment to entry in nil map`; **borrar no falla**. El range dio **4 órdenes distintos en 4 vueltas**. Dos goroutines escribiendo → `fatal error: concurrent map writes` (**fatal, no panic: no hay recover**). Y `invalid map key type []int`: un struct con un slice dentro deja de servir como clave. |
| 3 | `03-interfaces-y-el-nil-que-no-es-nil` | La función no falla, `err` se imprime `<nil>` y **aun así entra en el `if`**. El `%T` delata: `tipo = *main.MiError`, `valor = <nil>`. Le pasa a **cualquier** interfaz, también a `any` con un slice o un mapa dentro. `errors.As` devuelve **true** y luego leer el campo es SIGSEGV. Y la solución con `reflect` funciona **y es la equivocada**. |
| 4 | `04-punteros-y-receptores` | `Usuario does not implement Guardable (method Guardar has pointer receiver)`. Struct de **448 bytes**: valor **3,41 ns/op** contra **0,44 ns/op** por puntero (los dos con `//go:noinline`). `go vet`: `passes lock by value: contains sync.Mutex` — y el programa corre igual. Y pasar el struct por valor **no** protege el slice de dentro: `[999 2 3]`. |
| 5 | `05-goroutines-y-deadlock` | `fatal error: all goroutines are asleep - deadlock!`, con `goroutine 1 [chan send]`. El `wg.Add(1)` dentro de la goroutine: esperaba 1000, hizo **963**. Canal cerrado: recibir da cero y `false`; enviar da `panic: send on closed channel`. 100 fugas → **101 goroutines vivas** y el programa acaba sin decir nada. |
| 6 | `06-carreras-mutex-y-race` | Tres corridas: **183277 · 105352 · 124579**. Y la estadística que sustituye a la anécdota: de **500 corridas, 112 dieron exactamente 200000 — el 22,4 %**, con el peor caso en 100000 clavados. Con `-race` una corrida dio 200000 **y aun así** sacó `WARNING: DATA RACE`. Coste: mutex **51,0 ns** · atomic **13,7 ns** · RWMutex lecturas **41,6 ns** · sin proteger **1,22 ns**; `-race` lo multiplica por **~4,7x** (53 → 248 ns). Y la joya: **una carrera lógica con todo protegido, que `-race` pasa limpio** y aun así da 130556. |
| 7 | `07-errores-y-context` | `%w` y `%v` imprimen **el mismo mensaje** y `errors.Is` da **true / false**; `Unwrap` del segundo da `<nil>`. `errors.As` saca `Codigo = 429`. `WithTimeout(300ms)` cortó en 300 ms y el servidor se enteró: `[servidor] el cliente se fue: context canceled` — **la cancelación viaja**. `go vet`: `the cancel function ... should be called, not discarded`. Y el middleware con recover devuelve 500 y el servidor sigue vivo. |
| 8 | `08-el-ejercicio-en-vivo` | Las cuatro versiones, medidas: secuencial **2,41 s** (pico 1) → una goroutine por URL **200 ms** (**pico 12**) → semáforo K=4 **600 ms** (pico 4) → cancelable a 350 ms: **ok=4, cancelados=8, pico=4** (4+8=12, ni un error perdido). Todo pasa `-race`, y el test **demuestra** el límite contándolo desde el servidor: `k=1→1 · k=3→3 · k=5→5`. |

---

## Lo que estas corridas ya cambiaron

Tres cosas se escribieron mal antes de ejecutarlas, y ejecutar las corrigió. Queda apuntado porque
es el argumento de por qué esta serie no se puede guionizar de memoria:

1. **Episodio 1.** Se dio por hecho que `append` sobre `original[1:3]` **no** tocaba el original.
   Sí lo toca: la capacidad es 4. El vídeo tiene mejor giro por haberlo ejecutado, y el episodio 8
   lo cuenta como cierre de la serie.
2. **Episodio 4.** El primer benchmark daba **0,13 ns/op** al receptor por puntero — el compilador
   se lo había comido entero por *inlining*. Con `//go:noinline` en los dos métodos la comparación
   pasa a ser real. Publicar «41x más rápido» habría sido una cifra inventada con aspecto de medida.
3. **Episodio 6.** El guion citaba «tres corridas y la tercera dio el número bueno». Al reejecutar,
   ninguna de las tres lo dio: **es no determinista y la anécdota no se sostiene**. Se sustituyó por
   una prueba que repite la carrera **500 veces y cuenta**, que da el 22,4 % — un dato que sí aguanta
   una reejecución y que además es mejor: es la probabilidad de que el bug pase tus pruebas.

## Aviso para quien escriba o revise los guiones

- **Si reejecutas, revisa los guiones.** Las salidas con goroutines cambian en cada corrida, y el
  builder cita números literales. Lo que hay que citar es lo **estable** (el comportamiento) o lo
  **estadístico** (el 22,4 %), nunca una corrida concreta que no se pueda reproducir.
- El benchmark del episodio 4 son **nanosegundos**: la diferencia es real pero **irrelevante en casi
  todo el código real**. El cierre honesto es que se usa puntero por **mutación y consistencia**,
  no por velocidad. Vender ahí una ganancia de rendimiento sería mentir con datos ciertos.
- Si se cita un benchmark, **se dice la CPU**.
