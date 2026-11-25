#!/bin/bash
# Corre DENTRO de golang:1.26 (lo lanza ejecutar.sh). Compila la demo y enseña, con curl, lo que viaja por la red en cada librería.
set -u
rm -rf /w && mkdir -p /w && cp -r /trabajo/app /w/app && cd /w/app
mkdir -p static && cp /trabajo/htmx.min.js /trabajo/datastar.js static/
go mod init demo >/dev/null 2>&1
go get github.com/starfederation/datastar-go@v1.2.2 >/dev/null 2>&1
go mod tidy >/dev/null 2>&1
echo "### 1 · Las dos librerías (servidas desde la app, sin CDN)"
echo "\$ wc -c static/htmx.min.js static/datastar.js"
wc -c static/htmx.min.js static/datastar.js
echo "\$ gzip -9 -c static/htmx.min.js | wc -c ; gzip -9 -c static/datastar.js | wc -c"
gzip -9 -c static/htmx.min.js | wc -c
gzip -9 -c static/datastar.js | wc -c
echo "\$ go list -m all | wc -l   # módulos: net/http + datastar-go"
go list -m all | wc -l
go build -o /tmp/demo . || exit 1
/tmp/demo > /tmp/demo.log 2>&1 &
for i in $(seq 1 50); do curl -s -o /dev/null localhost:8080/htmx && break; sleep 0.1; done

echo
echo "### 2 · Buscar «go»: lo que responde el servidor"
echo "\$ curl -s 'localhost:8080/htmx/buscar?q=go'"
curl -s 'localhost:8080/htmx/buscar?q=go'; echo
echo "\$ curl -s 'localhost:8080/datastar/buscar?datastar={\"q\":\"go\"}'"
curl -s 'localhost:8080/datastar/buscar?datastar=%7B%22q%22%3A%22go%22%7D'; echo
echo "\$ curl -s -o /dev/null -w '%{size_download} bytes\n' …   # el tamaño de cada respuesta"
curl -s -o /dev/null -w 'htmx: %{size_download} bytes\n' 'localhost:8080/htmx/buscar?q=go'
curl -s -o /dev/null -w 'datastar: %{size_download} bytes\n' 'localhost:8080/datastar/buscar?datastar=%7B%22q%22%3A%22go%22%7D'

echo
echo "### 3 · El reloj: HTMX pregunta, Datastar empuja"
echo "\$ curl -s localhost:8080/htmx/hora   # HTMX: una petición por segundo"
curl -s localhost:8080/htmx/hora; echo
echo "\$ curl -sN --max-time 3 localhost:8080/datastar/hora   # Datastar: UNA conexión, el servidor manda cada segundo"
curl -sN --max-time 3 localhost:8080/datastar/hora
echo
kill %1 2>/dev/null
