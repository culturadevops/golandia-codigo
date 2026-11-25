FROM golang:1.27.2 AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /admision .

FROM scratch
COPY --from=build /admision /admision
USER 65532:65532
ENTRYPOINT ["/admision"]
