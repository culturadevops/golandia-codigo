# SALIDA — Goroutines vs hilos (golandia)

> Corrida del 2026-10-20 16:41. golang:1.27.2, contenedor con --cpus 2 --memory 4g. Cada línea es la corrida mediana de 5.

```
$ go version; gcc --version | head -1; nproc; ulimit -s
go version go1.27.2 linux/amd64
gcc (Debian 14.2.0-19) 14.2.0
24
8192

$ ./goroutines 1000   (mediana de 5)
goroutines n=1000 crear=0.8ms memoria=2.7MB porcada=2.78KB hilos_del_SO=4 pila_en_uso=2.2MB GOMAXPROCS=2
$ ./hilos 1000   (mediana de 5)
hilos      n=1000 crear=15.3ms memoria=8.3MB porcada=8.48KB hilos_del_SO=1001

$ ./goroutines 10000   (mediana de 5)
goroutines n=10000 crear=9.1ms memoria=26.3MB porcada=2.69KB hilos_del_SO=4 pila_en_uso=19.8MB GOMAXPROCS=2
$ ./hilos 10000   (mediana de 5)
hilos      n=10000 crear=177.3ms memoria=81.1MB porcada=8.31KB hilos_del_SO=10001

$ ./goroutines 100000   (mediana de 5)
goroutines n=100000 crear=108.3ms memoria=261.5MB porcada=2.68KB hilos_del_SO=4 pila_en_uso=195.6MB GOMAXPROCS=2
$ ./hilos 100000   (mediana de 5)
hilos      n=31890 crear=506.3ms memoria=258.4MB porcada=8.30KB hilos_del_SO=31891 FALLO en el hilo 31891: Resource temporarily unavailable

$ ulimit -s 256; ./hilos 100000   # pilas de 256 KB en vez de 8 MB
hilos      n=31893 crear=464.1ms memoria=258.3MB porcada=8.29KB hilos_del_SO=31894 FALLO en el hilo 31894: Resource temporarily unavailable

$ ./goroutines 1000000
goroutines n=1000000 crear=1800.2ms memoria=2622.0MB porcada=2.68KB hilos_del_SO=4 pila_en_uso=1953.4MB GOMAXPROCS=2
```
