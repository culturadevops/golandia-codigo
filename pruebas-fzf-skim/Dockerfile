# gl-grb-fzf:1 — fzf (Go) y skim (Rust), binarios OFICIALES de cada release, verificados con su suma.
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl xz-utils python3 time procps \
 && rm -rf /var/lib/apt/lists/*
ARG FZF=0.75.0
ARG SKIM=5.9.0
WORKDIR /tmp/d
RUN curl -fsSLO https://github.com/junegunn/fzf/releases/download/v${FZF}/fzf-${FZF}-linux_amd64.tar.gz \
 && curl -fsSLO https://github.com/junegunn/fzf/releases/download/v${FZF}/fzf_${FZF}_checksums.txt \
 && grep " fzf-${FZF}-linux_amd64.tar.gz$" fzf_${FZF}_checksums.txt | sha256sum -c - \
 && tar xzf fzf-${FZF}-linux_amd64.tar.gz -C /usr/local/bin fzf \
 && curl -fsSLO https://github.com/skim-rs/skim/releases/download/v${SKIM}/skim-x86_64-unknown-linux-gnu.tar.xz \
 && curl -fsSLO https://github.com/skim-rs/skim/releases/download/v${SKIM}/skim-x86_64-unknown-linux-gnu.tar.xz.sha256 \
 && echo "$(cut -d' ' -f1 skim-x86_64-unknown-linux-gnu.tar.xz.sha256)  skim-x86_64-unknown-linux-gnu.tar.xz" | sha256sum -c - \
 && mkdir sk && tar xJf skim-x86_64-unknown-linux-gnu.tar.xz -C sk && find sk -type f -name sk -exec cp {} /usr/local/bin/sk \; \
 && cd / && rm -rf /tmp/d && fzf --version && sk --version
WORKDIR /w
