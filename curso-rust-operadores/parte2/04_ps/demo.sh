# Listar procesos leyendo /proc, y el nombre de proceso que rompe el parseo ingenuo.
source /work/repo/comun.sh
tmp
cp /bin/sleep "./mi app (v2) prod"
echo '$ "./mi app (v2) prod" 300 &      # un proceso con espacios y paréntesis en el nombre'
"./mi app (v2) prod" 300 &
A=$!
echo '$ bash -c "sleep 0.2 & exec sleep 300" &   # su hijo queda zombi: nadie lo espera'
bash -c "sleep 0.2 & exec sleep 300" &
Z=$!
sleep 1
c 'ops-ps'
c 'ops-ps --ingenuo'
c 'ps -eo pid,ppid,stat,rss,comm'
kill $A $Z
wait 2>/dev/null
