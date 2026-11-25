#!/bin/bash
# Dentro de golang:1.27.2 (--cpus 2): goroutines esperando con time.Sleep, esperando datos de la red, o bloqueadas en una llamada al sistema.
cd /trabajo/bloqueo && go build -o /tmp/bloqueo .
for n in 1000 5000; do
  echo "\$ ./bloqueo sleep $n";   /tmp/bloqueo sleep $n
  echo "\$ ./bloqueo red $n";     /tmp/bloqueo red $n
  echo "\$ ./bloqueo syscall $n"; /tmp/bloqueo syscall $n
done
echo "\$ ./bloqueo syscall 12000"
/tmp/bloqueo syscall 12000 2>&1 | head -3
echo "(código de salida: ${PIPESTATUS[0]})"
