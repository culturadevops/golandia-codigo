# SALIDA — curso «Herramientas de línea de comandos en Go, de cero a release»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-10-28 18:10:36Z`
- `go version go1.27.2 linux/amd64`
- `jq: jq-1.7`
- `GitVersion:    2.18.1`
- `cobra: v1.10.2`
- `go-internal: v1.16.0`
- `Linux 6.18.33.2-microsoft-standard-WSL2 x86_64`

## Lección 01 — `01_flag`

`bash 01_flag/demo.sh` · código de salida `0`

```
== 1.1 · Lo normal: flags primero, URLs después ==
$ time vigia01 -timeout 1s $SRV/ok $SRV/lento
ARRIBA  200      0.5 ms  http://127.0.0.1:8080/ok
CAIDO   ---   1000.3 ms  http://127.0.0.1:8080/lento  (timeout)
[tardó 1.0 s]

== 1.2 · Trampa: un flag DESPUÉS de una URL ya no es un flag ==
$ time vigia01 -v $SRV/lento -timeout 1s
# entendí: timeout=10s v=true urls=["http://127.0.0.1:8080/lento" "-timeout" "1s"]
ARRIBA  200   3002.7 ms  http://127.0.0.1:8080/lento
CAIDO   ---      0.0 ms  -timeout  (Get "-timeout": unsupported protocol scheme "")
CAIDO   ---      0.0 ms  1s  (Get "1s": unsupported protocol scheme "")
[tardó 3.0 s]

== 1.3 · Trampa: un flag booleano no se lleva el valor que va detrás ==
$ vigia01 -v false $SRV/ok
# entendí: timeout=10s v=true urls=["false" "http://127.0.0.1:8080/ok"]
CAIDO   ---      0.1 ms  false  (Get "false": unsupported protocol scheme "")
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok
$ vigia01 -v=false $SRV/ok
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok

== 1.4 · Lo que flag te da gratis: -h, valores con tipo y el código 2 ==
$ vigia01 -h
Usage of vigia01:
  -timeout duration
    	tiempo máximo por endpoint (default 10s)
  -v	cuenta lo que entendió de los argumentos
  -version
    	muestra la versión y sale
[código de salida: 0]
$ vigia01 -timeout 5 $SRV/ok
invalid value "5" for flag -timeout: parse error
Usage of vigia01:
  -timeout duration
    	tiempo máximo por endpoint (default 10s)
  -v	cuenta lo que entendió de los argumentos
  -version
    	muestra la versión y sale
[código de salida: 2]
$ vigia01 --timeout=1s $SRV/ok
ARRIBA  200      0.5 ms  http://127.0.0.1:8080/ok
[código de salida: 0]
```

## Lección 02 — `02_subcomandos`

`bash 02_subcomandos/demo.sh` · código de salida `0`

```
== 2.1 · Cada subcomando con sus flags ==
$ vigia02 revisar -timeout 1s $SRV/ok $SRV/roto
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok
CAIDO   500      0.2 ms  http://127.0.0.1:8080/roto  (Internal Server Error)
[código de salida: 0]
$ vigia02 version
vigia02 dev
[código de salida: 0]

== 2.2 · Un flag delante del subcomando no es de nadie ==
$ vigia02 -timeout 1s revisar $SRV/ok
vigia02: subcomando desconocido "-timeout"

uso: vigia02 <subcomando> [flags]

subcomandos:
  revisar   revisa uno o varios endpoints HTTP
  version   muestra la versión

"vigia02 <subcomando> -h" da la ayuda de cada uno.
[código de salida: 2]

== 2.3 · Un subcomando mal escrito ==
$ vigia02 revsar $SRV/ok
vigia02: subcomando desconocido "revsar"

uso: vigia02 <subcomando> [flags]

subcomandos:
  revisar   revisa uno o varios endpoints HTTP
  version   muestra la versión

"vigia02 <subcomando> -h" da la ayuda de cada uno.
[código de salida: 2]

== 2.4 · La ayuda de un subcomando ==
$ vigia02 revisar -h
Usage of revisar:
  -timeout duration
    	tiempo máximo por endpoint (default 10s)
[código de salida: 0]

== 2.5 · Cuánto código es el despachador ==
$ grep -c "" 02_subcomandos/main.go
55
```

## Lección 03 — `03_cobra`

`bash 03_cobra/demo.sh` · código de salida `0`

```
== 3.1 · Lo mismo que la lección 2 ==
$ vigia03 revisar --timeout 1s $SRV/ok $SRV/roto
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok
CAIDO   500      0.1 ms  http://127.0.0.1:8080/roto  (Internal Server Error)
[código de salida: 0]

== 3.2 · Con Cobra (pflag) los flags van donde quieras ==
$ vigia03 revisar $SRV/ok -t 1s
ARRIBA  200      0.3 ms  http://127.0.0.1:8080/ok
[código de salida: 0]

== 3.3 · Un subcomando mal escrito: Cobra sugiere ==
$ vigia03 revsar $SRV/ok
Error: unknown command "revsar" for "vigia03"

Did you mean this?
	revisar

Run 'vigia03 --help' for usage.
[código de salida: 2]

== 3.4 · Ojo: en pflag un guion es un flag corto; -timeout es -t con el valor «imeout» ==
$ vigia03 revisar -timeout 1s $SRV/ok 2>&1 | head -1
Error: invalid argument "imeout" for "-t, --timeout" flag: time: invalid duration "imeout"
$ vigia03 revisar -timeout 1s $SRV/ok > /dev/null 2>&1
[código de salida: 2]

== 3.5 · Tamaño del binario (bytes) ==
  t02         8818880 bytes
  t03         9610571 bytes
  t02-sw      5972231 bytes
  t03-sw      6549767 bytes

== 3.6 · Paquetes que entran en el binario ==
  flag (02)       193 paquetes
  cobra (03)      200 paquetes

== 3.7 · Arranque: «version» lanzado 300 veces ==
  flag : mediana 1.26 ms   p90 1.46 ms   (300 arranques)
  cobra: mediana 1.31 ms   p90 1.46 ms   (300 arranques)
  flag : mediana 1.29 ms   p90 1.46 ms   (300 arranques)   (repetido, para ver el ruido)
  cobra: mediana 1.38 ms   p90 1.60 ms   (300 arranques)   (repetido, para ver el ruido)

== 3.8 · Líneas de código de cada versión ==
$ grep -c "" 02_subcomandos/main.go 03_cobra/main.go
02_subcomandos/main.go:55
03_cobra/main.go:47
```

## Lección 04 — `04_config`

`bash 04_config/demo.sh` · código de salida `0`

```
== 4.1 · Sin nada: valores por defecto ==
$ vigia config
timeout       10s    (defecto)
concurrencia  4      (defecto)
$ ingenuo04
A (el flag al final)      timeout=10s
B (comparar con defecto)  timeout=10s

== 4.2 · Con archivo vigia.json ==
$ cat vigia.json
{"timeout": "5s", "concurrencia": 8}
$ vigia config
timeout       5s     (archivo vigia.json)
concurrencia  8      (archivo vigia.json)
$ ingenuo04
A (el flag al final)      timeout=10s
B (comparar con defecto)  timeout=5s

== 4.3 · Archivo + entorno ==
$ VIGIA_TIMEOUT=3s vigia config
timeout       3s     (entorno VIGIA_TIMEOUT)
concurrencia  8      (archivo vigia.json)
$ VIGIA_TIMEOUT=3s ingenuo04
A (el flag al final)      timeout=10s
B (comparar con defecto)  timeout=3s

== 4.4 · Archivo + entorno + flag ==
$ VIGIA_TIMEOUT=3s vigia config -timeout 1s
timeout       1s     (flag -timeout)
concurrencia  8      (archivo vigia.json)
$ VIGIA_TIMEOUT=3s ingenuo04 -timeout 1s
A (el flag al final)      timeout=1s
B (comparar con defecto)  timeout=1s

== 4.5 · El flag puesto a mano con el MISMO valor que el defecto ==
$ VIGIA_TIMEOUT=3s vigia config -timeout 10s
timeout       10s    (flag -timeout)
concurrencia  8      (archivo vigia.json)
$ VIGIA_TIMEOUT=3s ingenuo04 -timeout 10s
A (el flag al final)      timeout=10s
B (comparar con defecto)  timeout=3s

== 4.6 · Un valor mal escrito en el entorno (falta la unidad) ==
$ VIGIA_TIMEOUT=3 vigia config
vigia config: VIGIA_TIMEOUT="3": time: missing unit in duration "3"
[código de salida: 2]
$ VIGIA_TIMEOUT=3 ingenuo04
A (el flag al final)      timeout=10s
B (comparar con defecto)  timeout=5s
[código de salida: 0]

== 4.7 · El archivo por defecto es opcional; el que pides, no ==
$ vigia config
timeout       10s    (defecto)
concurrencia  4      (defecto)
[código de salida: 0]
$ vigia config -config /no/existe.json
vigia config: open /no/existe.json: no such file or directory
[código de salida: 2]
```

## Lección 05 — `05_stdio`

`bash 05_stdio/demo.sh` · código de salida `0`

```
== 5.1 · En la terminal se ve todo junto ==
$ vigia revisar $SRV/ok $SRV/roto
revisando 2 endpoints (timeout 10s, concurrencia 4)
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok
CAIDO   500      0.4 ms  http://127.0.0.1:8080/roto  (Internal Server Error)
1 de 2 caídos

== 5.2 · Pero son dos canales: stderr fuera ==
$ vigia revisar $SRV/ok $SRV/roto 2>/dev/null
ARRIBA  200      0.3 ms  http://127.0.0.1:8080/ok
CAIDO   500      0.3 ms  http://127.0.0.1:8080/roto  (Internal Server Error)

== 5.3 · Contar líneas en un pipe ==
$ vigia revisar $SRV/ok $SRV/roto 2>/dev/null | wc -l
2
$ ingenuo05 $SRV/ok $SRV/roto 2>/dev/null | wc -l
3

== 5.4 · Guardar en un archivo: lo que queda dentro ==
$ vigia revisar $SRV/ok 2>/dev/null > "$TMP/bueno.txt"; cat -v "$TMP/bueno.txt"
ARRIBA  200      0.3 ms  http://127.0.0.1:8080/ok
$ ingenuo05 $SRV/ok > "$TMP/ingenuo.txt"; cat -v "$TMP/ingenuo.txt"
revisando 1 endpoints (timeout 10s)
^[[32mARRIBA^[[0m  200      0.3 ms  http://127.0.0.1:8080/ok

== 5.5 · Con una terminal de verdad delante (script le da una), vigia sí pinta ==
$ script -qc "vigia revisar $SRV/ok 2>/dev/null" /dev/null | cat -v
^[[32mARRIBA^[[0m  200      0.3 ms  http://127.0.0.1:8080/ok^M
$ NO_COLOR=1 script -qc "vigia revisar $SRV/ok 2>/dev/null" /dev/null | cat -v
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok^M

== 5.6 · Trampa: ¿es una terminal? /dev/null también es un dispositivo de caracteres ==
$ esterminal | cat
  os.ModeCharDevice: false   term.IsTerminal: false
$ esterminal > "$TMP/archivo.txt"
  os.ModeCharDevice: false   term.IsTerminal: false
$ esterminal > /dev/null
  os.ModeCharDevice: true    term.IsTerminal: false
$ script -qc "esterminal" /dev/null
  os.ModeCharDevice: true    term.IsTerminal: true

== 5.7 · Las URLs por stdin, con «-» ==
$ cat 05_stdio/urls.txt
# endpoints del servidor de pruebas (go run ./banco)
http://127.0.0.1:8080/ok
http://127.0.0.1:8080/roto
http://127.0.0.1:8080/lento?ms=300

# las líneas vacías y las que empiezan con # no cuentan
http://127.0.0.1:8080/ok?de=nuevo
$ vigia revisar - < 05_stdio/urls.txt
revisando 4 endpoints (timeout 10s, concurrencia 4)
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok
CAIDO   500      0.4 ms  http://127.0.0.1:8080/roto  (Internal Server Error)
ARRIBA  200    301.1 ms  http://127.0.0.1:8080/lento?ms=300
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok?de=nuevo
1 de 4 caídos
[código de salida: 1]
$ grep -v roto 05_stdio/urls.txt | vigia revisar - 2>/dev/null
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok
ARRIBA  200    301.0 ms  http://127.0.0.1:8080/lento?ms=300
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok?de=nuevo
[código de salida: 0]
```

## Lección 06 — `06_salida`

`bash 06_salida/demo.sh` · código de salida `0`

```
== 6.1 · Tres códigos, tres situaciones ==
$ vigia revisar $SRV/ok 2>/dev/null
ARRIBA  200      0.5 ms  http://127.0.0.1:8080/ok
[código de salida: 0]
$ vigia revisar $SRV/ok $SRV/roto 2>/dev/null
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok
CAIDO   500      0.5 ms  http://127.0.0.1:8080/roto  (Internal Server Error)
[código de salida: 1]
$ vigia revisar -timeout 3 $SRV/ok
invalid value "3" for flag -timeout: parse error
Usage of revisar:
  -concurrencia int
    	endpoints revisados a la vez (default 4)
  -config string
    	archivo de configuración (por defecto ./vigia.json)
  -informe string
    	además, guarda el informe en este archivo
  -json
    	salida JSON para scripts
  -solo-caidos
    	muestra solo los endpoints caídos
  -timeout duration
    	tiempo máximo por endpoint (default 10s)
[código de salida: 2]

== 6.2 · Ingenuo: log.Fatal en la configuración da el mismo 1 que un caído ==
$ ingenuo06 $SRV/roto
[código de salida: 1]
$ ingenuo06 -timeout 3 $SRV/ok
2026/10/28 18:09:41 time: missing unit in duration "3"
[código de salida: 1]

== 6.3 · Ingenuo: os.Exit se salta el defer, y el búfer se pierde ==
$ ingenuo06 $SRV/ok
ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok
[código de salida: 0]
$ ingenuo06 $SRV/ok $SRV/roto
[código de salida: 1]
$ ingenuo06 $SRV/ok $SRV/roto | wc -l
0

== 6.4 · En un script: if y case sobre el código ==
$ bash "$TMP/vigilar.sh" $SRV/ok
ARRIBA  200      0.5 ms  http://127.0.0.1:8080/ok
todo arriba
$ bash "$TMP/vigilar.sh" $SRV/ok $SRV/roto
ARRIBA  200      0.3 ms  http://127.0.0.1:8080/ok
CAIDO   500      0.4 ms  http://127.0.0.1:8080/roto  (Internal Server Error)
hay caídos: aviso a guardia
$ bash "$TMP/vigilar.sh" -timeout 3 $SRV/ok
el script está mal llamado: esto lo arregla quien lo escribió

== 6.5 · Trampa de Bash: en un pipe, el código es el del ÚLTIMO comando ==
$ vigia revisar $SRV/roto 2>/dev/null | tee "$TMP/informe.txt" > /dev/null
[código de salida: 0]
$ set -o pipefail; vigia revisar $SRV/roto 2>/dev/null | tee "$TMP/informe.txt" > /dev/null
[código de salida: 1]
$ vigia revisar $SRV/roto 2>/dev/null | tee "$TMP/informe.txt" > /dev/null; echo "PIPESTATUS: ${PIPESTATUS[*]}"
PIPESTATUS: 1 0

== 6.6 · set -e tampoco ve lo que pasa a la izquierda del pipe ==
$ bash -c "set -e; vigia revisar $SRV/roto 2>/dev/null | wc -l; echo sigo vivo"
1
sigo vivo
$ bash -c "set -eo pipefail; vigia revisar $SRV/roto 2>/dev/null | wc -l; echo sigo vivo"; echo "[código de salida: $?]"
1
[código de salida: 1]
```

## Lección 07 — `07_json`

`bash 07_json/demo.sh` · código de salida `0`

```
jq: jq-1.7

== 7.1 · La misma revisión, para personas y para scripts ==
$ vigia revisar $SRV/ok "$SRV/lento?ms=250" $SRV/roto 2>/dev/null
ARRIBA  200      0.5 ms  http://127.0.0.1:8080/ok
ARRIBA  200    251.2 ms  http://127.0.0.1:8080/lento?ms=250
CAIDO   500      0.4 ms  http://127.0.0.1:8080/roto  (Internal Server Error)
$ vigia revisar -json $SRV/ok "$SRV/lento?ms=250" $SRV/roto 2>/dev/null
[
  {
    "url": "http://127.0.0.1:8080/ok",
    "ok": true,
    "estado": 200,
    "latencia_ms": 0.486
  },
  {
    "url": "http://127.0.0.1:8080/lento?ms=250",
    "ok": true,
    "estado": 200,
    "latencia_ms": 251.047
  },
  {
    "url": "http://127.0.0.1:8080/roto",
    "ok": false,
    "estado": 500,
    "latencia_ms": 0.516,
    "error": "Internal Server Error"
  }
]

== 7.2 · jq: las URLs caídas, y la más lenta ==
$ vigia revisar -json $SRV/ok "$SRV/lento?ms=250" $SRV/roto 2>/dev/null | jq -r ".[] | select(.ok | not) | .url"
http://127.0.0.1:8080/roto
$ vigia revisar -json $SRV/ok "$SRV/lento?ms=250" $SRV/roto 2>/dev/null | jq -r "max_by(.latencia_ms) | \"\(.url) \(.latencia_ms) ms\""
http://127.0.0.1:8080/lento?ms=250 251.105 ms

== 7.3 · Ingenuo: time.Duration en JSON son nanosegundos ==
$ ingenuo07 "$SRV/lento?ms=250"
[{"url":"http://127.0.0.1:8080/lento?ms=250","ok":true,"latencia":251286075}]
$ ingenuo07 "$SRV/lento?ms=250" | jq ".[0].latencia > 1000"
true

== 7.4 · Ingenuo: sin caídos, el JSON es null y no [] ==
$ ingenuo07 -solo-caidos $SRV/ok
null
$ ingenuo07 -solo-caidos $SRV/ok | jq -r ".[].url"
jq: error (at <stdin>:1): Cannot iterate over null (null)
[código de salida: 5]
$ vigia revisar -json -solo-caidos $SRV/ok 2>/dev/null
[]
$ vigia revisar -json -solo-caidos $SRV/ok 2>/dev/null | jq -r ".[].url"
[código de salida: 0]

== 7.5 · Ingenuo: un aviso por stdout rompe el JSON entero ==
$ ingenuo07 -avisar $SRV/ok "$SRV/lento?ms=250" | jq length
jq: parse error: Invalid numeric literal at line 1, column 6
[código de salida: 5]
$ vigia revisar -json $SRV/ok "$SRV/lento?ms=250" | jq length
revisando 2 endpoints (timeout 10s, concurrencia 4)
2
```

## Lección 08 — `08_senales`

`bash 08_senales/demo.sh` · código de salida `0`

```
== 8.1 · Ingenuo, sin escuchar señales: Ctrl+C lo mata en el acto ==
$ ingenuo08 -informe informe.txt http://127.0.0.1:8080/ok http://127.0.0.1:8080/lento?ms=5000      # y a los 0,5 s: kill -INT
  salió con código 130, 1 ms después de la señal
  informe.txt: 34 bytes -> 200  0s  http://127.0.0.1:8080/ok|
  temporales que quedaron: 0

== 8.2 · Ingenuo con NotifyContext, pero la petición no lleva el ctx ==
$ ingenuo08 -escucha -informe informe.txt http://127.0.0.1:8080/ok http://127.0.0.1:8080/lento?ms=5000      # y a los 0,5 s: kill -INT
  stderr: interrumpido
  salió con código 130, 4506 ms después de la señal
  informe.txt: 83 bytes -> 200  0s  http://127.0.0.1:8080/ok|200  5.004s  http://127.0.
  temporales que quedaron: 0

== 8.3 · vigia: ctx en cada petición e informe atómico ==
$ vigia revisar -informe informe.txt http://127.0.0.1:8080/ok http://127.0.0.1:8080/lento?ms=5000      # y a los 0,5 s: kill -INT
  stderr: revisando 2 endpoints (timeout 10s, concurrencia 4)
  stderr: vigia: interrumpido, no escribo nada a medias
  salió con código 130, 1 ms después de la señal
  informe.txt: 31 bytes -> informe de ayer: 2 de 2 arriba|
  temporales que quedaron: 0

== 8.4 · vigia con SIGTERM (lo que manda docker stop o Kubernetes) ==
$ vigia revisar -informe informe.txt http://127.0.0.1:8080/ok http://127.0.0.1:8080/lento?ms=5000      # y a los 0,5 s: kill -TERM
  stderr: revisando 2 endpoints (timeout 10s, concurrencia 4)
  stderr: vigia: interrumpido, no escribo nada a medias
  salió con código 130, 1 ms después de la señal
  informe.txt: 31 bytes -> informe de ayer: 2 de 2 arriba|
  temporales que quedaron: 0

== 8.5 · Y sin interrumpir, el informe nuevo sustituye al viejo de una vez ==
$ vigia revisar -informe informe.txt $SRV/ok $SRV/roto 2>/dev/null >/dev/null
[código de salida: 1]
  informe.txt: 127 bytes -> ARRIBA  200      0.4 ms  http://127.0.0.1:8080/ok|CAIDO   50
  temporales que quedaron: 0
```

## Lección 09 — `09_tests`

`bash 09_tests/demo.sh` · código de salida `0`

```
== 9.1 · Los tests de vigia ==
$ go test -count=1 ./vigia
ok  	curso-cli/vigia	0.123s

== 9.2 · Uno por uno, con lo que tarda cada uno ==
$ go test -count=1 -v ./vigia 2>&1 | grep -E "^ *--- (PASS|FAIL)|^ok"
--- PASS: TestCodigosDeSalida (0.10s)
    --- PASS: TestCodigosDeSalida/todo_arriba (0.00s)
    --- PASS: TestCodigosDeSalida/uno_caído (0.00s)
    --- PASS: TestCodigosDeSalida/timeout (0.10s)
    --- PASS: TestCodigosDeSalida/sin_URLs (0.00s)
    --- PASS: TestCodigosDeSalida/flag_mal (0.00s)
    --- PASS: TestCodigosDeSalida/entorno_mal (0.00s)
    --- PASS: TestCodigosDeSalida/urls_por_stdin (0.00s)
    --- PASS: TestCodigosDeSalida/subcomando_mal (0.00s)
    --- PASS: TestCodigosDeSalida/ayuda (0.00s)
--- PASS: TestJSONLimpio (0.00s)
--- PASS: TestPrecedencia (0.00s)
--- PASS: TestAyudaGolden (0.00s)
--- PASS: TestScripts (0.00s)
    --- PASS: TestScripts/config (0.01s)
    --- PASS: TestScripts/salida (0.01s)
ok  	curso-cli/vigia	0.118s

== 9.3 · Un cambio en la ayuda sin actualizar el golden (en una copia del repo) ==
$ (cd "$TMP/repo" && go test -count=1 -run Ayuda ./vigia 2>&1 | head -8)
--- FAIL: TestAyudaGolden (0.00s)
    run_test.go:127: la ayuda cambió; si es a propósito, -update
        --- quería
        vigia revisa endpoints HTTP.
        
        uso:
          vigia revisar [flags] URL...     (o "-" para leer las URLs de stdin)
          vigia config  [flags]            la configuración efectiva, y de dónde sale
$ (cd "$TMP/repo" && go test -count=1 -run Ayuda ./vigia -update && git diff --no-index --stat vigia/testdata/ayuda.golden "$RAIZ/vigia/testdata/ayuda.golden")
ok  	curso-cli/vigia	0.002s
 {vigia => /work/repo/vigia}/testdata/ayuda.golden | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)

== 9.4 · Con tu entorno exportado (VIGIA_TIMEOUT=1ms): los tests de vigia no se enteran ==
$ VIGIA_TIMEOUT=1ms go test -count=1 ./vigia
ok  	curso-cli/vigia	0.120s

== 9.5 · El paquete ingenuo, que lee os.Getenv por su cuenta ==
$ go test -count=1 ./09_tests/ingenuo
ok  	curso-cli/09_tests/ingenuo	0.002s
$ VIGIA_TIMEOUT=1ms go test -count=1 ./09_tests/ingenuo
--- FAIL: TestTimeoutPorDefecto (0.00s)
    timeout_test.go:11: timeout por defecto: 1ms, quería 10s
FAIL
FAIL	curso-cli/09_tests/ingenuo	0.001s
FAIL

== 9.6 · Cobertura ==
$ go test -count=1 -cover ./vigia
ok  	curso-cli/vigia	0.122s	coverage: 71.1% of statements

== 9.7 · Trampa: defer srv.Close() en vez de t.Cleanup (en una copia del repo) ==
$ (cd "$TMP/repo" && grep -n "srv.Close" vigia/script_test.go)
25:	defer srv.Close()
$ (cd "$TMP/repo" && go test -count=1 -run Scripts ./vigia > "$TMP/test.txt" 2>&1; grep -E "^--- |^    --- |^(ok|FAIL)" "$TMP/test.txt")
--- FAIL: TestScripts (0.00s)
    --- FAIL: TestScripts/salida (0.00s)
FAIL
FAIL	curso-cli/vigia	0.012s
FAIL
$ grep -o "dial tcp [^ ]*: connect: connection refused" "$TMP/test.txt" | head -1
dial tcp 127.0.0.1:39925: connect: connection refused
```

## Lección 10 — `10_version`

`bash 10_version/demo.sh` · código de salida `0`

```
== 10.1 · go run: sin datos del repositorio ==
$ go run ./vigia version
vigia dev
  go            go1.27.2
  módulo        curso-cli (devel)
  CGO_ENABLED   1
  GOARCH        amd64
  GOOS          linux

== 10.2 · go build dentro de un repositorio git: commit, fecha y si está limpio ==
$ go build -o "$TMP/vigia" ./vigia && "$TMP/vigia" version
vigia dev
  go            go1.27.2
  módulo        curso-cli v0.0.0-20261028180953-d59f4b45821d
  CGO_ENABLED   1
  GOARCH        amd64
  GOOS          linux
  vcs.revision  d59f4b45821d71c4de2b4b446e4c5af150ee2fab
  vcs.time      2026-10-28T18:09:53Z
  vcs.modified  false

== 10.3 · Con una etiqueta, go build pone la versión del módulo solo ==
$ go build -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | head -4
vigia dev
  go            go1.27.2
  módulo        curso-cli v1.4.0
  CGO_ENABLED   1

== 10.4 · Un cambio sin commit: +dirty y vcs.modified ==
$ go build -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | grep -E "módulo|modified"
  módulo        curso-cli v1.4.0+dirty
  vcs.modified  true

== 10.5 · -ldflags -X: la versión que tú decides ==
$ go build -ldflags "-X main.version=1.4.0 -X main.commit=$(git rev-parse --short HEAD) -X main.fecha=2026-10-28" -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | head -2
vigia 1.4.0
  commit        d59f4b4 (2026-10-28)

== 10.6 · Trampa: un nombre mal escrito en -X no da error ==
$ go build -ldflags "-X main.Version=1.4.0" -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | head -1
vigia dev
[código de salida: 0]
$ go build -ldflags "-X curso-cli/vigia.version=1.4.0" -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | head -1
vigia dev
[código de salida: 0]

== 10.7 · Leerla desde fuera, sin ejecutar el binario ==
$ go build -o "$TMP/vigia" ./vigia && go version -m "$TMP/vigia" | grep -E "path|mod|vcs\.(revision|modified)|GOOS|GOARCH"
	path	curso-cli/vigia
	mod	curso-cli	v1.4.0	
	build	-buildmode=exe
	build	GOARCH=amd64
	build	GOOS=linux
	build	vcs.revision=d59f4b45821d71c4de2b4b446e4c5af150ee2fab
	build	vcs.modified=false
```

## Lección 11 — `11_cruzada`

`bash 11_cruzada/demo.sh` · código de salida `0`

```
== 11.1 · Cinco sistemas desde la misma máquina: GOOS y GOARCH ==
  binario                        normal        -s -w   ahorro
  vigia-darwin-amd64           10222928      6916128      32%
  vigia-darwin-arm64            9518978      6404274      32%
  vigia-linux-amd64            10071522      6762656      32%
  vigia-linux-arm64             9292726      6226080      33%
  vigia-windows-amd64.exe      10323456      7013376      32%

== 11.2 · Qué es cada uno (file) ==
  vigia-darwin-amd64: Mach-O 64-bit x86_64 executable, flags:<|DYLDLINK|PIE>
  vigia-darwin-arm64: Mach-O 64-bit arm64 executable, flags:<|DYLDLINK|PIE>
  vigia-linux-amd64: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked
  vigia-linux-arm64: ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), statically l
  vigia-windows-amd64.exe: PE32+ executable for MS Windows 10.00 (console), x86-64, 8 sections

== 11.3 · Trampa: sin -trimpath, el binario lleva las rutas de tu máquina ==
$ strings "$OUT/normal/vigia-linux-amd64" | grep -c "$RAIZ/"
6
$ strings "$OUT/normal/vigia-linux-amd64" | grep -m1 "$RAIZ/"
/work/repo/comun/comun.go
$ strings "$OUT/fino/vigia-linux-amd64" | grep -c "$RAIZ/"
0

== 11.4 · Trampa: compilado en Linux para Linux, con cgo, sale dinámico ==
$ go env CGO_ENABLED
1
$ file -b "$OUT/con-cgo" | cut -c1-90
ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib6
$ ldd "$OUT/con-cgo"
	linux-vdso.so.1 (0x0000717038889000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x000071703868a000)
	/lib64/ld-linux-x86-64.so.2 (0x000071703888b000)
$ file -b "$OUT/sin-cgo" | cut -c1-90
ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=TsnppGX
$ ldd "$OUT/sin-cgo"
	not a dynamic executable

== 11.5 · -s -w quita los símbolos, no la información de compilación ==
$ go version -m "$OUT/fino/vigia-linux-amd64" | grep -E "^\s+(path|build\s+(-trimpath|-ldflags|GOOS|GOARCH))"
	path	curso-cli/vigia
	build	-trimpath=true
	build	GOARCH=amd64
	build	GOOS=linux
$ "$OUT/fino/vigia-linux-amd64" version | head -3
vigia dev
  go            go1.27.2
  módulo        curso-cli (devel)
```

## Lección 12 — `12_release`

`bash 12_release/demo.sh` · código de salida `0`

```
== 12.1 · La versión de GoReleaser y la configuración ==
$ goreleaser --version 2>&1 | grep -E "^GitVersion"
GitVersion:    2.18.1
$ goreleaser check 2>&1 | tail -2
  • 1 configuration file(s) validated
  • thanks for using GoReleaser!

== 12.2 · Snapshot: construye todo y no publica nada ==
$ goreleaser release --snapshot --clean 2>&1 | grep -E "snapshot|building  |archiving|checksums|succeeded" | sed "s/  */ /g"
 • pipe skipped or partially skipped reason=disabled during snapshot mode
 • snapshotting
 • building snapshot... version=1.0.0
 • building paths=vigia binaries=vigia target=darwin_arm64_v8.0
 • building paths=vigia binaries=vigia target=darwin_amd64_v1
 • building paths=vigia binaries=vigia target=linux_amd64_v1
 • building paths=vigia binaries=vigia target=linux_arm64_v8.0
 • building paths=vigia binaries=vigia target=windows_amd64_v1
 • building paths=vigia binaries=vigia target=windows_arm64_v8.0
 • archiving name=dist/vigia_1.0.0_windows_arm64.zip
 • archiving name=dist/vigia_1.0.0_linux_arm64.tar.gz
 • archiving name=dist/vigia_1.0.0_darwin_amd64.tar.gz
 • archiving name=dist/vigia_1.0.0_linux_amd64.tar.gz
 • archiving name=dist/vigia_1.0.0_darwin_arm64.tar.gz
 • archiving name=dist/vigia_1.0.0_windows_amd64.zip
 • calculating checksums
 • release succeeded after 3s

== 12.3 · Lo que queda en dist/ ==
$ ls dist | grep -vE "_v1$|_v8.0$"
artifacts.json
checksums.txt
config.yaml
metadata.json
vigia_1.0.0_darwin_amd64.tar.gz
vigia_1.0.0_darwin_arm64.tar.gz
vigia_1.0.0_linux_amd64.tar.gz
vigia_1.0.0_linux_arm64.tar.gz
vigia_1.0.0_windows_amd64.zip
vigia_1.0.0_windows_arm64.zip
$ ls -l dist/*.tar.gz dist/*.zip | awk "{print \$5, \$9}"
2954540 dist/vigia_1.0.0_darwin_amd64.tar.gz
2722451 dist/vigia_1.0.0_darwin_arm64.tar.gz
2903813 dist/vigia_1.0.0_linux_amd64.tar.gz
2621765 dist/vigia_1.0.0_linux_arm64.tar.gz
2994114 dist/vigia_1.0.0_windows_amd64.zip
2676301 dist/vigia_1.0.0_windows_arm64.zip

== 12.4 · checksums.txt, y comprobarlo ==
$ cat dist/checksums.txt | cut -c1-16,65-
6d0021c31e8d23e2  vigia_1.0.0_darwin_amd64.tar.gz
3c02a47cc68ab320  vigia_1.0.0_darwin_arm64.tar.gz
90050820e7da0fb5  vigia_1.0.0_linux_amd64.tar.gz
b184d492eefd8853  vigia_1.0.0_linux_arm64.tar.gz
0255389b8f5184ef  vigia_1.0.0_windows_amd64.zip
ca326f99cdb72efc  vigia_1.0.0_windows_arm64.zip
$ (cd dist && sha256sum -c checksums.txt)
vigia_1.0.0_darwin_amd64.tar.gz: OK
vigia_1.0.0_darwin_arm64.tar.gz: OK
vigia_1.0.0_linux_amd64.tar.gz: OK
vigia_1.0.0_linux_arm64.tar.gz: OK
vigia_1.0.0_windows_amd64.zip: OK
vigia_1.0.0_windows_arm64.zip: OK

== 12.5 · El binario de Linux, sacado del archivo ==
$ mkdir -p "$TMP/x" && tar -xzf dist/vigia_1.0.0_linux_amd64.tar.gz -C "$TMP/x" && ls "$TMP/x" && "$TMP/x/vigia" version | head -3
README.md
vigia
vigia 1.0.0
  commit        e0cbd77 (2026-10-28T18:10:24Z)
  go            go1.27.2

== 12.6 · ¿Y el changelog? En snapshot no se genera ==
$ ls dist/CHANGELOG.md
ls: cannot access 'dist/CHANGELOG.md': No such file or directory
[código de salida: 2]

== 12.7 · Sin --snapshot: el changelog sale de los commits entre dos etiquetas ==
$ git log --oneline v0.9.0..v1.0.0 | cut -c9-
docs: README del curso
fix: Ctrl+C ya no deja el informe a medias
feat: salida -json para scripts
$ goreleaser release --clean --skip=publish,announce 2>&1 | grep -E "skipping|using tags|changelog|succeeded|failed" | sed "s/  */ /g"
 • skipping announce and publish...
 • using tags previous=v0.9.0 current=v1.0.0
 • generating changelog
 • release succeeded after 0s
$ cat dist/CHANGELOG.md
## Changelog
* 88a79ce52aae35b559737bf474c41b94f1e9d430 feat: salida -json para scripts
* 73e79ee648cef14668854451f1717ce716e8af9f fix: Ctrl+C ya no deja el informe a medias

== 12.8 · ¿Es reproducible? Dos snapshots seguidos, mismos checksums ==
$ diff "$TMP/a.txt" "$TMP/b.txt" > /dev/null && echo "idénticos" || diff "$TMP/a.txt" "$TMP/b.txt" | cut -c1-20,67- | head -6
idénticos

== 12.9 · Lo mismo quitando mod_timestamp de la configuración ==
$ diff "$TMP/c.txt" "$TMP/d.txt" > /dev/null && echo "idénticos" || echo "distintos: $(diff "$TMP/c.txt" "$TMP/d.txt" | grep -c "^<") de 6 archivos cambian de checksum"
distintos: 6 de 6 archivos cambian de checksum
```
