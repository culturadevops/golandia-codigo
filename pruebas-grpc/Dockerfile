# Imagen del banco de pruebas del vídeo 119 (curso-grpc-go).
#
# Por qué una imagen propia y no `golang:1.26` a pelo: `protoc` es un binario de C++ que no viene
# en la imagen de Go, y los plugins (`protoc-gen-go`, `protoc-gen-go-grpc`) y `grpcurl` son
# binarios de Go que hay que instalar. Instalarlos en CADA corrida costaba minuto y medio y
# dependía de la red; así se hace UNA vez y las mediciones salen del mismo entorno siempre.
#
# ⛔ Las versiones NO van a `latest`: una medición que no se puede repetir dentro de tres meses
# no es una medición. Las que salgan de aquí son las que se citan en el vídeo.
FROM golang:1.26

RUN apt-get update && apt-get install -y --no-install-recommends protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 \
    && go install github.com/fullstorydev/grpcurl/cmd/grpcurl@v1.9.3

ENV PATH="/go/bin:${PATH}"
ENV GOFLAGS=-buildvcs=false
