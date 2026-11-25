# Zig 0.17: binario estático (musl) ReleaseFast, compilado con el tarball oficial (golandia-zig:0.17.0, ver
# recursos/pruebas-go-vs/zig/Dockerfile.zig) y desplegado sobre scratch. El mismo binario sirve las dos variantes (MODO).
FROM golandia-zig:0.17.0 AS build
WORKDIR /src
COPY main.zig .
RUN zig build-exe main.zig -O ReleaseFast -target x86_64-linux-musl -femit-bin=/servidor && rm -f /*.o
FROM scratch
COPY --from=build /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
