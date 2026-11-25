# El chequeo de guardia: disco y memoria contra un umbral, con código de salida para cron.
source /work/repo/comun.sh
tmpshm
c 'ops-guardia /dev/shm 80 20'
c 'head -c 230000000 /dev/zero > relleno.bin'
c 'ops-guardia /dev/shm 80 20'
c 'rm relleno.bin'
echo '$ ops-comer 800 6 &              # algo que se come 800 MiB y los retiene'
ops-comer 800 6 > /dev/null &
P=$!
sleep 3
c 'ops-guardia /dev/shm 80 20'
c 'ops-guardia /dev/shm 80 20 || echo "ALERTA: avisar a la guardia"'
wait $P
c 'ops-guardia /dev/shm 80 20'
