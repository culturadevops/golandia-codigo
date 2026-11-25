#!/usr/bin/env bash
# Lección 12 — release con GoReleaser en modo snapshot: archivos, checksums y changelog.
source "$(dirname "$0")/../comun.sh"
TMP="$(mktemp -d)"
cp -r . "$TMP/repo"
cd "$TMP/repo" || exit 1
rm -rf .git dist
G="git -c user.name=curso -c user.email=curso@example.com"
git init -q
$G add -A
$G commit -qm "feat: vigia revisa endpoints HTTP"
$G tag v0.9.0
echo "// -json" >> vigia/revisar.go && $G commit -qam "feat: salida -json para scripts"
echo "// señales" >> vigia/main.go && $G commit -qam "fix: Ctrl+C ya no deja el informe a medias"
echo "// docs" >> README.md && $G commit -qam "docs: README del curso"
$G tag v1.0.0
# GoReleaser pide un remoto aunque no vaya a publicar; este no existe y no se contacta.
git remote add origin https://github.com/ejemplo/vigia.git

seccion "12.1 · La versión de GoReleaser y la configuración"
ej 'goreleaser --version 2>&1 | grep -E "^GitVersion"'
ej 'goreleaser check 2>&1 | tail -2'

seccion "12.2 · Snapshot: construye todo y no publica nada"
ej 'goreleaser release --snapshot --clean 2>&1 | grep -E "snapshot|building  |archiving|checksums|succeeded" | sed "s/  */ /g"'

seccion "12.3 · Lo que queda en dist/"
ej 'ls dist | grep -vE "_v1$|_v8.0$"'
ej 'ls -l dist/*.tar.gz dist/*.zip | awk "{print \$5, \$9}"'

seccion "12.4 · checksums.txt, y comprobarlo"
ej 'cat dist/checksums.txt | cut -c1-16,65-'
ej '(cd dist && sha256sum -c checksums.txt)'

seccion "12.5 · El binario de Linux, sacado del archivo"
ej 'mkdir -p "$TMP/x" && tar -xzf dist/vigia_1.0.0_linux_amd64.tar.gz -C "$TMP/x" && ls "$TMP/x" && "$TMP/x/vigia" version | head -3'

seccion "12.6 · ¿Y el changelog? En snapshot no se genera"
ejc 'ls dist/CHANGELOG.md'

seccion "12.7 · Sin --snapshot: el changelog sale de los commits entre dos etiquetas"
ej 'git log --oneline v0.9.0..v1.0.0 | cut -c9-'
ej 'goreleaser release --clean --skip=publish,announce 2>&1 | grep -E "skipping|using tags|changelog|succeeded|failed" | sed "s/  */ /g"'
ej 'cat dist/CHANGELOG.md'

seccion "12.8 · ¿Es reproducible? Dos snapshots seguidos, mismos checksums"
goreleaser release --snapshot --clean >/dev/null 2>&1 && cp dist/checksums.txt "$TMP/a.txt"
sleep 2
goreleaser release --snapshot --clean >/dev/null 2>&1 && cp dist/checksums.txt "$TMP/b.txt"
ej 'diff "$TMP/a.txt" "$TMP/b.txt" > /dev/null && echo "idénticos" || diff "$TMP/a.txt" "$TMP/b.txt" | cut -c1-20,67- | head -6'

seccion "12.9 · Lo mismo quitando mod_timestamp de la configuración"
grep -v "mod_timestamp" .goreleaser.yaml > "$TMP/sin-fecha.yaml"
goreleaser release --snapshot --clean -f "$TMP/sin-fecha.yaml" >/dev/null 2>&1 && cp dist/checksums.txt "$TMP/c.txt"
sleep 2
goreleaser release --snapshot --clean -f "$TMP/sin-fecha.yaml" >/dev/null 2>&1 && cp dist/checksums.txt "$TMP/d.txt"
ej 'diff "$TMP/c.txt" "$TMP/d.txt" > /dev/null && echo "idénticos" || echo "distintos: $(diff "$TMP/c.txt" "$TMP/d.txt" | grep -c "^<") de 6 archivos cambian de checksum"'

cd / && rm -rf "$TMP"
