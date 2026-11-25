$ grep -E 'x/text|width\.|language\.' seguro/main.go seguro/go.mod
"golang.org/x/text/width"
fmt.Println(width.Narrow.String("ＧＯ"))
golang.org/x/text v0.3.7
$ cd seguro && go run .
GO

$ grep -E 'x/text|width\.|language\.' afectado/main.go afectado/go.mod
"golang.org/x/text/language"
tags, _, err := language.ParseAcceptLanguage("es-MX,es;q=0.9,en;q=0.8")
golang.org/x/text v0.3.7
$ cd afectado && go run .
[es-MX es en] <nil>

$ cd seguro && govulncheck ./...
No vulnerabilities found.
Your code is affected by 0 vulnerabilities.
(código de salida: 0)
This scan also found 1 vulnerability in packages you import and 23 vulnerabilities in modules you require, but your code doesn't appear to call

$ cd afectado && govulncheck ./...
Vulnerability #1: GO-2022-1059
    Found in: golang.org/x/text@v0.3.7
    Fixed in: golang.org/x/text@v0.3.8
      #1: main.go:10:46: afectado.main calls language.ParseAcceptLanguage
Your code is affected by 1 vulnerability from 1 module.
(código de salida: 3)
    Denial of service via crafted Accept-Language header in
This scan also found 1 vulnerability in packages you import and 22 vulnerabilities in modules you require, but your code doesn't appear to call

$ cd norm && govulncheck ./...
Vulnerability #1: GO-2026-5970
    Found in: golang.org/x/text@v0.3.7
    Fixed in: golang.org/x/text@v0.39.0
      #1: main.go:10:29: norm.main calls norm.Form.String
Your code is affected by 1 vulnerability from 1 module.
(código de salida: 3)
    Infinite loop on invalid input in golang.org/x/text
This scan also found 1 vulnerability in packages you import and 22 vulnerabilities in modules you require, but your code doesn't appear to call

$ trivy fs --scanners vuln seguro
  vulnerabilidades: 3
   CVE-2022-32149 golang.org/x/text v0.3.7 -> 0.3.8 HIGH
   CVE-2026-56852 golang.org/x/text v0.3.7 -> 0.39.0 HIGH
   CVE-2026-56851 golang.org/x/text v0.3.7 -> 0.41.0 UNKNOWN

$ trivy fs --scanners vuln afectado
  vulnerabilidades: 3
   CVE-2022-32149 golang.org/x/text v0.3.7 -> 0.3.8 HIGH
   CVE-2026-56852 golang.org/x/text v0.3.7 -> 0.39.0 HIGH
   CVE-2026-56851 golang.org/x/text v0.3.7 -> 0.41.0 UNKNOWN

$ # el arreglo: subir x/text a la última versión
$ cd afectado && go get golang.org/x/text@latest && grep x/text go.mod
golang.org/x/text v0.43.0
$ cd afectado && govulncheck ./... | grep "Your code"
No vulnerabilities found.
Your code is affected by 0 vulnerabilities.
(código de salida: 0)
$ trivy fs --scanners vuln afectado
  vulnerabilidades: 0

$ docker run --rm aquasec/trivy:0.67.2 --version | head -1
Version: 0.67.2
$ govulncheck -version | grep -E "Scanner|DB updated"
Scanner: govulncheck@v1.9.0
DB updated: 2026-10-08 22:31:09 +0000 UTC
