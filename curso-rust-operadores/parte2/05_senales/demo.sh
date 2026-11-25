# Parar un proceso bien: SIGTERM, esperar, y SIGKILL solo si no sale.
source /work/repo/comun.sh
echo '$ bash -c '"'"'trap "echo [hijo] recibí TERM, cierro; exit 0" TERM; while :; do sleep 0.1; done'"'"' &'
bash -c 'trap "echo [hijo] recibí TERM, cierro; exit 0" TERM; while :; do sleep 0.1; done' &
P=$!
sleep 0.3
c 'ops-matar $P 3'
wait $P 2>/dev/null
echo '$ bash -c '"'"'trap "" TERM; while :; do sleep 0.1; done'"'"' &      # este ignora SIGTERM'
bash -c 'trap "" TERM; while :; do sleep 0.1; done' &
P=$!
sleep 0.3
c 'ops-matar $P 2'
wait $P 2>/dev/null
c 'ops-matar 999999'
