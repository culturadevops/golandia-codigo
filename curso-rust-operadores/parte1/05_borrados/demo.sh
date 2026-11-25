# Un archivo borrado que sigue ocupando espacio porque un proceso lo tiene abierto.
source /work/repo/comun.sh
tmpshm
c 'head -c 100000000 /dev/zero > app.log'
c 'ops-libre /dev/shm | grep -E "usado|uso"'
echo '$ sleep 600 3< app.log &        # un proceso que lo deja abierto'
sleep 600 3<app.log &
P=$!
sleep 0.3
c 'rm app.log'
c 'ls -la'
c 'du -sh .'
c 'ops-libre /dev/shm | grep -E "usado|uso"'
c 'ops-borrados'
echo '$ kill $P                       # cierro ese proceso'
kill $P
wait $P 2>/dev/null
c 'ops-libre /dev/shm | grep -E "usado|uso"'
c 'ops-borrados'
