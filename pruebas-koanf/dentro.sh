#!/bin/sh
set -e

echo "== go version =="
go version
echo
echo "== versiones clavadas =="
cd /lab/etapas/01-koanf && grep -E "koanf|cobra" go.mod | sed 's/^/  /'

echo
echo "== 1. kcli migrado a koanf: solo archivo kcli.yaml =="
cd /lab/etapas/01-koanf
go mod tidy >/tmp/t1.log 2>&1 || (cat /tmp/t1.log; exit 1)
go build -o /tmp/kcli-koanf .
/tmp/kcli-koanf serve start

echo
echo "== 2. la variable de entorno KCLI_SERVER_HOST gana al archivo =="
KCLI_SERVER_HOST=desde-entorno /tmp/kcli-koanf serve start

echo
echo "== 3. la bandera explicita gana a la variable de entorno =="
KCLI_SERVER_HOST=desde-entorno /tmp/kcli-koanf serve start --server.host desde-flag

echo
echo "== 4. el mismo binario con un JSON, sin recompilar =="
/tmp/kcli-koanf --config otro-formato.json serve start

echo
echo "== 5. CONTROL: koanf SIN la funcion de transformacion =="
echo "   (el equivalente exacto de olvidar SetEnvKeyReplacer en viper)"
cd /lab/etapas/02-sin-transformacion
go mod tidy >/tmp/t2.log 2>&1 || (cat /tmp/t2.log; exit 1)
go build -o /tmp/kcli-control .
KCLI_SERVER_HOST=desde-entorno /tmp/kcli-control serve start
echo "   --- y esto es lo que SI se ve al depurar: ---"
KCLI_SERVER_HOST=desde-entorno /tmp/kcli-control config view

echo
echo "== 6. VIPER contra KOANF: que expone cada una cuando falla =="
cd /lab/etapas/03-visibilidad
go mod tidy >/tmp/t3.log 2>&1 || (cat /tmp/t3.log; exit 1)
go build -o /tmp/vis .
KCLI_SERVER_HOST=desde-entorno /tmp/vis

echo
echo "== 7. tamano del binario: koanf contra viper =="
cd /lab/etapas/01-koanf && go build -o /tmp/bin-koanf . && ls -l /tmp/bin-koanf | awk '{print "  koanf: "$5" bytes"}'
cd /lab/etapas/04-viper && go mod tidy >/tmp/t4.log 2>&1 || (cat /tmp/t4.log; exit 1)
go build -o /tmp/bin-viper . && ls -l /tmp/bin-viper | awk '{print "  viper: "$5" bytes"}'

echo
echo "== 8. cuantas dependencias arrastra cada una =="
cd /lab/etapas/01-koanf && echo "  koanf: $(go list -m all | wc -l) modulos en el grafo"
cd /lab/etapas/04-viper && echo "  viper: $(go list -m all | wc -l) modulos en el grafo"
