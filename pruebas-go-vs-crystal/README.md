# Banco de «¿Go o Crystal?» (golandia 902)

Código del vídeo [«¿Go o Crystal? Lo medí: el lenguaje con cara de Ruby contra Go (2026)»](https://youtu.be/T35lhKysTgc) del canal Golandia.

---

El mismo contrato de toda la serie «¿Go o X? Lo medí» (/suma y /trabajo, mismo resultado que el servidor de Go de
`../pruebas-go-java-revancha/go`). Crystal con su librería estándar (`HTTP::Server`), sin frameworks.

| archivo | qué es |
|---|---|
| `src/servidor.cr` | el servidor; `HILOS=n` agranda el contexto de ejecución por defecto (`Fiber::ExecutionContext.default.resize`) |
| `Dockerfile.crystal` | `crystal build --release --no-debug --static` (musl) → scratch |
| `Dockerfile.crystal-defecto` | `crystal build --static`, sin `--release` (lo que sale si no pones nada) |
| `SALIDA-corrida-1.md`, `SALIDA-corrida-2.md` | el banco, dos veces (`SALIDA.md` = la que sale en pantalla) |
| `extra.py` | 1 contra 2 hilos según el número de clientes → SALIDA-extra.md. ⚠️ Escrito pero NO corrido: Docker dejó de responder el 2026-11-23 antes de poder correrlo |
| `SALIDA-demos.md` | núcleos que ve, hilos, tamaño del binario (sin cronómetro) |

```powershell
python investigacion\pruebas_go_vs_x_2026_11_23.py crystal 1     # con el candado _logs\banco.lock tomado
python investigacion\pruebas_go_vs_x_2026_11_23.py crystal 2
python recursos\pruebas-go-vs-crystal\demos.py                    # sin cronómetro
```

Todo en Docker con el prefijo `gl-902-` (contenedor `gl-902-srv`, red `gl-902-red` 10.90.2.0/24, puerto 9902).
