#!/bin/sh
# Demos de «¿Go o C?» (golandia 803). Se ejecuta dentro de gcc:16.2.0 con /d montado (solo lectura).
cd /tmp
for f in desborde liberado local; do
  echo "### gcc -O2 -Wall -Wextra $f.c"
  gcc -O2 -Wall -Wextra -o $f /d/demos/$f.c 2>&1
  echo "[código de salida de gcc: $?]"
done
echo "### ./desborde invitado"; ./desborde invitado; echo "[código de salida: $?]"
echo "### ./desborde invitado-xyz"; ./desborde invitado-xyz; echo "[código de salida: $?]"
echo "### ./liberado"; ./liberado; echo "[código de salida: $?]"
echo "### ./local"; ./local; echo "[código de salida: $?]"
for f in desborde liberado; do gcc -O1 -g -fsanitize=address -o $f-asan /d/demos/$f.c 2>/dev/null; done
echo "### gcc -fsanitize=address … && ./desborde-asan invitado-xyz"
./desborde-asan invitado-xyz > asan1.txt 2>&1; c=$?; head -3 asan1.txt; grep -m1 'in main' asan1.txt; echo "[código de salida: $c]"
echo "### gcc -fsanitize=address … && ./liberado-asan"
./liberado-asan > asan2.txt 2>&1; c=$?; head -3 asan2.txt; echo "[código de salida: $c]"
