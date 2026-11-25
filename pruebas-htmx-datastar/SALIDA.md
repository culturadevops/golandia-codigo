# SALIDA — ¿HTMX o Datastar? Con Go (golandia)

> Corrida del 2026-10-19 16:32. htmx 2.0.11 · Datastar 1.0.4 · datastar-go v1.2.2 · golang:1.26.

```
go version go1.26.9 linux/amd64
### 1 · Las dos librerías (servidas desde la app, sin CDN)
$ wc -c static/htmx.min.js static/datastar.js
52182 static/htmx.min.js
33553 static/datastar.js
85735 total
$ gzip -9 -c static/htmx.min.js | wc -c ; gzip -9 -c static/datastar.js | wc -c
16838
13380
$ go list -m all | wc -l   # módulos: net/http + datastar-go
21

### 2 · Buscar «go»: lo que responde el servidor
$ curl -s 'localhost:8080/htmx/buscar?q=go'
<ul id="resultados"><li>Go</li></ul>
$ curl -s 'localhost:8080/datastar/buscar?datastar={"q":"go"}'
event: datastar-patch-elements
data: elements <ul id="resultados"><li>Go</li></ul>



$ curl -s -o /dev/null -w '%{size_download} bytes\n' …   # el tamaño de cada respuesta
htmx: 36 bytes
datastar: 85 bytes

### 3 · El reloj: HTMX pregunta, Datastar empuja
$ curl -s localhost:8080/htmx/hora   # HTMX: una petición por segundo
21:32:47
$ curl -sN --max-time 3 localhost:8080/datastar/hora   # Datastar: UNA conexión, el servidor manda cada segundo
event: datastar-patch-elements
data: elements <span id="hora">21:32:47</span>


event: datastar-patch-elements
data: elements <span id="hora">21:32:48</span>


event: datastar-patch-elements
data: elements <span id="hora">21:32:49</span>



```

### 4 · En un navegador de verdad (Chromium, Playwright): 5 s quieto, escribe «ru», 2 s

```
htmx: librería 52182 bytes · peticiones del reloj en 5 s: 5 · búsquedas al escribir «ru»: 1 · resultados: ["Rust","Ruby"] · hora en pantalla: 21:33:14
datastar: librería 33553 bytes · peticiones del reloj en 5 s: 1 · búsquedas al escribir «ru»: 1 · resultados: ["Rust","Ruby"] · hora en pantalla: 21:33:21
```
