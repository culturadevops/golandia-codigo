FROM golang:1.27.2 AS compilar
WORKDIR /src
COPY go.mod .
COPY servicio ./servicio
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /servicio ./servicio

FROM scratch
COPY --from=compilar /servicio /servicio
ENTRYPOINT ["/servicio"]
