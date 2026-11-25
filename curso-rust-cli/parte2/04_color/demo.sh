source /work/repo/comun.sh
ver a_siempre.rs; proyecto a_siempre.rs
mostrar 'fuga > /tmp/informe.txt; cat -v /tmp/informe.txt'
mostrar 'fuga | grep -c "aws-access-key$"'
echo "# la herramienta completa: color solo si stdout es una terminal y no hay NO_COLOR"
proyecto ../../fuga/src/main.rs
mostrar 'fuga escanear datos/config.env 2>/dev/null | cat -v'
mostrar 'script -qec "fuga escanear datos/config.env" /dev/null | cat -v'
mostrar 'NO_COLOR=1 script -qec "fuga escanear datos/config.env" /dev/null | cat -v'
