source /work/repo/comun.sh
for i in $(seq 1 200000); do echo "linea $i AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"; done > /tmp/grande.env
ver a_progreso.rs; proyecto a_progreso.rs
mostrar 'fuga datos/config.env datos/app.yaml datos/README.md'
mostrar 'fuga datos/config.env datos/app.yaml datos/README.md | wc -l'
ver b_stderr.rs; proyecto b_stderr.rs
mostrar 'fuga datos/config.env datos/app.yaml datos/README.md | wc -l'
mostrar 'fuga datos/config.env datos/app.yaml datos/README.md 2>/dev/null'
ver c_pipe.rs; proyecto c_pipe.rs
mostrar 'wc -l /tmp/grande.env'
mostrar 'fuga /tmp/grande.env | head -2; echo "PIPESTATUS: ${PIPESTATUS[*]}"'
ver d_pipe_ok.rs; proyecto d_pipe_ok.rs
mostrar 'fuga /tmp/grande.env | head -2; echo "PIPESTATUS: ${PIPESTATUS[*]}"'
