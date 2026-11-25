source /work/repo/comun.sh
pkill -x sleep 2>/dev/null
ej comandos timeout
echo '$ pgrep -a sleep'; pgrep -a sleep || echo "(ninguno)"
pkill -x sleep 2>/dev/null; sleep 0.2
ej comandos timeout-kill
echo '$ pgrep -a sleep'; pgrep -a sleep || echo "(ninguno)"
