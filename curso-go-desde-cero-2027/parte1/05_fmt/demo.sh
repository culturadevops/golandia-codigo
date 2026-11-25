source /work/repo/comun.sh
ver a_verbos.go; proyecto p1_fmt a_verbos.go; correr p1_fmt
ver b_verbo_malo.go; proyecto p1_fmt b_verbo_malo.go; correr p1_fmt
echo '$ go vet .'
(cd /tmp/p/p1_fmt && go vet . 2>&1); echo "[codigo de salida: $?]"
