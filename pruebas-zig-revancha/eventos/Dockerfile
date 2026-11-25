# Contexto: recursos/pruebas-zig-revancha. Variante eventos (epoll a mano), mismo compilador y mismas banderas que la otra.
FROM golandia-zig:0.17.0 AS build
WORKDIR /src
COPY zig/main.zig eventos/eventos.zig ./
RUN zig build-exe eventos.zig -O ReleaseFast -target x86_64-linux-musl -femit-bin=/servidor && rm -f /*.o
FROM scratch
COPY --from=build /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
