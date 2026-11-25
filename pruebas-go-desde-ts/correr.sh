#!/bin/sh
# Corredor de las pruebas «vengo de TypeScript, que me va a doler en Go».
# Se ejecuta DENTRO de contenedores oficiales: en esta maquina App Control bloquea
# los binarios sin firmar, asi que nada de esto se compila en el host.
set -u

linea() { echo; echo "############ $1 ############"; }

linea "ENTORNO"
go version
node --version
echo "typescript $(tsc --version 2>/dev/null || npx --no-install tsc --version)"

# ─────────────────────────── TYPESCRIPT ───────────────────────────
linea "TS 1 · la excepcion que el tipo NO cuenta"
echo "== compila? =="
tsc --strict --noEmit /pruebas/ts/t1_excepciones.ts && echo "   tsc: SIN ERRORES (rc=0)"
echo "== y al ejecutarlo? =="
node --experimental-strip-types /pruebas/ts/t1_excepciones.ts 2>&1 | head -4

linea "TS 2 · uniones de literales: el compilador conoce los valores"
tsc --strict --noEmit /pruebas/ts/t2_uniones.ts 2>&1 | head -4
echo "   rc=$?"

linea "TS 3 · no hay valor cero: hay 'todavia no vale nada'"
tsc --strict --noEmit /pruebas/ts/t3_valor_cero.ts 2>&1 | head -4

linea "TS 4 · strictNullChecks: null es parte del tipo"
tsc --strict --noEmit /pruebas/ts/t4_null.ts 2>&1 | head -4

# ─────────────────────────── GO ───────────────────────────
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

linea "FIN"
