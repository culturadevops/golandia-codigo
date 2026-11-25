#!/bin/sh
set -e
echo "== go version =="
go version
cd /lab/etapas/01-dinero
go mod tidy >/tmp/t.log 2>&1 || (cat /tmp/t.log; exit 1)
echo
echo "== version de decimal =="
grep decimal go.mod | sed "s/^/  /"
echo
echo "== 1. el desvio acumulado, medido =="
go test -run TestDesvioAcumulado -v 2>&1 | head -14
echo
echo "== 2. NewFromFloat ya viene roto =="
go test -run TestNewFromFloatYaVieneRoto -v 2>&1 | head -10
echo
echo "== 3. la division no termina =="
go test -run TestDivisionNoTermina -v 2>&1 | head -10
echo
echo "== 4. el reparto que no cuadra =="
go test -run TestRepartoQueNoCuadra -v 2>&1 | head -10
echo
echo "== 5. EL PRECIO: float64 contra decimal contra centimos =="
go test -bench=. -benchmem -run=NADA 2>&1 | tail -12

echo
echo "== 6. ¿cuando duele NewFromFloat de verdad? =="
cd /lab/etapas/01-dinero
go test -run TestNewFromFloatCuandoSiDuele -v 2>&1 | head -14
