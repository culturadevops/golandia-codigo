#!/usr/bin/env bash
# Lección 9 — tests de la CLI: run() en el propio proceso, testscript y un golden.
source "$(dirname "$0")/../comun.sh"

seccion "9.1 · Los tests de vigia"
ej 'go test -count=1 ./vigia'

seccion "9.2 · Uno por uno, con lo que tarda cada uno"
ej 'go test -count=1 -v ./vigia 2>&1 | grep -E "^ *--- (PASS|FAIL)|^ok"'

seccion "9.3 · Un cambio en la ayuda sin actualizar el golden (en una copia del repo)"
TMP="$(mktemp -d)"
cp -r . "$TMP/repo"
sed -i 's/vigia revisa endpoints HTTP\./vigia revisa endpoints HTTP y HTTPS./' "$TMP/repo/vigia/run.go"
ej '(cd "$TMP/repo" && go test -count=1 -run Ayuda ./vigia 2>&1 | head -8)'
ej '(cd "$TMP/repo" && go test -count=1 -run Ayuda ./vigia -update && git diff --no-index --stat vigia/testdata/ayuda.golden "$RAIZ/vigia/testdata/ayuda.golden")'
rm -rf "$TMP"

seccion "9.4 · Con tu entorno exportado (VIGIA_TIMEOUT=1ms): los tests de vigia no se enteran"
ej 'VIGIA_TIMEOUT=1ms go test -count=1 ./vigia'

seccion "9.5 · El paquete ingenuo, que lee os.Getenv por su cuenta"
ej 'go test -count=1 ./09_tests/ingenuo'
ej 'VIGIA_TIMEOUT=1ms go test -count=1 ./09_tests/ingenuo'

seccion "9.6 · Cobertura"
ej 'go test -count=1 -cover ./vigia'

seccion "9.7 · Trampa: defer srv.Close() en vez de t.Cleanup (en una copia del repo)"
TMP="$(mktemp -d)"
cp -r . "$TMP/repo"
sed -i 's/t.Cleanup(srv.Close)/defer srv.Close()/' "$TMP/repo/vigia/script_test.go"
ej '(cd "$TMP/repo" && grep -n "srv.Close" vigia/script_test.go)'
ej '(cd "$TMP/repo" && go test -count=1 -run Scripts ./vigia > "$TMP/test.txt" 2>&1; grep -E "^--- |^    --- |^(ok|FAIL)" "$TMP/test.txt")'
ej 'grep -o "dial tcp [^ ]*: connect: connection refused" "$TMP/test.txt" | head -1'
rm -rf "$TMP"
