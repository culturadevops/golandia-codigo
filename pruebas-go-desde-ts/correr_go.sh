#!/bin/sh
set -u
linea() { echo; echo "############ $1 ############"; }
linea "ENTORNO GO"
go version
linea "GO 1 · el error es un valor, y el compilador te obliga a recogerlo"
echo "== ignorar el error, como se ignora una excepcion en TS =="
go run /pruebas/go/g1_errores.go 2>&1 | head -4
echo "== y asi es como se escribe de verdad =="
go run /pruebas/go/g1b_errores_bien.go 2>&1 | head -4
linea "GO 2 · aqui NO hay uniones de literales"
go run /pruebas/go/g2_uniones.go 2>&1 | head -6
linea "GO 3 · el valor cero, y por que duele en JSON"
go run /pruebas/go/g3_valor_cero.go 2>&1 | head -12
linea "GO 4 · nil no es null, y no se porta igual segun el tipo"
go run /pruebas/go/g4_nil.go 2>&1 | head -8
