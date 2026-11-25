# Enlaces duros que cuentan doble, y un enlace simbólico que hace un bucle.
source /work/repo/comun.sh
tmp
mkdir logs
c 'head -c 20000000 /dev/zero > logs/app.log'
c 'ln logs/app.log logs/app-copia1.log && ln logs/app.log logs/app-copia2.log'
c 'ls -li logs'
c 'ops-du logs'
c 'du -sh logs'
mkdir -p bucle/a
c 'ln -s .. bucle/a/arriba'
c 'ops-du bucle --seguir-enlaces'
c 'ops-du bucle'
