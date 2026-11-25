#!/usr/bin/env bash
# Demo del vídeo 460 «govulncheck vs Trivy» — se EJECUTA y su salida va a SALIDA.md.
# Dos programas con EXACTAMENTE la misma dependencia vieja (golang.org/x/text v0.3.7):
#   · «seguro»:   usa x/text/width, que no llega a ninguna función vulnerable;
#   · «afectado»: llama a language.ParseAcceptLanguage (la que nombra CVE-2022-32149).
# Los dos se escanean con govulncheck (mira qué funciones llama tu código) y con Trivy (mira el go.mod).
# Directorio temporal: no toca nada fuera.
set -u
GOVULN="$HOME/go/bin/govulncheck"
T=$(mktemp -d)
cd "$T"

mkdir seguro afectado norm
cat > seguro/main.go <<'EOF'
package main

import (
	"fmt"

	"golang.org/x/text/width"
)

func main() {
	fmt.Println(width.Narrow.String("ＧＯ"))
}
EOF
cat > afectado/main.go <<'EOF'
package main

import (
	"fmt"

	"golang.org/x/text/language"
)

func main() {
	tags, _, err := language.ParseAcceptLanguage("es-MX,es;q=0.9,en;q=0.8")
	fmt.Println(tags, err)
}
EOF
cat > norm/main.go <<'EOF2'
package main

import (
	"fmt"

	"golang.org/x/text/unicode/norm"
)

func main() {
	fmt.Println(norm.NFC.String("café"))
}
EOF2
for p in seguro afectado norm; do
  (cd "$p" && go mod init "demo/$p" >/dev/null 2>&1 && go get golang.org/x/text@v0.3.7 >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1)
done

for p in seguro afectado; do
  echo "\$ grep -E 'x/text|width\\.|language\\.' $p/main.go $p/go.mod"
  grep -h -E 'x/text|width\.|language\.' "$p/main.go" "$p/go.mod" | sed 's/^\t//; s/^require //'
  echo "\$ cd $p && go run ."
  (cd "$p" && go run . 2>&1)
  echo
done

for p in seguro afectado norm; do
  echo "\$ cd $p && govulncheck ./..."
  (cd "$p" && "$GOVULN" ./... > ../gv-$p.txt 2>&1; echo "(código de salida: $?)" >> ../gv-$p.txt)
  grep -E "^(Vulnerability #|    Found in|    Fixed in|      #1:|Your code|No vulnerabilities|\(código)" "gv-$p.txt"
  grep -E "^Vulnerability #" -A1 "gv-$p.txt" | grep -v -E "^(Vulnerability|--)" | sed 's/^ *//; s/^/    /'
  grep -E "^This scan also found" -A1 "gv-$p.txt" | tr '\n' ' ' | sed 's/  */ /g; s/ $/\n/'
  echo
done

for p in seguro afectado; do
  echo "\$ trivy fs --scanners vuln $p"
  MSYS_NO_PATHCONV=1 docker run --rm -v "$(cygpath -w "$T/$p"):/src" -v trivy-cache:/root/.cache aquasec/trivy:0.67.2 \
     fs --quiet --scanners vuln --format json /src 2>/dev/null \
  | python -c "
import json,sys
d=json.load(sys.stdin)
v=[x for r in d.get('Results',[]) for x in (r.get('Vulnerabilities') or [])]
print('  vulnerabilidades:', len(v))
for x in v: print('  ', x['VulnerabilityID'], x['PkgName'], x['InstalledVersion'], '->', x.get('FixedVersion'), x.get('Severity'))
"
  echo
done
echo '$ # el arreglo: subir x/text a la última versión'
echo '$ cd afectado && go get golang.org/x/text@latest && grep x/text go.mod'
(cd afectado && go get golang.org/x/text@latest >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1 && grep "golang.org/x/text" go.mod | sed 's/^	//; s/^require //')
echo '$ cd afectado && govulncheck ./... | grep "Your code"'
(cd afectado && "$GOVULN" ./... 2>&1 | grep -E "^Your code|^No vulnerabilities"; echo "(código de salida: ${PIPESTATUS[0]})")
echo '$ trivy fs --scanners vuln afectado'
MSYS_NO_PATHCONV=1 docker run --rm -v "$(cygpath -w "$T/afectado"):/src" -v trivy-cache:/root/.cache aquasec/trivy:0.67.2    fs --quiet --scanners vuln --format json /src 2>/dev/null | python -c "
import json,sys
d=json.load(sys.stdin)
print('  vulnerabilidades:', sum(len(r.get('Vulnerabilities') or []) for r in d.get('Results',[])))
"
echo
echo '$ docker run --rm aquasec/trivy:0.67.2 --version | head -1'
MSYS_NO_PATHCONV=1 docker run --rm aquasec/trivy:0.67.2 --version 2>/dev/null | head -1
echo '$ govulncheck -version | grep -E "Scanner|DB updated"'
"$GOVULN" -version | grep -E "Scanner|DB updated"
cd / && rm -rf "$T"
