#!/bin/sh
set -e
echo "== go version =="
go version
cd /lab/etapas/01-validar
go mod tidy >/tmp/t.log 2>&1 || (cat /tmp/t.log; exit 1)
echo
echo "== version =="
grep validator go.mod | sed "s/^/  /"
echo
echo "== 1. que atrapa =="
go test -run TestLoQueAtrapa -v 2>&1 | head -16
echo
echo "== 2. el error CRUDO, tal cual sale =="
go test -run TestElErrorCrudoEsIlegible -v 2>&1 | head -12
echo
echo "== 3. lo mismo, traducido a mano =="
go test -run TestMensajesDecentes -v 2>&1 | head -12
echo
echo "== 4. la trampa: required contra el valor cero =="
go test -run TestRequiredContraElValorCero -v 2>&1 | head -12
echo
echo "== 5. una etiqueta MAL ESCRITA =="
go test -run TestEtiquetaMalEscrita -v 2>&1 | head -10
echo
echo "== 6. un validador propio (RUC) =="
go test -run TestValidadorPropio -v 2>&1 | head -12
echo
echo "== 7. cuanto cuesta validar =="
go test -bench=. -benchmem -run=NADA 2>&1 | tail -8
