# Etapa 1: compilar con la imagen oficial de Go
FROM golang:1.27.2 AS compilar
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /tareas .

# Etapa 2: solo el binario, sin sistema operativo
FROM scratch
COPY --from=compilar /tareas /tareas
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/tareas"]
