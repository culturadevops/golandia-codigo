# SALIDA — Curso de Rust para operadores Linux 2027

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Cada comando va con su código de salida;
> los fallos y las alertas son a propósito: cada lección enseña una trampa.

- fecha: `2026-11-13 14:17:09Z`
- `rustc 1.99.1 (58cf8f9c3 2026-10-13)`
- `cargo 1.99.1 (5f94df478 2026-08-27)`
- `Linux 6.18.33.2-microsoft-standard-WSL2 x86_64`
- `docker run --memory=1g --memory-swap=1g --cpus=2 --shm-size=256m`

## parte1/01_du

`bash parte1/01_du/demo.sh` · código de salida `0`

```
$ truncate -s 1G datos/disco-virtual.img
[codigo de salida: 0]
$ head -c 5000000 /dev/urandom > datos/fotos.bin
[codigo de salida: 0]
$ ls -l datos
total 4884
-rw-r--r-- 1 root root 1073741824 Nov 13 14:14 disco-virtual.img
-rw-r--r-- 1 root root    5000000 Nov 13 14:14 fotos.bin
[codigo de salida: 0]
$ ops-du datos
archivos vistos:          2
suma ingenua de len():    1.0 GiB (1078741824 bytes)
len() sin repetir inode:  1.0 GiB (1078741824 bytes)
bloques en disco:         4.8 MiB (5001216 bytes)
[codigo de salida: 0]
$ du -sh datos
4.8M	datos
[codigo de salida: 0]
$ du -sh --apparent-size datos
1.1G	datos
[codigo de salida: 0]
```

## parte1/02_enlaces

`bash parte1/02_enlaces/demo.sh` · código de salida `0`

```
$ head -c 20000000 /dev/zero > logs/app.log
[codigo de salida: 0]
$ ln logs/app.log logs/app-copia1.log && ln logs/app.log logs/app-copia2.log
[codigo de salida: 0]
$ ls -li logs
total 58596
208425 -rw-r--r-- 3 root root 20000000 Nov 13 14:14 app-copia1.log
208425 -rw-r--r-- 3 root root 20000000 Nov 13 14:14 app-copia2.log
208425 -rw-r--r-- 3 root root 20000000 Nov 13 14:14 app.log
[codigo de salida: 0]
$ ops-du logs
archivos vistos:          3
suma ingenua de len():    57.2 MiB (60000000 bytes)
len() sin repetir inode:  19.1 MiB (20000000 bytes)
bloques en disco:         19.1 MiB (20000768 bytes)
[codigo de salida: 0]
$ du -sh logs
20M	logs
[codigo de salida: 0]
$ ln -s .. bucle/a/arriba
[codigo de salida: 0]
$ ops-du bucle --seguir-enlaces
error: File system loop found: bucle/a/arriba points to an ancestor bucle
archivos vistos:          0
suma ingenua de len():    0 B (0 bytes)
len() sin repetir inode:  0 B (0 bytes)
bloques en disco:         0 B (0 bytes)
errores:                  1
[codigo de salida: 0]
$ ops-du bucle
archivos vistos:          0
suma ingenua de len():    0 B (0 bytes)
len() sin repetir inode:  0 B (0 bytes)
bloques en disco:         0 B (0 bytes)
[codigo de salida: 0]
```

## parte1/03_grandes

`bash parte1/03_grandes/demo.sh` · código de salida `0`

```
$ ops-grandes /usr 5
 198.5 MiB  /usr/local/rustup/toolchains/1.99.1-x86_64-unknown-linux-gnu/lib/libLLVM.so.23.1-rust-1.99.1-stable
 150.9 MiB  /usr/local/rustup/toolchains/1.99.1-x86_64-unknown-linux-gnu/lib/librustc_driver-7ed0375ae352bd75.so
  64.8 MiB  /usr/local/rustup/toolchains/1.99.1-x86_64-unknown-linux-gnu/lib/rustlib/x86_64-unknown-linux-gnu/lib/libcore-1d288395a3394ed3.rmeta
  34.7 MiB  /usr/libexec/gcc/x86_64-linux-gnu/14/cc1plus
  34.5 MiB  /usr/local/rustup/toolchains/1.99.1-x86_64-unknown-linux-gnu/bin/cargo
22906 archivos revisados en 764 ms
[codigo de salida: 0]
$ find /usr -type f -printf "%s %p\n" | sort -rn | head -5
208167760 /usr/local/rustup/toolchains/1.99.1-x86_64-unknown-linux-gnu/lib/libLLVM.so.23.1-rust-1.99.1-stable
158240640 /usr/local/rustup/toolchains/1.99.1-x86_64-unknown-linux-gnu/lib/librustc_driver-7ed0375ae352bd75.so
67966594 /usr/local/rustup/toolchains/1.99.1-x86_64-unknown-linux-gnu/lib/rustlib/x86_64-unknown-linux-gnu/lib/libcore-1d288395a3394ed3.rmeta
36389344 /usr/libexec/gcc/x86_64-linux-gnu/14/cc1plus
36148320 /usr/local/rustup/toolchains/1.99.1-x86_64-unknown-linux-gnu/bin/cargo
[codigo de salida: 0]
$ find /usr -type f | wc -l
22906
[codigo de salida: 0]
```

## parte1/04_libre

`bash parte1/04_libre/demo.sh` · código de salida `0`

```
$ ops-libre /
/
  total:       1006.9 GiB
  usado:         42.7 GiB
  libre:        964.2 GiB  (f_bfree, cuenta lo reservado para root)
  disponible:   913.0 GiB  (f_bavail, lo que puede usar un usuario)
  reservado:     51.2 GiB
  uso:                 5%
[codigo de salida: 0]
$ df -h /
Filesystem      Size  Used Avail Use% Mounted on
overlay        1007G   43G  913G   5% /
[codigo de salida: 0]
$ ops-libre /dev/shm
/dev/shm
  total:        256.0 MiB
  usado:              0 B
  libre:        256.0 MiB  (f_bfree, cuenta lo reservado para root)
  disponible:   256.0 MiB  (f_bavail, lo que puede usar un usuario)
  reservado:          0 B
  uso:                 0%
[codigo de salida: 0]
$ df -h /dev/shm
Filesystem      Size  Used Avail Use% Mounted on
shm             256M     0  256M   0% /dev/shm
[codigo de salida: 0]
$ ops-libre /no/existe
error: /no/existe: No such file or directory (os error 2)
[codigo de salida: 1]
```

## parte1/05_borrados

`bash parte1/05_borrados/demo.sh` · código de salida `0`

```
$ head -c 100000000 /dev/zero > app.log
[codigo de salida: 0]
$ ops-libre /dev/shm | grep -E "usado|uso"
  usado:         95.4 MiB
  uso:                38%
[codigo de salida: 0]
$ sleep 600 3< app.log &        # un proceso que lo deja abierto
$ rm app.log
[codigo de salida: 0]
$ ls -la
total 0
drwx------ 2 root root 40 Nov 13 14:14 .
drwxrwxrwt 3 root root 60 Nov 13 14:14 ..
[codigo de salida: 0]
$ du -sh .
0	.
[codigo de salida: 0]
$ ops-libre /dev/shm | grep -E "usado|uso"
  usado:         95.4 MiB
  uso:                38%
[codigo de salida: 0]
$ ops-borrados
pid 731    sleep      fd 3     95.4 MiB  /dev/shm/ops.UUMK/app.log (deleted)
1 archivo(s) borrado(s) y abiertos: 95.4 MiB
[codigo de salida: 0]
$ kill $P                       # cierro ese proceso
$ ops-libre /dev/shm | grep -E "usado|uso"
  usado:              0 B
  uso:                 0%
[codigo de salida: 0]
$ ops-borrados
0 archivo(s) borrado(s) y abiertos: 0 B
[codigo de salida: 0]
```

## parte2/01_meminfo

`bash parte2/01_meminfo/demo.sh` · código de salida `0`

```
$ grep -E "^(MemTotal|MemFree|MemAvailable|Cached):" /proc/meminfo
MemTotal:       16101040 kB
MemFree:        10887924 kB
MemAvailable:   14438976 kB
Cached:          3500840 kB
[codigo de salida: 0]
$ ops-mem
/proc/meminfo
  MemTotal:        15.4 GiB
  MemFree:         10.4 GiB
  MemAvailable:    13.8 GiB
  Cached:           3.3 GiB
cgroup (/sys/fs/cgroup)
  memory.max:      1073741824
  memory.current:  156.8 MiB
  límite real:     1.0 GiB
  te quedan:       867.2 MiB
[codigo de salida: 0]
$ free -h
               total        used        free      shared  buff/cache   available
Mem:            15Gi       1.6Gi        10Gi        26Mi       3.6Gi        13Gi
Swap:          4.0Gi          0B       4.0Gi
[codigo de salida: 0]
```

## parte2/02_cgroup

`bash parte2/02_cgroup/demo.sh` · código de salida `0`

```
$ cat /sys/fs/cgroup/memory.max
1073741824
[codigo de salida: 0]
$ ops-mem
/proc/meminfo
  MemTotal:        15.4 GiB
  MemFree:         10.4 GiB
  MemAvailable:    13.8 GiB
  Cached:           3.3 GiB
cgroup (/sys/fs/cgroup)
  memory.max:      1073741824
  memory.current:  157.7 MiB
  límite real:     1.0 GiB
  te quedan:       866.3 MiB
[codigo de salida: 0]
$ ops-comer 500
reservados: 100 MiB
reservados: 200 MiB
reservados: 300 MiB
reservados: 400 MiB
reservados: 500 MiB
listo: 500 MiB en 5 bloques
[codigo de salida: 0]
$ ops-comer 2000
reservados: 100 MiB
reservados: 200 MiB
reservados: 300 MiB
reservados: 400 MiB
reservados: 500 MiB
reservados: 600 MiB
reservados: 700 MiB
reservados: 800 MiB
reservados: 900 MiB
reservados: 1000 MiB
/work/repo/comun.sh: line 9:   766 Killed                  ops-comer 2000
[codigo de salida: 137]
$ grep oom_kill /sys/fs/cgroup/memory.events
oom_kill 1
[codigo de salida: 0]
```

## parte2/03_cpu

`bash parte2/03_cpu/demo.sh` · código de salida `0`

```
$ nproc
24
[codigo de salida: 0]
$ cat /sys/fs/cgroup/cpu.max
200000 100000
[codigo de salida: 0]
$ ops-cpu 1
CPUs en /proc/stat:            24
available_parallelism():       2
cpu.max del cgroup:            200000 100000
/proc/stat, toda la máquina:   0.0 % = 0.00 núcleos
cpu.stat, este contenedor:     0.00 núcleos
veces que lo frenó el límite:  0
[codigo de salida: 0]
$ ops-quemar 2 4 &              # dos hilos al 100 % durante 4 s
$ ops-cpu 2
CPUs en /proc/stat:            24
available_parallelism():       2
cpu.max del cgroup:            200000 100000
/proc/stat, toda la máquina:   8.4 % = 2.02 núcleos
cpu.stat, este contenedor:     2.00 núcleos
veces que lo frenó el límite:  14
[codigo de salida: 0]
$ ops-quemar 8 4 &              # ocho hilos, con un límite de 2 CPUs
$ ops-cpu 2
CPUs en /proc/stat:            24
available_parallelism():       2
cpu.max del cgroup:            200000 100000
/proc/stat, toda la máquina:   8.5 % = 2.04 núcleos
cpu.stat, este contenedor:     2.00 núcleos
veces que lo frenó el límite:  20
[codigo de salida: 0]
$ cat /proc/loadavg
1.94 2.31 1.25 3/517 795
[codigo de salida: 0]
```

## parte2/04_ps

`bash parte2/04_ps/demo.sh` · código de salida `0`

```
$ "./mi app (v2) prod" 300 &      # un proceso con espacios y paréntesis en el nombre
$ bash -c "sleep 0.2 & exec sleep 300" &   # su hijo queda zombi: nadie lo espera
$ ops-ps
   PID   PPID EST        RSS  NOMBRE
     1      0   S    3.1 MiB  bash
   796      1   S    3.3 MiB  bash
   799    796   S    1.9 MiB  mi app (v2) pro
   800    796   S    1.9 MiB  sleep
   802    800   Z        0 B  sleep
   803    796   S    1.6 MiB  bash
   804    803   R    2.0 MiB  ops-ps
[codigo de salida: 0]
$ ops-ps --ingenuo
   PID   PPID EST        RSS  NOMBRE
     1      0   S    3.1 MiB  bash
   796      1   S    3.3 MiB  bash
   799   (v2) app    1.9 MiB  mi
   800    796   S    1.9 MiB  sleep
   802    800   Z        0 B  sleep
   805    796   S    1.8 MiB  bash
   806    805   R    2.0 MiB  ops-ps
[codigo de salida: 0]
$ ps -eo pid,ppid,stat,rss,comm
    PID    PPID STAT   RSS COMMAND
      1       0 Ss    3216 bash
    796       1 S     3372 bash
    799     796 S     1896 mi app (v2) pro
    800     796 S     1968 sleep
    802     800 Z        0 sleep
    807     796 S     1800 bash
    808     807 R     3732 ps
[codigo de salida: 0]
```

## parte2/05_senales

`bash parte2/05_senales/demo.sh` · código de salida `0`

```
$ bash -c 'trap "echo [hijo] recibí TERM, cierro; exit 0" TERM; while :; do sleep 0.1; done' &
$ ops-matar $P 3
SIGTERM enviado a 810
[hijo] recibí TERM, cierro
salió solo en 50 ms
[codigo de salida: 0]
$ bash -c 'trap "" TERM; while :; do sleep 0.1; done' &      # este ignora SIGTERM
$ ops-matar $P 2
SIGTERM enviado a 817
no salió en 2 s: SIGKILL a los 2008 ms
/work/repo/comun.sh: line 7:   817 Killed                  bash -c 'trap "" TERM; while :; do sleep 0.1; done'
[codigo de salida: 0]
$ ops-matar 999999
pid 999999: No such process (os error 3)
[codigo de salida: 1]
```

## parte3/01_seguir

`bash parte3/01_seguir/demo.sh` · código de salida `0`

```
$ cat escritor.sh
# 60 líneas, una cada 50 ms. En la 20 se rota con mv; en la 40, con copytruncate.
for i in $(seq 1 60); do
	echo "linea $i" >> app.log
	sleep 0.05
	if [ "$i" = 20 ]; then mv app.log app.log.1; fi
	if [ "$i" = 40 ]; then cp app.log app.log.2; : > app.log; fi
done
[codigo de salida: 0]
$ ops-seguir app.log 5 --ingenuo &   y a la vez:  bash escritor.sh
líneas leídas: 20 (primera: «linea 1», última: «linea 20»)
[codigo de salida: 0]
$ wc -l app.log.1 app.log.2 app.log
 20 app.log.1
 20 app.log.2
 20 app.log
 60 total
[codigo de salida: 0]
$ ops-seguir app.log 5  &   y a la vez:  bash escritor.sh
[rotado: el nombre ya es otro archivo, inode 208432 -> 208433]
[truncado: 9 bytes, menos que la posición 171: vuelvo al inicio]
líneas leídas: 59 (primera: «linea 1», última: «linea 60»)
[codigo de salida: 0]
$ wc -l app.log.1 app.log.2 app.log
 20 app.log.1
 20 app.log.2
 20 app.log
 60 total
[codigo de salida: 0]
```

## parte3/02_filtrar

`bash parte3/02_filtrar/demo.sh` · código de salida `0`

```
$ wc -l app.log
200000 app.log
[codigo de salida: 0]
$ head -3 app.log
2026-11-13T03:00:00 WARN pedido=0 servicio=pagos
2026-11-13T03:00:00 INFO pedido=1 servicio=pagos
2026-11-13T03:00:00 ERROR pedido=2 servicio=pagos
[codigo de salida: 0]
$ ops-filtrar app.log
200000 líneas, 5918 errores, en 14 ms
  2026-11-13T03:00   130 errores
  2026-11-13T03:01   134 errores
  2026-11-13T03:02   135 errores
  peor minuto: 2026-11-13T03:17 con 2051 errores
[codigo de salida: 0]
$ ops-filtrar app.log --lento
200000 líneas, 5918 errores, en 117543 ms (regex compilada en cada línea)
  2026-11-13T03:00   130 errores
  2026-11-13T03:01   134 errores
  2026-11-13T03:02   135 errores
  peor minuto: 2026-11-13T03:17 con 2051 errores
[codigo de salida: 0]
$ grep -c " ERROR " app.log
5918
[codigo de salida: 0]
```

## parte3/03_capturar

`bash parte3/03_capturar/demo.sh` · código de salida `0`

```
$ bash ruidoso.sh > solo-stdout.log; echo "código: $?"
aviso: lote 50 lento
aviso: lote 100 lento
aviso: lote 150 lento
aviso: lote 200 lento
aviso: lote 250 lento
aviso: lote 300 lento
aviso: lote 350 lento
aviso: lote 400 lento
fallo al subir el resultado
código: 3
[codigo de salida: 0]
$ grep -c fallo solo-stdout.log
0
[codigo de salida: 1]
$ bash ruidoso.sh 2>&1 | tee todo.log > /dev/null; echo "código: $?"
código: 0
[codigo de salida: 0]
$ ops-capturar app.log 2048 3 -- bash ruidoso.sh
capturadas 400 líneas de stdout y 9 de stderr; el comando salió con 3
[codigo de salida: 3]
$ ls -l app.log*
-rw-r--r-- 1 root root  374 Nov 13 14:17 app.log
-rw-r--r-- 1 root root 2030 Nov 13 14:17 app.log.1
-rw-r--r-- 1 root root 2032 Nov 13 14:17 app.log.2
-rw-r--r-- 1 root root 2030 Nov 13 14:17 app.log.3
[codigo de salida: 0]
$ cat app.log app.log.1 app.log.2 app.log.3 | wc -l
248
[codigo de salida: 0]
$ head -1 app.log.3
[out] procesando lote 159
[codigo de salida: 0]
$ tail -2 app.log
[err] aviso: lote 400 lento
[err] fallo al subir el resultado
[codigo de salida: 0]
```

## parte3/04_guardia

`bash parte3/04_guardia/demo.sh` · código de salida `0`

```
$ ops-guardia /dev/shm 80 20
OK     disco /dev/shm: 0% usado (umbral 80%), disponibles 256.0 MiB
OK     memoria: 97% disponible de 1.0 GiB (umbral 20%)
INFO   carga 1 min: 1.37 para 2 CPU(s)
[codigo de salida: 0]
$ head -c 230000000 /dev/zero > relleno.bin
[codigo de salida: 0]
$ ops-guardia /dev/shm 80 20
ALERTA disco /dev/shm: 86% usado (umbral 80%), disponibles 36.7 MiB
OK     memoria: 76% disponible de 1.0 GiB (umbral 20%)
INFO   carga 1 min: 1.37 para 2 CPU(s)
[codigo de salida: 1]
$ rm relleno.bin
[codigo de salida: 0]
$ ops-comer 800 6 &              # algo que se come 800 MiB y los retiene
$ ops-guardia /dev/shm 80 20
OK     disco /dev/shm: 0% usado (umbral 80%), disponibles 256.0 MiB
ALERTA memoria: 19% disponible de 1.0 GiB (umbral 20%)
INFO   carga 1 min: 1.34 para 2 CPU(s)
[codigo de salida: 1]
$ ops-guardia /dev/shm 80 20 || echo "ALERTA: avisar a la guardia"
OK     disco /dev/shm: 0% usado (umbral 80%), disponibles 256.0 MiB
ALERTA memoria: 19% disponible de 1.0 GiB (umbral 20%)
INFO   carga 1 min: 1.34 para 2 CPU(s)
ALERTA: avisar a la guardia
[codigo de salida: 0]
$ ops-guardia /dev/shm 80 20
OK     disco /dev/shm: 0% usado (umbral 80%), disponibles 256.0 MiB
OK     memoria: 97% disponible de 1.0 GiB (umbral 20%)
INFO   carga 1 min: 1.23 para 2 CPU(s)
[codigo de salida: 0]
```
