# Dentro de un contenedor, /proc/meminfo enseña la memoria de la máquina, no tu límite.
source /work/repo/comun.sh
c 'cat /sys/fs/cgroup/memory.max'
c 'ops-mem'
c 'ops-comer 500'
c 'ops-comer 2000'
c 'grep oom_kill /sys/fs/cgroup/memory.events'
