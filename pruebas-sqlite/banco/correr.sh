#!/bin/sh
# Banco SQLite vs Postgres. Se ejecuta DENTRO del contenedor de Go, en la misma red que el
# Postgres. El fichero de SQLite vive en el disco del contenedor (/datos), no en un volumen
# del host: montar el disco de Windows falsearia la medida.
#
# ⚖️ EQUIDAD: comparar SQLite en synchronous=NORMAL contra Postgres en synchronous_commit=on
# es comparar dos garantias distintas. Por eso el cara a cara va con LAS DOS parejas.
set -u
mkdir -p /datos
PGON="postgres://postgres:secreto@bancopg:5432/postgres?sslmode=disable"
PGOFF="$PGON&options=-c%20synchronous_commit%3Doff"
S=${SEGUNDOS:-4}

sq() { echo "file:/datos/p.db?_pragma=journal_mode($1)&_pragma=busy_timeout($2)&_pragma=synchronous($3)"; }

echo "############ ENTORNO ############"
go version
echo "postgres 17.10 (contenedor postgres:17-alpine)"
echo "modernc.org/sqlite (Go puro, sin cgo) · jackc/pgx · cada corrida ${S}s"

echo
echo "############ A · LA TRAMPA: no es un limite, es una opcion sin poner ############"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 8 -dsn "$(sq DELETE 0 FULL)" \
  -etiqueta "8 escritores · diario clasico · SIN busy_timeout"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 8 -dsn "$(sq WAL 0 FULL)" \
  -etiqueta "8 escritores · WAL · SIN busy_timeout"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 8 -dsn "$(sq WAL 5000 FULL)" \
  -etiqueta "8 escritores · WAL · CON busy_timeout 5s"

echo
echo "############ B · EL PRAGMA QUE NADIE MENCIONA (synchronous) ############"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 1 -dsn "$(sq WAL 5000 FULL)" \
  -etiqueta "8 escritores · WAL · pool 1 · synchronous FULL"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 1 -dsn "$(sq WAL 5000 NORMAL)" \
  -etiqueta "8 escritores · WAL · pool 1 · synchronous NORMAL"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 1 -dsn "$(sq WAL 5000 OFF)" \
  -etiqueta "8 escritores · WAL · pool 1 · synchronous OFF"

echo
echo "############ C · CARA A CARA, CON LA MISMA GARANTIA A CADA LADO ############"
echo "-- durabilidad estricta (cada commit al disco) --"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 1 -dsn "$(sq WAL 5000 FULL)" \
  -etiqueta "sqlite  WAL · synchronous FULL         · 8 escritores"
./banco -motor postgres -segundos $S -escritores 8 -maxabiertas 8 -dsn "$PGON" \
  -etiqueta "postgres synchronous_commit = on       · 8 escritores"
echo "-- durabilidad relajada (se puede perder el ultimo commit si cae la maquina) --"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 1 -dsn "$(sq WAL 5000 NORMAL)" \
  -etiqueta "sqlite  WAL · synchronous NORMAL       · 8 escritores"
./banco -motor postgres -segundos $S -escritores 8 -maxabiertas 8 -dsn "$PGOFF" \
  -etiqueta "postgres synchronous_commit = off      · 8 escritores"

echo
echo "############ D · LEER MIENTRAS ALGUIEN ESCRIBE (4 escriben + 4 leen) ############"
./banco -motor sqlite -segundos $S -escritores 4 -lectores 4 -maxabiertas 8 -dsn "$(sq DELETE 5000 FULL)" \
  -etiqueta "sqlite  diario clasico"
./banco -motor sqlite -segundos $S -escritores 4 -lectores 4 -maxabiertas 8 -dsn "$(sq WAL 5000 FULL)" \
  -etiqueta "sqlite  WAL"
./banco -motor postgres -segundos $S -escritores 4 -lectores 4 -maxabiertas 8 -dsn "$PGON" \
  -etiqueta "postgres"

echo
echo "############ E · UN SOLO ESCRITOR (el caso de casi todo el mundo) ############"
./banco -motor sqlite -segundos $S -escritores 1 -maxabiertas 1 -dsn "$(sq WAL 5000 NORMAL)" \
  -etiqueta "sqlite  WAL · NORMAL · 1 escritor"
./banco -motor postgres -segundos $S -escritores 1 -maxabiertas 1 -dsn "$PGOFF" \
  -etiqueta "postgres synchronous_commit off · 1 escritor"

echo
echo "############ FIN ############"
