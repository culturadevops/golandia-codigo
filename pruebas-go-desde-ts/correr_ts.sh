#!/bin/sh
set -u
linea() { echo; echo "############ $1 ############"; }
linea "ENTORNO TYPESCRIPT"
node --version
npm ls -g --depth=0 2>/dev/null | grep -i typescript || tsc --version
linea "TS 1 · la excepcion que el tipo NO cuenta"
echo "== compila? =="
if tsc --strict --noEmit /pruebas/ts/t1_excepciones.ts; then echo "   tsc: SIN ERRORES (rc=0)"; fi
echo "== y al ejecutarlo? =="
node --experimental-strip-types /pruebas/ts/t1_excepciones.ts 2>&1 | head -4
linea "TS 2 · uniones de literales: el compilador conoce los valores"
tsc --strict --noEmit /pruebas/ts/t2_uniones.ts 2>&1 | head -4
linea "TS 3 · no hay valor cero: hay 'todavia no vale nada'"
tsc --strict --noEmit /pruebas/ts/t3_valor_cero.ts 2>&1 | head -4
linea "TS 4 · strictNullChecks: null es parte del tipo"
tsc --strict --noEmit /pruebas/ts/t4_null.ts 2>&1 | head -4
