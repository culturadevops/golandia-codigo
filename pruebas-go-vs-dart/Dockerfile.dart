FROM dart:3.13.7 AS build
WORKDIR /app
COPY pubspec.yaml /app/
RUN dart pub get
COPY bin /app/bin
RUN dart compile exe bin/servidor.dart -o /app/servidor
FROM scratch
COPY --from=build /runtime/ /
COPY --from=build /app/servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
