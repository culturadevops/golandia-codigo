# nubecita en una imagen, en dos etapas.
#
# ⚠️ POR QUE HAY DOCKERFILE desde el episodio 3: el Application Control de la maquina de
# desarrollo empezo a bloquear tambien el binario del servidor recien compilado
# («Una directiva de Control de aplicaciones bloqueo este archivo»). Correr en contenedor quita
# de en medio esa fragilidad y, de paso, es como lo va a arrancar cualquiera:
#
#     docker build -t nubecita .
#     docker run --rm -p 4566:4566 nubecita
#
# La segunda etapa es `scratch`: NADA, ni siquiera un shell. Se puede porque el binario de Go es
# estatico con CGO_ENABLED=0, asi que la imagen entera pesa lo que pesa el binario. Es el mejor
# argumento de Go para contenedores, y aqui se ve sin trucos.

FROM golang:1.26 AS construir
WORKDIR /src
# Primero los modulos: asi la capa de dependencias se reutiliza mientras no cambien.
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
# -tags y -ldflags: sin la tabla de simbolos ni la informacion de depuracion.
# ⛔ El binario tiene que ser ESTATICO (CGO_ENABLED=0) o `scratch` no lo puede ejecutar.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /nubecita .

FROM scratch
COPY --from=construir /nubecita /nubecita
EXPOSE 4566
# 0.0.0.0 y no 127.0.0.1: dentro del contenedor, escuchar solo en la interfaz local dejaria
# fuera al que publica el puerto. Es el fallo numero uno al meter un servidor en un contenedor.
ENTRYPOINT ["/nubecita", "-escuchar", "0.0.0.0"]
