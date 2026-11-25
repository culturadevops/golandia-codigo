FROM swift:6.4.1-noble AS build
WORKDIR /build
COPY Package.swift /build/
RUN swift package resolve
COPY Sources /build/Sources
RUN swift build -c release --static-swift-stdlib && cp "$(swift build -c release --show-bin-path)/Servidor" /servidor && strip /servidor
FROM ubuntu:noble
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libjemalloc2 tzdata && rm -rf /var/lib/apt/lists/*
COPY --from=build /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
