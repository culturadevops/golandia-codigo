# Banco golandia 959: restic (Go) contra rustic (Rust), los binarios OFICIALES de sus imágenes, en el mismo contenedor.
FROM restic/restic:0.20.0 AS restic
FROM ghcr.io/rustic-rs/rustic:v0.12.0 AS rustic
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends time diffutils coreutils findutils && rm -rf /var/lib/apt/lists/*
COPY --from=restic /usr/bin/restic /usr/local/bin/restic
COPY --from=rustic /rustic /usr/local/bin/rustic
