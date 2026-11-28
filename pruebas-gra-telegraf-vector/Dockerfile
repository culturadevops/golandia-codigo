# Banco golandia 960: Telegraf (Go) contra Vector (Rust), los binarios OFICIALES de sus imágenes, en el mismo contenedor.
FROM telegraf:1.40.4 AS telegraf
FROM timberio/vector:0.62.0-debian
RUN apt-get update && apt-get install -y --no-install-recommends time procps jq coreutils && rm -rf /var/lib/apt/lists/*
COPY --from=telegraf /usr/bin/telegraf /usr/local/bin/telegraf
ENTRYPOINT []
