# Banco de «¿Go o Scala Native?» (golandia 865)

Código del vídeo [«¿Go o Scala Native? Lo medí: arranca como Go, pero con objetos (2026)»](https://youtu.be/jP3u1IDPAP4) del canal Golandia.

---

Continuación del 806 (`../pruebas-go-vs-scala`, Scala en la JVM con `jdk.httpserver`). Aquí **el mismo `src/Servidor.scala`**
se compila a la JVM y a Scala Native 0.5.12. Como Scala Native no trae `com.sun.net.httpserver`, el HTTP va a mano sobre
`java.net.ServerSocket` (HTTP/1.1 con keep-alive, un hilo por conexión) y el cálculo pasa por un semáforo con tantos permisos
como núcleos ve el runtime, o los de la variable `HILOS`. Go es el servidor de siempre (`../pruebas-go-java-revancha/go`).

| archivo | qué es |
|---|---|
| `Dockerfile.scala-jvm` | el jar con scala-cli (Scala 3.10.0, Java 25) sobre eclipse-temurin:25-jre |
| `Dockerfile.scala-native` | binario release-full **tal cual**: se cuelga bajo carga (ver SALIDA-extra.md) |
| `Dockerfile.scala-native-sintrap` | release-full recompilado con `SCALANATIVE_GC_TRAP_BASED_YIELDPOINTS=0` (el que se mide) |
| `Dockerfile.scala-native-defecto` | `scala-cli package --native` sin `--native-mode`: compila en debug |
| `Dockerfile.scala-native-commix` | prueba con Commix; descartado (más lento y lo mata el límite de 1 GiB) |
| `SALIDA.md` | el banco (corrida 2, la buena). `SALIDA-corrida-1.md`: la 1, para ver que se sostiene |
| `SALIDA-extra.md` | el cuelgue del release tal cual y el modo por defecto (+ `SALIDA-extra-corrida-1.md`) |
| `SALIDA-demos.md` | modos de scala-cli, `ldd`, el jar, y cuántos núcleos ve cada uno |

```powershell
mkdir <proyecto>\_logs\banco.lock          # candado: nadie más mide a la vez
python investigacion\pruebas_go_vs_lote_2026_11.py scala-native   # → SALIDA.md
python recursos\pruebas-go-vs-scala-native\extra.py               # → SALIDA-extra.md
rmdir <proyecto>\_logs\banco.lock
python recursos\pruebas-go-vs-scala-native\demos.py               # → SALIDA-demos.md (sin cronómetro)
```

⚠️ En esta máquina Docker se quedó sin rangos de red por las redes viejas de otros proyectos: la red del banco se crea antes a
mano con `docker network create --subnet 10.86.5.0/24 v865-red` (el script ignora que ya exista).
