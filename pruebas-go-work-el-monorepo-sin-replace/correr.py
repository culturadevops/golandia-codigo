# -*- coding: utf-8 -*-
"""go work — banco ejecutable (golandia, vídeo 367). Escribe SALIDA.md.

Un repo con dos módulos (app importa saludo), sin proxy (GOPROXY=off: nada se baja de internet):
  1  sin nada: app no encuentra saludo;
  2  con `replace` en go.mod: funciona... y el go.mod queda cambiado (el commit que se cuela);
  3  la CI que solo tiene app: el replace apunta a un directorio que no existe;
  4  con `go work`: funciona sin tocar go.mod, y un cambio en saludo se ve al momento;
  5  GOWORK=off vuelve al modo de un solo módulo; y git status: go.work es lo único nuevo.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
from banco_go import Banco  # noqa: E402

E = {"GOPROXY": "off"}
with Banco(AQUI, "go work en Go") as b:
    b.titulo("La versión de Go")
    b.sh("go version")
    b.sh("bash /w/preparar.sh", mostrar="bash preparar.sh   # dos módulos: app importa saludo", env=E)
    b.sh("cd repo && find . -name '*.go' -o -name go.mod | grep -v .git | sort && cat app/go.mod",
         mostrar="find . -name '*.go' -o -name go.mod && cat app/go.mod")

    b.titulo("1 · Sin nada")
    b.sh("cd repo/app && go build ./... 2>&1 | head -2", mostrar="cd app && go build ./...", env=E)

    b.titulo("2 · Con replace")
    b.sh("cd repo/app && go mod edit -replace ejemplo.com/saludo=../saludo && go run .",
         mostrar="go mod edit -replace ejemplo.com/saludo=../saludo && go run .", env=E)
    b.sh("cd repo && git diff app/go.mod | tail -3", mostrar="git diff app/go.mod")

    b.titulo("3 · La CI, que solo clona app")
    b.sh("rm -rf /tmp/ci && mkdir /tmp/ci && cp -r repo/app /tmp/ci/ && cd /tmp/ci/app && go build ./... 2>&1 | head -2",
         mostrar="(en la CI) go build ./...", env=E)
    b.sh("cd repo && git checkout -q app/go.mod && git status --short && echo '(limpio)'",
         mostrar="git checkout app/go.mod && git status --short")

    b.titulo("4 · Con go work")
    b.sh("cd repo && go work init ./app ./saludo && cat go.work", mostrar="go work init ./app ./saludo && cat go.work", env=E)
    b.sh("cd repo && go run ./app && go env GOWORK", mostrar="go run ./app && go env GOWORK", env=E)
    b.sh("cd repo && sed -i 's/Hola, /¡Hola, /' saludo/saludo.go && go run ./app",
         mostrar="(cambio el saludo en saludo/saludo.go) && go run ./app", env=E)

    b.titulo("5 · GOWORK=off y lo que queda en git")
    b.sh("cd repo/app && GOWORK=off go build ./... 2>&1 | head -1", mostrar="cd app && GOWORK=off go build ./...", env=E)
    b.sh("cd repo && git status --short", mostrar="git status --short")

    b.titulo("6 · Los módulos del workspace")
    b.sh("cd repo && go list -m", mostrar="go list -m", env=E)
    b.sh("cd repo && go work edit -json", mostrar="go work edit -json", env=E)

    b.titulo("7 · Un tercer módulo: go work use")
    b.sh("cd repo && mkdir -p util && printf 'module ejemplo.com/util\n\ngo 1.27\n' > util/go.mod && "
         "printf 'package util\n\nfunc Mayus(s string) string { return s + \"!\" }\n' > util/util.go && "
         "go work use ./util && cat go.work && go list -m",
         mostrar="mkdir util && (go.mod + util.go)\n$ go work use ./util && cat go.work && go list -m", env=E)
    b.sh("cd repo && mkdir -p herramientas/cli && printf 'module ejemplo.com/cli\n\ngo 1.27\n' > herramientas/cli/go.mod && "
         "printf 'package main\n\nfunc main() {}\n' > herramientas/cli/main.go && go work use -r . && cat go.work",
         mostrar="go work use -r .   # busca todos los go.mod por debajo", env=E)

    b.titulo("8 · ./... desde la raíz del workspace")
    b.sh("cd repo && go build ./... 2>&1 | head -2; echo \"exit=$?\"", mostrar="go build ./...", env=E)
    b.sh("cd repo && go vet ./app/... ./saludo/... && echo 'vet: sin problemas'",
         mostrar="go vet ./app/... ./saludo/...", env=E)

    b.titulo("9 · go mod tidy dentro de un workspace")
    b.sh("cd repo/app && go mod tidy 2>&1 | head -3", mostrar="cd app && go mod tidy", env=E)

    b.titulo("10 · replace en go.work, sobre un módulo del propio workspace")
    b.sh("cd repo && go work edit -replace ejemplo.com/saludo=./saludo && grep replace go.work && git diff --stat app/go.mod; echo '(app/go.mod sin cambios)'",
         mostrar="go work edit -replace ejemplo.com/saludo=./saludo && grep replace go.work\n$ git diff --stat app/go.mod", env=E)
    b.sh("cd repo && go list -m 2>&1", mostrar="go list -m", env=E)
    b.sh("cd repo && go work edit -dropreplace ejemplo.com/saludo && go list -m",
         mostrar="go work edit -dropreplace ejemplo.com/saludo && go list -m", env=E)
    b.sh("cd repo && ls go.work*", mostrar="ls go.work*")

    b.titulo("11 · Desde una subcarpeta también lo encuentra")
    b.sh("cd repo/app && go env GOWORK && cd ../saludo && go env GOWORK", mostrar="cd app && go env GOWORK\n$ cd ../saludo && go env GOWORK", env=E)

    b.titulo("12 · Quitar un módulo: -dropuse")
    b.sh("cd repo && go work edit -dropuse ./herramientas/cli && go work edit -dropuse ./util && go list -m",
         mostrar="go work edit -dropuse ./herramientas/cli && go work edit -dropuse ./util && go list -m", env=E)

    b.titulo("13 · Otro fichero de workspace con GOWORK")
    b.sh("cd repo && cp go.work /tmp/solo.work && GOWORK=/tmp/solo.work go list -m 2>&1",
         mostrar="cp go.work /tmp/solo.work && GOWORK=/tmp/solo.work go list -m", env=E)
    b.sh("cd repo && cp go.work solo-app.work && GOWORK=$PWD/solo-app.work go work edit -dropuse ./saludo && "
         "GOWORK=$PWD/solo-app.work go list -m && echo '--- y el go.work normal sigue igual:' && go list -m; rm -f solo-app.work",
         mostrar="cp go.work solo-app.work && GOWORK=$PWD/solo-app.work go work edit -dropuse ./saludo\n$ GOWORK=$PWD/solo-app.work go list -m\n$ go list -m", env=E)

    b.titulo("14 · go test en dos módulos a la vez")
    b.sh("cd repo && printf 'package saludo\n\nimport \"testing\"\n\nfunc TestHola(t *testing.T) {\n\tif Hola(\"Go\") != \"¡Hola, Go\" {\n\t\tt.Fatal(Hola(\"Go\"))\n\t}\n}\n' > saludo/saludo_test.go && "
         "go test ./app/... ./saludo/... 2>&1",
         mostrar="go test ./app/... ./saludo/...", env=E)
    b.sh("rm -rf repo", guardar=False)
