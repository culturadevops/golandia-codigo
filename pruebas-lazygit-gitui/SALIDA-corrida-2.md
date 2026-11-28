# lazygit (Go) contra gitui (Rust) · corrida 2

> Corrida del 2026-11-27 09:51. Cada bloque, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_lazygit_gitui.py 2`.

- Máquina (docker info): 24 CPU · 16487469056 B · 6.18.33.2-microsoft-standard-WSL2 · 29.6.2
- Contenedor: `--cpus 2 --memory 1g`, imagen gl-grb-git:1 (debian:trixie-slim + binarios oficiales + git + pyte)

## 0 · Versiones, binarios y repos

```
commit=e0fbfd50c36e1ef88c52297ca34750e640b4f5f4, build date=2026-11-07T23:20:44Z, build source=binaryRelease, version=0.66.3, os=linux, arch=amd64, git version=2.47.3
gitui 0.28.1
git version 2.47.3
-rwxr-xr-x 1 1001 1001 14899808 Mar 24  2026 /usr/local/bin/gitui
-rwxr-xr-x 1 1001 1001 20627618 Nov  7 23:26 /usr/local/bin/lazygit
f6149b9ae203397158b0c89c13cfde718e7121d3d3cd2ebc597f93d6628d9b5b  gitui-linux-x86_64.tar.gz
/datos/repo-100000: 100000 commits, 5000 archivos, 30 cambios, .git 56M
/datos/repo-500000: 500000 commits, 5000 archivos, 30 cambios, .git 273M
```

## A · ¿Ven lo mismo? (primera pantalla de cada una, repo de 100.000)

Si alguna no llega a pintar el estado, el log o el primer commit, el banco para.

- lazygit: estado 0.02 s · log 0.359 s · fondo 0.582 s
  ```
  ╭─[1]─Status──────────────────────────────────────────────────────╮╭─[0]─Patch──────────────────────────────────────────────────────────────────────────────────
  │repo-100000 → main                                               ││commit bcbdff50e1f8f2b0ce18e331633d9ae9d1bbe250                                             
  ╰─────────────────────────────────────────────────────────────────╯│Author: banco <banco@example.com>                                                           
  ╭─[2]─Files - Worktrees - Submodules──────────────────────────────╮│Date:   Sun Sep 13 12:27:40 2020 +0000                                                      
  │▼ /                                                              ▐│                                                                                            
  │  ?? nuevo-00.txt                                                ▐│    commit 000001: cambia el archivo 00001                                                  
  ```
- gitui: estado 0.017 s · log 0.048 s · fondo 0.01 s
  ```
    Status [1]  |  Log [2]  |  Files [3]  |  Stashing [4]  |  Stashes [5]                                                                                         
   ───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  ┌Commit 1/100000────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  │a4e419b 2020-09-13 banco                 commit 000045: cambia el archivo 00045                                                                                
  │8f7a2ff 2020-09-13 banco                 commit 000044: cambia el archivo 00044                                                                                
  │53ea286 2020-09-13 banco                 commit 000043: cambia el archivo 00043                                                                                
  ```

## B · Arranque, log, ir al primer commit y RAM (mediana de 5; cada repetición, proceso nuevo)

| commits | herramienta | estado s | log s | ir al 1.er commit s | RAM al ver el log MiB | RAM en reposo MiB | RAM tras ir al final MiB | VmHWM final MiB | procesos |
|---|---|---|---|---|---|---|---|---|---|
| 100.000 | lazygit | 0.030 | 0.340 | 0.543 | 105.0 | 24.3 | 242.5 | 220.7 | [1, 2] |
| 100.000 | gitui | 0.018 | 0.048 | 0.011 | 26.5 | 17.6 | 17.6 | 47.9 | [1] |
| 500.000 | lazygit | 0.040 | 1.743 | 3.579 | 24.3 | 24.4 | 1085.2 | 1029.7 | [1] |
| 500.000 | gitui | 0.019 | 0.050 | 0.335 | 72.0 | 167.8 | 37.8 | 251.4 | [1] |

Crudo (una línea por repetición):

```
{"herr": "lazygit", "estado": 0.02, "mem_estado": {"arbol_MiB": 62.9, "procesos": 4, "VmHWM": 22.4, "VmRSS": 22.4}, "log": 0.34, "mem_log": {"arbol_MiB": 105.0, "procesos": 2, "VmHWM": 23.9, "VmRSS": 23.9}, "mem_reposo": {"arbol_MiB": 24.2, "procesos": 1, "VmHWM": 24.2, "VmRSS": 24.2}, "fondo": 0.575, "pulsaciones": 6, "mem_fondo": {"arbol_MiB": 265.0, "procesos": 2, "VmHWM": 243.2, "VmRSS": 243.2}, "rep": 0}
{"herr": "lazygit", "estado": 0.03, "mem_estado": {"arbol_MiB": 59.7, "procesos": 4, "VmHWM": 23.1, "VmRSS": 23.1}, "log": 0.346, "mem_log": {"arbol_MiB": 105.0, "procesos": 2, "VmHWM": 24.0, "VmRSS": 24.0}, "mem_reposo": {"arbol_MiB": 24.6, "procesos": 1, "VmHWM": 24.6, "VmRSS": 24.6}, "fondo": 0.543, "pulsaciones": 6, "mem_fondo": {"arbol_MiB": 242.0, "procesos": 2, "VmHWM": 220.3, "VmRSS": 220.3}, "rep": 1}
{"herr": "lazygit", "estado": 0.03, "mem_estado": {"arbol_MiB": 49.0, "procesos": 4, "VmHWM": 22.5, "VmRSS": 22.5}, "log": 0.341, "mem_log": {"arbol_MiB": 24.1, "procesos": 1, "VmHWM": 24.1, "VmRSS": 24.1}, "mem_reposo": {"arbol_MiB": 24.1, "procesos": 1, "VmHWM": 24.1, "VmRSS": 24.1}, "fondo": 0.552, "pulsaciones": 6, "mem_fondo": {"arbol_MiB": 264.9, "procesos": 2, "VmHWM": 243.0, "VmRSS": 243.0}, "rep": 2}
{"herr": "lazygit", "estado": 0.03, "mem_estado": {"arbol_MiB": 51.6, "procesos": 5, "VmHWM": 22.9, "VmRSS": 22.9}, "log": 0.325, "mem_log": {"arbol_MiB": 105.2, "procesos": 2, "VmHWM": 24.3, "VmRSS": 24.3}, "mem_reposo": {"arbol_MiB": 24.3, "procesos": 1, "VmHWM": 24.3, "VmRSS": 24.3}, "fondo": 0.52, "pulsaciones": 5, "mem_fondo": {"arbol_MiB": 231.8, "procesos": 2, "VmHWM": 210.0, "VmRSS": 210.0}, "rep": 3}
{"herr": "lazygit", "estado": 0.026, "mem_estado": {"arbol_MiB": 64.7, "procesos": 4, "VmHWM": 22.8, "VmRSS": 22.8}, "log": 0.315, "mem_log": {"arbol_MiB": 24.6, "procesos": 1, "VmHWM": 24.6, "VmRSS": 24.6}, "mem_reposo": {"arbol_MiB": 24.7, "procesos": 1, "VmHWM": 24.7, "VmRSS": 24.7}, "fondo": 0.536, "pulsaciones": 6, "mem_fondo": {"arbol_MiB": 242.5, "procesos": 2, "VmHWM": 220.7, "VmRSS": 220.7}, "rep": 4}
{"herr": "gitui", "estado": 0.018, "mem_estado": {"arbol_MiB": 13.2, "procesos": 1, "VmHWM": 23.8, "VmRSS": 13.1}, "log": 0.048, "mem_log": {"arbol_MiB": 26.3, "procesos": 1, "VmHWM": 38.6, "VmRSS": 26.3}, "mem_reposo": {"arbol_MiB": 17.4, "procesos": 1, "VmHWM": 47.9, "VmRSS": 17.4}, "fondo": 0.01, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.4, "procesos": 1, "VmHWM": 47.9, "VmRSS": 17.4}, "rep": 0}
{"herr": "gitui", "estado": 0.016, "mem_estado": {"arbol_MiB": 12.9, "procesos": 1, "VmHWM": 23.9, "VmRSS": 12.9}, "log": 0.044, "mem_log": {"arbol_MiB": 26.5, "procesos": 1, "VmHWM": 38.9, "VmRSS": 26.5}, "mem_reposo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 47.9, "VmRSS": 17.6}, "fondo": 0.011, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 47.9, "VmRSS": 17.6}, "rep": 1}
{"herr": "gitui", "estado": 0.018, "mem_estado": {"arbol_MiB": 12.9, "procesos": 1, "VmHWM": 23.9, "VmRSS": 12.9}, "log": 0.048, "mem_log": {"arbol_MiB": 26.5, "procesos": 1, "VmHWM": 39.1, "VmRSS": 26.5}, "mem_reposo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 48.1, "VmRSS": 17.6}, "fondo": 0.01, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 48.1, "VmRSS": 17.6}, "rep": 2}
{"herr": "gitui", "estado": 0.02, "mem_estado": {"arbol_MiB": 13.0, "procesos": 1, "VmHWM": 24.3, "VmRSS": 13.0}, "log": 0.048, "mem_log": {"arbol_MiB": 26.5, "procesos": 1, "VmHWM": 39.5, "VmRSS": 26.5}, "mem_reposo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 48.4, "VmRSS": 17.6}, "fondo": 0.011, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.7, "procesos": 1, "VmHWM": 48.4, "VmRSS": 17.7}, "rep": 3}
{"herr": "gitui", "estado": 0.019, "mem_estado": {"arbol_MiB": 12.9, "procesos": 1, "VmHWM": 24.0, "VmRSS": 12.9}, "log": 0.049, "mem_log": {"arbol_MiB": 26.5, "procesos": 1, "VmHWM": 39.0, "VmRSS": 26.5}, "mem_reposo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 47.9, "VmRSS": 17.6}, "fondo": 0.011, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.7, "procesos": 1, "VmHWM": 47.9, "VmRSS": 17.7}, "rep": 4}
{"herr": "lazygit", "estado": 0.034, "mem_estado": {"arbol_MiB": 163.2, "procesos": 4, "VmHWM": 23.1, "VmRSS": 23.1}, "log": 1.748, "mem_log": {"arbol_MiB": 24.5, "procesos": 1, "VmHWM": 24.5, "VmRSS": 24.5}, "mem_reposo": {"arbol_MiB": 24.7, "procesos": 1, "VmHWM": 24.7, "VmRSS": 24.7}, "fondo": 5.585, "pulsaciones": 54, "mem_fondo": {"arbol_MiB": 1085.6, "procesos": 2, "VmHWM": 1030.6, "VmRSS": 1019.4}, "rep": 0}
{"herr": "lazygit", "estado": 0.066, "mem_estado": {"arbol_MiB": 162.0, "procesos": 3, "VmHWM": 23.3, "VmRSS": 23.3}, "log": 1.774, "mem_log": {"arbol_MiB": 24.2, "procesos": 1, "VmHWM": 24.2, "VmRSS": 24.2}, "mem_reposo": {"arbol_MiB": 24.3, "procesos": 1, "VmHWM": 24.3, "VmRSS": 24.3}, "fondo": 5.616, "pulsaciones": 55, "mem_fondo": {"arbol_MiB": 928.0, "procesos": 2, "VmHWM": 1031.5, "VmRSS": 860.3}, "rep": 1}
{"herr": "lazygit", "estado": 0.035, "mem_estado": {"arbol_MiB": 103.1, "procesos": 5, "VmHWM": 22.6, "VmRSS": 22.6}, "log": 1.739, "mem_log": {"arbol_MiB": 24.3, "procesos": 1, "VmHWM": 24.3, "VmRSS": 24.3}, "mem_reposo": {"arbol_MiB": 24.4, "procesos": 1, "VmHWM": 24.4, "VmRSS": 24.4}, "fondo": 3.55, "pulsaciones": 34, "mem_fondo": {"arbol_MiB": 1083.9, "procesos": 2, "VmHWM": 1029.6, "VmRSS": 1016.3}, "rep": 2}
{"herr": "lazygit", "estado": 0.041, "mem_estado": {"arbol_MiB": 161.9, "procesos": 4, "VmHWM": 22.6, "VmRSS": 22.6}, "log": 1.734, "mem_log": {"arbol_MiB": 24.3, "procesos": 1, "VmHWM": 24.3, "VmRSS": 24.3}, "mem_reposo": {"arbol_MiB": 24.3, "procesos": 1, "VmHWM": 24.3, "VmRSS": 24.3}, "fondo": 3.578, "pulsaciones": 34, "mem_fondo": {"arbol_MiB": 1085.2, "procesos": 2, "VmHWM": 1029.6, "VmRSS": 1017.6}, "rep": 3}
{"herr": "lazygit", "estado": 0.04, "mem_estado": {"arbol_MiB": 100.2, "procesos": 4, "VmHWM": 22.8, "VmRSS": 22.8}, "log": 1.743, "mem_log": {"arbol_MiB": 24.6, "procesos": 1, "VmHWM": 24.6, "VmRSS": 24.6}, "mem_reposo": {"arbol_MiB": 24.6, "procesos": 1, "VmHWM": 24.6, "VmRSS": 24.6}, "fondo": 3.579, "pulsaciones": 34, "mem_fondo": {"arbol_MiB": 1095.1, "procesos": 2, "VmHWM": 1029.7, "VmRSS": 1027.6}, "rep": 4}
{"herr": "gitui", "estado": 0.013, "mem_estado": {"arbol_MiB": 12.7, "procesos": 1, "VmHWM": 31.5, "VmRSS": 12.7}, "log": 0.044, "mem_log": {"arbol_MiB": 71.9, "procesos": 1, "VmHWM": 128.4, "VmRSS": 71.9}, "mem_reposo": {"arbol_MiB": 167.8, "procesos": 1, "VmHWM": 167.8, "VmRSS": 167.8}, "fondo": 0.346, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.8, "procesos": 1, "VmHWM": 251.4, "VmRSS": 37.8}, "rep": 0}
{"herr": "gitui", "estado": 0.016, "mem_estado": {"arbol_MiB": 13.2, "procesos": 1, "VmHWM": 31.5, "VmRSS": 13.2}, "log": 0.046, "mem_log": {"arbol_MiB": 72.0, "procesos": 1, "VmHWM": 128.8, "VmRSS": 72.0}, "mem_reposo": {"arbol_MiB": 167.8, "procesos": 1, "VmHWM": 167.8, "VmRSS": 167.8}, "fondo": 0.325, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.8, "procesos": 1, "VmHWM": 251.7, "VmRSS": 37.8}, "rep": 1}
{"herr": "gitui", "estado": 0.019, "mem_estado": {"arbol_MiB": 12.8, "procesos": 1, "VmHWM": 31.5, "VmRSS": 12.8}, "log": 0.051, "mem_log": {"arbol_MiB": 72.0, "procesos": 1, "VmHWM": 128.6, "VmRSS": 72.0}, "mem_reposo": {"arbol_MiB": 167.7, "procesos": 1, "VmHWM": 167.7, "VmRSS": 167.7}, "fondo": 0.335, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.8, "procesos": 1, "VmHWM": 251.4, "VmRSS": 37.8}, "rep": 2}
{"herr": "gitui", "estado": 0.019, "mem_estado": {"arbol_MiB": 13.2, "procesos": 1, "VmHWM": 31.2, "VmRSS": 13.2}, "log": 0.05, "mem_log": {"arbol_MiB": 72.0, "procesos": 1, "VmHWM": 128.5, "VmRSS": 72.0}, "mem_reposo": {"arbol_MiB": 167.8, "procesos": 1, "VmHWM": 167.8, "VmRSS": 167.8}, "fondo": 0.339, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.8, "procesos": 1, "VmHWM": 251.2, "VmRSS": 37.8}, "rep": 3}
{"herr": "gitui", "estado": 0.02, "mem_estado": {"arbol_MiB": 13.0, "procesos": 1, "VmHWM": 31.9, "VmRSS": 13.0}, "log": 0.051, "mem_log": {"arbol_MiB": 72.2, "procesos": 1, "VmHWM": 129.2, "VmRSS": 72.2}, "mem_reposo": {"arbol_MiB": 167.8, "procesos": 1, "VmHWM": 168.0, "VmRSS": 167.8}, "fondo": 0.33, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 38.0, "procesos": 1, "VmHWM": 252.1, "VmRSS": 38.0}, "rep": 4}
```

FIN
