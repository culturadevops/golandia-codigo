# SALIDA — Go práctico 2027 · parte 1

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Ids, tiempos y tamaños son los de esta corrida.

- fecha: `2026-11-23 12:02:57Z`
- imagen Go: `golang:1.27.2` · Postgres: `postgres:18.6-alpine` · `Docker 29.6.2`

## versiones

```
$ go version
go version go1.27.2 linux/amd64
$ cat go.mod
module tareas

go 1.27.2
$ ls
go.mod
main.go
manejadores.go
memoria.go
tareas.go
$ go vet ./...
```

## arrancar

```
$ ./tareas &
```

## crear

```
$ curl -si -X POST localhost:8080/tareas -d '{"titulo":"comprar pan"}'
HTTP/1.1 201 Created
Content-Type: application/json
Location: /tareas/1
Date: Mon, 23 Nov 2026 12:02:56 GMT
Content-Length: 46

{"id":1,"titulo":"comprar pan","hecha":false}
$ curl -s -X POST localhost:8080/tareas -d '{"titulo":"  llamar al dentista  "}'
{"id":2,"titulo":"llamar al dentista","hecha":false}
$ curl -s localhost:8080/tareas
[{"id":1,"titulo":"comprar pan","hecha":false},{"id":2,"titulo":"llamar al dentista","hecha":false}]
```

## validar

```
$ curl -si -X POST localhost:8080/tareas -d '{"titulo":"   "}'
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json
Date: Mon, 23 Nov 2026 12:02:56 GMT
Content-Length: 45

{"error":"el título no puede estar vacío"}
$ curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '{"titulo":"'$(printf 'a%.0s' $(seq 121))'"}'
{"error":"el título tiene 121 caracteres; el máximo es 120"}
HTTP 422
$ curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '{"titulo":"pan"'
{"error":"JSON inválido: unexpected EOF"}
HTTP 400
$ curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '{"titel":"pan"}'
{"error":"JSON inválido: json: unknown field \"titel\""}
HTTP 400
$ curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '[1,2,3]'
{"error":"JSON inválido: json: cannot unmarshal array into Go value of type struct { Titulo string \"json:\\\"titulo\\\"\" }"}
HTTP 400
```

## obtener

```
$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/1
{"id":1,"titulo":"comprar pan","hecha":false}
HTTP 200
$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/99
{"error":"tarea no encontrada"}
HTTP 404
$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/abc
{"error":"id inválido: \"abc\""}
HTTP 400
$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/-5
{"error":"id inválido: \"-5\""}
HTTP 400
```

## marcar

```
$ curl -s -w 'HTTP %{http_code}\n' -X PATCH localhost:8080/tareas/1 -d '{"hecha":true}'
{"id":1,"titulo":"comprar pan","hecha":true}
HTTP 200
$ curl -s -w 'HTTP %{http_code}\n' -X PATCH localhost:8080/tareas/1 -d '{}'
{"error":"falta el campo hecha"}
HTTP 422
$ curl -s localhost:8080/tareas
[{"id":1,"titulo":"comprar pan","hecha":true},{"id":2,"titulo":"llamar al dentista","hecha":false}]
```

## borrar

```
$ curl -si -X DELETE localhost:8080/tareas/1
HTTP/1.1 204 No Content
Date: Mon, 23 Nov 2026 12:02:56 GMT

$ curl -s -w 'HTTP %{http_code}\n' -X DELETE localhost:8080/tareas/1
{"error":"tarea no encontrada"}
HTTP 404
```

## metodos

```
$ curl -si -X PUT localhost:8080/tareas
HTTP/1.1 405 Method Not Allowed
Allow: GET, HEAD, POST
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Mon, 23 Nov 2026 12:02:56 GMT
Content-Length: 19

Method Not Allowed
$ curl -si localhost:8080/
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Mon, 23 Nov 2026 12:02:56 GMT
Content-Length: 19

404 page not found
$ curl -s -o /dev/null -w '%{http_code}\n' -I localhost:8080/tareas
200
```

## reiniciar

```
$ kill -TERM $PID
[la API terminó con código 143]
$ cat api.log
2026/11/23 12:02:56 escuchando en :8080
$ ./tareas &
$ curl -s localhost:8080/tareas
[]
$ kill -TERM $PID
[la API terminó con código 143]

[demo.sh terminó con código 0]
```
