FROM ghcr.io/gleam-lang/gleam:v1.19.2-erlang-alpine AS build
WORKDIR /app
COPY servidor/gleam.toml /app/
RUN gleam deps download
COPY servidor/src /app/src
RUN gleam export erlang-shipment && erl -noshell -eval 'io:format("~s~n",[erlang:system_info(otp_release)]), halt().'
FROM erlang:29-alpine
COPY --from=build /app/build/erlang-shipment /app
EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh", "run"]
