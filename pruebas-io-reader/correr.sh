#!/bin/bash
# Dentro de golang:1.27.2: compila las demos de io.Reader y las corre. Los archivos de prueba van en /tmp del contenedor.
set -e
cd /trabajo/lector && go build -o /tmp/lector .
echo '$ go version'; go version
echo; echo '$ ./lector propio'; /tmp/lector propio
/tmp/lector generar /tmp/datos.txt 500
echo; echo '$ ls -l datos.txt'; ls -l /tmp/datos.txt | awk '{print $5, $9}'
echo; echo '$ ./lector todo datos.txt     # os.ReadFile + sha256.Sum256'; /tmp/lector todo /tmp/datos.txt
echo; echo '$ ./lector stream datos.txt   # io.Copy(sha256.New(), archivo)'; /tmp/lector stream /tmp/datos.txt
echo; echo '$ sha256sum datos.txt'; sha256sum /tmp/datos.txt | cut -c1-64
echo; echo '$ ./lector cadena datos.txt   # archivo -> gzip -> contador -> hash + archivo.gz'; /tmp/lector cadena /tmp/datos.txt
echo; echo '$ ./lector lineas datos.txt   # bufio.Scanner, línea a línea'; /tmp/lector lineas /tmp/datos.txt
echo; echo '$ wc -l datos.txt'; wc -l < /tmp/datos.txt
echo; echo '$ ./lector linea-larga   # una línea de 100 KB'; /tmp/lector linea-larga
echo; echo '$ ./lector limite'; /tmp/lector limite
echo; echo '$ go doc -all bufio | grep "MaxScanTokenSize = "'; (cd /tmp && go doc -all bufio | grep "MaxScanTokenSize = ")
echo; echo '$ ./lector trampa'; /tmp/lector trampa
echo; echo '$ ./lector bucle   # mismo Reader, dos bucles'; /tmp/lector bucle
echo; echo '$ # ¿cuántos tipos de la librería estándar tienen un método Read([]byte)? (sin tests, vendor ni testdata)'
R='^func \([^)]*\) Read\(\w+ \[\]byte\)'
echo "\$ grep -rE '$R' \$(go env GOROOT)/src --include=*.go --exclude=*_test.go | grep -v -e /vendor/ -e /testdata/ | wc -l"
grep -rE "$R" $(go env GOROOT)/src --include=*.go --exclude=*_test.go | grep -v -e /vendor/ -e /testdata/ | wc -l
echo '$ # los 10 paquetes con más'
grep -rE "$R" $(go env GOROOT)/src --include=*.go --exclude=*_test.go | grep -v -e /vendor/ -e /testdata/ | sed -E 's#.*/src/##; s#/[^/]*\.go:.*##' | sort | uniq -c | sort -rn | head -10
