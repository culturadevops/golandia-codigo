FROM rust:1.99.1-trixie
RUN apt-get update && apt-get install -y --no-install-recommends jq musl-tools \
 && rm -rf /var/lib/apt/lists/* && rustup target add x86_64-unknown-linux-musl
