# SALIDA — pruebas-guardduty-go

Generado por `python ejecutar_todo.py`. **No se edita a mano.**

- fecha: 2026-09-30 18:58 UTC
- go version go1.26.5 windows/amd64
- moto 5.2.3 (emulador de AWS, en http://127.0.0.1:5077)
- Windows 10

Los eventos de `testdata/eventos.json` están **construidos a mano** con los campos que documenta AWS. No salen de una cuenta real. El emulador no implementa los findings.

## 1 · go vet

`$ go vet ./...` → código 0

```

```

## 2 · go test

`$ go test -count=1 -v ./...` → código 0

```
=== RUN   TestDecidir
--- PASS: TestDecidir (0.00s)
=== RUN   TestIgnoraOtrosEventos
--- PASS: TestIgnoraOtrosEventos (0.00s)
=== RUN   TestMensajeConSaltoDeLinea
--- PASS: TestMensajeConSaltoDeLinea (0.00s)
PASS
ok  	github.com/culturadevops/pruebas-guardduty-go/alerta	0.144s
?   	github.com/culturadevops/pruebas-guardduty-go/cmd/lambda	[no test files]
?   	github.com/culturadevops/pruebas-guardduty-go/cmd/leer	[no test files]
?   	github.com/culturadevops/pruebas-guardduty-go/cmd/supresion	[no test files]
=== RUN   TestParseTipo
--- PASS: TestParseTipo (0.00s)
=== RUN   TestParseTipoMalo
--- PASS: TestParseTipoMalo (0.00s)
=== RUN   TestTramoDe
--- PASS: TestTramoDe (0.00s)
PASS
ok  	github.com/culturadevops/pruebas-guardduty-go/hallazgo	0.134s
=== RUN   TestValidar
--- PASS: TestValidar (0.00s)
=== RUN   TestEntradaArchiva
--- PASS: TestEntradaArchiva (0.00s)
PASS
ok  	github.com/culturadevops/pruebas-guardduty-go/supresion	0.150s
```

## 3 · leer los eventos

`$ go run ./cmd/leer` → código 0

```
Recon:EC2/PortProbeUnprotectedPort
  propósito=Recon recurso=EC2 familia=PortProbeUnprotectedPort
  severidad=2.0 → Low · detectó: la inteligencia de AWS · veces=41
  acción: registro

UnauthorizedAccess:IAMUser/InstanceCredentialExfiltration.OutsideAWS
  propósito=UnauthorizedAccess recurso=IAMUser familia=InstanceCredentialExfiltration mecanismo=OutsideAWS
  severidad=8.0 → High · detectó: la inteligencia de AWS · veces=1
  acción: avisar-ya

CryptoCurrency:EC2/BitcoinTool.B!DNS
  propósito=CryptoCurrency recurso=EC2 familia=BitcoinTool mecanismo=B artefacto=DNS
  severidad=8.0 → High · detectó: la inteligencia de AWS · veces=1
  acción: prueba

AttackSequence:IAM/CompromisedCredentials
  propósito=AttackSequence recurso=IAM familia=CompromisedCredentials
  severidad=9.2 → Critical · detectó: la inteligencia de AWS · veces=1
  acción: avisar-ya

mensaje para la notificación (la descripción trae un salto de línea):
{"accion":"avisar-ya","cuenta":"\u003cCUENTA\u003e","descripcion":"Credentials created exclusively for an EC2 instance\nwere used from an IP address outside AWS.","proposito":"UnauthorizedAccess","recurso":"IAMUser","region":"us-east-1","tipo":"UnauthorizedAccess:IAMUser/InstanceCredentialExfiltration.OutsideAWS","tramo":"High","veces":1}
¿JSON válido? true
con SetEscapeHTML(false): {"cuenta":"<CUENTA>"}
```

## 4 · reglas de supresión contra el emulador

`$ go run ./cmd/supresion` → código 0

```
detector: ac5cf1f1abec426803cc71e8b72e6115 (32 caracteres)
detectores en esta región: 1

✓ creada: escaner-de-vulnerabilidades
✓ creada: bastion-ssh
✗ rechazada antes de llamar a la API · todo-el-portscan: una supresión solo por tipo apaga ese tipo en TODA la cuenta
✗ rechazada antes de llamar a la API · todo-kubernetes: un tipo con * no es un tipo concreto

escaner-de-vulnerabilidades · acción=ARCHIVE · rango=1
    resource.instanceDetails.imageId = [ami-999999999]
    type = [Recon:EC2/Portscan]
bastion-ssh · acción=ARCHIVE · rango=2
    resource.instanceDetails.tags.value = [devops]
    type = [UnauthorizedAccess:EC2/SSHBruteForce]

tras borrar bastion-ssh, GetFilter: operation error GuardDuty: GetFilter, https response error StatusCode: 400, RequestID: dMfVjzODL9zwspEgRRZnXlep4BKUB8KEcXhvlQJwPI2nDDo7j5zR, BadRequestException: The request is rejected since no such resource found.

ListFindings: HTTP 404 · el emulador no implementa los findings
```

## 5 · el handler de Lambda, para Linux arm64

`$ GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o _build/bootstrap ./cmd/lambda` → código 0

```
bootstrap: 7.4 MB
```
