# Banco de «Squirrel en Go: consultas SQL dinámicas sin concatenar strings» (golandia 881)

Código del vídeo [«Squirrel en Go: consultas SQL dinámicas sin concatenar strings (y sus 2 trampas)»](https://youtu.be/xt-VYWTpsFk) del canal Golandia.

---

El mismo buscador de productos (filtros opcionales, orden elegido, paginación) de cuatro formas, ejecutado contra un Postgres 18.6
desechable con 100.000 productos (`app/esquema.sql`, datos deterministas):

| forma | dónde |
|---|---|
| A · `fmt.Sprintf` (insegura, para ver la inyección) | `app/main.go`, `conSprintf` |
| B · a mano, con `$1, $2…` | `app/main.go`, `aMano` |
| C · Squirrel v1.5.4 con `PlaceholderFormat(sq.Dollar)` | `app/main.go`, `conSquirrel` |
| D · sqlc 1.31.1 con `sqlc.narg` | `sqlc/query.sql` → `app/bd/` (generado; `exportar.go` es a mano) |

```powershell
python recursos\pruebas-squirrel\correr.py     # → SALIDA.md (sin cronómetro: no hace falta el candado)
```

⛔ Las inyecciones (incluida la que borra la tabla `notas`) solo se ejecutan contra el contenedor `gl-881-pg` (`--rm`, datos en
tmpfs, sin volumen). Red `gl-881-red` (10.88.1.0/24); caché de módulos en el volumen `gl-881-gomod`. El guion borra contenedor y red.
