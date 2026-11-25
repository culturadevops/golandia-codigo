#!/usr/bin/env bash
# Lección 11 — compilación cruzada para Linux, Windows y macOS, con los tamaños medidos.
source "$(dirname "$0")/../comun.sh"
OUT="$(mktemp -d)"

seccion "11.1 · Cinco sistemas desde la misma máquina: GOOS y GOARCH"
for par in linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64; do
	os="${par%/*}" arch="${par#*/}" ext=""
	[ "$os" = windows ] && ext=".exe"
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -o "$OUT/normal/vigia-$os-$arch$ext" ./vigia
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags="-s -w" -o "$OUT/fino/vigia-$os-$arch$ext" ./vigia
done
printf '  %-24s %12s %12s %8s\n' "binario" "normal" "-s -w" "ahorro"
for f in "$OUT"/normal/*; do
	n=$(basename "$f")
	a=$(stat -c %s "$f")
	b=$(stat -c %s "$OUT/fino/$n")
	printf '  %-24s %12s %12s %7s%%\n' "$n" "$a" "$b" "$(( (a - b) * 100 / a ))"
done

seccion "11.2 · Qué es cada uno (file)"
for f in "$OUT"/fino/*; do
	echo "  $(basename "$f"): $(file -b "$f" | cut -c1-70)"
done

seccion "11.3 · Trampa: sin -trimpath, el binario lleva las rutas de tu máquina"
ej 'strings "$OUT/normal/vigia-linux-amd64" | grep -c "$RAIZ/"'
ej 'strings "$OUT/normal/vigia-linux-amd64" | grep -m1 "$RAIZ/"'
ej 'strings "$OUT/fino/vigia-linux-amd64" | grep -c "$RAIZ/"'

seccion "11.4 · Trampa: compilado en Linux para Linux, con cgo, sale dinámico"
ej 'go env CGO_ENABLED'
go build -o "$OUT/con-cgo" ./vigia
CGO_ENABLED=0 go build -o "$OUT/sin-cgo" ./vigia
ej 'file -b "$OUT/con-cgo" | cut -c1-90'
ej 'ldd "$OUT/con-cgo"'
ej 'file -b "$OUT/sin-cgo" | cut -c1-90'
ej 'ldd "$OUT/sin-cgo"'

seccion "11.5 · -s -w quita los símbolos, no la información de compilación"
ej 'go version -m "$OUT/fino/vigia-linux-amd64" | grep -E "^\s+(path|build\s+(-trimpath|-ldflags|GOOS|GOARCH))"'
ej '"$OUT/fino/vigia-linux-amd64" version | head -3'

rm -rf "$OUT"
