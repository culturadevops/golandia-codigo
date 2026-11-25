# Imagen de la demo: el job de Go encima de postgres:18, para tener pg_dump/pg_restore/psql 18 (los del servidor).
FROM golang:1.27.2 AS build
WORKDIR /src
COPY respaldo/ .
RUN CGO_ENABLED=0 go build -o /respaldo .

FROM postgres:18
COPY --from=build /respaldo /usr/local/bin/respaldo
