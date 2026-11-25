# La imagen de despliegue de Java: el jar + el JRE oficial que lo ejecuta.
FROM eclipse-temurin:21-jre
WORKDIR /app
COPY _bin/servidor.jar /app/servidor.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/servidor.jar"]
