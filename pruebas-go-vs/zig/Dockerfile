# Zig: binario estático (musl) optimizado, compilado con el tarball oficial, y desplegado solo sobre scratch.
FROM golandia-zig:0.17.0 AS build
WORKDIR /src
COPY main.zig .
RUN mkdir /out && zig build-exe main.zig -O ReleaseFast -target x86_64-linux-musl -femit-bin=/out/servidor && rm -f /out/*.o
FROM scratch
COPY --from=build /out/servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
