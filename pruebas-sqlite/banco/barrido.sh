#!/bin/sh
# G · LA BARRIDA: donde se cruzan las dos curvas.
# Misma garantia a los dos lados (durabilidad relajada), misma carga, solo cambia
# cuantos escritores compiten. Es la pregunta del video: ¿a partir de cuantos?
set -u
mkdir -p /datos
PGOFF="postgres://postgres:secreto@bancopg:5432/postgres?sslmode=disable&options=-c%20synchronous_commit%3Doff"
SQ="file:/datos/b.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
S=${SEGUNDOS:-4}

echo "############ G · BARRIDA DE ESCRITORES (misma garantia a los dos lados) ############"
for n in 1 2 4 8 16 32; do
  ./banco -motor sqlite   -segundos $S -escritores $n -maxabiertas 1 -dsn "$SQ" \
    -etiqueta "sqlite   $(printf %2d $n) escritores"
  ./banco -motor postgres -segundos $S -escritores $n -maxabiertas $n -dsn "$PGOFF" \
    -etiqueta "postgres $(printf %2d $n) escritores"
done
echo "############ FIN ############"
