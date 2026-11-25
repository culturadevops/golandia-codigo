# Imagen de la API: compila en golang y copia SOLO el binario a scratch.
# Medido en SALIDA.md (lección 9): tamaño, usuario y qué pasa sin CGO_ENABLED=0.

FROM golang:1.27.2 AS compilar
WORKDIR /src
# Primero solo go.mod y go.sum: si no cambian, la descarga de módulos sale de la caché.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0: binario estático. Sin esto, enlaza con la libc del compilador y en scratch
# no arranca, con un mensaje que no ayuda: «no such file or directory».
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /api ./cmd/api

FROM scratch
# Certificados por si la API llama a algo por HTTPS. No hay shell, ni curl, ni usuario root.
COPY --from=compilar /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=compilar /api /api
USER 65532:65532
ENV PUERTO=8080
EXPOSE 8080
# Sin curl en la imagen, la sonda es el propio binario.
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 CMD ["/api", "salud"]
# Forma exec (con corchetes): /api es el PID 1 y recibe el SIGTERM de docker stop.
ENTRYPOINT ["/api"]
