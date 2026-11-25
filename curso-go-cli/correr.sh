#!/usr/bin/env bash
# Corre DENTRO del contenedor (lo lanza ejecutar_todo.py). Copia el repo a /work para no
# tocar el original, y ejecuta el demo.sh de cada lección entre dos marcas que el script
# de Python usa para partir la salida.
#
#   bash correr.sh              todas las lecciones
#   bash correr.sh 05 08        solo esas
set -u
rm -rf /work/repo
cp -r /src /work/repo
cd /work/repo || exit 1
rm -rf .git

echo "@@VERSIONES"
go version
echo "jq: $(jq --version)"
goreleaser --version 2>&1 | grep -E "^GitVersion"
echo "cobra: $(go list -m -f '{{.Version}}' github.com/spf13/cobra)"
echo "go-internal: $(go list -m -f '{{.Version}}' github.com/rogpeppe/go-internal)"
uname -srm

lecciones=("$@")
if [ ${#lecciones[@]} -eq 0 ]; then
	lecciones=($(ls -d [0-9][0-9]_* | cut -c1-2))
fi
for n in "${lecciones[@]}"; do
	dir=$(ls -d "${n}"_*)
	echo "@@INICIO $n $dir"
	bash "$dir/demo.sh" 2>&1
	echo "@@FIN $n $?"
done
