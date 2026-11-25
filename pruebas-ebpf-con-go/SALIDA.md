# SALIDA — eBPF con Go (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-10 18:31 UTC
- imagen: fab-go-ebpf:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## El kernel: ¿soporta eBPF?

`$ uname -r && ls -l /sys/kernel/btf/vmlinux && zcat /proc/config.gz | grep BPF`

```
6.18.33.2-microsoft-standard-WSL2
6677359 /sys/kernel/btf/vmlinux
CONFIG_BPF_SYSCALL=y
CONFIG_BPF_JIT=y
CONFIG_DEBUG_INFO_BTF=y
```

## Sin escribir nada: el tracepoint en crudo

`$ echo 1 > /sys/kernel/tracing/events/sock/inet_sock_set_state/enable
$ curl 127.0.0.1:8080
$ cat /sys/kernel/tracing/trace`

```
family=AF_INET sport=0 dport=8080 saddr=127.0.0.1 daddr=127.0.0.1 oldstate=TCP_CLOSE newstate=TCP_SYN_SENT
family=AF_INET sport=44000 dport=8080 saddr=127.0.0.1 daddr=127.0.0.1 oldstate=TCP_SYN_SENT newstate=TCP_ESTABLISHED
family=AF_INET sport=44000 dport=8080 saddr=127.0.0.1 daddr=127.0.0.1 oldstate=TCP_ESTABLISHED newstate=TCP_FIN_WAIT1
```

## Compilar: bpf2go + go build

`$ go generate   # bpf2go compila conexiones.c con clang
$ go build -o conexiones .`

```
go version go1.27.2 linux/amd64
require github.com/cilium/ebpf v0.22.0

5863831 conexiones
3724 conexiones_bpfel.go
5232 conexiones_bpfel.o
```

## El verificador del kernel

`$ __u16 puerto = saliente ? a->dport : a->sport;   # elegir directamente sobre el contexto
$ ./conexiones`

```
2026/10/10 18:30:40 cargar el programa eBPF: assign values: field Conexion: program conexion: load program: permission denied: dereference of modified ctx ptr R1 off=26 disallowed (24 line(s) omitted)
```

## Ver las conexiones de la máquina

`$ ./conexiones -durante 6s   # y, mientras, 3 curl al servidor local y 1 a 1.1.1.1`

```
escuchando 6s…
saliente curl       pid 15785  127.0.0.1 → 127.0.0.1:8080
entrante 127.0.0.1:8080 → 127.0.0.1:60228
saliente curl       pid 15786  127.0.0.1 → 127.0.0.1:8080
entrante 127.0.0.1:8080 → 127.0.0.1:60240
saliente curl       pid 15787  127.0.0.1 → 127.0.0.1:8080
entrante 127.0.0.1:8080 → 127.0.0.1:60250
saliente curl       pid 15788  172.17.0.3 → 1.1.1.1:443
entrante 192.168.65.7:2376 → 192.168.65.1:34524
entrante 192.168.65.7:2376 → 192.168.65.1:52175
entrante 192.168.65.7:2376 → 192.168.65.1:52891
entrante 192.168.65.7:2376 → 192.168.65.1:50518
entrante 192.168.65.7:2376 → 192.168.65.1:57790
entrante 192.168.65.7:2376 → 192.168.65.1:20158
entrante 192.168.65.7:2376 → 192.168.65.1:59552
entrante 192.168.65.7:2376 → 192.168.65.1:28987
entrante 192.168.65.7:2376 → 192.168.65.1:55972
entrante 192.168.65.7:2376 → 192.168.65.1:26306
entrante 192.168.65.7:2376 → 192.168.65.1:32621
entrante 192.168.65.7:2376 → 192.168.65.1:57467
entrante 192.168.65.7:2376 → 192.168.65.1:57728
entrante 192.168.65.7:2376 → 192.168.65.1:58438
entrante 192.168.65.7:2376 → 192.168.65.1:42068
entrante 192.168.65.7:2376 → 192.168.65.1:42613
entrante 192.168.65.7:2376 → 192.168.65.1:50682
entrante 192.168.65.7:2376 → 192.168.65.1:36883
entrante 192.168.65.7:2376 → 192.168.65.1:46602
entrante 192.168.65.7:2376 → 192.168.65.1:43084
entrante 192.168.65.7:2376 → 192.168.65.1:61813
entrante 192.168.65.7:2376 → 192.168.65.1:17265
entrante 192.168.65.7:2376 → 192.168.65.1:53099
entrante 192.168.65.7:2376 → 192.168.65.1:64652
entrante 192.168.65.7:2376 → 192.168.65.1:53955
entrante 192.168.65.7:2376 → 192.168.65.1:24350
entrante 192.168.65.7:2376 → 192.168.65.1:29297
entrante 192.168.65.7:2376 → 192.168.65.1:44670
entrante 192.168.65.7:2376 → 192.168.65.1:62964
entrante 192.168.65.7:2376 → 192.168.65.1:46385
entrante 192.168.65.7:2376 → 192.168.65.1:24642
entrante 192.168.65.7:2376 → 192.168.65.1:52963
entrante 192.168.65.7:2376 → 192.168.65.1:64076
entrante 192.168.65.7:2376 → 192.168.65.1:33537
entrante 192.168.65.7:2376 → 192.168.65.1:25386
entrante 192.168.65.7:2376 → 192.168.65.1:61326
entrante 192.168.65.7:2376 → 192.168.65.1:42796
entrante 192.168.65.7:2376 → 192.168.65.1:16801
entrante 192.168.65.7:2376 → 192.168.65.1:16711
entrante 192.168.65.7:2376 → 192.168.65.1:35665
entrante 192.168.65.7:2376 → 192.168.65.1:35319
entrante 192.168.65.7:2376 → 192.168.65.1:36068
entrante 192.168.65.7:2376 → 192.168.65.1:40221
entrante 192.168.65.7:2376 → 192.168.65.1:62004
entrante 192.168.65.7:2376 → 192.168.65.1:36579
entrante 192.168.65.7:2376 → 192.168.65.1:16200
entrante 192.168.65.7:2376 → 192.168.65.1:44786
--- conexiones por puerto del servicio (mapa del kernel)
puerto 443   1
puerto 2376  47
puerto 8080  6
```

## Una conexión que falla

`$ ./conexiones -durante 4s   # y curl a 127.0.0.1:9, donde no escucha nadie`

```
curl: código 7
saliente curl       pid 15868  127.0.0.1 → 127.0.0.1:9
puerto 9     1
```

## Otro contenedor, que no sabe nada

`$ ./conexiones -durante 10s   # mientras OTRO contenedor hace curl a 1.1.1.1 y example.com`

```
saliente curl       pid 15995  172.17.0.4 → 1.1.1.1:443
saliente curl       pid 15957  172.17.0.4 → 172.66.147.243:80
--- conexiones por puerto del servicio (mapa del kernel)
puerto 80    1
puerto 443   1
```

## Sin privilegios

`$ docker run (sin --privileged) ... ./conexiones`

```
2026/10/10 18:31:08 cargar el programa eBPF: assign values: field Conexion: program conexion: map por_puerto: map create: operation not permitted (MEMLOCK may be too low, consider rlimit.RemoveMemlock)
```
