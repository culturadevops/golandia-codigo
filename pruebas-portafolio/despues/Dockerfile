# Imagen mínima: se compila en golang:1.26 y se entrega en scratch (solo el binario).
FROM golang:1.26 AS compilar
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /escaner ./cmd/escaner

FROM scratch
COPY --from=compilar /escaner /escaner
USER 65534:65534
ENTRYPOINT ["/escaner"]
