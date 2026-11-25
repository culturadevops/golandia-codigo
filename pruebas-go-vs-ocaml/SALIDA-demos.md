# SALIDA-demos — ¿Go o OCaml 5? (golandia 903), sin cronómetro

> 2026-11-23 07:06. Regenerable con `python recursos/pruebas-go-vs-ocaml/demos.py`.

## 1 · Núcleos que ve y hilos que arranca, en un contenedor con --cpus 2

```
$ docker run --cpus 2 -e DOMINIOS=1 gl-903-ocaml:1
LISTO_EN_MS 3.949 NUCLEOS 24 DOMINIOS 1
# sigue vivo · Threads:	1
```

```
$ docker run --cpus 2 -e DOMINIOS=2 gl-903-ocaml:1
LISTO_EN_MS 1.792 NUCLEOS 24 DOMINIOS 2
# sigue vivo · Threads:	4
```

```
$ docker run --cpus 2 -e DOMINIOS=24 gl-903-ocaml:1
LISTO_EN_MS 1.750 NUCLEOS 24 DOMINIOS 24
# sigue vivo · Threads:	48
```

## 2 · El motor de Eio dentro de Docker

```
$ docker run --cpus 2 -e EIO_BACKEND=linux gl-903-ocaml:1   # exigir io_uring
Fatal error: exception Failure("io_uring is not available (permission denied)")
# el servidor se cayó
```

```
$ docker run --cpus 2 --security-opt seccomp=seccomp-io_uring.json -e EIO_BACKEND=linux gl-903-ocaml:1
LISTO_EN_MS 0.135 NUCLEOS 24 DOMINIOS 1
# sigue vivo · Threads:	1
```

Sin EIO_BACKEND, el mismo contenedor arranca igual (sección 1): Eio prueba io_uring, no puede, y usa posix sin decir nada.

## 3 · Versiones y lo que necesita el binario

```
$ ocamlopt -version; opam list --installed ...
5.5.2
cohttp-eio 6.3.0
dune       3.25.0
eio        1.7
eio_linux  1.7
eio_posix  1.7
ocaml      5.5.2
```

```
$ ldd /servidor
linux-vdso.so.1
	libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
	/lib64/ld-linux-x86-64.so.2
```

