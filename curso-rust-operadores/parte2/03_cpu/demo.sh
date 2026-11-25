# Uso de CPU muestreando /proc/stat, y cuántas CPUs tienes de verdad.
source /work/repo/comun.sh
c 'nproc'
c 'cat /sys/fs/cgroup/cpu.max'
c 'ops-cpu 1'
echo '$ ops-quemar 2 4 &              # dos hilos al 100 % durante 4 s'
ops-quemar 2 4 &
sleep 0.5
c 'ops-cpu 2'
wait
echo '$ ops-quemar 8 4 &              # ocho hilos, con un límite de 2 CPUs'
ops-quemar 8 4 &
sleep 0.5
c 'ops-cpu 2'
wait
c 'cat /proc/loadavg'
