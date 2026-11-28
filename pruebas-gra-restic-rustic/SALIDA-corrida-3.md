# restic (Go) contra rustic (Rust) · corrida 3

> Corrida del 2026-11-27 08:17, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_restic_rustic.py 3`. Orden: restic → rustic.

## 0 · Versiones (binarios oficiales de sus imágenes)

- `restic 0.20.0 compiled with go1.27.4 on linux/amd64`
- `rustic v0.12.0`
- contenedores ajenos vivos al empezar: ninguno · CPU de Windows: 2 %

## 1 · ¿Leen y escriben el mismo formato? (compat.sh, sobre go/src/net: 461 archivos)

```
## A · repo de restic
restic init rc=0
restic backup rc=0
rustic snapshots rc=0
   rustic ve: 1 snapshot
rustic check --read-data rc=0
rustic restore rc=0
   diferencias tras restaurar con rustic: 0 (archivos 461 de 461)
rustic backup (en repo de restic) rc=0
restic snapshots rc=0
   restic ve: 2 snapshots
restic check --read-data rc=0
restic restore (snapshot de rustic) rc=0
   diferencias tras restaurar con restic: 0
## B · repo de rustic
rustic init rc=0
rustic backup rc=0
restic snapshots rc=0
   restic ve: 1 snapshots
restic check --read-data rc=0
restic restore rc=0
   diferencias tras restaurar con restic: 0
restic backup (en repo de rustic) rc=0
rustic snapshots rc=0
   rustic ve: 2 snapshot
rustic check --read-data rc=0
## config de cada repo
   restic init → {"version":2,"id":"c56e229e12667bc1aa37880a07f7b218e27d8ea95bdac43322ef4d065b353074","chunker_polynomial":"2521c780d8504d"}
   rustic init → {"version":2,"id":"5bf217bbe0d5a4cf5e15996106ced5aefc0a92d5f780960b9296323e968022c7","chunker_polynomial":"304c82fd271909"}
```

## 2 · restic: salida cruda de medir.sh

```
PASO init seg=2.85 rss_kb=94188 user=2.82 sys=0.04 rc=0
PASO inicial seg=4.64 rss_kb=131532 user=5.06 sys=1.39 rc=0
REPO inicial bytes=370906849 archivos=26
PASO sin_cambios seg=0.88 rss_kb=75116 user=0.68 sys=0.08 rc=0
  inicial> no parent snapshot found, will read all files
  inicial> Files:       15667 new,     0 changed,     0 unmodified
  inicial> Added to the repository: 854.487 MiB (352.934 MiB stored)
  inicial> processed 15667 files, 850.520 MiB in 0:04
  inicial> snapshot 27b301c3 saved
CAMBIO modificados=310 borrados=161 nuevo=32MiB
PASO incremental seg=1.01 rss_kb=95308 user=0.77 sys=0.18 rc=0
REPO incremental bytes=405938073 archivos=32
PASO restaurar seg=2.36 rss_kb=170848 user=1.57 sys=1.86 rc=0
PASO check seg=0.78 rss_kb=75308 user=0.57 sys=0.05 rc=0
PASO check_datos seg=2.41 rss_kb=94332 user=1.55 sys=0.26 rc=0
SNAPSHOTS 3
CONFIG {"version":2,"id":"7a12e8f75ff27fe7179450dd106c00988b7f6bc60f6942569fa54ccf74928be9","chunker_polynomial":"34ef37dfa1fa33"}
  incremental> Files:           1 new,   310 changed, 15196 unmodified
  incremental> Added to the repository: 38.419 MiB (33.375 MiB stored)
  incremental> processed 15507 files, 880.479 MiB in 0:00
RESTAURADO archivos=15507 bytes=923248762 diferencias=0
ORIGEN archivos=15507 bytes=923248762
```

## 2 · rustic: salida cruda de medir.sh

```
PASO init seg=0.20 rss_kb=139180 user=0.16 sys=0.04 rc=0
PASO inicial seg=9.40 rss_kb=139208 user=8.34 sys=10.23 rc=0
REPO inicial bytes=368462563 archivos=16
PASO sin_cambios seg=1.15 rss_kb=139672 user=0.58 sys=0.52 rc=0
  inicial> [INFO] Files:       15667 new, 0 changed, 0 unchanged
  inicial> [INFO] Added to the repo: 349.9 MiB (raw: 854.5 MiB)
  inicial> [INFO] processed 15667 files, 850.5 MiB
  inicial> [INFO] snapshot 4890e55c successfully saved.
CAMBIO modificados=310 borrados=161 nuevo=32MiB
PASO incremental seg=1.05 rss_kb=139152 user=0.71 sys=0.77 rc=0
REPO incremental bytes=403488799 archivos=22
PASO restaurar seg=4.28 rss_kb=591644 user=1.30 sys=4.06 rc=0
PASO check seg=0.58 rss_kb=139404 user=0.45 sys=0.49 rc=0
PASO check_datos seg=2.26 rss_kb=139548 user=1.35 sys=2.21 rc=0
SNAPSHOTS 3 snapshot
CONFIG {"version":2,"id":"915c6aa9ca6f954e2c580dc6aef1892e2acafcc32f3732763ef0621e16875f3c","chunker_polynomial":"3fa45cd7181d65"}
  incremental> [INFO] Files:       1 new, 310 changed, 15196 unchanged
  incremental> [INFO] Added to the repo: 33.3 MiB (raw: 38.4 MiB)
  incremental> [INFO] processed 15507 files, 880.5 MiB
RESTAURADO archivos=15507 bytes=923248762 diferencias=0
ORIGEN archivos=15507 bytes=923248762
```

## 3 · ¿Llegaron los dos al mismo estado?

- carpeta tras el cambio: 15507 archivos, 923248762 bytes (restic) · 15507 archivos, 923248762 bytes (rustic)
- restaurado: restic 15507 archivos / 923248762 bytes / 0 diferencias · rustic 15507 / 923248762 / 0
- ✅ sí: misma carpeta de partida, mismo cambio y restauración idéntica (diff -r = 0) en los dos.

## 4 · La tabla (2 CPU, 1 GiB; seg = reloj; CPU = user+sys; RAM = máximo residente del proceso)

| paso | restic seg | rustic seg | restic CPU s | rustic CPU s | restic RAM MiB | rustic RAM MiB |
|---|---:|---:|---:|---:|---:|---:|
| init | 2.85 | 0.20 | 2.86 | 0.20 | 92.0 | 135.9 |
| inicial | 4.64 | 9.40 | 6.45 | 18.57 | 128.4 | 135.9 |
| sin_cambios | 0.88 | 1.15 | 0.76 | 1.10 | 73.4 | 136.4 |
| incremental | 1.01 | 1.05 | 0.95 | 1.48 | 93.1 | 135.9 |
| restaurar | 2.36 | 4.28 | 3.43 | 5.36 | 166.8 | 577.8 |
| check | 0.78 | 0.58 | 0.62 | 0.94 | 73.5 | 136.1 |
| check_datos | 2.41 | 2.26 | 1.81 | 3.56 | 92.1 | 136.3 |

| repositorio | restic MiB | rustic MiB |
|---|---:|---:|
| tras el inicial | 353.7 | 351.4 |
| tras el incremental | 387.1 | 384.8 |
| carpeta de partida | 850.5 | (la misma) |

> Terminó 2026-11-27 08:19. CPU de Windows al final: 2 %.
