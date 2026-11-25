# Kotlin: se compila a un jar con el runtime de Kotlin dentro, y se despliega sobre el JRE oficial (Temurin 21).
FROM golandia-kotlinc:2.4.21 AS build
WORKDIR /src
COPY Servidor.kt .
RUN mkdir /out && kotlinc Servidor.kt -include-runtime -d /out/servidor.jar
FROM eclipse-temurin:21-jre
COPY --from=build /out/servidor.jar /app/servidor.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/servidor.jar"]
