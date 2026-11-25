# Cuánto ocupa un directorio: el tamaño que dice el archivo frente a los bloques que ocupa.
source /work/repo/comun.sh
tmp
mkdir datos
c 'truncate -s 1G datos/disco-virtual.img'
c 'head -c 5000000 /dev/urandom > datos/fotos.bin'
c 'ls -l datos'
c 'ops-du datos'
c 'du -sh datos'
c 'du -sh --apparent-size datos'
