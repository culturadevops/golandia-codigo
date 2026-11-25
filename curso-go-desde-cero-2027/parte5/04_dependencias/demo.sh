source /work/repo/comun.sh
ver a_uuid.go; proyecto p5_dep a_uuid.go
cd /tmp/p/p5_dep
echo '$ go run .   # sin pedir la dependencia'
go run . 2>&1; echo "[codigo de salida: $?]"
echo '$ go get github.com/google/uuid@v1.6.0'
go get github.com/google/uuid@v1.6.0 2>&1 | sed 's#^go: ##'; echo "[codigo de salida: $?]"
echo '$ cat go.mod'; cat go.mod
echo '$ go mod tidy && cat go.mod'; go mod tidy 2>&1; cat go.mod
echo '$ wc -l go.sum'; wc -l < go.sum
correr p5_dep
