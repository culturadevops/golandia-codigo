# El host OFICIAL de Azure Functions (el que genera `func init --worker-runtime custom --docker`: azure-functions/dotnet:4),
# con la versión fijada, y el custom handler de Go compilado estático al lado de host.json.
FROM golang:1.27.2 AS build
WORKDIR /src
COPY go/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/handler .

FROM mcr.microsoft.com/azure-functions/dotnet:4.1036.0.1
ENV AzureWebJobsScriptRoot=/home/site/wwwroot \
    AzureFunctionsJobHost__Logging__Console__IsEnabled=true \
    FUNCTIONS_WORKER_RUNTIME=custom
COPY app/ /home/site/wwwroot/
COPY --from=build /out/handler /home/site/wwwroot/handler
