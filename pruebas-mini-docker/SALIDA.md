# SALIDA — «Escribí Docker en Go» (golandia 681)

Ejecutado el 2026-10-26 con `bash correr.sh`. Salida completa en `salidas/dentro.txt` (mini, dentro del contenedor
`--privileged`) y `salidas/docker.txt` (Docker de verdad, desde el anfitrión). Las escenas de terminal copian de aquí.

## Versiones

| Pieza | Versión |
|---|---|
| Docker (Docker Desktop, WSL2) | 29.6.2, cgroup v2, driver cgroupfs |
| Kernel | 6.18.33.2-microsoft-standard-WSL2 |
| Go | go1.27.2 linux/amd64 (imagen `golang:1.27.2`, Debian 13 trixie) |
| rootfs | `alpine-minirootfs-3.24.3-x86_64.tar.gz`, sha256 `f8e1cafb…a53407` comprobado (8,7 MB, 516 entradas) |
| Imagen para comparar | `alpine:3.24.3` |

## Cómo corre

- `mini/main.go`: **158 líneas** (144 sin contar las vacías), solo biblioteca estándar (0 dependencias externas).
  Binario con `-s -w`: **2.023.584 bytes**.
- `comer/main.go`: reserva memoria de 8 en 8 MB (tocando cada página) hasta 512 MB.
- Contenedor externo: `docker run --rm --privileged --pids-limit 300 --memory 1g golang:1.27.2`, código montado `:ro`,
  todo lo escrito en `mktemp -d`.
- Antes de nada, el truco de `docker:dind`: mover los procesos a `/sys/fs/cgroup/init` y escribir `+memory +pids` en
  `cgroup.subtree_control` (si no, el cgroup hijo no tiene esos controladores). Resultado: `subtree_control: memory pids`.

## 1 · UTS

```
fuera:  668f731b9f82
caja
fuera, después: 668f731b9f82
```

## 2 · PID

```
dentro, mi PID: 1
PID   COMMAND
    1 ps -o pid,args
fuera, el mismo sleep:
   1195 ./mini.bin run rootfs /bin/sleep 2
   1202 /bin/sleep 2
fuera, procesos visibles: 6
```

## 3 · mount (pivot_root + /proc propio + /dev mínimo)

```
fuera:  PRETTY_NAME="Debian GNU/Linux 13 (trixie)"
dentro: PRETTY_NAME="Alpine Linux v3.24"
ls /: bin dev etc home lib media mnt opt proc root run sbin srv sys tmp usr var
montajes:
overlay / overlay rw,relatime,lowerdir=/var/lib/desktop-containerd/...   (línea larga, recortada)
proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /dev tmpfs rw,nosuid,relatime,mode=755 0 0
fuera, montajes del contenedor en mi tabla: 0
```

Tres montajes dentro. La raíz sale como `overlay` porque el rootfs vive dentro del disco del contenedor externo.

## 4-6 · cgroups: memoria

```
== 4 · sin límite ==
reservados 512 MB
terminé: nadie me paró
[mini] memory.peak=541507584 oom_kill=0

== 5 · memory.max=64M y memory.swap.max=0 ==
reservados 40 MB
reservados 48 MB
reservados 56 MB
[mini] el proceso murió por la señal killed
[mini] memory.peak=67108864 oom_kill=1
código de salida: 137 · tardó 196 ms        (otra corrida: 199 ms)
Memory cgroup out of memory: Killed process 16897 (comer) total-vm:1294556kB, anon-rss:64320kB

== 6 · memory.max=64M SIN memory.swap.max=0 (cgroup hecho a mano) ==
/dev/sdc  partition  4194304   (4 GB de swap en la VM de WSL2)
reservados 512 MB
terminé: nadie me paró
memory.peak=67108864 memory.swap.peak=475865088 oom_kill=0
```

- Con el límite, `comer` llega a **56 MB** y el kernel lo mata (SIGKILL, código 137) al tocar los 64 MB: `memory.peak` = 67.108.864 = 64 MiB exactos.
- **Sin `memory.swap.max=0` nadie muere**: los 512 MB entran, 64 en RAM y **~454 MiB (475.865.088 bytes) en swap**.
- Docker hace lo mismo (`salidas/docker.txt`): `--memory 64m` solo → `memory.swap.max=67108864` (otros 64 MB de swap);
  con `--memory-swap 64m` → `memory.swap.max=0`.

## 7 · cgroups: bomba fork con pids.max=20

`sh -c 'b(){ b | b & }; b 2>/dev/null; exec sleep 5'` — vive 5 s; al morir el PID 1, el kernel mata todo el namespace.

```
a los 2 s: pids.current=20 · rechazos=11
fuera, ls / mientras tanto: 1 ms
fuera, procesos totales: 25
[mini] memory.peak=5111808 oom_kill=0 pids.peak=20 pids.max_rechazos=11
después: cgroups mini que quedan: 0
```

Otras corridas: 11 y 12 rechazos, `ls /` en 1 y 9 ms. Siempre `pids.peak=20`.

## 8-9 · red

```
== sin namespace de red ==
1: lo    inet 127.0.0.1/8 scope host lo
2: eth0    inet 172.17.0.2/16 brd 172.17.255.255 scope global
1 packets transmitted, 1 packets received, 0% packet loss

== con -red (CLONE_NEWNET) ==
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN qlen 1000
PING 1.1.1.1 (1.1.1.1): 56 data bytes
ping: sendto: Network unreachable
ping salió con 1
```

Sin `-red`, el contenedor usa la red del anfitrión tal cual (misma eth0, misma IP). Con `-red`, solo un loopback apagado.
Docker por defecto (`docker.txt`): `eth0 172.17.0.2/16`, `default via 172.17.0.1 dev eth0`, `nameserver 192.168.65.7`
(par veth al puente, ruta, NAT y DNS: nada de eso está en mini).

## 10 · la raíz es una carpeta compartida

```
mini:   hola                                   (el 2.º contenedor lee lo que escribió el 1.º)
        fuera, en la carpeta del rootfs: escrito-por-el-primero
docker: escrito
        cat: can't open '/escrito-por-el-primero': No such file or directory
docker image inspect alpine:3.24.3 → 1 capa(s)
```

## 11 · capabilities y seccomp

```
mini (dentro del --privileged)          | docker run alpine:3.24.3 (perfil por defecto)
CapEff:  000001ffffffffff               | CapEff:  00000000a80425fb
Seccomp_filters: 1 (heredado)           | Seccomp_filters: 2
hostname otro: permitido                | hostname: sethostname: Operation not permitted
mount tmpfs: permitido                  | mount: permission denied (are you root?)
unshare -U: permitido                   | unshare: unshare(0x10000000): Operation not permitted
/dev: null urandom zero                 | /dev: core fd full mqueue null ptmx pts random shm stderr stdin stdout tty urandom zero
```

- `000001ffffffffff` = las 41 capabilities; `a80425fb` = 14 (las de Docker por defecto).
- Seccomp: Docker Desktop ya pone 1 filtro a todo (lo hereda el `--privileged` y mini). El perfil por defecto de Docker
  añade el 2.º. Con `--security-opt seccomp=unconfined`: `Seccomp_filters: 1` y `unshare -U: permitido` → el que bloquea
  `unshare -U` es el seccomp de Docker, no las capabilities.

## 12 · arranque

```
mini run /bin/true (dentro del contenedor, 3 rondas de 20): 3 · 2 · 3 ms
docker run --rm alpine:3.24.3 true (desde Windows, 3 rondas de 10): 486 · 609 · 575 ms
```

⚠️ No es la misma comparación: el `docker run` cruza CLI de Windows → daemon → containerd → runc, prepara overlay y red;
y la máquina tenía 59 contenedores de otros agentes corriendo. Otra corrida dio 432 y 1.079 ms. Sirve como orden de
magnitud, no como benchmark.

## Fuentes de reglas citadas

- namespaces(7), pivot_root(2), clone(2) `CLONE_INTO_CGROUP`: man7.org
- cgroup v2, `memory.max`, `memory.swap.max`, `pids.max`, regla de «no internal processes»: docs.kernel.org/admin-guide/cgroup-v2.html
- Docker: `--memory-swap` (docs.docker.com/engine/containers/resource_constraints), perfil seccomp por defecto
  (docs.docker.com/engine/security/seccomp), capabilities por defecto (docs.docker.com/engine/containers/run/#runtime-privilege-and-linux-capabilities)
- Go: `syscall.SysProcAttr.UseCgroupFD` / `CgroupFD` (pkg.go.dev/syscall)
