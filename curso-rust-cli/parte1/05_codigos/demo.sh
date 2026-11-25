source /work/repo/comun.sh
ver a_exit.rs; proyecto a_exit.rs
mostrar 'fuga datos/config.env'
mostrar 'fuga datos/README.md'
ver b_exitcode.rs; proyecto b_exitcode.rs
mostrar 'fuga datos/config.env'
mostrar 'fuga datos/README.md && echo "limpio: sigo con el despliegue"'
mostrar 'fuga datos/config.env && echo "limpio: sigo con el despliegue" || echo "FRENO: hay secretos"'
