# El laboratorio del curso: Go, Bash, jq, file y GoReleaser, todo con versión fija.
FROM goreleaser/goreleaser:v2.18.1 AS goreleaser

FROM golang:1.27.2
RUN apt-get update -qq \
 && apt-get install -y -qq --no-install-recommends jq file \
 && rm -rf /var/lib/apt/lists/*
COPY --from=goreleaser /usr/bin/goreleaser /usr/local/bin/goreleaser
ENV GOTOOLCHAIN=local GOFLAGS=-buildvcs=auto
WORKDIR /work
