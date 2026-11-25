#!/bin/sh
# Corre las cuatro pruebas del video `go-y-el-sistema-operativo` dentro de un
# contenedor de Go sobre Linux, que es donde estas cosas tienen sentido
# (/proc, SIGTERM, ldd). Desde Windows:
#
#   MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/pruebas" -w /pruebas \
#       golang:alpine sh /pruebas/correr.sh
set -e
cd /pruebas

echo "############ ENTORNO ############"
go version
cat /etc/os-release | head -1

echo
echo "############ 1 · PROCESOS (os/exec + context) ############"
go run 01_procesos.go

echo
echo "############ 2 · SENALES (apagado ordenado) ############"
go run 02_senales.go

echo
echo "############ 3 · LEER EL SISTEMA (/proc, cero dependencias) ############"
go run 03_leer_el_sistema.go

echo
echo "############ 4 · EL BINARIO ESTATICO ############"
sh 04_binario_estatico.sh
