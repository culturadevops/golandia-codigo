# SALIDA-demos — ¿Go o Kotlin Native? (golandia 879)

> Corrida del 2026-11-18 10:19. Sin cronómetro de carga.

## 1 · ¿Cuántos núcleos ve cada uno? (contenedor con --cpus 2)

```
$ docker run --cpus 2 gl-879-kotlin-jvm:1
LISTO_EN_MS 208.387 PLATAFORMA jvm NUCLEOS 2 HILOS Default
$ docker run --cpus 2 gl-879-kotlin-graal:1
LISTO_EN_MS 2.507 PLATAFORMA jvm NUCLEOS 2 HILOS Default
$ docker run --cpus 2 gl-879-kotlin-native:1
LISTO_EN_MS 3.644 PLATAFORMA native NUCLEOS 24 HILOS Default
$ nproc   # en un contenedor de 2 CPU
24
```

## 2 · ¿De qué depende cada binario? (ldd)

```
$ ldd /servidor   # gl-879-kotlin-native:1
	linux-vdso.so.1 (0x000075e2a4c57000)
	libresolv.so.2 => /lib/x86_64-linux-gnu/libresolv.so.2 (0x000075e2a4c3c000)
	libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6 (0x000075e2a4b4c000)
	libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x000075e2a4b47000)
	libutil.so.1 => /lib/x86_64-linux-gnu/libutil.so.1 (0x000075e2a4b42000)
	libcrypt.so.1 => /lib/x86_64-linux-gnu/libcrypt.so.1 (0x000075e2a4b04000)
	librt.so.1 => /lib/x86_64-linux-gnu/librt.so.1 (0x000075e2a4aff000)
	libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x000075e2a4afa000)
	libgcc_s.so.1 => /lib/x86_64-linux-gnu/libgcc_s.so.1 (0x000075e2a4acd000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x000075e2a48d9000)
	/lib64/ld-linux-x86-64.so.2 (0x000075e2a4c59000)
```

```
$ ldd /servidor   # gl-879-kotlin-graal:1
	linux-vdso.so.1 (0x00007fc1b18c5000)
	libz.so.1 => /lib/x86_64-linux-gnu/libz.so.1 (0x00007fc1b189c000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007fc1b16a8000)
	/lib64/ld-linux-x86-64.so.2 (0x00007fc1b18c7000)
```

## 3 · GraalVM: versión y recolector

```
$ native-image --version
native-image 25.0.2 2026-01-20
GraalVM Runtime Environment GraalVM CE 25.0.2+10.1 (build 25.0.2+10-jvmci-b01)
Substrate VM GraalVM CE 25.0.2+10.1 (build 25.0.2+10, serial gc)
```

## 4 · Las tareas de Gradle que enlazan el binario nativo

```
$ gradle tasks --all | grep -i 'link.*LinuxX64'
linkDebugExecutableLinuxX64 - Links an executable 'debugExecutable' for a target 'linuxX64'.
linkDebugTestLinuxX64 - Links a test executable 'debugTest' for a target 'linuxX64'.
linkLinuxX64 - Links all binaries for target 'linuxX64'.
linkReleaseExecutableLinuxX64 - Links an executable 'releaseExecutable' for a target 'linuxX64'.
runDebugExecutableLinuxX64 - Executes Kotlin/Native executable debugExecutable for target linuxX64
runReleaseExecutableLinuxX64 - Executes Kotlin/Native executable releaseExecutable for target linuxX64
```

