#!/bin/sh
set -e
echo "== go version =="
go version
cd /lab/etapas/01-fuga
go mod tidy >/tmp/t.log 2>&1 || (cat /tmp/t.log; exit 1)
echo
echo "== version de goleak =="
grep goleak go.mod | sed 's/^/  /'
echo
echo "== 1. el test con la fuga, SIN goleak: pasa =="
go test -run TestConFuga_SinGoleak -v 2>&1 | tail -6
echo
echo "== 2. el MISMO test con goleak: falla, y dice donde =="
go test -run TestConFuga_ConGoleak -v 2>&1 | head -24
echo
echo "== 3. la version sana con goleak: pasa =="
go test -run TestSano_ConGoleak -v 2>&1 | tail -6

echo
echo "== 4. CODIGO REAL DEL CANAL: el chequeo de salud del video de goroutines =="
cd /lab/etapas/02-codigo-real
go mod tidy >/tmp/t2.log 2>&1 || (cat /tmp/t2.log; exit 1)
echo "   --- la version concurrente, que el video dio por buena ---"
go test -run TestConcurrente -v 2>&1 | head -20
echo "   --- la del limite de tiempo, donde estaba la sospecha ---"
go test -run TestConTimeout -v 2>&1 | head -26

echo
echo "== 5. el MISMO test, con goleak mirando ANTES del Close =="
cd /lab/etapas/02-codigo-real
go test -run TestConTimeout_GoleakPrimero -v 2>&1 | head -24
