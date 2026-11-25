#!/bin/sh
# e09 · empaquetar: los tamanos de verdad y el precio del cgo en la compilacion cruzada.
set -u
cd /app
t() { echo; echo "--- $1 ---"; }

echo "############ e09 · empaquetar y repartir ############"

t "1 · el binario tal cual (lo que se midio en el episodio 1)"
CGO_ENABLED=1 go build -o /tmp/normal ./etapas/e01-ventana
echo "    normal:                 $(stat -c%s /tmp/normal) bytes"

t "2 · quitando la informacion de depuracion"
CGO_ENABLED=1 go build -ldflags="-s -w" -o /tmp/pelado ./etapas/e01-ventana
echo "    con -ldflags='-s -w':   $(stat -c%s /tmp/pelado) bytes"

t "3 · y comprimido, que es lo que se descarga de verdad"
gzip -9 -c /tmp/pelado > /tmp/pelado.gz
echo "    ese mismo, comprimido:  $(stat -c%s /tmp/pelado.gz) bytes"

t "4 · el precio del cgo: apagarlo"
if CGO_ENABLED=0 go build -o /tmp/sincgo ./etapas/e01-ventana 2>/tmp/e0.txt; then
    echo "    [?] compilo sin cgo, lo cual seria raro"
else
    echo "    NO compila. Lo que dice:"
    grep -v "^#" /tmp/e0.txt | head -4 | sed 's/^/      /'
fi

t "5 · el precio del cgo: cruzar a Windows desde Linux"
if CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o /tmp/win.exe ./etapas/e01-ventana 2>/tmp/e1.txt; then
    echo "    [?] cruzo, lo cual seria raro sin toolchain"
else
    echo "    NO compila. Lo que dice:"
    grep -v "^#" /tmp/e1.txt | head -4 | sed 's/^/      /'
fi

t "6 · y para comparar: un programa SIN Fyne si cruza gratis"
mkdir -p /tmp/hola && printf 'package main\nimport "fmt"\nfunc main(){fmt.Println("hola")}\n' > /tmp/hola/main.go
cd /tmp/hola && go mod init hola >/dev/null 2>&1
if CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /tmp/hola.exe . 2>/tmp/e2.txt; then
    echo "    hola.exe para Windows:  $(stat -c%s /tmp/hola.exe) bytes  [ok] sin instalar nada"
else
    echo "    [X] no cruzo:"; head -3 /tmp/e2.txt | sed 's/^/      /'
fi
echo
echo "############ FIN ############"
