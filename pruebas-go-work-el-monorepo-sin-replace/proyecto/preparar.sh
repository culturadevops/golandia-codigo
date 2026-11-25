#!/usr/bin/env bash
# Deja el repo de ejemplo en su estado inicial: dos módulos, sin replace ni go.work.
set -e
cd /w
rm -rf repo && mkdir -p repo/saludo repo/app && cd repo
git init -q -b main . && git config user.email demo@example.com && git config user.name demo
cat > saludo/go.mod <<'X'
module ejemplo.com/saludo

go 1.27
X
cat > saludo/saludo.go <<'X'
package saludo

func Hola(n string) string { return "Hola, " + n }
X
cat > app/go.mod <<'X'
module ejemplo.com/app

go 1.27

require ejemplo.com/saludo v0.1.0
X
cat > app/main.go <<'X'
package main

import (
	"fmt"

	"ejemplo.com/saludo"
)

func main() { fmt.Println(saludo.Hola("Golandia")) }
X
git add -A && git commit -qm "dos módulos" && echo "repo listo"
