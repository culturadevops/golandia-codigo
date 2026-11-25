#!/usr/bin/env bash
# demo.sh — la medición del vídeo «go test -fuzz contra mi propio parser».
#
# ⛔ ÉTICO: se fuzzea CÓDIGO PROPIO, que es exactamente para lo que está la herramienta.
#
# El parser tiene un fallo plantado a propósito, de una clase muy real: se fía de la longitud
# que viene DENTRO del mensaje. Así nacen media docena de vulnerabilidades al año.
set -u
cd /lab || exit 1

echo "======================================================================"
echo " 1 · LAS PRUEBAS NORMALES. Las que escribiría cualquiera."
echo "======================================================================"
go test -run "^Test" -v ./... 2>&1 | grep -E "^(=== RUN|--- (PASS|FAIL)|PASS|FAIL|ok)" | sed 's/^/   /'
echo
echo "   >>> el parser está probado. Y tiene un agujero."

echo
echo "======================================================================"
echo " 2 · EL FUZZ. Seis líneas, y a buscar."
echo "======================================================================"
echo "   \$ go test -fuzz='^FuzzParsea$' -fuzztime=60s"
echo
t0=$(date +%s%N)
go test -fuzz='^FuzzParsea$' -fuzztime=60s -run "^$" . > /tmp/fuzz.log 2>&1
SAL=$?
t1=$(date +%s%N)
echo "   salió con $SAL en $(( (t1-t0)/1000000 )) ms"
grep -q "will not fuzz" /tmp/fuzz.log && { echo "⛔ EL FUZZ NO LLEGÓ A EJECUTARSE:"; cat /tmp/fuzz.log; exit 1; }
echo
sed 's/^/   /' /tmp/fuzz.log | head -32

echo
echo "======================================================================"
echo " 3 · 🥇 LA ENTRADA QUE LO TUMBA, y dónde la guarda Go"
echo "======================================================================"
echo "   ficheros nuevos en testdata:"
find testdata -type f 2>/dev/null | sed 's/^/     /'
echo
for f in $(find testdata -type f 2>/dev/null); do
  echo "   contenido de $f:"
  cat "$f" | sed 's/^/     /'
  echo
done

echo "======================================================================"
echo " 4 · Y AHORA ESE FICHERO ES UNA PRUEBA MÁS, para siempre"
echo "======================================================================"
echo "   \$ go test -run FuzzParsea   (sin -fuzz: solo ejecuta el corpus guardado)"
go test -run "FuzzParsea$" . 2>&1 | head -14 | sed 's/^/   /'

echo
echo "======================================================================"
echo " 5 · EL MISMO FUZZ CONTRA LA VERSIÓN CORREGIDA"
echo "======================================================================"
echo "   \$ go test -fuzz='^FuzzParseaBien$' -fuzztime=60s"
t0=$(date +%s%N)
go test -fuzz='^FuzzParseaBien$' -fuzztime=60s -run "^$" . > /tmp/fuzz2.log 2>&1
SAL2=$?
t1=$(date +%s%N)
echo "   salió con $SAL2 en $(( (t1-t0)/1000000 )) ms"
grep -E "elapsed|execs|new interesting|PASS|ok " /tmp/fuzz2.log | tail -6 | sed 's/^/   /'

echo
echo "======================================================================"
echo " 6 · ⛔ ¿Y SIN CORPUS DE SEMILLAS? ¿Lo encuentra igual?"
echo "======================================================================"
rm -rf testdata
cp parser_test.go /tmp/con-semillas.go
# Se quitan las tres llamadas a f.Add del fuzz principal.
sed -i '/^\tf\.Add(\[\]byte{0x01, 0x00, 0x05/d; /^\tf\.Add(\[\]byte{0x02, 0x00, 0x00})$/d; /^\tf\.Add(\[\]byte{0x07, 0x00, 0x03/d' parser_test.go
echo -n "   semillas que quedan en FuzzParsea: "
awk '/func FuzzParsea\(/,/^}/' parser_test.go | grep -c "f.Add" || echo 0
t0=$(date +%s%N)
go test -fuzz='^FuzzParsea$' -fuzztime=60s -run "^$" . > /tmp/fuzz3.log 2>&1
SAL3=$?
t1=$(date +%s%N)
echo "   salió con $SAL3 en $(( (t1-t0)/1000000 )) ms"
grep -E "elapsed|execs|Failing input|panic:|goroutine" /tmp/fuzz3.log | head -6 | sed 's/^/   /'
cp /tmp/con-semillas.go parser_test.go

echo
echo "== fin =="
