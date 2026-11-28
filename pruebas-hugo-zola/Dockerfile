# gl-grb-hugo:1 — Hugo (Go) y Zola (Rust), binarios OFICIALES de cada release.
# Hugo se verifica con su checksums.txt; Zola no publica sumas: se anota la del tar descargado.
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 time procps \
 && rm -rf /var/lib/apt/lists/*
ARG HUGO=0.170.1
ARG ZOLA=0.23.6
WORKDIR /tmp/d
RUN curl -fsSLO https://github.com/gohugoio/hugo/releases/download/v${HUGO}/hugo_${HUGO}_linux-amd64.tar.gz \
 && curl -fsSLO https://github.com/gohugoio/hugo/releases/download/v${HUGO}/hugo_${HUGO}_checksums.txt \
 && grep " hugo_${HUGO}_linux-amd64.tar.gz$" hugo_${HUGO}_checksums.txt | sha256sum -c - \
 && tar xzf hugo_${HUGO}_linux-amd64.tar.gz -C /usr/local/bin hugo \
 && curl -fsSLO https://github.com/getzola/zola/releases/download/v${ZOLA}/zola-v${ZOLA}-x86_64-unknown-linux-gnu.tar.gz \
 && sha256sum zola-v${ZOLA}-x86_64-unknown-linux-gnu.tar.gz | tee /zola-tar.sha256 \
 && tar xzf zola-v${ZOLA}-x86_64-unknown-linux-gnu.tar.gz -C /usr/local/bin zola \
 && cd / && rm -rf /tmp/d && hugo version && zola --version
WORKDIR /w
