# restic (Go) contra rustic (Rust) · corrida 2

> Corrida del 2026-11-26 22:17, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_restic_rustic.py 2`. Orden: rustic → restic.

## 0 · Versiones (binarios oficiales de sus imágenes)

- `restic 0.20.0 compiled with go1.27.4 on linux/amd64`
- `rustic v0.12.0`
- contenedores ajenos vivos al empezar: ninguno · CPU de Windows: 9 %

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
   restic init → {"version":2,"id":"d40ad23745caf657b5ae02ec5ddb674f72de5395235fb54c2d198ef2a1950ee9","chunker_polynomial":"260dc595fa9a37"}
   rustic init → {"version":2,"id":"75860b435d086ece3e187b36a80e21056acf0d4a6f7fceec8e95b531b3a4c270","chunker_polynomial":"25bd82dfb488df"}
```

## 2 · rustic: salida cruda de medir.sh

```
PASO init seg=0.20 rss_kb=139256 user=0.15 sys=0.04 rc=0
PASO inicial seg=8.20 rss_kb=139740 user=7.37 sys=8.78 rc=0
REPO inicial bytes=368666213 archivos=16
PASO sin_cambios seg=0.65 rss_kb=139676 user=0.42 sys=0.23 rc=0
  inicial> [INFO] Files:       15667 new, 0 changed, 0 unchanged
  inicial> [INFO] Added to the repo: 350.1 MiB (raw: 854.5 MiB)
  inicial> [INFO] processed 15667 files, 850.5 MiB
  inicial> [INFO] snapshot be4c0a02 successfully saved.
CAMBIO modificados=310 borrados=161 nuevo=32MiB
PASO incremental seg=0.83 rss_kb=139600 user=0.60 sys=0.68 rc=0
REPO incremental bytes=403694339 archivos=22
PASO restaurar seg=4.30 rss_kb=595552 user=1.24 sys=3.48 rc=0
PASO check seg=0.54 rss_kb=139820 user=0.44 sys=0.44 rc=0
PASO check_datos seg=1.98 rss_kb=139816 user=1.22 sys=1.91 rc=0
SNAPSHOTS 3 snapshot
CONFIG {"version":2,"id":"4396f14a90762caaf0b7d64f11aaa63f4cdefddc2752d20de7727cbf8d40eedc","chunker_polynomial":"37ffad44f095b3"}
  incremental> [INFO] Files:       1 new, 310 changed, 15196 unchanged
  incremental> [INFO] Added to the repo: 33.3 MiB (raw: 38.4 MiB)
  incremental> [INFO] processed 15507 files, 880.5 MiB
RESTAURADO archivos=15507 bytes=923248762 diferencias=0
ORIGEN archivos=15507 bytes=923248762
```

## 2 · restic: salida cruda de medir.sh

```
PASO init seg=3.34 rss_kb=93964 user=3.30 sys=0.04 rc=0
PASO inicial seg=3.81 rss_kb=123368 user=4.26 sys=0.91 rc=0
REPO inicial bytes=370985945 archivos=27
PASO sin_cambios seg=0.89 rss_kb=75624 user=0.69 sys=0.07 rc=0
  inicial> no parent snapshot found, will read all files
  inicial> Files:       15667 new,     0 changed,     0 unmodified
  inicial> Added to the repository: 854.489 MiB (353.008 MiB stored)
  inicial> processed 15667 files, 850.520 MiB in 0:03
  inicial> snapshot 5123121e saved
CAMBIO modificados=310 borrados=161 nuevo=32MiB
PASO incremental seg=0.99 rss_kb=112048 user=0.81 sys=0.13 rc=0
REPO incremental bytes=406017299 archivos=33
PASO restaurar seg=2.26 rss_kb=173728 user=1.43 sys=1.72 rc=0
PASO check seg=1.32 rss_kb=75752 user=0.64 sys=0.02 rc=0
PASO check_datos seg=1.31 rss_kb=97776 user=1.41 sys=0.14 rc=0
SNAPSHOTS 3
CONFIG {"version":2,"id":"44742556e698de9a173fe7c780a59aa18fc06e1c5d1dc9041d63299021d97bbb","chunker_polynomial":"3fa2e77de07617"}
  incremental> Files:           1 new,   310 changed, 15196 unmodified
  incremental> Added to the repository: 38.419 MiB (33.376 MiB stored)
  incremental> processed 15507 files, 880.479 MiB in 0:00
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
| init | 3.34 | 0.20 | 3.34 | 0.19 | 91.8 | 136.0 |
| inicial | 3.81 | 8.20 | 5.17 | 16.15 | 120.5 | 136.5 |
| sin_cambios | 0.89 | 0.65 | 0.76 | 0.65 | 73.9 | 136.4 |
| incremental | 0.99 | 0.83 | 0.94 | 1.28 | 109.4 | 136.3 |
| restaurar | 2.26 | 4.30 | 3.15 | 4.72 | 169.7 | 581.6 |
| check | 1.32 | 0.54 | 0.66 | 0.88 | 74.0 | 136.5 |
| check_datos | 1.31 | 1.98 | 1.55 | 3.13 | 95.5 | 136.5 |

| repositorio | restic MiB | rustic MiB |
|---|---:|---:|
| tras el inicial | 353.8 | 351.6 |
| tras el incremental | 387.2 | 385.0 |
| carpeta de partida | 850.5 | (la misma) |

> Terminó 2026-11-26 22:19. CPU de Windows al final: 4 %.
