FROM virtuslab/scala-cli:1.18.0 AS build
WORKDIR /src
COPY src/Servidor.scala /src/
RUN scala-cli --power package --server=false --assembly --preamble=false Servidor.scala -o /servidor.jar
FROM eclipse-temurin:25-jre
COPY --from=build /servidor.jar /app/servidor.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/servidor.jar"]
