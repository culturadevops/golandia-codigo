#!/bin/sh
# Lección 10: el antes y el después del servicio, con tráfico.
# Cliente y servidor comparten las mismas CPU: las cifras sirven para comparar versiones.
cd "$(dirname "$0")/.." && . ./comun.sh

paso "go build -o \$T/carga ./10_carga"
for v in V1 V3 V5; do
	paso "\$T/carga -version $v -clientes 8 -duracion 5s"
done
