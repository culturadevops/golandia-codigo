source /work/repo/comun.sh
rm -rf /tmp/rel && mkdir -p /tmp/rel/src && cp /work/repo/base/Cargo.toml /work/repo/base/Cargo.lock /tmp/rel/ && cp /work/repo/fuga/src/main.rs /tmp/rel/src/
cd /tmp/rel
export CARGO_TARGET_DIR=/tmp/rel/target
P=(--config profile.release.strip=true --config 'profile.release.lto="fat"' --config profile.release.codegen-units=1)
cargo build -q --locked --release "${P[@]}" 2>&1
cargo build -q --locked --release --target x86_64-unknown-linux-musl "${P[@]}" 2>&1
G=target/release/fuga; M=target/x86_64-unknown-linux-musl/release/fuga
echo '$ file fuga-gnu'; file -b $G | cut -d, -f1-2,4-5
echo '$ ldd fuga-gnu'; ldd $G | sed -E 's/ \(0x[0-9a-f]+\)//'
echo '$ file fuga-musl'; file -b $M | cut -d, -f1-2,4-5
echo '$ ldd fuga-musl'; ldd $M 2>&1; echo "[codigo de salida: $?]"
echo '$ ls -l'; printf '%s fuga-gnu\n%s fuga-musl\n' "$(stat -c %s $G)" "$(stat -c %s $M)"
# un sistema de archivos vacío: solo el binario (lo que sería FROM scratch)
rm -rf /tmp/vacio && mkdir -p /tmp/vacio && cp $G /tmp/vacio/fuga-gnu && cp $M /tmp/vacio/fuga-musl
echo '$ chroot vacio /fuga-gnu --version'; chroot /tmp/vacio /fuga-gnu --version 2>&1; echo "[codigo de salida: $?]"
echo '$ chroot vacio /fuga-musl --version'; chroot /tmp/vacio /fuga-musl --version 2>&1; echo "[codigo de salida: $?]"
cp /work/repo/datos/config.env /tmp/vacio/
echo '$ chroot vacio /fuga-musl escanear /config.env'; chroot /tmp/vacio /fuga-musl escanear /config.env 2>&1; echo "[codigo de salida: $?]"
