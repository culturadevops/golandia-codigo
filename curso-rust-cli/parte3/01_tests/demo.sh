source /work/repo/comun.sh
cp /work/repo/fuga/src/main.rs /tmp/sin_borde.rs
sed -i 's/r"\\bAKIA\[0-9A-Z\]{16}\\b"/r"AKIA[0-9A-Z]{16}"/' /tmp/sin_borde.rs
echo '$ grep -n "aws-access-key\", Regex" src/main.rs   # versión sin \b'
grep -n 'aws-access-key", Regex' /tmp/sin_borde.rs
rm -rf /tmp/p/fuga; mkdir -p /tmp/p/fuga/src
cp /work/repo/base/Cargo.toml /work/repo/base/Cargo.lock /tmp/p/fuga/; cp /tmp/sin_borde.rs /tmp/p/fuga/src/main.rs
echo '$ cargo test'
(set -o pipefail; cd /tmp/p/fuga && cargo test -q --locked 2>&1 | grep -v "^\s*Running\|^$"); echo "[codigo de salida: $?]"
echo '$ fuga escanear datos/notas.txt   # con el patrón sin \b'
(cd /tmp/p/fuga && cargo build -q --locked) && (cd /work/repo && $CARGO_TARGET_DIR/debug/fuga escanear datos/notas.txt 2>&1); echo "[codigo de salida: $?]"
echo "# con \b, la herramienta completa"
cp /work/repo/fuga/src/main.rs /tmp/p/fuga/src/main.rs
echo '$ cargo test'
(set -o pipefail; cd /tmp/p/fuga && cargo test -q --locked 2>&1 | grep -v "^\s*Running\|^$"); echo "[codigo de salida: $?]"
(cd /tmp/p/fuga && cargo build -q --locked); ln -sf $CARGO_TARGET_DIR/debug/fuga /usr/local/bin/fuga
mostrar 'fuga escanear datos/notas.txt'
