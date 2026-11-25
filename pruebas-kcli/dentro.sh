#!/bin/sh
set -e

echo "== go version =="
go version

echo
echo "== 1. kcli: arbol de subcomandos anidados (como kubectl) =="
cd /lab/etapas/01-kcli
go mod tidy >/tmp/tidy.log 2>&1 || (cat /tmp/tidy.log; exit 1)
go build -o /tmp/kcli .
/tmp/kcli --help
echo "---"
/tmp/kcli serve --help

echo
echo "== 2. solo archivo de configuracion (kcli.yaml), sin flags ni env =="
/tmp/kcli serve start

echo
echo "== 3. con variable de entorno KCLI_SERVER_HOST (con el replacer puesto en root.go) =="
KCLI_SERVER_HOST=desde-entorno /tmp/kcli serve start

echo
echo "== 4. la variable de entorno pierde contra un flag explicito =="
KCLI_SERVER_HOST=desde-entorno /tmp/kcli serve start --host desde-flag

echo
echo "== 4b. el mismo binario, un archivo JSON en vez de YAML, sin tocar el codigo =="
/tmp/kcli --config otro-formato.json serve start

echo
echo "== 5. config view: la configuracion ya resuelta =="
KCLI_SERVER_HOST=desde-entorno /tmp/kcli config view

echo
echo "== 6. autocompletado real generado por cobra (primeras 15 lineas) =="
/tmp/kcli completion bash | head -15

echo
echo "== 7. EL CONTROL: mismo config.yaml, misma variable de entorno, una sola linea de setup distinta =="
cd /lab/etapas/02-replacer-control
go mod tidy >/tmp/tidy2.log 2>&1 || (cat /tmp/tidy2.log; exit 1)
go build -o /tmp/replacer-on ./cmd/replacer-on
go build -o /tmp/replacer-off ./cmd/replacer-off

echo "--- con viper.SetEnvKeyReplacer(\".\", \"_\") ---"
KCLI_SERVER_HOST=desde-entorno /tmp/replacer-on
echo "--- sin esa linea (todo lo demas identico) ---"
KCLI_SERVER_HOST=desde-entorno /tmp/replacer-off

echo
echo "== 8. cobra-cli: el generador oficial de estructura =="
go install github.com/spf13/cobra-cli@v1.3.0 >/tmp/install-cobra-cli.log 2>&1 || (cat /tmp/install-cobra-cli.log; exit 1)
GOBIN="$(go env GOPATH)/bin"
mkdir -p /tmp/generado && cd /tmp/generado
go mod init generado >/dev/null 2>&1
"$GOBIN/cobra-cli" init --author "demo" 2>&1
echo "--- arbol que crea cobra-cli init ---"
find . -type f | sort

echo
echo "== 9. viper.New(): dos configuraciones a la vez, sin pisarse =="
cd /lab/etapas/03-viper-instances
go mod tidy >/tmp/tidy3.log 2>&1 || (cat /tmp/tidy3.log; exit 1)
go run .
