#!/usr/bin/env bash
# Demo del vídeo 465 «Módulos falsos en Go» — se EJECUTA y su salida va a SALIDA.md.
#   1) go.sum: los hashes de una dependencia real (github.com/google/uuid);
#   2) go mod verify con la caché intacta, y después de tocar un archivo de la caché;
#   3) un go.sum con un hash alterado: el go command se niega a usar el módulo;
#   4) el paquete falso github.com/boltdb-go/bolt: SOLO se consulta en la base oficial de vulnerabilidades
#      (vuln.go.dev). ⛔ Nunca se descarga.
# Caché de módulos propia en el directorio temporal (GOMODCACHE): no toca la del sistema.
set -u
T=$(mktemp -d)
export GOMODCACHE="$T/cache" GOFLAGS=-modcacherw PYTHONUTF8=1
cd "$T"
mkdir app && cd app
go mod init demo/app >/dev/null 2>&1
cat > main.go <<'EOF'
package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() { fmt.Println(uuid.NewString() != "") }
EOF
go get github.com/google/uuid@v1.6.0 >/dev/null 2>&1

echo '$ cat go.sum'
cat go.sum
echo
echo '$ go mod verify'
go mod verify 2>&1
echo
echo '$ echo "// cambio" >> $GOMODCACHE/github.com/google/uuid@v1.6.0/uuid.go && go mod verify'
chmod u+w "$GOMODCACHE/github.com/google/uuid@v1.6.0/uuid.go"
echo "// cambio" >> "$GOMODCACHE/github.com/google/uuid@v1.6.0/uuid.go"
go mod verify 2>&1 | sed "s|(.*uuid@v1.6.0)|($GOMODCACHE/github.com/google/uuid@v1.6.0)|"
echo "(código de salida: ${PIPESTATUS[0]})"
echo

# 3) go.sum alterado, con la caché vacía
chmod -R u+w "$GOMODCACHE" && rm -rf "$GOMODCACHE"
sed -i 's|^\(github.com/google/uuid v1.6.0 h1:\)....|\1AAAA|' go.sum
echo '$ grep "uuid v1.6.0 h1" go.sum   # alguien cambió el hash'
grep "uuid v1.6.0 h1" go.sum
echo '$ go build .'
go build . 2>&1 | sed "s|$GOMODCACHE|\$GOMODCACHE|g"
echo "(código de salida: ${PIPESTATUS[0]})"
echo

# 4) el módulo falso, solo en la base de vulnerabilidades
echo '$ curl -s https://vuln.go.dev/index/modules.json | grep -o boltdb...'
curl -s https://vuln.go.dev/index/modules.json | python -c "
import json,sys
for m in json.load(sys.stdin):
    if 'boltdb' in m['path']:
        print(' ', m['path'], [v['id'] for v in m.get('vulns',[])])
"
ID=$(curl -s https://vuln.go.dev/index/modules.json | python -c "
import json,sys
print(next(v['id'] for m in json.load(sys.stdin) if m['path']=='github.com/boltdb-go/bolt' for v in m['vulns']))")
echo "\$ curl -s https://vuln.go.dev/ID/$ID.json"
curl -s "https://vuln.go.dev/ID/$ID.json" | python -c "
import json,sys
d=json.load(sys.stdin)
print('  id:', d['id'])
print('  resumen:', d.get('summary'))
print('  publicado:', d.get('published','')[:10])
for a in d.get('affected',[]):
    print('  módulo:', a['package']['name'], '· afectadas: desde', a['ranges'][0]['events'][0].get('introduced'), '(todas)')
print('  detalle:', (d.get('details') or '')[:160])
"
cd / && rm -rf "$T"
