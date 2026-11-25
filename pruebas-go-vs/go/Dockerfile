# Go: se compila en la imagen oficial y se despliega el binario estático, solo, sobre scratch.
FROM golang:1.27.2 AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/servidor .
FROM scratch
COPY --from=build /out/servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
