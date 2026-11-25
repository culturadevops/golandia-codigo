# Pyronaut 0.1.2 empaquetado como dice su guía para servicios: un FAT JAR (`pyronaut build --jar`) que se ejecuta con
# `java -jar` sobre el GraalVM JDK que provisiona `pyronaut setup`. Misma base que el 848 (python:3.13.17-slim, Debian 13).
# docker build -t gl-prac-pyronaut:1 .
FROM python:3.13.17-slim AS sdk
RUN pip install --no-cache-dir 'pyronaut==0.1.2' && pyronaut setup && pyronaut --version

FROM sdk AS build
WORKDIR /p
COPY proyecto/ /p/
RUN pyronaut install && pyronaut build --jar && ls -la dist/

FROM debian:trixie-slim
COPY --from=sdk /root/.local/share/pyronaut/sdks/graalvm-25.4.4.1.1+1.1 /opt/graalvm
COPY --from=build /p/dist/banco-1.0.0.jar /app/banco.jar
EXPOSE 8080
ENTRYPOINT ["/opt/graalvm/bin/java", "-jar", "/app/banco.jar"]
