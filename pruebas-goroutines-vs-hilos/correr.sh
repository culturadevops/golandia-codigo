#!/bin/bash
# Dentro del contenedor (golang:1.27.2, --cpus 2): compila los dos programas y los corre con 1.000, 10.000 y 100.000.
# Cada medida, 5 veces; se imprime la corrida mediana por tiempo de creación.
set -e
cd /trabajo/goroutines && go build -o /tmp/gor . 
gcc -O2 -pthread -o /tmp/hil /trabajo/hilos/hilos.c
echo "\$ go version; gcc --version | head -1; nproc; ulimit -s"
go version; gcc --version | head -1; nproc; ulimit -s
mediana() { sort -t= -k3 -n | sed -n 3p; }
for n in 1000 10000 100000; do
  echo
  echo "\$ ./goroutines $n   (mediana de 5)"
  for i in 1 2 3 4 5; do /tmp/gor $n; done | sort -t'=' -k3,3 -g | sed -n 3p
  echo "\$ ./hilos $n   (mediana de 5)"
  for i in 1 2 3 4 5; do /tmp/hil $n; done | sort -t'=' -k3,3 -g | sed -n 3p
done
echo
echo "\$ ulimit -s 256; ./hilos 100000   # pilas de 256 KB en vez de 8 MB"
(ulimit -s 256; /tmp/hil 100000)
echo
echo "\$ ./goroutines 1000000"
/tmp/gor 1000000
