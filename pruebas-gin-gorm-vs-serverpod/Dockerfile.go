# Gin + GORM: binario estático sobre scratch, igual que el Go de toda la serie «¿Go o X?».
FROM golang:1.27.2 AS build
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /servidor .
FROM scratch
COPY --from=build /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
