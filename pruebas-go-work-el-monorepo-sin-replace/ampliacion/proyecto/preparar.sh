#!/usr/bin/env bash
# Dos módulos del mismo repo que usan golang.org/x/text en versiones DISTINTAS:
# app pide v0.14.0 y saludo v0.21.0. No dependen entre sí.
set -e
cd /w
rm -rf repo && mkdir -p repo/saludo repo/app && cd repo
git init -q -b main . && git config user.email demo@example.com && git config user.name demo
cat > saludo/saludo.go <<'X'
package saludo

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Hola(n string) string { return "Hola, " + cases.Title(language.Spanish).String(n) }
X
cat > app/main.go <<'X'
package main

import (
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() { fmt.Println(cases.Upper(language.Spanish).String("listo")) }
X
(cd saludo && printf 'module ejemplo.com/saludo\n\ngo 1.27\n' > go.mod && go get golang.org/x/text@v0.21.0 >/dev/null 2>&1 && go mod tidy)
(cd app && printf 'module ejemplo.com/app\n\ngo 1.27\n' > go.mod && go get golang.org/x/text@v0.14.0 >/dev/null 2>&1 && go mod tidy)
git add -A && git commit -qm "dos módulos, x/text en dos versiones" && echo "repo listo"
