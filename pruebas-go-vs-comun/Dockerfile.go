FROM golang:1.27.2 AS build
WORKDIR /src
COPY . /src
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /servidor .
FROM scratch
COPY --from=build /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
