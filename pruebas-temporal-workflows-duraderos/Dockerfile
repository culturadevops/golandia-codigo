# Los tres binarios (worker, iniciar, sintemporal) en una imagen. golang:1.27.2 compila; la imagen final es scratch.
FROM golang:1.27.2 AS build
WORKDIR /src
COPY go/go.mod go/go.sum ./
RUN go mod download
COPY go/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/ ./cmd/...

FROM scratch
COPY --from=build /out/ /bin/
