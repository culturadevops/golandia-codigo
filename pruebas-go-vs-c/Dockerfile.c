FROM gcc:16.2.0 AS build
RUN apt-get update && apt-get install -y --no-install-recommends libmicrohttpd-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY src/servidor.c /src/
RUN gcc -O2 -Wall -Wextra -o /servidor servidor.c -lmicrohttpd && strip /servidor
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends libmicrohttpd12t64 && rm -rf /var/lib/apt/lists/*
COPY --from=build /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
