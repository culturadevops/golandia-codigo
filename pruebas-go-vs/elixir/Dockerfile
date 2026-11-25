# Elixir: release de Mix compilada en la imagen oficial (Debian 13), desplegada sobre Debian 13 slim con lo mínimo
# que necesita la BEAM. El release lleva su propio ERTS dentro.
FROM elixir:1.20.4-otp-29 AS build
ENV MIX_ENV=prod
WORKDIR /src
RUN mix local.hex --force && mix local.rebar --force
COPY mix.exs ./
RUN mix deps.get --only prod && mix deps.compile
COPY lib lib
RUN mix release && mkdir /out && cp -r _build/prod/rel/servidor/. /out/
FROM debian:13.1-slim
RUN apt-get update && apt-get install -y --no-install-recommends libstdc++6 libncurses6 libsctp1 openssl ca-certificates \
 && rm -rf /var/lib/apt/lists/*
COPY --from=build /out /app
ENV LANG=C.UTF-8
EXPOSE 8080
ENTRYPOINT ["/app/bin/servidor", "start"]
