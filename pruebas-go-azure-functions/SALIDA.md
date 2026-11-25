# SALIDA — Go en Azure Functions con custom handler, en LOCAL (golandia)

> Corrida del 2026-11-23 07:41. Regenerable con `python correr.py`. ⛔ Nada en la nube real: el host oficial de Functions en un contenedor, Core Tools y Azurite, todo en esta máquina.

- Core Tools `func --version`: 4.17.0 · host: `mcr.microsoft.com/azure-functions/dotnet:4.1036.0.1` (HOST_VERSION=4.635.1 en su entorno) · Azurite 3.38.0 · `go version go1.27.2 linux/amd64`
- Handler: net/http y nada más (go/main.go), compilado estático (CGO_ENABLED=0).

## 1 · Lo que genera `func init proy --worker-runtime custom --docker` (Core Tools 4.17.0)

```
.dockerignore
.gitignore
.vscode
Dockerfile
host.json
local.settings.json
--- Dockerfile
# DisableDockerDetector "Base docker file that is consumed by the other Dockerfiles in this directory. It is not intended to be used directly."
# FROM mcr.microsoft.com/azure-functions/dotnet:4-appservice
FROM mcr.microsoft.com/azure-functions/dotnet:4
ENV AzureWebJobsScriptRoot=/home/site/wwwroot \
    AzureFunctionsJobHost__Logging__Console__IsEnabled=true

COPY --from=installer-env ["/home/site/wwwroot", "/home/site/wwwroot"]
--- host.json (customHandler)
  "customHandler": {
    "description": {
      "defaultExecutablePath": "",
      "workingDirectory": "",
      "arguments": []
    }
  }
}
```

¿Construye tal cual? (`docker build` del Dockerfile generado, en una carpeta temporal del contenedor de Core Tools no hay Docker: se copia aquí a _generado/):

```
#3 ERROR: pull access denied, repository does not exist or may require authorization: server message: insufficient_scope: authorization failed
ERROR: failed to build: failed to solve: installer-env: failed to resolve source metadata for docker.io/library/installer-env:latest: pull access denied, repository does not exist or may require authorization: server message: insufficient_scope: authorization failed
```

## 2 · `func start` (Core Tools) con el handler de Go

host.json del proyecto (el mismo en `func start` y en la imagen):
```
{
  "version": "2.0",
  "logging": {
    "logLevel": {
      "default": "Warning",
      "Host.Startup": "Information",
      "Function": "Information",
      "Azure.Core": "Error"
    }
  },
  "extensionBundle": {
    "id": "Microsoft.Azure.Functions.ExtensionBundle",
    "version": "[4.*, 5.0.0)"
  },
  "customHandler": {
    "description": {
      "defaultExecutablePath": "handler",
      "workingDirectory": "",
      "arguments": []
    },
    "enableForwardingHttpRequest": true
  }
}
```

Salida de `func start` (las líneas de las funciones y del handler):
```
Core Tools Version:       4.17.0+fec3d82b386c4e84ecb615cf165effabb3841859 (64-bit)
Function Runtime Version: 4.1055.300.26527
[2026-11-23T12:41:41.076Z] 0 functions found (Custom)
[2026-11-23T12:41:41.161Z] 2026/11/23 12:41:41 [handler] escuchando en :38347 (pid 230)
[2026-11-23T12:41:41.171Z] Worker process started and initialized.
[2026-11-23T12:41:41.193Z] Host started (138ms)
Functions:
hola: [GET,POST] http://localhost:7071/api/hola
procesar: queueTrigger
```

`/api/hola?nombre=func-start`, pedido desde dentro del contenedor →
```
{"hola":"func-start","invocacion":2,"metodo":"GET","pid":230,"ruta":"/api/hola","vivo_desde_s":0}
```

### 2b · `func start` con AzureWebJobsStorage vacío (solo la función HTTP importa… ¿o no?)

```
Secret initialization from Blob storage failed due to missing both an Azure Storage connection string and a SAS connection uri. For Blob Storage, please provide at least one of these. If you intend to use files for secrets, add an App Setting key 'AzureWebJobsSecretStorageType' with value 'Files'.

/api/hola responde: no
```

## 3 · El host oficial en un contenedor (2 CPU, 1 GiB), con Azurite

Tres peticiones seguidas a /api/hola:
```
{"hola":"golandia-1","invocacion":2,"metodo":"GET","pid":28,"ruta":"/api/hola","vivo_desde_s":0}
{"hola":"golandia-2","invocacion":3,"metodo":"GET","pid":28,"ruta":"/api/hola","vivo_desde_s":0}
{"hola":"golandia-3","invocacion":4,"metodo":"GET","pid":28,"ruta":"/api/hola","vivo_desde_s":0}
```

Procesos dentro del contenedor (pid, nombre, memoria residente):
```
1 bash VmRSS: 3144 kB
28 handler VmRSS: 6636 kB
41 sh VmRSS: 1656 kB
9 Microsoft.Azure VmRSS: 185184 kB
```

El handler escucha en el puerto que le da el host: 37459 (FUNCTIONS_CUSTOMHANDLER_PORT; el host escucha en el 80 y reenvía).

## 4 · Disparador de cola (Azurite) → handler de Go → blob (binding de salida)

```
cola pedidos creada
mensaje 0a687e90-6684-42cb-b969-bcda7c56f4b9 en pedidos: {"id":"p-42","total":1000}
```

Lo que escribió el handler en su log (vía el host):
```
Executing 'Functions.procesar' (Reason='New queue message detected on 'pedidos'.', Id=5f8a03db-e814-46b2-bd50-b05cd8da11b7)
[handler] capas de texto quitadas: 2 → id=p-42 total=1000.00
[handler] /procesar Data.pedido="\"{\\\"id\\\":\\\"p-42\\\",\\\"total\\\":1000}\"" · claves de Metadata=[NextVisibleTime PopReceipt sys total DequeueCount ExpirationTime Id InsertionTime]
Executed 'Functions.procesar' (Succeeded, Id=5f8a03db-e814-46b2-bd50-b05cd8da11b7, Duration=164ms)
```

Blobs en el contenedor «facturas» (los escribió el HOST con lo que devolvió el handler en Outputs.factura):
```
b73cf88a-c5a8-4f72-b1a7-b020037d99f8.json: {"hecha_por_pid":28,"iva":160,"pedido":"p-42","total":1000}
```

## 5 · ¿Y si el proceso de Go muere? (`kill -9` al handler dentro del contenedor)

```
antes: {"hola":"antes","invocacion":5,"metodo":"GET","pid":28,"ruta":"/api/hola","vivo_desde_s":3}
kill -9 28
+0.04 s → 200 {"hola":"despues","invocacion":1,"metodo":"GET","pid":69,"ruta":"/api/hola","vivo_desde_s":0}
```

```
2026/11/23 12:42:23 [handler] escuchando en :37459 (pid 28)
Language Worker Process exited. Pid=28.
Microsoft.Azure.WebJobs.Script.Workers.WorkerProcessExitException: /home/site/wwwroot/handler exited with code 137 (0x89)
---> System.Exception: 2026/11/23 12:42:23 [handler] escuchando en :37459 (pid 28),[handler] capas de texto quitadas: 2 → id=p-42 total=1000.00,[handler] /procesar Data.pedido="\"{\\\"id\\\":\\\"p-42\\\",\\\"total\\\":1000}\"" · claves de Metadata=[NextVisibleTime PopReceipt sys total DequeueCount ExpirationTime Id InsertionTime]
2026/11/23 12:42:27 [handler] escuchando en :37459 (pid 69)
```

## 6 · Lo que pesa cada imagen (`docker image inspect .Size`)

| imagen | tamaño |
|---|---|
| host oficial + handler de Go (`gl-prac-azfunc:1`) | 454.3 MiB |
| base del host sola (`mcr.microsoft.com/azure-functions/dotnet:4.1036.0.1`) | 451.5 MiB |
| el handler solo (FROM scratch) (`gl-prac-azfunc-sologo:1`) | 2.7 MiB |

El binario: `-rwxr-xr-x 1 root root 6611104 Nov 23 12:10 /home/site/wwwroot/handler`

