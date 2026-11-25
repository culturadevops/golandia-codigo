$ cat main.go
package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	total := 0
	precio, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("precio inválido:", err)
		os.Exit(1)
	}
	fmt.Println("con impuesto:", precio*118/100)
}

$ go build -o precios .
# demo/precios
.\main.go:10:2: declared and not used: total
(código de salida: 1)

$ go build -o precios . && ./precios 100
con impuesto: 118

$ GOOS=linux GOARCH=amd64 go build -o precios-linux .
$ ls -l precios-linux | awk "{print \$5, \$9}"
2408924 precios-linux
$ file precios-linux
precios-linux: ELF 64-bit LSB executable, x86-64
