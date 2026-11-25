#!/bin/bash
# Dentro de golang:1.27.2 (--cpus 2): las diferencias v1/v2 y el banco de velocidad.
# Banco: 5 pasadas de cada uno (-count 5); se imprime la mediana de ns/op por benchmark.
cd /trabajo/diferencias
echo '$ go version'; go version
echo; echo '$ go run .'; go run . 2>&1
cd /trabajo/velocidad
echo; echo '$ GOEXPERIMENT=nojsonv2 go test -bench . -benchmem -count 5   # encoding/json con su implementación vieja'
GOEXPERIMENT=nojsonv2 go test -bench . -benchmem -count 5 2>&1 | python3 /trabajo/mediana.py
echo; echo '$ go test -bench . -benchmem -count 5   # Go 1.27 por defecto: encoding/json ya va sobre v2'
go test -bench . -benchmem -count 5 2>&1 | python3 /trabajo/mediana.py
