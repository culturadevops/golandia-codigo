FROM debian:trixie AS build
RUN apt-get update -qq && apt-get install -y -qq --no-install-recommends clang curl ca-certificates git >/dev/null
RUN curl -sSL https://github.com/odin-lang/Odin/releases/download/dev-2026-11a/odin-linux-amd64-dev-2026-11a.tar.gz | tar xz -C /opt \
 && ln -s /opt/odin-linux-amd64-*/odin /usr/local/bin/odin
RUN git clone -q https://github.com/laytan/odin-http /src/odin-http && git -C /src/odin-http checkout -q 5e301bef8f
COPY src /src/servidor
WORKDIR /src/servidor
RUN odin build . -o:speed -out:/servidor
FROM debian:trixie-slim
COPY --from=build /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
