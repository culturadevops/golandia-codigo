$ go version
go1.26.5
$ ollama --version
ollama version is 0.40.5

$ ls app && cat app/go.mod
go.mod
go.sum
main.go
module demo/servicio

go 1.26.5

require github.com/google/uuid v1.6.0

$ ./dockerfile-ia -modelo qwen2.5-coder:7b -errores cola app
== intento 1 (3.8 s)
   | # Usamos una imagen base de Go
   | FROM golang:1.26.5-alpine AS builder
   | 
   | # Establecemos el directorio de trabajo
   | WORKDIR /app
   | 
   | # Copiamos el archivo go.mod y go.sum
   | COPY go.mod go.sum ./
   | 
   | # Descargamos las dependencias
   | RUN go mod download
   | 
   | # Copiamos el código fuente
   | COPY . .
   | 
   | # Compilamos la aplicación
   | RUN go build -o main .
   | 
   | # Imagen final
   | FROM alpine:latest
   | 
   | # Establecemos el directorio de trabajo
   | WORKDIR /app
   | 
   | # Copiamos el binario compilado desde la etapa de construcción
   | COPY --from=builder /app/main .
   | 
   | # Exponemos el puerto 8080
   | EXPOSE 8080
   | 
   | # Definimos la variable de entorno PORT
   | ENV PORT 8080
   | 
   | # Comando para ejecutar la aplicación
   | CMD ["./main"]
   -> GET /salud -> 200 {"estado":"ok","peticion":"ba83d095"}

RESULTADO: funciona en el intento 1 · 3.9 s en total · imagen de 9 MB · modelo qwen2.5-coder:7b
(código de salida: 0)

$ ./dockerfile-ia -modelo qwen2.5-coder:7b -errores relevantes app
== intento 1 (3.6 s)
   | # Usamos una imagen base de Go
   | FROM golang:1.26.5-alpine AS builder
   | 
   | # Establecemos el directorio de trabajo
   | WORKDIR /app
   | 
   | # Copiamos el archivo go.mod y go.sum
   | COPY go.mod go.sum ./
   | 
   | # Descargamos las dependencias
   | RUN go mod download
   | 
   | # Copiamos el código fuente
   | COPY . .
   | 
   | # Compilamos la aplicación
   | RUN go build -o main .
   | 
   | # Imagen final
   | FROM alpine:latest
   | 
   | # Establecemos el directorio de trabajo
   | WORKDIR /app
   | 
   | # Copiamos el binario compilado desde la etapa de construcción
   | COPY --from=builder /app/main .
   | 
   | # Exponemos el puerto 8080
   | EXPOSE 8080
   | 
   | # Definimos la variable de entorno PORT
   | ENV PORT 8080
   | 
   | # Comando para ejecutar la aplicación
   | CMD ["./main"]
   -> GET /salud -> 200 {"estado":"ok","peticion":"585c8e67"}

RESULTADO: funciona en el intento 1 · 3.6 s en total · imagen de 9 MB · modelo qwen2.5-coder:7b
(código de salida: 0)

$ ./dockerfile-ia -modelo qwen2.5-coder:1.5b -errores relevantes app
== intento 1 (6.0 s)
   | # Use an official Go runtime as a parent image
   | FROM golang:1.26.5
   | 
   | # Set the working directory in the container
   | WORKDIR /app
   | 
   | # Copy the current directory contents into the container at /app
   | COPY . /app
   | 
   | # Install any needed packages specified in go.mod and go.sum
   | RUN go mod download
   | 
   | # Make port 8080 available to the world outside this container
   | EXPOSE 8080
   | 
   | # Define environment variable
   | ENV PORT 8080
   | 
   | # Run the app when the container launches
   | CMD ["go", "run", "main.go"]
   -> GET /salud -> 200 {"estado":"ok","peticion":"ee8f7e4b"}

RESULTADO: funciona en el intento 1 · 6.1 s en total · imagen de 312 MB · modelo qwen2.5-coder:1.5b
(código de salida: 0)

$ # la versión del 7B, revisada por una persona (Dockerfile.revisado)
$ docker build -t ia-demo:revisado app && docker run -d -p 18080:8080 ia-demo:revisado && curl /salud
  build: ok
  GET /salud -> {"estado":"ok","peticion":"38f26b39"}
  usuario de la imagen: 65532
  ¿hay shell en la imagen?: no
  imagen: 3 MB

