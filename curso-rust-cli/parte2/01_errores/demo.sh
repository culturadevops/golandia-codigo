source /work/repo/comun.sh
ver a_box.rs; proyecto a_box.rs
mostrar 'fuga datos/config.env datos/no-existe.env datos/app.yaml'
ver b_anyhow.rs; proyecto b_anyhow.rs
mostrar 'fuga datos/config.env datos/no-existe.env datos/app.yaml'
echo "# la herramienta completa: 0 limpio, 1 hay secretos, 3 no pude trabajar, 2 lo pone clap"
proyecto ../../fuga/src/main.rs
mostrar 'fuga escanear datos/config.env'
mostrar 'fuga escanear datos/no-existe.env'
mostrar 'fuga escanear datos/README.md'
