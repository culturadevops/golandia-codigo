# SALIDA-demos — ¿Go o Odin? (golandia 905), sin cronómetro

> 2026-11-23 07:05. Regenerable con `python recursos/pruebas-go-vs-odin/demos.py`.

## 1 · El servidor con el perfil seccomp por defecto de Docker, y con io_uring permitido

```
$ docker run --cpus 2 gl-905-odin:1
LISTO_EN_MS 0.006 NUCLEOS 24 HILOS 0
/src/odin-http/server.odin(131:2) runtime assertion: acquire_err == nil
# el servidor se cayó
```

```
$ docker run --cpus 2 --security-opt seccomp=unconfined gl-905-odin:1
LISTO_EN_MS 0.008 NUCLEOS 24 HILOS 0
# sigue vivo · Threads:	24
```

```
$ docker run --cpus 2 --security-opt seccomp=seccomp-io_uring.json gl-905-odin:1   # el de Docker + io_uring
LISTO_EN_MS 0.005 NUCLEOS 24 HILOS 0
# sigue vivo · Threads:	24
```

```
$ docker run --cpus 2 --security-opt seccomp=seccomp-io_uring.json -e HILOS=2 gl-905-odin:1
LISTO_EN_MS 0.008 NUCLEOS 24 HILOS 2
# sigue vivo · Threads:	2
```

## 2 · Lo que necesita el binario

```
$ ldd /servidor
linux-vdso.so.1
	libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
	/lib64/ld-linux-x86-64.so.2
```

