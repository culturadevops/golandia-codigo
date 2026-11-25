# C# compilado nativo (Native AOT): el SDK oficial con las herramientas de AOT, y la imagen mínima (chiseled) de Microsoft.
FROM mcr.microsoft.com/dotnet/sdk:10.0-noble-aot AS build
WORKDIR /src
COPY Servidor.csproj Program.cs ./
RUN dotnet publish -c Release -p:PublishAot=true -o /out && rm -f /out/*.dbg
FROM mcr.microsoft.com/dotnet/runtime-deps:10.0-noble-chiseled
COPY --from=build /out/Servidor /app/Servidor
EXPOSE 8080
ENTRYPOINT ["/app/Servidor"]
