# Banco de «¿Go o Kotlin Native?» (golandia 879)

Código del vídeo [«¿Go o Kotlin Native? Lo medí: arranca como Go, pero con objetos (2026)»](https://youtu.be/buUkfJ9uzlQ) del canal Golandia.

---

Continuación del ¿Go o Kotlin? (486, `../pruebas-go-vs/kotlin`, Kotlin en la JVM con `com.sun.net.httpserver`). Aquí **el mismo
`src/commonMain/kotlin/Servidor.kt`** se compila de tres formas: a la JVM (jar con todo dentro), del jar a binario con GraalVM
`native-image`, y con Kotlin/Native (linuxX64). Como Kotlin/Native no tiene el JDK, el servidor web de Java no existe ahí: el
HTTP va con **Ktor 3.6.1 y el motor CIO**, que es Kotlin puro y compila a los dos lados. Solo cambia `Main.kt` (dos líneas por
plataforma: núcleos y variable `HILOS`). Go es el servidor de siempre (`../pruebas-go-java-revancha/go`).

| archivo | qué es |
|---|---|
| `build.gradle.kts` | Kotlin 2.4.21 multiplataforma: `jvm()` + `linuxX64()`, tarea `jarCompleto` |
| `Dockerfile.kotlin-jvm` | `gradle jarCompleto` → eclipse-temurin:25-jre |
| `Dockerfile.kotlin-graal` | el mismo jar → `native-image --no-fallback` (GraalVM CE 25.0.2) → debian:trixie-slim |
| `Dockerfile.kotlin-native` | `gradle linkReleaseExecutableLinuxX64` → debian:trixie-slim |
| `Dockerfile.kotlin-native-debug` | `gradle linkDebugExecutableLinuxX64` (el que usa `runDebugExecutableLinuxX64`) |
| `SALIDA-corrida-1.md`, `SALIDA-corrida-2.md` | el banco, dos veces |
| `SALIDA-extra.md` | debug contra release de Kotlin/Native |
| `SALIDA-demos.md` | núcleos que ve cada uno, `ldd`, tareas de Gradle (sin cronómetro) |

```powershell
mkdir <proyecto>\_logs\banco.lock          # candado: nadie más mide a la vez
python investigacion\pruebas_go_vs_nativo_2026_11.py kotlin-native 1
python investigacion\pruebas_go_vs_nativo_2026_11.py kotlin-native 2
python recursos\pruebas-go-vs-kotlin-native\extra.py
rmdir <proyecto>\_logs\banco.lock
python recursos\pruebas-go-vs-kotlin-native\demos.py            # sin cronómetro
```

Todo en Docker con el prefijo `gl-879-` (contenedor `gl-879-srv`, red `gl-879-red` con subred 10.87.9.0/24, puerto 9879).
