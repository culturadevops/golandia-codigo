# Pruebas — Go y el sistema operativo (vídeo `go-y-el-sistema-operativo`)

> Nace de la crítica de un espectador (2026-08-06): *«el 90 % de los vídeos de Golandia son sobre
> herramientas para hacer páginas web con Go, cuando ese no es ni por asomo el uso principal de Go
> […] sería mucho más interesante tocar manejo/interacción con el sistema, microservicios,
> infraestructura, apps (no web)»*.
>
> **Todo lo que sale en pantalla se ejecutó el 2026-08-06** dentro de `golang:alpine`
> (`go version go1.26.5 linux/amd64`), porque estas cosas —`/proc`, `SIGTERM`, `ldd`— solo tienen
> sentido en Linux. Salida completa y sin retocar en [`salida-real.txt`](salida-real.txt).

## Cómo se reproduce

```sh
cd canales/golang/recursos/pruebas-sistema
MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/pruebas" -w //pruebas \
    golang:alpine sh //pruebas/correr.sh
```

## 1 · Procesos con plazo (`01_procesos.go`)

No es «ejecutar un comando»: es que el proceso hijo **muere solo** cuando vence el plazo.

```
a) el sistema dice: Linux 6.18.33.2-microsoft-standard-WSL2
b) pedi 30 segundos de sleep y volvi en 1.0s
   el error que devuelve: signal: killed
   motivo del contexto:   context deadline exceeded
```

👉 `exec.CommandContext` + `context.WithTimeout`. Sin hilos, sin vigilantes, sin señales a mano.

## 2 · Apagado ordenado (`02_senales.go`)

Docker y Kubernetes **no matan** tu contenedor: mandan `SIGTERM` y esperan. Lo que hagas en esos
segundos decide si pierdes peticiones a medias.

```
servidor arrancado, PID 1188 — esperando trabajo
[el sistema manda SIGTERM]
recibida la senal: dejo de aceptar peticiones nuevas
terminando las peticiones en vuelo...
cerrada la base de datos
adios limpiamente
```

👉 `signal.NotifyContext` son **dos líneas** para lo que en otros lenguajes es un manejador global.

## 3 · Leer la máquina sin dependencias (`03_leer_el_sistema.go`)

En Linux el núcleo publica su estado como ficheros de texto. Un agente de monitorización en Go es
`os.ReadFile` y a contar.

```
carga (1, 5 y 15 minutos): 0.00  0.00  0.00
procesos ejecutandose:     1/501
memoria total:             15.4 GiB
memoria disponible:        14.0 GiB (91%)
nucleos que ve el proceso: 24
```

## 4 · El binario estático (`04_binario_estatico.sh`) — 🥇 el argumento de verdad

```
== a) compilar para esta maquina ==
   servidor: 8345950 bytes

== b) que necesita para arrancar? ==
/lib/ld-musl-x86_64.so.1: servidor: Not a valid dynamic program

== c) compilar para OTRA maquina, sin salir de esta ==
   servidor-arm64     7777885 bytes
   servidor-mac       7933938 bytes
   servidor.exe       8395776 bytes

== d) cuantos destinos sabe hacer este compilador, sin instalar nada ==
47
```

**Las tres frases que sostienen el bloque:**

- `ldd` responde **«Not a valid dynamic program»**: no es que le falte una librería, es que **no
  enlaza con ninguna**. Ese fichero de 8,3 MB se copia a un servidor vacío y arranca.
- Desde un Linux x86 salieron, en tres órdenes, los binarios para **ARM64, macOS y Windows**. Sin
  instalar ningún compilador cruzado, sin contenedores, sin máquinas virtuales.
- `go tool dist list` devuelve **47 combinaciones** de sistema y arquitectura que este mismo
  compilador sabe generar.

⚠️ **Precisión obligatoria en la narración:** el binario es estático **porque no hay `cgo` de por
medio**. En cuanto un paquete usa cgo (por ejemplo un driver de base de datos en C), deja de serlo.
En el vídeo se dice.

## 5 · El microservicio, pesado de verdad (`05_microservicio.sh`)

Las cuatro piezas juntas —servidor HTTP, ruta de salud y apagado ordenado— metidas en una imagen
que parte de `scratch`, o sea de **nada**: ni distribución, ni consola, ni gestor de paquetes.
⚠️ Esta prueba corre en el **host** (necesita docker), no dentro del contenedor de Go.

```
== lo que pesa el microservicio entero, en disco ==
   golandia/micro:scratch  8.32MB

== con que se compara ==
   python:3.13-alpine (imagen BASE, sin tu codigo)  70.1MB
   node:22-alpine (imagen BASE, sin tu codigo)  232MB

== y arranca? ==
   GET /salud -> vivo
   mandando SIGTERM (lo mismo que hace docker stop / kubernetes)...
   parado limpiamente
```

⚠️ **La comparación es asimétrica y en el vídeo se dice en voz alta:** se compara el **resultado
final** de Go (8,32 MB, con el código dentro) contra el **punto de partida** de los otros dos
(imágenes base, todavía sin una línea de código). Se presenta así a propósito, porque ese *es* el
argumento: en Go no hay punto de partida que arrastrar. Pero decirlo es obligatorio.

## Lo que este material NO demuestra

- No se midió rendimiento de nada: no hay comparación de velocidad contra Python o Node, y en el
  vídeo no se afirma ninguna.
- Los 8,3 MB del binario son de un servidor HTTP mínimo con la biblioteca estándar; un programa real
  pesa más. Se presenta como lo que es.

## 📦 Dónde están los binarios

Los binarios que produce `04_binario_estatico.sh` (`servidor`, `servidor-arm64`, `servidor-mac`,
`servidor.exe`) **no se quedan aquí**: `recursos/` se respalda a S3 y ahí solo sube material para
recrear (canon §10). Archivados en `_ARCHIVO/binarios-pruebas-golang-2026-08-06/`.
Se regeneran con `correr.sh`, y sus tamaños están en la salida de arriba.
