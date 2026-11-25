# SALIDA — la revancha de Zig: hilo por conexión contra pool de hilos, y Go (golandia)

> Corrida del 2026-11-06, con el candado `_logs/banco.lock` tomado en cada sección. Regenerable con `python investigacion/pruebas_zig_revancha.py imagenes igual carga control pelicula informe`. ⛔ Ninguna cifra de blog.

- go: `go version go1.27.2 linux/amd64` (net/http) · imagen 2.4 MiB
- zig: `0.17.0` (std.Io Threaded, solo librería estándar) · imagen 1.1 MiB
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey` con 4 CPU desde otro contenedor. Contenedor nuevo en cada corrida, 5 s de calentamiento y 15 s medidos, 3 corridas en orden rotado. Se da la MEDIANA y, entre corchetes, el rango (mínimo–máximo) de las 3.
- zig-hilos: una tarea por conexión (`Io.Group.concurrent`, como el 487). zig-pool: 2 hilos fijos + cola (`Io.Queue`).
- «peticiones/s» = el `Requests/sec` de hey. ⚠️ En zig-pool con keep-alive ese número no compara nada: solo N clientes reciben respuesta (ver «errores») y hey espera 20 s a los demás, así que divide entre 20 s; lo que importa ahí son los clientes sin respuesta. p50/p99: solo de las respuestas que llegaron.
- Los dos Zig con backlog 4096 (el de Go; el de Zig por defecto es 128) y keep-alive.
- CPU/memoria/hilos: `investigacion/muestreo.py`, cada 0,5 s, del cgroup del contenedor (100 % = un núcleo; techo 200 %; el reloj del auxiliar tiene 10 ms de resolución, así que un punto suelto puede pasar de 200 %). «Frenadas»: `nr_throttled` de cpu.stat en los 15 s medidos: cuántos periodos de 100 ms el límite de 2 CPU detuvo al contenedor.

## 0 · ¿Hacen lo mismo?

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 · `LISTO_EN_MS 0.036`
- zig-hilos: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 · `LISTO_EN_MS 0.015 modo=hilos hilos=2`
- zig-pool: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 · `LISTO_EN_MS 0.016 modo=pool hilos=2`

## 1 · Régimen keep-alive (lo que hace hey por defecto: cada cliente reutiliza su conexión)

### /suma?n=200000 · 16 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **14976** [14956–15012] | 1.0 ms | 2.6 ms | 224659 | 0 | 197 % | 10.8 MiB | 4 | 44 |
| zig-hilos | **7149** [6766–7351] | 0.3 ms | 81.2 ms | 107375 | 0 | 199 % | 9.6 MiB | 17 | 150 |
| zig-pool | **9078** [9031–9080] | 0.2 ms | 0.2 ms | 181607 | 14 | 110 % | 3.0 MiB | 3 | 0 |

### /suma?n=200000 · 300 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **14688** [14633–14713] | 20.0 ms | 37.6 ms | 220650 | 0 | 196 % | 27.9 MiB | 5 | 54 |
| zig-hilos | **9802** [8267–11692] | 2.1 ms | 196.1 ms | 147111 | 0 | 198 % | 314.4 MiB | 1033 | 149 |
| zig-pool | **4411** [16–4436] | 0.2 ms | 0.3 ms | 88255 | 299 | 54 % | 8.0 MiB | 3 | 0 |

### /trabajo?n=20000 · 16 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **479** [470–482] | 25.9 ms | 121.7 ms | 7191 | 0 | 194 % | 36.9 MiB | 4 | 31 |
| zig-hilos | **475** [471–482] | 3.9 ms | 99.8 ms | 7164 | 0 | 199 % | 38.5 MiB | 17 | 150 |
| zig-pool | **697** [694–697] | 2.1 ms | 2.7 ms | 13925 | 14 | 143 % | 8.0 MiB | 3 | 0 |

### /trabajo?n=20000 · 300 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **493** [493–497] | 637.7 ms | 1711.8 ms | 7684 | 0 | 197 % | 125.8 MiB | 5 | 49 |
| zig-hilos | **415** [390–452] | 396.9 ms | 11298.8 ms | 6446 | 0 | 199 % | 597.3 MiB | 404 | 155 |
| zig-pool | **702** [701–703] | 2.1 ms | 2.7 ms | 13817 | 298 | 141 % | 11.9 MiB | 3 | 0 |

### /espera?ms=100 · 300 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **2945** [2941–2946] | 101.2 ms | 106.3 ms | 44400 | 0 | 13 % | 21.2 MiB | 6 | 0 |
| zig-hilos | **2960** [2955–2961] | 100.6 ms | 106.8 ms | 44457 | 0 | 28 % | 96.0 MiB | 301 | 0 |
| zig-pool | **15** [15–15] | 101.1 ms | — | 12 | 298 | 0 % | 9.6 MiB | 3 | 0 |

## 2 · Régimen conexión nueva por petición (`hey -disable-keepalive`)

### /suma?n=200000 · 16 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **13582** [13567–13644] | 1.2 ms | 2.0 ms | 203788 | 0 | 191 % | 11.4 MiB | 4 | 43 |
| zig-hilos | **7886** [7699–7976] | 0.5 ms | 66.7 ms | 118592 | 0 | 199 % | 15.6 MiB | 25 | 149 |
| zig-pool | **15929** [15903–15934] | 1.0 ms | 1.8 ms | 238964 | 0 | 192 % | 5.6 MiB | 3 | 148 |

### /suma?n=200000 · 300 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **13434** [13407–13452] | 22.4 ms | 38.3 ms | 201807 | 0 | 191 % | 13.7 MiB | 5 | 58 |
| zig-hilos | **7197** [6991–7258] | 7.8 ms | 100.9 ms | 108209 | 0 | 198 % | 144.5 MiB | 427 | 150 |
| zig-pool | **15568** [15565–15621] | 18.6 ms | 23.4 ms | 233805 | 0 | 191 % | 7.8 MiB | 3 | 148 |

### /trabajo?n=20000 · 16 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **460** [458–463] | 33.1 ms | 89.1 ms | 6916 | 0 | 194 % | 32.3 MiB | 5 | 31 |
| zig-hilos | **377** [374–378] | 8.7 ms | 102.3 ms | 5676 | 0 | 199 % | 39.0 MiB | 27 | 150 |
| zig-pool | **841** [837–841] | 18.9 ms | 21.2 ms | 12624 | 0 | 196 % | 8.9 MiB | 3 | 53 |

### /trabajo?n=20000 · 300 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **480** [479–482] | 578.9 ms | 1543.4 ms | 7480 | 0 | 196 % | 119.6 MiB | 4 | 53 |
| zig-hilos | **314** [303–366] | 896.9 ms | 2202.5 ms | 5028 | 0 | 199 % | 517.3 MiB | 520 | 155 |
| zig-pool | **834** [832–835] | 359.4 ms | 365.7 ms | 12806 | 0 | 195 % | 11.2 MiB | 3 | 51 |

### /espera?ms=100 · 300 conexiones

| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |
|---|---|---|---|---|---|---|---|---|---|
| go | **2879** [2851–2879] | 103.0 ms | 115.8 ms | 43247 | 0 | 16 % | 19.4 MiB | 5 | 0 |
| zig-hilos | **2917** [2910–2925] | 101.9 ms | 110.5 ms | 43859 | 0 | 35 % | 125.3 MiB | 382 | 0 |
| zig-pool | **20** [20–20] | 15058.9 ms | 15067.6 ms | 598 | 0 | 0 % | 8.0 MiB | 3 | 0 |

## 3 · ¿El muestreo ensucia? (mismas corridas con y sin el auxiliar, intercaladas)

| configuración | variante | con muestreo (mediana) | sin muestreo (mediana) | diferencia |
|---|---|---|---|---|
| keepalive-suma-c16 | go | 14980 [14980, 14910, 15022] | 14983 [15048, 14983, 14961] | -0.0 % |
| keepalive-suma-c16 | zig-hilos | 6818 [6766, 7280, 6818] | 6607 [6607, 6814, 6542] | +3.2 % |
| keepalive-suma-c16 | zig-pool | 8883 [9049, 4435, 8883] | 9039 [9039, 9094, 8915] | -1.7 % |
| nueva-trabajo-c16 | go | 446 [462, 446, 443] | 451 [459, 445, 451] | -1.1 % |
| nueva-trabajo-c16 | zig-hilos | 473 [454, 535, 473] | 462 [415, 495, 462] | +2.4 % |
| nueva-trabajo-c16 | zig-pool | 828 [831, 816, 828] | 824 [836, 824, 824] | +0.5 % |

## 4 · Las series en el tiempo (para las gráficas)

3 s en reposo, 12 s con 300 conexiones, 5 s después; un punto cada 0,5 s. Ficheros: `series/pelicula-*.json` y `.csv`.

| serie | peticiones/s | CPU media | CPU máx. | memoria máx. | hilos máx. | conexiones máx. |
|---|---|---|---|---|---|---|
| pelicula-nueva-trabajo-c300-go | 466 | 120.3 % | 203.9 % | 124.75 MiB | 5 | 301 |
| pelicula-nueva-trabajo-c300-zig-hilos | 487 | 121.7 % | 233.8 % | 549.25 MiB | 681 | 300 |
| pelicula-nueva-trabajo-c300-zig-pool | 813 | 120.6 % | 204.0 % | 10.92 MiB | 3 | 301 |
| pelicula-nueva-trabajo-c300-zig-eventos | 951 | 120.1 % | 203.2 % | 11.52 MiB | 2 | 301 |
| pelicula-keepalive-suma-c300-go | 14090 | 117.9 % | 203.4 % | 26.65 MiB | 5 | 597 |
| pelicula-keepalive-suma-c300-zig-hilos | 10846 | 118.7 % | 236.0 % | 210.24 MiB | 684 | 571 |
| pelicula-keepalive-suma-c300-zig-pool | 16 | 0.2 % | 7.4 % | 7.5 MiB | 3 | 327 |
| pelicula-keepalive-suma-c300-zig-eventos | 16445 | 117.3 % | 203.2 % | 10.55 MiB | 2 | 587 |
| pelicula-keepalive-espera-c300-go | 2942 | 3.8 % | 7.2 % | 22.18 MiB | 6 | 330 |
| pelicula-keepalive-espera-c300-zig-hilos | 2962 | 7.5 % | 15.1 % | 95.86 MiB | 301 | 300 |
| pelicula-keepalive-espera-c300-zig-pool | 27 | 0.1 % | 0.3 % | 7.74 MiB | 3 | 300 |
| pelicula-keepalive-espera-c300-zig-eventos | 2940 | 1.7 % | 3.1 % | 7.4 MiB | 2 | 300 |


# Segunda parte: lo raro, comprobado (2026-11-05, tras reiniciar la máquina)

Binarios de la primera parte sin tocar. `zig-eventos` (eventos/eventos.zig): 2 hilos, cada uno con su epoll y su socket SO_REUSEPORT, escrito a mano sobre std.os.linux porque std.Io.Evented no compila en 0.17.0. Mismo /suma y /trabajo.

## F · LA TABLA DEL VÍDEO: las cuatro variantes en la misma sesión (zig-eventos con arena por hilo)

### keepalive-suma-c16

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **14300** [14281–14434] | 1.0 ms | 2.8 ms | 0 | 197 % | 9.7 MiB | 4 | 41 |
| zig-hilos | **11695** [11506–11851] | 0.2 ms | 80.4 ms | 0 | 200 % | 9.4 MiB | 17 | 150 |
| zig-eventos | **16725** [16672–16739] | 1.0 ms | 2.4 ms | 0 | 197 % | 4.7 MiB | 2 | 10 |

### keepalive-suma-c300

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **13753** [13746–13852] | 21.2 ms | 39.4 ms | 0 | 196 % | 29.4 MiB | 5 | 59 |
| zig-hilos | **11994** [11809–12077] | 2.1 ms | 105.5 ms | 0 | 198 % | 301.0 MiB | 990 | 149 |
| zig-eventos | **16362** [16346–16434] | 19.6 ms | 22.2 ms | 0 | 196 % | 10.7 MiB | 2 | 5 |

### keepalive-trabajo-c16

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **448** [445–448] | 29.3 ms | 121.2 ms | 0 | 195 % | 36.2 MiB | 4 | 32 |
| zig-hilos | **758** [742–784] | 2.8 ms | 91.8 ms | 0 | 199 % | 38.3 MiB | 17 | 150 |
| zig-eventos | **968** [961–973] | 18.4 ms | 38.1 ms | 0 | 198 % | 8.0 MiB | 2 | 5 |

### keepalive-trabajo-c300

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **470** [467–471] | 671.6 ms | 1794.7 ms | 0 | 196 % | 130.1 MiB | 5 | 42 |
| zig-hilos | **657** [638–657] | 398.3 ms | 1693.8 ms | 0 | 197 % | 692.8 MiB | 432 | 154 |
| zig-eventos | **953** [951–966] | 318.1 ms | 370.1 ms | 0 | 196 % | 13.4 MiB | 2 | 6 |

### keepalive-espera-c300

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **2953** [2949–2955] | 101.0 ms | 102.9 ms | 0 | 7 % | 21.3 MiB | 5 | 0 |
| zig-hilos | **2966** [2965–2970] | 100.5 ms | 101.9 ms | 0 | 13 % | 101.7 MiB | 319 | 0 |
| zig-eventos | **2953** [2950–2958] | 101.0 ms | 102.0 ms | 0 | 3 % | 8.0 MiB | 2 | 0 |

### nueva-suma-c16

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **12675** [12445–13136] | 1.3 ms | 2.5 ms | 0 | 197 % | 11.3 MiB | 4 | 40 |
| zig-hilos | **9815** [9688–9846] | 0.5 ms | 65.4 ms | 0 | 198 % | 16.7 MiB | 27 | 149 |
| zig-pool | **15531** [15424–15576] | 1.0 ms | 1.9 ms | 0 | 198 % | 5.8 MiB | 3 | 149 |
| zig-eventos | **15126** [14856–15131] | 1.0 ms | 2.0 ms | 0 | 191 % | 5.7 MiB | 2 | 0 |

### nueva-suma-c300

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **12608** [12580–13038] | 23.8 ms | 41.7 ms | 0 | 191 % | 13.9 MiB | 5 | 60 |
| zig-hilos | **9979** [9804–10012] | 5.7 ms | 93.5 ms | 0 | 198 % | 131.1 MiB | 372 | 149 |
| zig-pool | **15138** [15113–15152] | 19.2 ms | 23.7 ms | 0 | 197 % | 7.7 MiB | 3 | 148 |
| zig-eventos | **15229** [15109–15229] | 21.0 ms | 40.4 ms | 0 | 187 % | 10.0 MiB | 2 | 8 |

### nueva-trabajo-c16

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **436** [434–450] | 34.7 ms | 93.9 ms | 0 | 194 % | 30.5 MiB | 4 | 40 |
| zig-hilos | **538** [537–569] | 5.4 ms | 93.6 ms | 0 | 199 % | 38.0 MiB | 29 | 149 |
| zig-pool | **816** [815–825] | 19.5 ms | 21.9 ms | 0 | 197 % | 8.6 MiB | 3 | 58 |
| zig-eventos | **938** [932–940] | 16.8 ms | 32.9 ms | 0 | 192 % | 8.3 MiB | 2 | 18 |

### nueva-trabajo-c300

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **462** [453–467] | 622.6 ms | 1547.8 ms | 0 | 196 % | 115.6 MiB | 5 | 51 |
| zig-hilos | **465** [464–486] | 598.7 ms | 1495.5 ms | 0 | 199 % | 550.0 MiB | 661 | 155 |
| zig-pool | **819** [816–820] | 365.2 ms | 378.6 ms | 0 | 195 % | 11.1 MiB | 3 | 59 |
| zig-eventos | **957** [947–963] | 309.4 ms | 546.3 ms | 0 | 195 % | 11.8 MiB | 2 | 15 |

### nueva-espera-c300

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **2934** [2933–2938] | 101.8 ms | 104.4 ms | 0 | 12 % | 19.9 MiB | 6 | 0 |
| zig-hilos | **2951** [2948–2951] | 101.1 ms | 104.1 ms | 0 | 21 % | 122.1 MiB | 371 | 0 |
| zig-pool | **20** [20–20] | 15031.6 ms | 15033.7 ms | 0 | 0 % | 8.0 MiB | 3 | 0 |
| zig-eventos | **2937** [2934–2940] | 101.7 ms | 104.5 ms | 0 | 6 % | 8.6 MiB | 2 | 0 |

## A · zig-eventos contra Go (intercalados; Go es el ancla contra la primera parte)

### keepalive-suma-c16 (Go en la primera parte: 14976 pet/s; ahora 14343: -4.2 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **14343** [14267–14465] | 1.0 ms | 2.8 ms | 0 | 197 % | 10.3 MiB | 4 | 38 |
| zig-eventos | **16760** [16634–16761] | 0.9 ms | 2.0 ms | 0 | 198 % | 5.2 MiB | 2 | 3 |

### keepalive-suma-c300 (Go en la primera parte: 14688 pet/s; ahora 13820: -5.9 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **13820** [13726–13886] | 21.2 ms | 39.4 ms | 0 | 197 % | 28.9 MiB | 5 | 77 |
| zig-eventos | **16251** [16239–16284] | 21.2 ms | 23.8 ms | 0 | 196 % | 10.7 MiB | 2 | 2 |

### keepalive-trabajo-c16 (Go en la primera parte: 479 pet/s; ahora 447: -6.5 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **447** [446–450] | 29.2 ms | 120.7 ms | 0 | 194 % | 33.3 MiB | 4 | 37 |
| zig-eventos | **908** [903–914] | 17.0 ms | 35.3 ms | 0 | 198 % | 33.7 MiB | 2 | 9 |

### keepalive-trabajo-c300 (Go en la primera parte: 493 pet/s; ahora 469: -4.8 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **469** [469–471] | 668.9 ms | 1774.5 ms | 0 | 196 % | 125.2 MiB | 5 | 41 |
| zig-eventos | **905** [900–916] | 332.0 ms | 458.0 ms | 0 | 196 % | 571.2 MiB | 2 | 5 |

### keepalive-espera-c300 (Go en la primera parte: 2945 pet/s; ahora 2952: +0.2 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **2952** [2950–2952] | 101.1 ms | 103.0 ms | 0 | 7 % | 21.3 MiB | 6 | 0 |
| zig-eventos | **2951** [2950–2953] | 101.0 ms | 102.0 ms | 0 | 4 % | 10.0 MiB | 2 | 0 |

### nueva-suma-c16 (Go en la primera parte: 13582 pet/s; ahora 12938: -4.7 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **12938** [12885–12954] | 1.3 ms | 2.3 ms | 0 | 197 % | 11.4 MiB | 4 | 46 |
| zig-eventos | **14903** [14832–14974] | 1.1 ms | 2.1 ms | 0 | 190 % | 5.8 MiB | 2 | 1 |

### nueva-suma-c300 (Go en la primera parte: 13434 pet/s; ahora 12402: -7.7 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **12402** [12364–12654] | 24.2 ms | 41.3 ms | 0 | 191 % | 13.7 MiB | 5 | 61 |
| zig-eventos | **15043** [14896–15089] | 20.9 ms | 40.5 ms | 0 | 189 % | 12.0 MiB | 2 | 14 |

### nueva-trabajo-c16 (Go en la primera parte: 460 pet/s; ahora 418: -9.1 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **418** [415–423] | 36.4 ms | 97.8 ms | 0 | 194 % | 31.5 MiB | 5 | 31 |
| zig-eventos | **767** [762–778] | 20.3 ms | 40.6 ms | 0 | 191 % | 8.4 MiB | 2 | 21 |

### nueva-trabajo-c300 (Go en la primera parte: 480 pet/s; ahora 445: -7.2 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **445** [445–450] | 645.3 ms | 1526.2 ms | 0 | 196 % | 129.3 MiB | 5 | 51 |
| zig-eventos | **793** [790–808] | 362.2 ms | 542.1 ms | 0 | 195 % | 11.6 MiB | 2 | 38 |

### nueva-espera-c300 (Go en la primera parte: 2879 pet/s; ahora 2925: +1.6 %)

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **2925** [2922–2930] | 101.9 ms | 105.1 ms | 0 | 11 % | 19.3 MiB | 6 | 0 |
| zig-eventos | **2932** [2929–2933] | 101.7 ms | 103.9 ms | 0 | 6 % | 9.0 MiB | 2 | 0 |

## B · Causa: keep-alive y pool bloqueante (/suma, 15 s medidos, variando hilos y conexiones)

| hilos del pool · conexiones | respuestas por corrida | peticiones/s | errores (clientes sin respuesta) |
|---|---|---|---|
| hilos2-c2 | [169243, 169125, 170003] | 11283 (respuestas ÷ 15 s) | [0, 0, 0] |
| hilos2-c3 | [170522, 163882, 167938] | 11196 (respuestas ÷ 15 s) | [1, 1, 1] |
| hilos2-c16 | [167109, 170273, 168448] | 11230 (respuestas ÷ 15 s) | [14, 14, 14] |
| hilos16-c16 | [175463, 182424, 179210] | 11947 (respuestas ÷ 15 s) | [0, 1, 0] |

## C · Causa: /espera?ms=100 sin keep-alive, 300 conexiones, pool de N hilos que duermen

| hilos | peticiones/s [rango] | teoría N ÷ 0,1 s | p50 | memoria máx. |
|---|---|---|---|---|
| 2 | **20** [20–20] | 20 | 15029.4 ms | 8.3 MiB |
| 8 | **80** [80–80] | 80 | 3708.7 ms | 10.0 MiB |
| 32 | **318** [318–318] | 320 | 902.3 ms | 16.6 MiB |
| 300 | **2952** [2948–2954] | 3000 | 101.1 ms | 95.8 MiB |

## D · Diagnóstico: hilos contra conexiones contra tareas (binario diag/main.zig, solo para esto)

«vivas_max» = aceptadas menos terminadas (cuenta las que esperan hilo); «corriendo_max» = dentro de atender().

| corrida | hilos máx. | conexiones máx. | último DIAG |
|---|---|---|---|
| diag-keepalive-suma-c300-r1 | 965 | 947 | `DIAG aceptadas=4067 terminadas=4067 vivas=0 vivas_max=947 corriendo_max=797` |
| diag-keepalive-suma-c300-r2 | 809 | 649 | `DIAG aceptadas=3214 terminadas=3214 vivas=0 vivas_max=773 corriendo_max=696` |
| diag-nueva-trabajo-c300-r1 | 758 | 300 | `DIAG aceptadas=10065 terminadas=10065 vivas=0 vivas_max=480 corriendo_max=468` |
| diag-nueva-trabajo-c300-r2 | 651 | 300 | `DIAG aceptadas=10412 terminadas=10412 vivas=0 vivas_max=484 corriendo_max=478` |
| diag-keepalive-espera-c300-r1 | 302 | 300 | `DIAG aceptadas=601 terminadas=601 vivas=0 vivas_max=300 corriendo_max=300` |
| diag-keepalive-espera-c300-r2 | 302 | 300 | `DIAG aceptadas=601 terminadas=601 vivas=0 vivas_max=300 corriendo_max=300` |

## E · ¿Es el límite de CPU? Lo mismo con `--cpuset-cpus 2,4` (dos núcleos enteros, sin cuota; hey en 8-11)

### keepalive-suma-c16

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **14185** [14065–14189] | 1.0 ms | 2.9 ms | 0 | 194 % | 8.8 MiB | 4 | 0 |
| zig-hilos | **16568** [16540–16650] | 0.9 ms | 2.6 ms | 0 | 198 % | 7.1 MiB | 17 | 0 |
| zig-eventos | **16735** [16689–16781] | 0.9 ms | 2.0 ms | 0 | 198 % | 2.7 MiB | 2 | 0 |

### nueva-trabajo-c16

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **416** [416–419] | 36.1 ms | 100.6 ms | 0 | 190 % | 32.4 MiB | 5 | 0 |
| zig-hilos | **697** [697–700] | 19.6 ms | 79.9 ms | 0 | 191 % | 39.5 MiB | 38 | 0 |
| zig-pool | **802** [790–804] | 19.9 ms | 21.9 ms | 0 | 196 % | 6.5 MiB | 3 | 0 |
| zig-eventos | **779** [778–788] | 20.3 ms | 39.7 ms | 0 | 191 % | 6.9 MiB | 2 | 0 |

### nueva-suma-c300

| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |
|---|---|---|---|---|---|---|---|---|
| go | **12636** [12619–12638] | 23.7 ms | 42.8 ms | 0 | 190 % | 12.5 MiB | 5 | 0 |
| zig-hilos | **14750** [14673–14929] | 19.3 ms | 132.7 ms | 0 | 187 % | 118.9 MiB | 380 | 0 |
| zig-pool | **15286** [15096–15344] | 19.5 ms | 21.1 ms | 0 | 191 % | 5.2 MiB | 3 | 0 |
| zig-eventos | **15161** [15150–15259] | 21.6 ms | 39.7 ms | 0 | 188 % | 7.9 MiB | 2 | 0 |

## G · Lo que hay detrás del límite de CPU (ejecutado 2026-11-05)

```
$ docker run --rm --cpus 2 golang:1.27.2 sh -c 'go run main.go; cat /sys/fs/cgroup/cpu.max'
GOMAXPROCS 2 NumCPU 24
cpu.max: 200000 100000
$ docker run --rm golang:1.27.2 go run main.go
GOMAXPROCS 24 NumCPU 24
```

## H · Notas de lectura

- Para comparar variantes, usa la sección F: las cuatro en la misma sesión. Entre sesiones Go se movió un 5 % (la primera parte, antes de un corte de luz, salió más alta) y zig-hilos con `--cpus 2` mucho más: /suma keep-alive 16 conexiones, 7.149 en la primera parte y 11.695 en la F. Dentro de una sesión su rango es estrecho; entre sesiones no: es la variante más sensible a la cuota.
- zig-eventos de la carga A y de la sección E retenía un arena por conexión (571 MiB en keep-alive /trabajo con 300). Fallo mío, no del método: la F y las películas usan un arena por hilo (13,4 MiB).
- El control de la sección 3: el 4435 de zig-pool no es el muestreo. Con keep-alive el reparto depende de qué conexiones toman los dos hilos; si uno se queda con una conexión ociosa, atiende la mitad.
- «Frenadas» no explica sola la latencia: zig-pool sin keep-alive también marca ~149 y su p99 es 1,9 ms. La prueba de causa es la sección E (sin cuota), no el contador.
- Los servidores Zig entienden lo justo de HTTP; net/http de Go es completo. Parte de la ventaja viene de ahí (sin aislar).

