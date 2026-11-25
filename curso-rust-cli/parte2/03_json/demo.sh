source /work/repo/comun.sh
ver a_hashmap.rs; proyecto a_hashmap.rs
for i in 1 2 3; do mostrar 'fuga'; done
mostrar 'fuga > a.json; fuga > b.json; diff -q a.json b.json'
echo "# la herramienta completa: --json con BTreeMap"
proyecto ../../fuga/src/main.rs
mostrar 'fuga escanear --json datos/config.env datos/app.yaml'
mostrar 'fuga escanear --json datos/*.env datos/*.yaml datos/*.pem > a.json; fuga escanear --json datos/*.env datos/*.yaml datos/*.pem > b.json; diff -q a.json b.json && echo idénticos'
mostrar 'fuga escanear --json datos/* | jq -r ".hallazgos[] | \"\(.archivo):\(.linea) \(.tipo)\""'
mostrar 'fuga escanear --json datos/* | jq ".por_tipo"'
mostrar 'fuga escanear --json datos/README.md | jq ".hallazgos | length"'
