# Espacio libre con statvfs(2): libre no es lo mismo que disponible.
source /work/repo/comun.sh
c 'ops-libre /'
c 'df -h /'
c 'ops-libre /dev/shm'
c 'df -h /dev/shm'
c 'ops-libre /no/existe'
