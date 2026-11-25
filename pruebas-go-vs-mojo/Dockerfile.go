FROM golang:1.27.2 AS build
WORKDIR /src
COPY go-computo /src
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /computo .
FROM scratch AS final
COPY --from=build /computo /computo
ENTRYPOINT ["/computo"]
FROM debian:trixie-slim AS medir
RUN apt-get update -qq && apt-get install -y -qq time >/dev/null && rm -rf /var/lib/apt/lists/*
COPY --from=build /computo /computo
