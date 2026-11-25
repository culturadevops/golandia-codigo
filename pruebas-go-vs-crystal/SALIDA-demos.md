# SALIDA-demos — ¿Go o Crystal? (golandia 902), sin cronómetro

> 2026-11-23 07:06. Regenerable con `python recursos/pruebas-go-vs-crystal/demos.py`.

## 1 · Núcleos que ve y hilos, en un contenedor con --cpus 2

La imagen final es scratch (no tiene shell): se mira el mismo binario desde la etapa de compilación (alpine).

```
$ docker run --cpus 2 gl-902-crystal (sin HILOS)
LISTO_EN_MS 0.021 NUCLEOS 24 HILOS 1 (por defecto)
# sigue vivo · Threads:	17
```

```
$ docker run --cpus 2 -e HILOS=2 gl-902-crystal
LISTO_EN_MS 0.054 NUCLEOS 24 HILOS 2
# sigue vivo · Threads:	17
```

## 2 · El binario

```
$ ldd /servidor; ls -l /servidor (--release --no-debug --static) /tmp/con-debug (--release --static)
/lib/ld-musl-x86_64.so.1 (0x7cc4a1380000)
22959776 /servidor
24720960 /tmp/con-debug
```

```
$ ls -l /servidor   # crystal build --static, sin --release
26339536 /servidor
```

## 3 · Versión

```
$ crystal --version
Crystal 1.22.3 [3462c0774] (2026-11-13)

LLVM: 22.1.3
Default target: x86_64-unknown-linux-musl
```

