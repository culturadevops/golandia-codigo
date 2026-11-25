# C# normal (JIT): se publica en el SDK oficial y se despliega sobre la imagen oficial con el runtime de ASP.NET Core.
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src
COPY Servidor.csproj Program.cs ./
RUN dotnet publish -c Release -o /out
FROM mcr.microsoft.com/dotnet/aspnet:10.0
COPY --from=build /out /app
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["dotnet", "Servidor.dll"]
