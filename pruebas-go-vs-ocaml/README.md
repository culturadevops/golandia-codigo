# Banco de «¿Go o OCaml 5?» (golandia 903)

Código del vídeo [«¿Go o OCaml 5? Lo medí: dominios en varios núcleos contra gorrutinas (2026)»](https://youtu.be/Ap4SJfWCKkU) del canal Golandia.

---

El mismo contrato de toda la serie «¿Go o X? Lo medí» (/suma y /trabajo, mismo resultado que el servidor de Go de
`../pruebas-go-java-revancha/go`). OCaml 5.5.2 no trae servidor HTTP: se usa cohttp-eio 6.3.0 sobre Eio 1.7, que reparte
las conexiones entre dominios (multinúcleo de OCaml 5) con `~additional_domains`.

| archivo | qué es |
|---|---|
| `src/servidor.ml`, `src/dune` | el servidor; `DOMINIOS=n` (por defecto 1, lo que hace cohttp-eio si no le pasas nada) |
| `Dockerfile.ocaml` | `opam install dune eio_main cohttp-eio mtime` + `dune build --profile release` → debian:trixie-slim |
| `SALIDA-corrida-1.md`, `SALIDA-corrida-2.md` | el banco, dos veces (`SALIDA.md` = la que sale en pantalla) |
| `SALIDA-demos.md` | núcleos que ve, hilos por dominio, el motor de Eio en Docker (sin cronómetro) |

Eio usa io_uring si puede. En Docker el perfil seccomp por defecto lo bloquea y Eio cae al motor posix sin avisar; la
variante «+ io_uring» usa `../pruebas-go-vs-odin/seccomp-io_uring.json` y `EIO_BACKEND=linux`.

```powershell
python investigacion\pruebas_go_vs_x_2026_11_23.py ocaml 1     # con el candado _logs\banco.lock tomado
python investigacion\pruebas_go_vs_x_2026_11_23.py ocaml 2
python recursos\pruebas-go-vs-ocaml\demos.py                    # sin cronómetro
```

Todo en Docker con el prefijo `gl-903-` (contenedor `gl-903-srv`, red `gl-903-red` 10.90.3.0/24, puerto 9903).
