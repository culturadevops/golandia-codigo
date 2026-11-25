# Banco de «¿Go o Odin?» (golandia 905)

Código del vídeo [«¿Go o Odin? Lo medí: el C moderno sin recolector contra Go (2026)»](https://youtu.be/bqnLK1d-JzM) del canal Golandia.

---

El mismo contrato de toda la serie «¿Go o X? Lo medí» (/suma y /trabajo, mismo resultado que el servidor de Go de
`../pruebas-go-java-revancha/go`). Odin (dev-2026-11a) no trae servidor HTTP en su librería: se usa odin-http (laytan,
commit 5e301bef8f), que en Linux usa io_uring y reparte las conexiones entre hilos. Sin recolector de basura: lo de cada
petición va al `context.temp_allocator`, que odin-http libera al terminar la petición.

| archivo | qué es |
|---|---|
| `src/servidor.odin` | el servidor; `HILOS=n` (si no, odin-http usa los núcleos que ve); `NOMBRES=fmt` arma «item-N» con fmt.tprintf |
| `Dockerfile.odin` | descarga de Odin + `git clone` de odin-http + `odin build -o:speed` → debian:trixie-slim |
| `seccomp-io_uring.json` | el perfil seccomp por defecto de Docker (github.com/moby/profiles, default.json) MÁS `io_uring_setup`, `io_uring_enter` e `io_uring_register` |
| `SALIDA-corrida-2.md` | el banco con el servidor definitivo (`SALIDA.md` = copia, la que sale en pantalla). La repetición (`odin 1`) se colgó 4 h el 2026-11-23 y no hay SALIDA-corrida-1.md: ver `_previa-fmt/LEEME.md` |
| `_previa-fmt/` | la primera corrida, con el servidor anterior (nombres con fmt.tprintf) |
| `SALIDA-demos.md` | el servidor con el seccomp de Docker (se cae), hilos que arranca, ldd (sin cronómetro) |

```powershell
python investigacion\pruebas_go_vs_x_2026_11_23.py odin 1     # con el candado _logs\banco.lock tomado
python investigacion\pruebas_go_vs_x_2026_11_23.py odin 2
python recursos\pruebas-go-vs-odin\demos.py                    # sin cronómetro
```

Todo en Docker con el prefijo `gl-905-` (contenedor `gl-905-srv`, red `gl-905-red` 10.90.5.0/24, puerto 9905).

⚠️ Si se repite: ponle un timeout duro al hey y al contenedor; la corrida del 2026-11-23 13:22 se quedó colgada 4 h.
