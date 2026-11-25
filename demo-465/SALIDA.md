$ cat go.sum
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=

$ go mod verify
all modules verified

$ echo "// cambio" >> $GOMODCACHE/github.com/google/uuid@v1.6.0/uuid.go && go mod verify
github.com/google/uuid v1.6.0: dir has been modified (/tmp/tmp.OuNVj0BUYt/cache/github.com/google/uuid@v1.6.0)
(código de salida: 1)

$ grep "uuid v1.6.0 h1" go.sum   # alguien cambió el hash
github.com/google/uuid v1.6.0 h1:AAAAJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
$ go build .
go: downloading github.com/google/uuid v1.6.0
verifying github.com/google/uuid@v1.6.0: checksum mismatch
	downloaded: h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
	go.sum:     h1:AAAAJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=

SECURITY ERROR
This download does NOT match an earlier download recorded in go.sum.
The bits may have been replaced on the origin server, or an attacker may
have intercepted the download attempt.

For more information, see 'go help module-auth'.
(código de salida: 1)

$ curl -s https://vuln.go.dev/index/modules.json | grep -o boltdb...
  github.com/boltdb-go/bolt ['GO-2025-3451']
$ curl -s https://vuln.go.dev/ID/GO-2025-3451.json
  id: GO-2025-3451
  resumen: Vulnerability in github.com/boltdb-go/bolt
  publicado: 2025-02-05
  módulo: github.com/boltdb-go/bolt · afectadas: desde 0 (todas)
  detalle: This module is a malicious typosquat, attempting to take advantage of confusion with the github.com/boltdb/bolt module.
