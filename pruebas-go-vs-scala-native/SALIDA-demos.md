# SALIDA-demos — ¿Go o Scala Native? (golandia 865)

> Corrida del 2026-11-16 16:45. Sin cronómetro de carga: los tiempos «Total» son los que imprime scala-cli y se dan solo como orientación (la comparación de build es la de SALIDA.md, con el candado).

## 1 · Sin --native-mode: ¿en qué modo compila?

```
$ scala-cli --power package --native Servidor.scala -o servidor
[info] Linking (multithreadingEnabled=detect) (889 ms)
[info] Optimizing (debug mode) (638 ms)
[info] Compiling to native code (912 ms)
[info] Linking native code (immix gc, none lto) (83 ms)
[info] Total (3437 ms)
6385560 /tmp/s
[codigo de salida: 0]
```

## 2 · Con --native-mode release-full

```
$ scala-cli --power package --native --native-mode release-full Servidor.scala -o servidor
[info] Linking (multithreadingEnabled=detect) (916 ms)
[info] Optimizing (release-full mode) (5630 ms)
[info] Compiling to native code (7397 ms)
[info] Linking native code (immix gc, none lto) (84 ms)
[info] Total (15190 ms)
5751616 /tmp/s
$ ldd servidor
	linux-vdso.so.1 (0x000076b194f43000)
	libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6 (0x000076b1949c3000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x000076b1947cf000)
	/lib64/ld-linux-x86-64.so.2 (0x000076b194f45000)
[codigo de salida: 0]
```

## 3 · El mismo archivo, en la JVM

```
$ scala-cli --power package --assembly Servidor.scala -o servidor.jar
Wrote /tmp/s.jar
9856657 /tmp/s.jar
[codigo de salida: 0]
```

## 4 · Líneas del servidor

```
$ grep -cv '^ *//' Servidor.scala
91
102
[codigo de salida: 0]
```

## 5 · ¿Cuántos núcleos ve cada uno? (contenedor con --cpus 2)

```
$ docker run --cpus 2 v865-scala-jvm:1
LISTO_EN_MS 67.152 NUCLEOS 2 HILOS 2
$ docker run --cpus 2 v865-scala-native-sintrap:1
LISTO_EN_MS 0.176 NUCLEOS 24 HILOS 24
$ nproc   # en el mismo contenedor de 2 CPU
24
```

