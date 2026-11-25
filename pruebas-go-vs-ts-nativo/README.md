# Banco de «¿Go o TypeScript compilado a nativo?» (golandia 880)

Código del vídeo [«¿Go o TypeScript compilado a nativo? Bun AOT y scriptc, lo medí (2026)»](https://youtu.be/94D-EpJk5jw) del canal Golandia.

---

El mismo `src/servidor.ts` (node:http, el contrato de toda la serie: /suma y /trabajo) compilado de tres formas, contra el
servidor de Go de siempre (`../pruebas-go-java-revancha/go`):

| variante | qué es | versión fijada |
|---|---|---|
| `bun --compile` | Bun entero + tu código empaquetado, con JIT | `oven/bun:1.4.4` (1.4.4+38afee6c8) |
| `bun --compile --aot` | código de máquina en la compilación, sin JIT | `oven/bun:canary@sha256:c4574411…` (1.4.5-canary.1+4ce14ba8b) |
| `scriptc` | Vercel Labs: TypeScript a binario nativo con los tipos, sin motor JS | 0.4.14 (tar de GitHub, sha256 comprobado) + clang 19 |

⚠️ Bun AOT **no está en la estable 1.4.4**: `bun build --compile --aot` se acepta, no avisa y sale el mismo ejecutable que sin
`--aot` (109.123.760 bytes los dos; SALIDA-demos.md §3). Está en canary: por eso se fija por digest.

Los tres son un solo hilo de TypeScript (no se usa `node:cluster`): Go sí usa los 2 núcleos del contenedor.

| archivo | qué es |
|---|---|
| `SALIDA-corrida-1.md`, `SALIDA-corrida-2.md` | el banco, dos veces, con el candado |
| `SALIDA-comprobar.md` | respuesta entera igual a Go (n = 1.000, 20.000, 50.000), CPU con carga y memoria tras 3.000 peticiones (`comprobar.py`, con candado) |
| `SALIDA-demos.md` | cobertura: 11 programas (`cobertura/`) con los tres compiladores, `scriptc coverage`, `ldd` |

```powershell
mkdir <proyecto>\_logs\banco.lock
python investigacion\pruebas_go_vs_nativo_2026_11.py ts-nativo 1
python investigacion\pruebas_go_vs_nativo_2026_11.py ts-nativo 2
rmdir <proyecto>\_logs\banco.lock
python recursos\pruebas-go-vs-ts-nativo\demos.py      # sin cronómetro
```

Todo en Docker con el prefijo `gl-880-` (contenedor `gl-880-srv`, red `gl-880-red` con subred 10.88.0.0/24, puerto 9880).
