# SALIDA — Goroutines bloqueadas en una llamada al sistema (golandia)

> Corrida del 2026-10-20 16:51. golang:1.27.2, contenedor con --cpus 2 --memory 4g. Hilos contados a 1 s de una espera de 2 s.

```
$ ./bloqueo sleep 1000
sleep    goroutines=1000 hilos_del_SO=5
$ ./bloqueo red 1000
red      goroutines=1000 hilos_del_SO=5
$ ./bloqueo syscall 1000
syscall  goroutines=1000 hilos_del_SO=1004
$ ./bloqueo sleep 5000
sleep    goroutines=5000 hilos_del_SO=5
$ ./bloqueo red 5000
red      goroutines=5000 hilos_del_SO=6
$ ./bloqueo syscall 5000
syscall  goroutines=5000 hilos_del_SO=5004
$ ./bloqueo syscall 12000
syscall  goroutines=12000 hilos_del_SO=8691
runtime: program exceeds 10000-thread limit
fatal error: thread exhaustion
(código de salida: 2)
```
