#!/bin/sh
# El cálculo de la serie compilado a WebAssembly (wasip1) y ejecutado con wasmtime; y en nativo, para comparar.
cd /out
echo "### wasmtime --version"; wasmtime --version
echo "### ls -l *.wasm (bytes)"; ls -l *.wasm | awk '{printf "%10d  %s\n", $5, $9}'
for w in go tinygo tinygo-opt2; do
  echo "### wasmtime $w.wasm suma 300"; /usr/bin/time -f "memoria máx: %M KiB" wasmtime $w.wasm suma 300 2>&1
  echo "### wasmtime $w.wasm trabajo 300"; /usr/bin/time -f "memoria máx: %M KiB" wasmtime $w.wasm trabajo 300 2>&1
done
for b in go-nativo tinygo-nativo tinygo-nativo-opt2; do
  echo "### ./$b suma 300"; /usr/bin/time -f "memoria máx: %M KiB" ./$b suma 300 2>&1
  echo "### ./$b trabajo 300"; /usr/bin/time -f "memoria máx: %M KiB" ./$b trabajo 300 2>&1
done
