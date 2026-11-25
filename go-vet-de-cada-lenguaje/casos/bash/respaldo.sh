#!/bin/bash
# Respalda los .log de un directorio en otro. Uso: respaldo.sh ORIGEN DESTINO
origen=$1
destino=$2

cd $destino
cp $origen/*.log .
for f in $(ls *.log); do
  gzip $f
done
echo "respaldo listo en $(pwd)"
