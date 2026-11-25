source /work/repo/comun.sh
rm -rf /tmp/rel && mkdir -p /tmp/rel/src && cp /work/repo/base/Cargo.toml /work/repo/base/Cargo.lock /tmp/rel/ && cp /work/repo/fuga/src/main.rs /tmp/rel/src/
cd /tmp/rel
export CARGO_TARGET_DIR=/tmp/rel/target
tam() { stat -c %s "$1"; }
cargo build -q --locked 2>&1
printf '%-44s %10s\n' "perfil" "bytes"
printf '%-44s %10s\n' "debug (cargo build)" "$(tam target/debug/fuga)"
cargo build -q --locked --release 2>&1
printf '%-44s %10s\n' "release" "$(tam target/release/fuga)"
cp target/release/fuga /tmp/fuga-release
b() { cargo build -q --locked --release "$@" 2>&1; }
b --config profile.release.strip=true
printf '%-44s %10s\n' "+ strip = true" "$(tam target/release/fuga)"
b --config profile.release.strip=true --config 'profile.release.lto="fat"' --config profile.release.codegen-units=1
printf '%-44s %10s\n' "+ lto = fat, codegen-units = 1" "$(tam target/release/fuga)"
cp target/release/fuga /tmp/fuga-lto
b --config profile.release.strip=true --config 'profile.release.lto="fat"' --config profile.release.codegen-units=1 --config 'profile.release.opt-level="z"'
printf '%-44s %10s\n' "+ opt-level = z" "$(tam target/release/fuga)"
cp target/release/fuga /tmp/fuga-z
b --config profile.release.strip=true --config 'profile.release.lto="fat"' --config profile.release.codegen-units=1 --config 'profile.release.opt-level="z"' --config 'profile.release.panic="abort"'
printf '%-44s %10s\n' "+ panic = abort" "$(tam target/release/fuga)"
# velocidad: un archivo de 400.000 líneas con 1 de cada 4 «sucia»
for i in $(seq 1 100000); do printf 'APP=%d\nAWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE\nuser=app\nruta=/srv/app/%d\n' $i $i; done > /tmp/grande.env
echo "\$ wc -l grande.env"; wc -l < /tmp/grande.env
TIMEFORMAT='%R s'
for v in "debug:target/debug/fuga" "release:/tmp/fuga-release" "lto:/tmp/fuga-lto" "opt-level z:/tmp/fuga-z"; do
  n=${v%%:*}; f=${v#*:}
  t=$( { time FUGA_MAX=1000000 $f escanear /tmp/grande.env > /dev/null 2>&1; } 2>&1 )
  t2=$( { time FUGA_MAX=1000000 $f escanear /tmp/grande.env > /dev/null 2>&1; } 2>&1 )
  t3=$( { time FUGA_MAX=1000000 $f escanear /tmp/grande.env > /dev/null 2>&1; } 2>&1 )
  printf '%-12s %s · %s · %s\n' "$n" "$t" "$t2" "$t3"
done
echo "\$ FUGA_MAX=1000000 fuga escanear grande.env | wc -l"
FUGA_MAX=1000000 /tmp/fuga-release escanear /tmp/grande.env 2>/dev/null | wc -l
