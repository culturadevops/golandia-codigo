#!/bin/sh
# Genera gen/ desde proto/. Se corre dentro de la imagen curso-grpc:1.
set -e
cd /curso
protoc -I proto --go_out=. --go_opt=module=curso-grpc --go-grpc_out=. --go-grpc_opt=module=curso-grpc inventario/inventario.proto
protoc -I proto --go_out=. --go_opt=module=curso-grpc l01/antes.proto l01/despues.proto l01/bien.proto
echo generado
