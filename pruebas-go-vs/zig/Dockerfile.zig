# Imagen para compilar Zig: el tarball oficial de ziglang.org, versión fija y suma comprobada.
FROM debian:13.1-slim
RUN apt-get update && apt-get install -y --no-install-recommends curl xz-utils ca-certificates && rm -rf /var/lib/apt/lists/* \
 && curl -sSL https://ziglang.org/download/0.17.0/zig-x86_64-linux-0.17.0.tar.xz -o /tmp/zig.tar.xz \
 && echo "1cbe9df9f27e6b78d14ccbca43b6703a404ef79ef1c463de901d7f088d4e2026  /tmp/zig.tar.xz" | sha256sum -c - \
 && tar -xJf /tmp/zig.tar.xz -C /opt && mv /opt/zig-x86_64-linux-0.17.0 /opt/zig && rm /tmp/zig.tar.xz
ENV PATH="/opt/zig:${PATH}"
