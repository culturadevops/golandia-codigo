#!/usr/bin/env bash
# dentro.sh — corre DENTRO del contenedor --privileged que lanza correr.sh. No lo ejecutes en tu máquina.
#
# ⛔ Todo lo que se escribe va dentro de un directorio de mktemp; el código fuente se monta de solo lectura en /demo.
# ⛔ La bomba fork corre con pids.max=20 en su cgroup, dentro de un contenedor con --pids-limit 300, y muere a los 5 s.
set -u
W=$(mktemp -d /tmp/mini.XXXXXX)
cd "$W" || exit 1
R=rootfs
ALP=3.24.3
SHA=f8e1cafba9b1b283b5d30d06f9885c8d2bc161ffe9083ade8e2e157813a53407
seccion() { echo; echo "== $* =="; }

seccion "0 · versiones"
go version
uname -r
echo "cgroup: $(stat -fc %T /sys/fs/cgroup)"
grep PRETTY /etc/os-release

seccion "0 · rootfs: alpine-minirootfs-$ALP"
curl -fsSLo alp.tgz "https://dl-cdn.alpinelinux.org/alpine/v3.24/releases/x86_64/alpine-minirootfs-$ALP-x86_64.tar.gz"
echo "$SHA  alp.tgz" | sha256sum -c - || exit 1
mkdir "$R" && tar -xzf alp.tgz -C "$R"
echo "rootfs: $(du -sh $R | cut -f1), $(find $R | wc -l) entradas"

seccion "0 · compilo mini y comer"
cp -r /demo/mini /demo/comer .
(cd mini && CGO_ENABLED=0 go build -buildvcs=false -ldflags "-s -w" -o ../mini.bin .) || exit 1
(cd comer && CGO_ENABLED=0 go build -buildvcs=false -o ../$R/usr/local/bin/comer .) || exit 1
echo "main.go: $(wc -l < mini/main.go) líneas · binario mini (-s -w): $(stat -c %s mini.bin) bytes"
go version -m mini.bin | grep -c dep | sed 's/^/dependencias externas: /'
M=./mini.bin

seccion "0 · cgroup v2: activo memory y pids para los hijos (el truco de docker:dind)"
mkdir -p /sys/fs/cgroup/init
xargs -rn1 < /sys/fs/cgroup/cgroup.procs > /sys/fs/cgroup/init/cgroup.procs 2>/dev/null
echo "+memory +pids" > /sys/fs/cgroup/cgroup.subtree_control
echo "subtree_control: $(cat /sys/fs/cgroup/cgroup.subtree_control)"

seccion "1 · UTS: hostname propio"
echo "fuera:  $(hostname)"
$M run -hostname caja $R /bin/hostname
echo "fuera, después: $(hostname)"

seccion "2 · PID: el comando es el PID 1"
$M run $R /bin/sh -c 'echo "dentro, mi PID: $$"; ps -o pid,args'
$M run $R /bin/sleep 2 2>/dev/null &
sleep 0.5
echo "fuera, el mismo sleep:"; ps -eo pid,args | grep "[s]leep 2$"
echo "fuera, procesos visibles: $(ps -e --no-headers | wc -l)"
wait

seccion "3 · mount: otra raíz y un /proc propio"
echo "fuera:  $(grep PRETTY /etc/os-release)"
$M run $R /bin/sh -c 'echo "dentro: $(grep PRETTY /etc/os-release)"; echo "ls /: $(ls / | tr "\n" " ")"; echo "montajes:"; cat /proc/mounts; echo "procesos en /proc: $(ls -d /proc/[0-9]* | tr "\n" " ")"'
$M run $R /bin/sleep 2 2>/dev/null &
sleep 0.5
echo "fuera, montajes del contenedor en mi tabla: $(grep -c "$W" /proc/mounts)"
wait

seccion "4 · cgroups: memoria sin límite"
$M run $R /usr/local/bin/comer | tail -2

seccion "5 · cgroups: memory.max=64M (y memory.swap.max=0)"
inicio=$(date +%s%N)
$M run -mem 64M $R /usr/local/bin/comer | tail -3
echo "código de salida: ${PIPESTATUS[0]} · tardó $(( ($(date +%s%N) - inicio) / 1000000 )) ms"
dmesg 2>/dev/null | grep -i "killed process" | tail -1

seccion "6 · cgroups: memory.max=64M SIN memory.swap.max=0 (cgroup hecho a mano)"
grep -E "^/|Filename" /proc/swaps
mkdir /sys/fs/cgroup/conswap
echo 64M > /sys/fs/cgroup/conswap/memory.max
sh -c "echo \$\$ > /sys/fs/cgroup/conswap/cgroup.procs; exec $R/usr/local/bin/comer" | tail -2
echo "memory.peak=$(cat /sys/fs/cgroup/conswap/memory.peak) memory.swap.peak=$(cat /sys/fs/cgroup/conswap/memory.swap.peak) oom_kill=$(grep oom_kill /sys/fs/cgroup/conswap/memory.events | cut -d' ' -f2)"
rmdir /sys/fs/cgroup/conswap

seccion "7 · cgroups: bomba fork con pids.max=20 (vive 5 s)"
$M run -pids 20 $R /bin/sh -c 'b(){ b | b & }; b 2>/dev/null; exec sleep 5' &
sleep 2
cg=$(ls -d /sys/fs/cgroup/mini-* | head -1)
echo "a los 2 s: pids.current=$(cat $cg/pids.current) · rechazos=$(grep max $cg/pids.events | cut -d' ' -f2)"
inicio=$(date +%s%N); ls / >/dev/null; echo "fuera, ls / mientras tanto: $(( ($(date +%s%N) - inicio) / 1000000 )) ms"
echo "fuera, procesos totales: $(ps -e --no-headers | wc -l)"
wait
echo "después: cgroups mini que quedan: $(ls -d /sys/fs/cgroup/mini-* 2>/dev/null | wc -l)"

seccion "8 · red: sin namespace de red (ve la red del anfitrión)"
$M run $R /bin/sh -c 'ip -o -4 addr | cut -c1-60; ping -c1 -W2 1.1.1.1 | tail -2' 2>/dev/null

seccion "9 · red: con -red (CLONE_NEWNET)"
$M run -red $R /bin/sh -c 'ip -o link | cut -c1-70; ip -o -4 addr; ping -c1 -W2 1.1.1.1; echo "ping salió con $?"' 2>&1 | grep -v "^\[mini\]"

seccion "10 · la raíz es una carpeta compartida: lo que escribe un contenedor lo ve el siguiente"
$M run $R /bin/sh -c 'echo hola > /escrito-por-el-primero' 2>/dev/null
$M run $R /bin/cat /escrito-por-el-primero 2>/dev/null
echo "fuera, en la carpeta del rootfs: $(ls $R | grep escrito)"

seccion "11 · capabilities y seccomp dentro de mini"
echo "fuera (el contenedor --privileged): $(grep -E '^Seccomp_filters:' /proc/self/status)"
$M run $R /bin/sh -c 'grep -E "^(CapEff|NoNewPrivs|Seccomp|Seccomp_filters):" /proc/self/status; hostname otro && echo "hostname otro: permitido"; mount -t tmpfs none /mnt && echo "mount tmpfs: permitido"; unshare -U true && echo "unshare -U (namespace de usuario): permitido"' 2>/dev/null

seccion "12 · arranque: 3 rondas de 20 veces mini run /bin/true"
for ronda in 1 2 3; do
  inicio=$(date +%s%N)
  for i in $(seq 20); do $M run $R /bin/true 2>/dev/null; done
  echo "ronda $ronda: $(( ($(date +%s%N) - inicio) / 20000000 )) ms por contenedor"
done

cd / && rm -rf -- "$W"
