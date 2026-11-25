# Banco de «¿Go o Mojo?» (golandia 904)

Código del vídeo [«¿Go o Mojo? Lo medí: rápido para calcular, pero sin servidor web (2026)»](https://youtu.be/dbbpy_H6LLc) del canal Golandia.

---

Mojo 1.2.0 (la estable, `pip install mojo`) no trae TCP en su librería estándar: el ejemplo de `std.net` de su propia
documentación no compila (sí en la nightly 1.3.0.dev, y solo TCP, sin HTTP), y lightbug_http, el framework web de la
comunidad, está archivado. Así que no hay servidor web: se miden LAS MISMAS dos funciones de /suma y /trabajo como
programa, con el mismo arnés en Go, Mojo y Python 3.13.

| archivo | qué es |
|---|---|
| `src/computo.mojo` | suma y trabajo en Mojo + el arnés (`una`, `medir`, nada) |
| `go-computo/main.go` | suma y trabajo copiados del servidor de Go de la serie + el mismo arnés |
| `py-computo/computo.py` | lo mismo en Python 3.13 |
| `Dockerfile.go`, `Dockerfile.mojo`, `Dockerfile.python` | etapas `build`, `final` (lo que despliegas) y `medir` (+ GNU time) |
| `banco.py` | el banco → `SALIDA-corrida-<n>.md` (`SALIDA.md` = la que sale en pantalla) |
| `demos.py` | versión, licencia, std.net en la estable y en la nightly, lightbug_http, ldd → `SALIDA-demos.md` |

```powershell
python recursos\pruebas-go-vs-mojo\banco.py 1     # con el candado _logs\banco.lock tomado
python recursos\pruebas-go-vs-mojo\banco.py 2
python recursos\pruebas-go-vs-mojo\demos.py        # sin cronómetro
```

Imágenes con el prefijo `gl-904-`.
