cd /tmp && rm -rf hola && mkdir hola && cd hola
echo '$ go version'; go version
echo '$ go env GOOS GOARCH GOPATH'; go env GOOS GOARCH GOPATH
echo '$ go mod init hola'; go mod init hola 2>&1
echo '$ cat go.mod'; cat go.mod
cat > main.go <<'GO'
package main

import "fmt"

func main() {
	fmt.Println("hola, Golandia")
}
GO
echo '$ cat main.go'; cat main.go
echo '$ go run .'; go run . 2>&1
echo '$ go build -o hola .'; go build -buildvcs=false -o hola . 2>&1
echo '$ ls -l hola'; ls -l hola | awk '{print $5, $9}'
echo '$ ./hola'; ./hola
echo '$ go build -ldflags="-s -w" -o hola-chico .'; go build -buildvcs=false -ldflags="-s -w" -o hola-chico . 2>&1
echo '$ ls -l hola-chico'; ls -l hola-chico | awk '{print $5, $9}'
echo '$ GOOS=windows go build -o hola.exe .'; GOOS=windows go build -buildvcs=false -o hola.exe . 2>&1
echo '$ file hola hola.exe'; (command -v file >/dev/null && file hola hola.exe) || (head -c 4 hola | od -c | head -1; head -c 2 hola.exe | od -c | head -1)
