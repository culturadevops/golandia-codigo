source /work/repo/comun.sh
ver precio.go; ver precio_test.go
proyecto p5_tests precio.go precio_test.go
cd /tmp/p/p5_tests
echo '$ go test'
go test 2>&1 | sed -E 's/[0-9]+\.[0-9]+s/X.XXXs/'; echo "[codigo de salida: ${PIPESTATUS[0]}]"
echo '$ go test -v -run "TestDescuento/negativo"'
go test -v -run "TestDescuento/negativo" 2>&1 | sed -E 's/\([0-9.]+s\)/(X.XXs)/; s/[0-9]+\.[0-9]+s/X.XXXs/'; echo "[codigo de salida: ${PIPESTATUS[0]}]"
cp "$AQUI_LECCION/precio_arreglado.go" main.go
echo '$ # arreglo: if porcentaje < 0 { porcentaje = 0 }'
echo '$ go test -cover'
go test -cover 2>&1 | sed -E 's/[0-9]+\.[0-9]+s/X.XXXs/'; echo "[codigo de salida: ${PIPESTATUS[0]}]"
