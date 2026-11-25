# SALIDA — tipos y permisos: Deno 2.9.8 frente a Go 1.27.2 (golandia 699)

> Corrida del 2026-10-29 08:02. Regenerable con `python ejecutar_demos.py`. Cada bloque es la salida literal del contenedor (`docker run --rm`).

## 1 · Permisos

### Deno sin permisos: el servidor no arranca

```
$ cd /app && timeout 10 deno run servidor.ts
error: Uncaught (in promise) NotCapable: Requires net access to "0.0.0.0:8080", run again with the --allow-net flag
Deno.serve({
     ^
    at listen (ext:deno_net/01_net.js:1:7011)
    at serveInner (ext:deno_http/00_serve.ts:1:16770)
    at Object.serve (ext:deno_http/00_serve.ts:1:15240)
    at file:///app/servidor.ts:6:6
[código de salida: 1]
```

### Deno con permiso de red solo para su puerto

```
$ deno run --allow-net=0.0.0.0:8080 /d/permisos.ts
leer /etc/passwd: NotCapable: Requires read access to "/etc/passwd", run again with the --allow-read flag
leer la variable HOME: NotCapable: Requires env access to "HOME", run again with the --allow-env flag
conectar a evil.example:443: NotCapable: Requires net access to "evil.example:443", run again with the --allow-net flag
ejecutar sh: NotCapable: Requires run access to "sh", run again with the --allow-run flag
escuchar en 0.0.0.0:8080: PERMITIDO
[código de salida: 0]
```

### El binario de deno compile se lleva los permisos con los que lo compilaste

```
$ deno compile --quiet --allow-net=0.0.0.0:8080 -o /tmp/permisos /d/permisos.ts && /tmp/permisos
leer /etc/passwd: NotCapable: Requires read access to "/etc/passwd", specify the required permissions during compilation using `deno compile --allow-read`
leer la variable HOME: NotCapable: Requires env access to "HOME", specify the required permissions during compilation using `deno compile --allow-env`
conectar a evil.example:443: NotCapable: Requires net access to "evil.example:443", specify the required permissions during compilation using `deno compile --allow-net`
ejecutar sh: NotCapable: Requires run access to "sh", specify the required permissions during compilation using `deno compile --allow-run`
escuchar en 0.0.0.0:8080: PERMITIDO
[código de salida: 0]
```

### Deno con -A (todos los permisos): lo mismo que Go

```
$ deno run -A /d/permisos.ts 2>&1 | grep -v evil
leer /etc/passwd: PERMITIDO
leer la variable HOME: PERMITIDO
ejecutar sh: PERMITIDO
escuchar en 0.0.0.0:8080: PERMITIDO
[código de salida: 0]
```

### Go: el binario puede todo lo que pueda su usuario

```
$ cd /tmp && go run /d/go_permisos/main.go
leer /etc/passwd: PERMITIDO, 18 líneas
leer la variable HOME: PERMITIDO, /root
ejecutar sh: PERMITIDO, uid 0
[código de salida: 0]
```

## 2 · Tipos

### deno run NO revisa tipos: el error se ejecuta

```
$ deno run /d/error_tipo.ts
n + 1 = 51
[código de salida: 0]
```

### deno check sí lo para

```
$ deno check /d/error_tipo.ts
Check ../d/error_tipo.ts
TS2322 [ERROR]: Type 'string | null' is not assignable to type 'number'.
  Type 'null' is not assignable to type 'number'.
const n: number = q.get("n");
      ^
    at file:///d/error_tipo.ts:3:7

error: Type checking failed.
[código de salida: 1]
```

### deno compile también revisa tipos antes de compilar

```
$ deno compile -o /tmp/x /d/error_tipo.ts
Check ../d/error_tipo.ts
TS2322 [ERROR]: Type 'string | null' is not assignable to type 'number'.
  Type 'null' is not assignable to type 'number'.
const n: number = q.get("n");
      ^
    at file:///d/error_tipo.ts:3:7

error: Type checking failed.

  info: The program failed type-checking, but it still might work correctly.
  hint: Re-run with --no-check to skip type-checking.
[código de salida: 1]
```

### Go: el mismo error no compila

```
$ cd /tmp && go run /d/go_tipo/main.go
# command-line-arguments
/d/go_tipo/main.go:11:14: cannot use q.Get("n") (value of type string) as int value in variable declaration
[código de salida: 1]
```

## 3 · Anexo: los hilos de deno serve --parallel y el tamaño en disco

### --parallel con el contenedor limitado a 2 CPU

```
$ cd /app && timeout 3 deno serve --parallel --port 8080 --allow-net paralelo.ts; true
deno serve: Listening on http://0.0.0.0:8080/ with 2 threads
[código de salida: 0]
```

Sin `--cpus` (la máquina tiene 24 hilos), la misma orden dice:

```
deno serve: Listening on http://0.0.0.0:8080/ with 24 threads
```

### El límite de memoria de V8 con deno run (contenedor de 1 GiB)

Para buscar por qué el binario de `deno compile` usó más memoria bajo carga.

```
$ deno run /d/heap.ts
heap_size_limit=536 MiB total_heap_size=5 MiB nucleos=2
[código de salida: 0]
```

### El mismo límite en el binario de deno compile

```
$ deno compile --quiet -o /tmp/heap /d/heap.ts && /tmp/heap
heap_size_limit=536 MiB total_heap_size=5 MiB nucleos=2
[código de salida: 0]
```

Tamaño de las imágenes en disco, descomprimidas (`docker images`); la tabla 2 de SALIDA.md es lo que se descarga, comprimido (`docker image inspect .Size`):

```
v699-deno:1  261MB
v699-go:1  8.41MB
v699-deno-compile:1  178MB
```

