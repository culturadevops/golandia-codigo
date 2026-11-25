# La funcion, empaquetada como la empaquetarias para Lambda de verdad: un binario estatico
# `bootstrap` sobre el runtime `provided.al2023`. Aqui lo dejamos en `scratch` porque solo tiene
# que correr en nuestra red de Docker, pero el binario es el mismo.
FROM golang:1.26 AS construir
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bootstrap .

FROM scratch
COPY --from=construir /bootstrap /bootstrap
ENTRYPOINT ["/bootstrap"]
