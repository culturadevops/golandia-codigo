# SALIDA — io.Reader en Go (golandia)

> Corrida del 2026-10-20 16:55. golang:1.27.2, contenedor con --cpus 2 --memory 4g. Archivo de prueba: 500 MB de texto generado.

```
$ go version
go version go1.27.2 linux/amd64

$ ./lector propio
HOLA, GOLANDIA

$ ls -l datos.txt
524287995 /tmp/datos.txt

$ ./lector todo datos.txt     # os.ReadFile + sha256.Sum256
sha256=f8addf37446c2b3d7b062a7ee751c1ca6bbcae6a1be42f8fe657ee6ba3131cdf bytes=524287995 pico_de_memoria=503.4 MB

$ ./lector stream datos.txt   # io.Copy(sha256.New(), archivo)
sha256=f8addf37446c2b3d7b062a7ee751c1ca6bbcae6a1be42f8fe657ee6ba3131cdf bytes=524287995 pico_de_memoria=2.6 MB

$ sha256sum datos.txt
f8addf37446c2b3d7b062a7ee751c1ca6bbcae6a1be42f8fe657ee6ba3131cdf

$ ./lector cadena datos.txt   # archivo -> gzip -> contador -> hash + archivo.gz
comprimido=1538181 bytes sha256_del_gz=f816314483caf52c pico_de_memoria=3.4 MB

$ ./lector lineas datos.txt   # bufio.Scanner, línea a línea
lineas=9198035 err=<nil> pico_de_memoria=2.5 MB

$ wc -l datos.txt
9198035

$ ./lector linea-larga   # una línea de 100 KB
sin Buffer:   líneas leídas=1 err=bufio.Scanner: token too long
con Buffer 1MB: líneas leídas=3 err=<nil>

$ ./lector limite
LimitReader(1024) sobre 5000 bytes: leí 1024 err=<nil>

$ go doc -all bufio | grep "MaxScanTokenSize = "
	MaxScanTokenSize = 64 * 1024

$ ./lector trampa
Read:     pedí 8, llegaron 4: "0123" err=<nil>
ReadFull: pedí 8, llegaron 8: "01234567" err=<nil>
último Read: n=3 "fin" err=EOF   <- datos Y EOF a la vez

$ ./lector bucle   # mismo Reader, dos bucles
err primero: "hola, goland"
n primero:   "hola, golandia"

$ # ¿cuántos tipos de la librería estándar tienen un método Read([]byte)? (sin tests, vendor ni testdata)
$ grep -rE '^func \([^)]*\) Read\(\w+ \[\]byte\)' $(go env GOROOT)/src --include=*.go --exclude=*_test.go | grep -v -e /vendor/ -e /testdata/ | wc -l
107
$ # los 10 paquetes con más
     18 net/http
      7 testing/iotest
      7 net
      5 net/http/internal/http2
      5 io
      4 archive/tar
      3 mime/multipart
      2 testing/fstest
      2 net/http/httputil
      2 internal/poll
```
