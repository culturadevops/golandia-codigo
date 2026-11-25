# SALIDA — go work — ampliación (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-11 02:57 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

`$ bash preparar.sh   # app pide x/text v0.14.0, saludo v0.21.0`

```
repo listo
```

`$ grep x/text app/go.mod saludo/go.mod`

```
app/go.mod:require golang.org/x/text v0.14.0
saludo/go.mod:require golang.org/x/text v0.21.0
```

## 1 · Qué versión usa app dentro del workspace

`$ GOWORK=off go list -m golang.org/x/text;  go work init ./app ./saludo;  go list -m golang.org/x/text`

```
solo app:      golang.org/x/text v0.14.0
en el workspace: golang.org/x/text v0.21.0
require golang.org/x/text v0.14.0
```

`$ go build && GOWORK=off go build   # y go version -m de cada binario`

```
compilado en el workspace: x/text	v0.21.0
compilado solo:            x/text	v0.14.0
```

## 2 · go.work.sum

`$ git status --short && cat go.work.sum`

```
?? go.work
?? go.work.sum
---
golang.org/x/mod v0.17.0/go.mod h1:hTbmBsO62+eylJbnUtE2MGJUyE7QWk4xUqP
golang.org/x/sync v0.10.0/go.mod h1:Czt+wKu1gCyEFDUtn0jG5QVvpJ6rzVqr5a
golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d/go.mod h1:aiJ
```

## 3 · go work sync

`$ go work sync && git diff app/go.mod`

```
-require golang.org/x/text v0.14.0
+require golang.org/x/text v0.21.0
-golang.org/x/text v0.14.0 h1:ScX5w1eTa3QqT8oi6+ziP7dTV1S2+ALU0bI+0zXK
-golang.org/x/text v0.14.0/go.mod h1:18ZOQIKpY8NJVqYksKHtTdi31H5itFRjB
+golang.org/x/text v0.21.0 h1:zyQAAkrwaneQ066sspRyJaG9VNi/YJ1NfzcGB3hZ
```

## 4 · go work vendor

`$ go work vendor && ls vendor && head vendor/modules.txt`

```
golang.org
modules.txt
1
## workspace
# golang.org/x/text v0.21.0
3.4M
```
