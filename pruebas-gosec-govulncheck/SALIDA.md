# SALIDA — gosec y govulncheck no son lo mismo (golandia 683)

Generado 2026-10-26 16:47 UTC con `python correr.py`. Docker 29.6.2. Código vulnerable A PROPÓSITO (app/): nunca se despliega.

## 1 · La misma app con dos tipos de fallo

### versiones

`$ go version; go version -m gosec; govulncheck -version`

```
go version go1.27.2 linux/amd64
	mod	github.com/securego/gosec/v2	v2.30.0	h1:67aF6LHa7SAL6QEc4es857GrXTT50gUqmrl0XPhtDx0=
Scanner: govulncheck@v1.9.0
DB: https://vuln.go.dev
DB updated: 2026-10-21 00:27:29 +0000 UTC
```
(código de salida 0)

### las dependencias

`$ cat go.mod`

```
module ejemplo/tienda

go 1.27

require (
	github.com/golang-jwt/jwt/v4 v4.5.0
	golang.org/x/net v0.20.0
)
```
(código de salida 0)

### lo que el código usa de cada una

`$ grep -n 'jwt\.\|html\.' main.go`

```
60:// Dependencia 1: jwt.Parse SÍ se llama (github.com/golang-jwt/jwt/v4 v4.5.0).
62:	t, err := jwt.Parse(r.Header.Get("Authorization"), func(t *jwt.Token) (any, error) {
74:	fmt.Fprintf(w, "<p>Hola, %s</p>", html.EscapeString(r.URL.Query().Get("nombre")))
```
(código de salida 0)

### gosec ./...

`$ gosec ./...`

```
main.go:23   G201  MEDIUM SQL string formatting
main.go:24   G701  HIGH   SQL injection via taint analysis: (*database/sql.DB).Query
main.go:37   G304  MEDIUM Potential file inclusion via variable
main.go:37   G703  HIGH   Path traversal via taint analysis: os.ReadFile
main.go:42   G104  LOW    Errors unhandled
main.go:42   G705  MEDIUM XSS via taint analysis: (net/http.ResponseWriter).Write
main.go:50   G404  HIGH   Use of weak random number generator (math/rand or math/rand/v2 instead
main.go:57   G402  HIGH   TLS InsecureSkipVerify set to true.
main.go:74   G705  MEDIUM XSS via taint analysis: fmt.Fprintf
main.go:80   G104  LOW    Errors unhandled
main.go:90   G114  MEDIUM Use of net/http serve function that has no support for setting timeout
Files: 1 · Lines: 91 · Nosec: 0 · Issues: 11
```
(código de salida 1)

> gosec mira solo los paquetes de tu módulo: ninguna línea menciona jwt ni x/net.

### gosec: el resumen

`$ gosec ./... | tail`

```
Summary:
  Gosec  : dev
  Files  : 1
  Lines  : 91
  Nosec  : 0
  Issues : 11
```
(código de salida 0)

### govulncheck ./...

`$ govulncheck ./...`

```
=== Symbol Results ===

Vulnerability #1: GO-2026-6809
    Distrust compromised intermediate certificates in crypto/x509
  More info: https://pkg.go.dev/vuln/GO-2026-6809
  Standard library
    Found in: crypto/x509@go1.27.2
    Fixed in: crypto/x509@go1.27.3
    Example traces found:
      #1: main.go:90:33: tienda.main calls http.ListenAndServe, which eventually calls x509.Certificate.Verify

Vulnerability #2: GO-2025-3553
    Excessive memory allocation during header parsing in
    github.com/golang-jwt/jwt
  More info: https://pkg.go.dev/vuln/GO-2025-3553
  Module: github.com/golang-jwt/jwt/v4
    Found in: github.com/golang-jwt/jwt/v4@v4.5.0
    Fixed in: github.com/golang-jwt/jwt/v4@v4.5.2
    Example traces found:
      #1: main.go:62:21: tienda.sesion calls jwt.Parse, which eventually calls jwt.Parser.ParseUnverified

Vulnerability #3: GO-2024-3250
    Improper error handling in ParseWithClaims and bad documentation may cause
    dangerous situations in github.com/golang-jwt/jwt
  More info: https://pkg.go.dev/vuln/GO-2024-3250
  Module: github.com/golang-jwt/jwt/v4
    Found in: github.com/golang-jwt/jwt/v4@v4.5.0
    Fixed in: github.com/golang-jwt/jwt/v4@v4.5.1
    Example traces found:
      #1: main.go:62:21: tienda.sesion calls jwt.Parse

Your code is affected by 3 vulnerabilities from 1 module and the Go standard library.
This scan also found 9 vulnerabilities in packages you import and 10
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
```
(código de salida 3)

### govulncheck por módulo (de -format json)

`$ govulncheck -format json ./... | clasificar`

```
github.com/golang-jwt/jwt/v4: llamadas 2 ['GO-2024-3250', 'GO-2025-3553'] · importadas sin llamar 0 · solo en go.mod 0
golang.org/x/net: llamadas 0 [] · importadas sin llamar 9 · solo en go.mod 10
stdlib: llamadas 1 ['GO-2026-6809'] · importadas sin llamar 0 · solo en go.mod 0
```

> «llamadas» = tu código llega a la función vulnerable (las que hacen fallar con código 3). «importadas sin llamar» = importas el paquete pero no la función. «solo en go.mod» = el módulo está, el paquete ni se importa.

### el arreglo: jwt v4.5.2 y golang:1.27.3 (x/net sigue en v0.20.0)

`$ go get github.com/golang-jwt/jwt/v4@v4.5.2 && govulncheck ./...`

```
go version go1.27.3 linux/amd64
	github.com/golang-jwt/jwt/v4 v4.5.2
	golang.org/x/net v0.20.0
=== Symbol Results ===

No vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 9 vulnerabilities in packages you import and 10
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
(código de salida: 0)
```

> x/net sigue vieja, pero tu código no llama a nada vulnerable: govulncheck sale con 0.

## 2 · Falsos positivos, #nosec y -exclude/-include

### #nosec · a-sin-anotacion

`$ gosec ./...`

```
13:	datos, err := os.ReadFile(origen) 
18:	os.WriteFile("/usr/share/nginx/html/estado.json", datos, 0o644) 

[gosec] 2026/10/26 16:47:54 Including rules: default
[gosec] 2026/10/26 16:47:54 Excluding rules: default
[gosec] 2026/10/26 16:47:54 Including analyzers: default
[gosec] 2026/10/26 16:47:54 Excluding analyzers: default
[gosec] 2026/10/26 16:47:54 Import directory: /ns/a-sin-anotacion
[gosec] 2026/10/26 16:47:54 Checking package: main
[gosec] 2026/10/26 16:47:54 Checking file: main.go
[main.go:18] - G703 (CWE-22): Path traversal via taint analysis: os.WriteFile (Confidence: HIGH, Severity: HIGH)
[main.go:13] - G703 (CWE-22): Path traversal via taint analysis: os.ReadFile (Confidence: HIGH, Severity: HIGH)
[main.go:13] - G304 (CWE-22): Potential file inclusion via variable (Confidence: HIGH, Severity: MEDIUM)
[main.go:18] - G306 (CWE-276): Expect WriteFile permissions to be 0600 or less (Confidence: HIGH, Severity: MEDIUM)
[main.go:18] - G104 (CWE-703): Errors unhandled (Confidence: HIGH, Severity: LOW)
  Nosec  : 0
  Issues : 5
```
(código de salida 0)

### #nosec · b-nosec-pelado

`$ gosec ./...`

```
13:	datos, err := os.ReadFile(origen) // #nosec
18:	os.WriteFile("/usr/share/nginx/html/estado.json", datos, 0o644) // #nosec

[gosec] 2026/10/26 16:47:55 Including rules: default
[gosec] 2026/10/26 16:47:55 Excluding rules: default
[gosec] 2026/10/26 16:47:55 Including analyzers: default
[gosec] 2026/10/26 16:47:55 Excluding analyzers: default
[gosec] 2026/10/26 16:47:55 Import directory: /ns/b-nosec-pelado
[gosec] 2026/10/26 16:47:55 Checking package: main
[gosec] 2026/10/26 16:47:55 Checking file: main.go
  Nosec  : 2
  Issues : 0
```
(código de salida 0)

### #nosec · c-solo-g304

`$ gosec ./...`

```
13:	datos, err := os.ReadFile(origen) // #nosec G304 -- la ruta la elige el operador
18:	os.WriteFile("/usr/share/nginx/html/estado.json", datos, 0o644) // #nosec G306 -- estado.json es público: lo sirve nginx

[gosec] 2026/10/26 16:47:55 Including rules: default
[gosec] 2026/10/26 16:47:55 Excluding rules: default
[gosec] 2026/10/26 16:47:55 Including analyzers: default
[gosec] 2026/10/26 16:47:55 Excluding analyzers: default
[gosec] 2026/10/26 16:47:55 Import directory: /ns/c-solo-g304
[gosec] 2026/10/26 16:47:55 Checking package: main
[gosec] 2026/10/26 16:47:55 Checking file: main.go
[main.go:18] - G703 (CWE-22): Path traversal via taint analysis: os.WriteFile (Confidence: HIGH, Severity: HIGH)
[main.go:13] - G703 (CWE-22): Path traversal via taint analysis: os.ReadFile (Confidence: HIGH, Severity: HIGH)
[main.go:18] - G104 (CWE-703): Errors unhandled (Confidence: HIGH, Severity: LOW)
  Nosec  : 2
  Issues : 3
```
(código de salida 0)

### #nosec · d-justificado

`$ gosec ./...`

```
13:	datos, err := os.ReadFile(origen) // #nosec G304 G703 -- la ruta la elige el operador
18:	os.WriteFile("/usr/share/nginx/html/estado.json", datos, 0o644) // #nosec G306 G703 -- estado.json es público y la ruta es fija

[gosec] 2026/10/26 16:47:55 Including rules: default
[gosec] 2026/10/26 16:47:55 Excluding rules: default
[gosec] 2026/10/26 16:47:55 Including analyzers: default
[gosec] 2026/10/26 16:47:55 Excluding analyzers: default
[gosec] 2026/10/26 16:47:55 Import directory: /ns/d-justificado
[gosec] 2026/10/26 16:47:55 Checking package: main
[gosec] 2026/10/26 16:47:55 Checking file: main.go
[main.go:18] - G104 (CWE-703): Errors unhandled (Confidence: HIGH, Severity: LOW)
  Nosec  : 2
  Issues : 1
```
(código de salida 0)

### -nosec: ignorar todas las anotaciones (auditoría)

`$ gosec -nosec ./...`

```
[gosec] 2026/10/26 16:47:55 Including rules: default
[gosec] 2026/10/26 16:47:55 Excluding rules: default
[gosec] 2026/10/26 16:47:55 Including analyzers: default
[gosec] 2026/10/26 16:47:55 Excluding analyzers: default
[gosec] 2026/10/26 16:47:55 Import directory: /ns/b-nosec-pelado
[gosec] 2026/10/26 16:47:55 Checking package: main
[gosec] 2026/10/26 16:47:55 Checking file: main.go
[main.go:18] - G703 (CWE-22): Path traversal via taint analysis: os.WriteFile (Confidence: HIGH, Severity: HIGH)
[main.go:13] - G703 (CWE-22): Path traversal via taint analysis: os.ReadFile (Confidence: HIGH, Severity: HIGH)
[main.go:13] - G304 (CWE-22): Potential file inclusion via variable (Confidence: HIGH, Severity: MEDIUM)
[main.go:18] - G306 (CWE-276): Expect WriteFile permissions to be 0600 or less (Confidence: HIGH, Severity: MEDIUM)
[main.go:18] - G104 (CWE-703): Errors unhandled (Confidence: HIGH, Severity: LOW)
  Nosec  : 0
  Issues : 5
```
(código de salida 0)

### -track-suppressions: el motivo queda en el informe

`$ gosec -track-suppressions -fmt json ./...`

```
G703 línea 18 · inSource: estado.json es público y la ruta es fija
G703 línea 13 · inSource: la ruta la elige el operador
G304 línea 13 · inSource: la ruta la elige el operador
G306 línea 18 · inSource: estado.json es público y la ruta es fija
G104 línea 18 · sin suprimir
```
(código de salida 1)

### filtros sobre la app

`$ gosec <filtro> ./...`

```
todo                                        11  G104 G114 G201 G304 G402 G404 G701 G703 G705
-exclude=G104                                9  G114 G201 G304 G402 G404 G701 G703 G705
-exclude=G104,G304                           8  G114 G201 G402 G404 G701 G703 G705
-include=G201,G304,G402,G404,G701,G703       6  G201 G304 G402 G404 G701 G703
-severity high                               4  G402 G404 G701 G703
-severity high -confidence high              3  G402 G701 G703
```

### gosec sobre github.com/spf13/cobra@v1.10.1

`$ gosec ./...`

```
bash_completions.go:702  G304  MEDIUM Potential file inclusion via variable
bash_completionsV2.go:471  G304  MEDIUM Potential file inclusion via variable
completions.go:958-959 G302  MEDIUM Expect file permissions to be 0600 or less
completions.go:958-959 G304  MEDIUM Potential file inclusion via variable
completions.go:958  G703  HIGH   Path traversal via taint analysis: os.OpenFile
man_docs.go:72   G304  MEDIUM Potential file inclusion via variable
md_docs.go:145  G304  MEDIUM Potential file inclusion via variable
rest_docs.go:157  G304  MEDIUM Potential file inclusion via variable
yaml_docs.go:72   G304  MEDIUM Potential file inclusion via variable
fish_completions.go:285  G304  MEDIUM Potential file inclusion via variable
powershell_completions.go:321  G304  MEDIUM Potential file inclusion via variable
zsh_completions.go:71   G304  MEDIUM Potential file inclusion via variable
Files: 18 · Lines: 6897 · Nosec: 1 · Issues: 12

por regla:
G304  10
G302  1
G703  1
```

### gosec -tests sobre github.com/spf13/cobra@v1.10.1 (también los _test.go)

`$ gosec -tests ./...`

```
Issues: 35
por regla:
G304  21
G104  8
G302  2
G703  2
G204  1
G301  1
```

## 3 · gosec dentro de golangci-lint

### versión y el gosec que trae dentro

`$ golangci-lint version; go version -m golangci-lint | grep gosec`

```
golangci-lint has version 2.15.0 built with go1.27.0 from 25c5369a on 2026-10-16T11:03:09Z
	dep	github.com/securego/gosec/v2	v2.29.0	h1:pF2HSLcnY5voqpxQumEe0O5YAAbvF4Syu4qeIewpLyc=
```
(código de salida 0)

### solo gosec, opciones por defecto

`$ golangci-lint run --no-config --default none --enable gosec ./...`

```
main.go:23:7: G201: SQL string formatting (gosec)
	q := fmt.Sprintf("SELECT nombre FROM productos WHERE id = '%s'", r.URL.Query().Get("id"))
	     ^
main.go:24:24: G701: SQL injection via taint analysis (gosec)
	filas, err := db.Query(q)
	                      ^
main.go:37:16: G304: Potential file inclusion via variable (gosec)
	datos, err := os.ReadFile(ruta)
	              ^
main.go:42:2: G104: Errors unhandled (gosec)
	w.Write(datos)
	^
main.go:50:17: G404: Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) (gosec)
		b[i] = letras[rand.Intn(len(letras))]
		              ^
main.go:57:51: G402: TLS InsecureSkipVerify set to true. (gosec)
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	                                                 ^
main.go:74:13: G705: XSS via taint analysis (gosec)
	fmt.Fprintf(w, "<p>Hola, %s</p>", html.EscapeString(r.URL.Query().Get("nombre")))
	           ^
main.go:80:3: G104: Errors unhandled (gosec)
		resp.Body.Close()
		^
main.go:90:14: G114: Use of net/http serve function that has no support for setting timeouts (gosec)
	fmt.Println(http.ListenAndServe(":8080", nil))
	            ^
9 issues:
* gosec: 9
```
(código de salida 1)

### lo mismo con --uniq-by-line=false

`$ golangci-lint run --no-config --default none --enable gosec --uniq-by-line=false ./...`

```
main.go:23:7: G201: SQL string formatting (gosec)
main.go:24:24: G701: SQL injection via taint analysis (gosec)
main.go:37:16: G304: Potential file inclusion via variable (gosec)
main.go:37:27: G703: Path traversal via taint analysis (gosec)
main.go:42:2: G104: Errors unhandled (gosec)
main.go:42:9: G705: XSS via taint analysis (gosec)
main.go:50:17: G404: Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) (gosec)
main.go:57:51: G402: TLS InsecureSkipVerify set to true. (gosec)
main.go:74:13: G705: XSS via taint analysis (gosec)
main.go:80:3: G104: Errors unhandled (gosec)
main.go:90:14: G114: Use of net/http serve function that has no support for setting timeouts (gosec)
11 issues:
* gosec: 11
```
(código de salida 0)

### .golangci.yml

`$ cat .golangci.yml`

```
version: "2"

linters:
  default: none
  enable:
    - gosec
  settings:
    gosec:
      excludes:
        - G104 # errores sin revisar: ya los ve errcheck

issues:
  max-same-issues: 0 # por defecto 3: esconde el resto de avisos con el mismo texto
  uniq-by-line: false # por defecto true: si una línea tiene dos reglas, muestra solo una
```
(código de salida 0)

### con .golangci.yml

`$ golangci-lint run ./...`

```
main.go:23:7: G201: SQL string formatting (gosec)
	q := fmt.Sprintf("SELECT nombre FROM productos WHERE id = '%s'", r.URL.Query().Get("id"))
	     ^
main.go:24:24: G701: SQL injection via taint analysis (gosec)
	filas, err := db.Query(q)
	                      ^
main.go:37:16: G304: Potential file inclusion via variable (gosec)
	datos, err := os.ReadFile(ruta)
	              ^
main.go:37:27: G703: Path traversal via taint analysis (gosec)
	datos, err := os.ReadFile(ruta)
	                         ^
main.go:42:9: G705: XSS via taint analysis (gosec)
	w.Write(datos)
	       ^
main.go:50:17: G404: Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) (gosec)
		b[i] = letras[rand.Intn(len(letras))]
		              ^
main.go:57:51: G402: TLS InsecureSkipVerify set to true. (gosec)
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	                                                 ^
main.go:74:13: G705: XSS via taint analysis (gosec)
	fmt.Fprintf(w, "<p>Hola, %s</p>", html.EscapeString(r.URL.Query().Get("nombre")))
	           ^
main.go:90:14: G114: Use of net/http serve function that has no support for setting timeouts (gosec)
	fmt.Println(http.ListenAndServe(":8080", nil))
	            ^
9 issues:
* gosec: 9
```
(código de salida 1)

### cobra en golangci-lint: max-same-issues y los _test.go

`$ golangci-lint run --no-config --default none --enable gosec [opciones] ./... | contar por regla`

```
--- 
      3 G304
      3 G104
      1 G302
      1 G301
      1 G204
--- --max-same-issues 0
     10 G304
      8 G104
      1 G302
      1 G301
      1 G204
--- --max-same-issues 0 --tests=false
      9 G304
      1 G302
```
(código de salida 0)

> Por defecto golangci-lint muestra como mucho 3 avisos con el mismo texto, y analiza también los _test.go (gosec suelto no, salvo con -tests).

## 4 · El SARIF de gosec

### gosec -no-fail -fmt sarif -out gosec.sarif ./...

`$ gosec -no-fail -fmt sarif -out gosec.sarif ./...`

```
version: 2.1.0 · tool: gosec dev
results: 11 · niveles: {'error': 9, 'warning': 2}
reglas: G104 G114 G201 G304 G402 G404 G701 G703 G705
código con -no-fail: 0
```

## 5 · El pipeline, ejecutado en local con act

### act --version

`$ act --version`

```
act version 0.2.89
```
(código de salida 0)

### el workflow

`$ cat .github/workflows/seguridad.yml`

```
# gosec revisa TU código; govulncheck, las dependencias (y la versión de Go) que tu código llama.
# No se sustituyen: un trabajo para cada uno. Versiones fijas: el pipeline cambia cuando tú lo decides.
name: seguridad-go
on:
  push:
  pull_request:
  schedule:
    - cron: "0 6 * * *" # la base de vulnerabilidades cambia aunque tu código no

permissions:
  contents: read

jobs:
  gosec:
    runs-on: ubuntu-latest
    container: golang:1.27.3
    permissions:
      contents: read
      security-events: write # para subir el SARIF a Code Scanning
    steps:
      - uses: actions/checkout@v4
      - name: gosec a SARIF (no rompe)
        run: go run github.com/securego/gosec/v2/cmd/gosec@v2.30.0 -quiet -no-fail -fmt sarif -out gosec.sarif ./...
      - name: subir a Code Scanning
        if: ${{ !env.ACT }} # act no tiene Code Scanning: este paso solo corre en GitHub
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif
      - name: gosec (rompe con severidad alta)
        run: go run github.com/securego/gosec/v2/cmd/gosec@v2.30.0 -quiet -severity high -confidence medium ./...

  govulncheck:
    runs-on: ubuntu-latest
    container: golang:1.27.3
    steps:
      - uses: actions/checkout@v4
      - name: govulncheck (rompe si tu código llama a algo vulnerable)
        run: go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
```
(código de salida 0)

### los trabajos

`$ act -l`

```
Stage  Job ID       Job name     Workflow name  Workflow file  Events                    
0      govulncheck  govulncheck  seguridad-go   seguridad.yml  push,pull_request,schedule
0      gosec        gosec        seguridad-go   seguridad.yml  pull_request,schedule,push
time="2026-10-26T16:48:05Z" level=info msg="Using docker host 'unix:///var/run/docker.sock', and daemon socket 'unix:///var/run/docker.sock'"
```
(código de salida 0)

### act push

`$ act push`

```
[seguridad-go/govulncheck]   ✅  Success - Main actions/checkout@v4
[seguridad-go/govulncheck]   ❌  Failure - Main govulncheck (rompe si tu código llama a algo vulnerable)
[seguridad-go/govulncheck] 🏁  Job failed
[seguridad-go/gosec      ]   ✅  Success - Main actions/checkout@v4
[seguridad-go/gosec      ]   ✅  Success - Main gosec a SARIF (no rompe)
[seguridad-go/gosec      ]   ❌  Failure - Main gosec (rompe con severidad alta)
[seguridad-go/gosec      ] 🏁  Job failed
(act tardó 377.5 s)
```

### lo que dijo cada trabajo (del registro de act)

`$ grep … act.log`

```
[seguridad-go/govulncheck]   | Vulnerability #1: GO-2025-3553
[seguridad-go/govulncheck]   | Vulnerability #2: GO-2024-3250
[seguridad-go/govulncheck]   | Your code is affected by 2 vulnerabilities from 1 module.
[seguridad-go/gosec      ]   | [/w/main.go:50] - G404 (CWE-338): Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) (Confidence: MEDIUM, Severity: HIGH)
[seguridad-go/gosec      ]   | [/w/main.go:57] - G402 (CWE-295): TLS InsecureSkipVerify set to true. (Confidence: HIGH, Severity: HIGH)
[seguridad-go/gosec      ]   | [/w/main.go:24] - G701 (CWE-89): SQL injection via taint analysis: (*database/sql.DB).Query (Confidence: HIGH, Severity: HIGH)
[seguridad-go/gosec      ]   | [/w/main.go:37] - G703 (CWE-22): Path traversal via taint analysis: os.ReadFile (Confidence: HIGH, Severity: HIGH)
[seguridad-go/gosec      ]   |   Files  : 1
[seguridad-go/gosec      ]   |   Issues : 4
```
