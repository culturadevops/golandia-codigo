# -*- coding: utf-8 -*-
"""Build tags en Go — banco ejecutable (golandia, vídeo 368). Escribe SALIDA.md.

Un programa con un fichero por sistema (ruta_linux.go, ruta_windows.go, ruta_darwin.go), una
edición «pro» con etiqueta propia, una herramienta con //go:build ignore:
  1  qué ficheros entran en cada sistema (go list con GOOS);
  2  tres binarios desde la misma máquina, y la cabecera que delata cada uno (ELF, MZ, Mach-O);
  3  -tags pro;
  4  //go:build ignore;
  5  la trampa: un //go:build DESPUÉS del package se ignora y el error sale en otra parte;
  6  la sintaxis vieja // +build y lo que hace gofmt.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
from banco_go import Banco  # noqa: E402

with Banco(AQUI, "Build tags en Go") as b:
    b.titulo("La versión de Go")
    b.sh("go version")
    b.sh("ls *.go && head -4 edicion_pro.go", mostrar="ls *.go && head -4 edicion_pro.go")

    b.titulo("1 · Qué ficheros entran en cada sistema")
    b.sh("for s in linux windows darwin; do echo \"$s: $(GOOS=$s go list -f '{{.GoFiles}}' .)\"; done",
         mostrar="for s in linux windows darwin; do GOOS=$s go list -f '{{.GoFiles}}' .; done")

    b.titulo("2 · Tres binarios desde la misma máquina")
    b.sh("for s in linux windows darwin; do GOOS=$s GOARCH=amd64 go build -o app-$s . ; done; "
         "for s in linux windows darwin; do echo \"app-$s: $(head -c 4 app-$s | od -An -tx1)\"; done",
         mostrar="for s in linux windows darwin; do GOOS=$s GOARCH=amd64 go build -o app-$s .; done\n"
                 "$ head -c 4 app-* | od -An -tx1")
    b.sh("./app-linux", mostrar="./app-linux")

    b.titulo("3 · Una etiqueta propia: -tags pro")
    b.sh("go list -tags pro -f '{{.GoFiles}}' . && go build -tags pro -o app-pro . && ./app-pro",
         mostrar="go list -tags pro -f '{{.GoFiles}}' . && go build -tags pro -o app-pro . && ./app-pro")

    b.titulo("4 · //go:build ignore")
    b.sh("head -1 generar.go && go list -f '{{.IgnoredGoFiles}}' . && go run generar.go",
         mostrar="head -1 generar.go && go list -f '{{.IgnoredGoFiles}}' . && go run generar.go")

    b.titulo("5 · La trampa: el //go:build después del package")
    b.sh("cp edicion_pro.go /tmp/bk && printf 'package main\n\n//go:build pro\n\nconst edicion = \"pro\"\n' > edicion_pro.go && "
         "cat edicion_pro.go && go build -o /dev/null . 2>&1 | head -3; cp /tmp/bk edicion_pro.go",
         mostrar="cat edicion_pro.go   # el //go:build, debajo del package\n$ go build .")

    b.titulo("6 · La sintaxis vieja // +build")
    b.sh("printf '// +build linux\n\npackage main\n' > viejo.go && gofmt -d viejo.go; rm -f viejo.go",
         mostrar="cat viejo.go   # // +build linux\n$ gofmt -d viejo.go")

    b.titulo("7 · Cuántas plataformas conoce Go")
    b.sh("go tool dist list | wc -l && go tool dist list | grep -c '^linux/' && go tool dist list | grep -E '^(windows|darwin)/' | tr '\n' ' '",
         mostrar="go tool dist list | wc -l\n$ go tool dist list | grep -c '^linux/'\n$ go tool dist list | grep -E '^(windows|darwin)/'")

    b.titulo("8 · La arquitectura: amd64 contra arm64")
    b.sh("GOOS=linux GOARCH=amd64 go build -o app-amd64 . && GOOS=linux GOARCH=arm64 go build -o app-arm64 . && "
         "for a in amd64 arm64; do echo \"app-$a: magia $(head -c 4 app-$a | od -An -tx1) · máquina $(od -An -tx1 -j18 -N2 app-$a)\"; done; rm -f app-amd64 app-arm64",
         mostrar="GOOS=linux GOARCH=arm64 go build -o app-arm64 .\n$ od -An -tx1 -j18 -N2 app-amd64 app-arm64   # el campo «máquina» de ELF")

    b.titulo("9 · Lo que pesa cada binario")
    b.sh("for s in linux windows darwin; do GOOS=$s GOARCH=amd64 go build -o app-$s . ; done; ls -l app-linux app-windows app-darwin | awk '{print $5, $9}'",
         mostrar="ls -l app-*")

    b.titulo("10 · Tests con etiqueta: //go:build integration")
    b.sh("printf 'package main\n\nimport \"testing\"\n\nfunc TestRuta(t *testing.T) {\n\tif rutaConfig() == \"\" {\n\t\tt.Fatal(\"vacía\")\n\t}\n}\n' > ruta_test.go && "
         "printf '//go:build integration\n\npackage main\n\nimport \"testing\"\n\nfunc TestContraElDisco(t *testing.T) {\n\tt.Log(\"esto solo corre con -tags integration\")\n}\n' > disco_test.go && "
         "echo '--- go test -v .' && go test -v . 2>&1 | grep -E '^(=== RUN|--- |ok|PASS|FAIL)' ; "
         "echo '--- go test -v -tags integration .' && go test -v -tags integration . 2>&1 | grep -E '^(=== RUN|--- |ok|PASS|FAIL)|solo corre'; rm -f ruta_test.go disco_test.go",
         mostrar="head -1 disco_test.go   # //go:build integration\n$ go test -v .\n$ go test -v -tags integration .")

    b.titulo("11 · Dos etiquetas a la vez: -tags pro,debug")
    b.sh("printf '//go:build pro && debug\n\npackage main\n\nfunc init() { println(\"traza activada\") }\n' > traza.go && "
         "echo \"-tags pro:       $(go list -tags pro -f '{{.GoFiles}}' .)\" && echo \"-tags pro,debug: $(go list -tags pro,debug -f '{{.GoFiles}}' .)\" && "
         "go run -tags pro,debug . ; rm -f traza.go",
         mostrar="head -1 traza.go   # //go:build pro && debug\n$ go list -tags pro -f '{{.GoFiles}}' .\n$ go list -tags pro,debug -f '{{.GoFiles}}' .\n$ go run -tags pro,debug .")

    b.titulo("12 · cgo, unix y la versión de Go como condición")
    b.sh("printf '//go:build cgo\n\npackage main\n' > con_cgo.go && printf '//go:build unix\n\npackage main\n' > solo_unix.go && "
         "printf '//go:build go1.99\n\npackage main\n' > futuro.go && printf '//go:build go1.27\n\npackage main\n' > actual.go && "
         "echo \"CGO_ENABLED=1:  $(CGO_ENABLED=1 go list -f '{{.GoFiles}}' .)\" && echo \"CGO_ENABLED=0:  $(CGO_ENABLED=0 go list -f '{{.GoFiles}}' .)\" && "
         "for s in linux darwin windows; do echo \"$s: $(CGO_ENABLED=0 GOOS=$s go list -f '{{.GoFiles}}' . | grep -o solo_unix.go || echo '(sin solo_unix.go)')\"; done; "
         "echo \"go1.99: $(go list -f '{{.IgnoredGoFiles}}' . | grep -o futuro.go)  ← ignorado\"; echo \"go1.27: $(go list -f '{{.GoFiles}}' . | grep -o actual.go)  ← entra\"; "
         "rm -f con_cgo.go solo_unix.go futuro.go actual.go",
         mostrar="go list con //go:build cgo, //go:build unix, //go:build go1.99 y //go:build go1.27")
    b.sh("echo \"nativo:        CGO_ENABLED=$(go env CGO_ENABLED)\"; echo \"GOOS=windows:  CGO_ENABLED=$(GOOS=windows go env CGO_ENABLED)\"; echo \"GOARCH=arm64:  CGO_ENABLED=$(GOARCH=arm64 go env CGO_ENABLED)\"",
         mostrar="go env CGO_ENABLED; GOOS=windows go env CGO_ENABLED; GOARCH=arm64 go env CGO_ENABLED")

    b.titulo("13 · go vet ve la trampa")
    b.sh("cp edicion_pro.go /tmp/bk && printf 'package main\n\n//go:build pro\n\nconst edicion = \"pro\"\n' > edicion_pro.go && "
         "go vet -tags pro . 2>&1 | head -4; cp /tmp/bk edicion_pro.go",
         mostrar="go vet -tags pro .   # con el //go:build debajo del package")

    b.titulo("14 · La librería estándar usa lo mismo")
    b.sh("cd $(go env GOROOT)/src/os && echo \"ficheros en os: $(ls *.go | grep -vc _test.go)\" && echo \"_windows.go: $(ls *_windows.go | grep -vc _test)\" && echo \"_unix.go: $(ls *_unix.go | grep -vc _test)\" && "
         "echo \"linux:   $(GOOS=linux go list -f '{{len .GoFiles}}' os) ficheros\" && echo \"windows: $(GOOS=windows go list -f '{{len .GoFiles}}' os) ficheros\"",
         mostrar="cd $(go env GOROOT)/src/os && ls *_windows.go *_unix.go\n$ GOOS=linux go list -f '{{len .GoFiles}}' os\n$ GOOS=windows go list -f '{{len .GoFiles}}' os")

    b.titulo("15 · Si te olvidas de un sistema")
    b.sh("mv ruta_darwin.go /tmp/ && GOOS=darwin go build -o /dev/null . 2>&1 | head -3; mv /tmp/ruta_darwin.go .",
         mostrar="mv ruta_darwin.go /tmp/   # me olvido de Mac\n$ GOOS=darwin go build .")
    b.sh("mv ruta_darwin.go /tmp/ && GOOS=linux go build -o /dev/null . && echo 'linux: compila sin quejarse'; mv /tmp/ruta_darwin.go .",
         mostrar="GOOS=linux go build .   # en tu máquina no se nota")

    b.titulo("16 · GOFLAGS: dejar la etiqueta puesta")
    b.sh("GOFLAGS=-tags=pro go run .", mostrar="GOFLAGS=-tags=pro go run .")

    b.titulo("17 · La alternativa: -ldflags -X cambia un valor, no el código")
    b.sh("mkdir -p /tmp/v && cd /tmp/v && printf 'module v\n\ngo 1.27\n' > go.mod && "
         "printf 'package main\n\nimport \"fmt\"\n\nvar version = \"dev\"\n\nfunc main() { fmt.Println(\"versión:\", version) }\n' > main.go && "
         "go run . && go run -ldflags '-X main.version=1.4.2' .",
         mostrar="cat main.go   # var version = \"dev\"\n$ go run .\n$ go run -ldflags '-X main.version=1.4.2' .")

    b.titulo("18 · runtime.GOOS contra las etiquetas")
    b.sh("mkdir -p /tmp/r && cd /tmp/r && printf 'module r\n\ngo 1.27\n' > go.mod && "
         "printf 'package main\n\nimport (\n\t\"fmt\"\n\t\"runtime\"\n)\n\nfunc main() {\n\tif runtime.GOOS == \"windows\" {\n\t\tfmt.Println(\"en Windows\")\n\t} else {\n\t\tfmt.Println(\"en\", runtime.GOOS)\n\t}\n}\n' > main.go && "
         "go run . && GOOS=windows go build -o /dev/null . && echo 'windows: compila'",
         mostrar="cat main.go   # if runtime.GOOS == \"windows\" { ... }\n$ go run .\n$ GOOS=windows go build .")
    b.sh("mkdir -p /tmp/s && cd /tmp/s && printf 'module s\n\ngo 1.27\n' > go.mod && "
         "printf 'package main\n\nimport (\n\t\"fmt\"\n\t\"runtime\"\n\t\"syscall\"\n)\n\nfunc main() {\n\tif runtime.GOOS == \"windows\" {\n\t\tdll, _ := syscall.LoadDLL(\"kernel32.dll\")\n\t\tfmt.Println(dll)\n\t}\n}\n' > main.go && "
         "GOOS=windows go build -o /dev/null . && echo 'windows: compila'; GOOS=linux go build -o /dev/null . 2>&1 | head -3",
         mostrar="cat main.go   # syscall.LoadDLL dentro de un if runtime.GOOS\n$ GOOS=windows go build .\n$ GOOS=linux go build .")
    b.sh("rm -f app-*", guardar=False)
