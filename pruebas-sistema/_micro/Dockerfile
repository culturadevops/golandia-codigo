FROM golang:alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /micro .

# scratch = imagen VACIA. Ni distribucion, ni shell, ni gestor de paquetes.
FROM scratch
COPY --from=build /micro /micro
EXPOSE 8080
ENTRYPOINT ["/micro"]
