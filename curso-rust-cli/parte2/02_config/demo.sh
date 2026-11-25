source /work/repo/comun.sh
cd /tmp && rm -rf cfg && mkdir cfg && cd cfg
ver a_ingenuo.rs; proyecto a_ingenuo.rs
cd /tmp/cfg && echo 'max = 3' > fuga.toml
echo '$ cat fuga.toml'; cat fuga.toml
echo '$ fuga'; fuga; echo "[codigo de salida: $?]"
echo '$ fuga --max 10'; fuga --max 10; echo "[codigo de salida: $?]"
echo "# la herramienta completa: Option<usize> + env = FUGA_MAX + fuga.toml + 10"
proyecto ../../fuga/src/main.rs
cd /tmp/cfg && cp /work/repo/datos/config.env . && for i in 1 2 3 4 5 6 7 8 9 10 11 12; do echo "K$i=AKIAIOSFODNN7EXAMPLE"; done >> config.env
c() { echo "\$ $1"; eval "$1" 2>&1; echo "[codigo de salida: $?]"; }
rm fuga.toml
c 'fuga escanear config.env | wc -l'
echo 'max = 3' > fuga.toml; echo '$ echo "max = 3" > fuga.toml'
c 'fuga escanear config.env | wc -l'
c 'FUGA_MAX=5 fuga escanear config.env | wc -l'
c 'FUGA_MAX=5 fuga escanear config.env --max 7 | wc -l'
c 'fuga escanear config.env --max 10 | wc -l'
c 'FUGA_MAX=cinco fuga escanear config.env'
echo 'maxx = 3' > fuga.toml; echo '$ echo "maxx = 3" > fuga.toml'
c 'fuga escanear config.env'
ver b_sin_deny.rs; proyecto b_sin_deny.rs
mostrar 'fuga'
