source /work/repo/comun.sh
ver a_cli.rs
correr_tests() { echo "\$ $1"; (set -o pipefail; cd /tmp/p/fuga && eval "$1" 2>&1 | grep -E "^test |test result|panicked|Unexpected|code=|stdout=|warning: use of deprecated"); echo "[codigo de salida: $?]"; }
rm -rf /tmp/p/fuga; mkdir -p /tmp/p/fuga/src /tmp/p/fuga/tests
cp /work/repo/base/Cargo.toml /work/repo/base/Cargo.lock /tmp/p/fuga/; cp /work/repo/fuga/src/main.rs /tmp/p/fuga/src/
rm -rf /tmp/datos && cp -r /work/repo/datos /tmp/datos   # los tests corren en /tmp/p/fuga: ../../datos
cp "$AQUI_LECCION/a_cli.rs" /tmp/p/fuga/tests/cli.rs
correr_tests 'cargo test -q --locked --test cli'
correr_tests 'FUGA_MAX=1 cargo test -q --locked --test cli'
ver b_cli.rs
cp "$AQUI_LECCION/b_cli.rs" /tmp/p/fuga/tests/cli.rs
correr_tests 'FUGA_MAX=1 cargo test -q --locked --test cli'
