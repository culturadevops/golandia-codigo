# lazygit (Go) contra gitui (Rust) · corrida 1

> Corrida del 2026-11-27 09:35. Cada bloque, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_lazygit_gitui.py 1`.

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

- lazygit: estado 0.016 s · log 0.319 s · fondo 0.546 s
  ```
  ╭─[1]─Status──────────────────────────────────────────────────────╮╭─[0]─Patch──────────────────────────────────────────────────────────────────────────────────
  │repo-100000 → main                                               ││commit bcbdff50e1f8f2b0ce18e331633d9ae9d1bbe250                                             
  ╰─────────────────────────────────────────────────────────────────╯│Author: banco <banco@example.com>                                                           
  ╭─[2]─Files - Worktrees - Submodules──────────────────────────────╮│Date:   Sun Sep 13 12:27:40 2020 +0000                                                      
  │▼ /                                                              ▐│                                                                                            
  │  ?? nuevo-00.txt                                                ▐│    commit 000001: cambia el archivo 00001                                                  
  ```
- gitui: estado 0.014 s · log 0.045 s · fondo 0.01 s
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
| 100.000 | lazygit | 0.027 | 0.331 | 0.530 | 24.2 | 24.5 | 264.5 | 242.7 | [1, 2] |
| 100.000 | gitui | 0.018 | 0.047 | 0.010 | 26.4 | 17.5 | 17.6 | 48.0 | [1] |
| 500.000 | lazygit | 0.038 | 1.735 | 4.376 | 24.4 | 24.6 | 1090.3 | 1030.1 | [1, 2] |
| 500.000 | gitui | 0.018 | 0.049 | 0.324 | 72.1 | 170.2 | 37.8 | 251.4 | [1] |

Crudo (una línea por repetición):

```
{"herr": "lazygit", "estado": 0.027, "mem_estado": {"arbol_MiB": 59.7, "procesos": 4, "VmHWM": 23.2, "VmRSS": 23.2}, "log": 0.331, "mem_log": {"arbol_MiB": 105.6, "procesos": 2, "VmHWM": 24.4, "VmRSS": 24.4}, "mem_reposo": {"arbol_MiB": 24.5, "procesos": 1, "VmHWM": 24.5, "VmRSS": 24.5}, "fondo": 0.583, "pulsaciones": 6, "mem_fondo": {"arbol_MiB": 269.2, "procesos": 2, "VmHWM": 247.4, "VmRSS": 247.4}, "rep": 0}
{"herr": "lazygit", "estado": 0.027, "mem_estado": {"arbol_MiB": 69.8, "procesos": 4, "VmHWM": 22.9, "VmRSS": 22.9}, "log": 0.349, "mem_log": {"arbol_MiB": 24.1, "procesos": 2, "VmHWM": 24.1, "VmRSS": 24.1}, "mem_reposo": {"arbol_MiB": 24.4, "procesos": 1, "VmHWM": 24.4, "VmRSS": 24.4}, "fondo": 0.517, "pulsaciones": 5, "mem_fondo": {"arbol_MiB": 221.2, "procesos": 2, "VmHWM": 199.5, "VmRSS": 199.5}, "rep": 1}
{"herr": "lazygit", "estado": 0.032, "mem_estado": {"arbol_MiB": 55.8, "procesos": 6, "VmHWM": 22.5, "VmRSS": 22.5}, "log": 0.325, "mem_log": {"arbol_MiB": 105.3, "procesos": 2, "VmHWM": 24.4, "VmRSS": 24.4}, "mem_reposo": {"arbol_MiB": 24.5, "procesos": 1, "VmHWM": 24.5, "VmRSS": 24.5}, "fondo": 0.539, "pulsaciones": 6, "mem_fondo": {"arbol_MiB": 283.0, "procesos": 2, "VmHWM": 261.2, "VmRSS": 261.2}, "rep": 2}
{"herr": "lazygit", "estado": 0.023, "mem_estado": {"arbol_MiB": 62.7, "procesos": 4, "VmHWM": 22.4, "VmRSS": 22.4}, "log": 0.314, "mem_log": {"arbol_MiB": 23.9, "procesos": 1, "VmHWM": 23.9, "VmRSS": 23.9}, "mem_reposo": {"arbol_MiB": 24.2, "procesos": 1, "VmHWM": 24.2, "VmRSS": 24.2}, "fondo": 0.53, "pulsaciones": 6, "mem_fondo": {"arbol_MiB": 242.3, "procesos": 2, "VmHWM": 220.6, "VmRSS": 220.6}, "rep": 3}
{"herr": "lazygit", "estado": 0.025, "mem_estado": {"arbol_MiB": 61.3, "procesos": 4, "VmHWM": 22.5, "VmRSS": 22.5}, "log": 0.343, "mem_log": {"arbol_MiB": 24.2, "procesos": 1, "VmHWM": 24.2, "VmRSS": 24.2}, "mem_reposo": {"arbol_MiB": 24.5, "procesos": 1, "VmHWM": 24.5, "VmRSS": 24.5}, "fondo": 0.518, "pulsaciones": 5, "mem_fondo": {"arbol_MiB": 264.5, "procesos": 2, "VmHWM": 242.7, "VmRSS": 242.7}, "rep": 4}
{"herr": "gitui", "estado": 0.013, "mem_estado": {"arbol_MiB": 12.8, "procesos": 1, "VmHWM": 23.9, "VmRSS": 12.8}, "log": 0.042, "mem_log": {"arbol_MiB": 26.4, "procesos": 1, "VmHWM": 38.6, "VmRSS": 26.4}, "mem_reposo": {"arbol_MiB": 17.5, "procesos": 1, "VmHWM": 48.0, "VmRSS": 17.5}, "fondo": 0.009, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.5, "procesos": 1, "VmHWM": 48.0, "VmRSS": 17.5}, "rep": 0}
{"herr": "gitui", "estado": 0.016, "mem_estado": {"arbol_MiB": 12.8, "procesos": 1, "VmHWM": 23.8, "VmRSS": 12.8}, "log": 0.046, "mem_log": {"arbol_MiB": 26.4, "procesos": 1, "VmHWM": 38.9, "VmRSS": 26.4}, "mem_reposo": {"arbol_MiB": 17.5, "procesos": 1, "VmHWM": 47.9, "VmRSS": 17.5}, "fondo": 0.01, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 47.9, "VmRSS": 17.6}, "rep": 1}
{"herr": "gitui", "estado": 0.019, "mem_estado": {"arbol_MiB": 12.9, "procesos": 1, "VmHWM": 24.0, "VmRSS": 12.9}, "log": 0.049, "mem_log": {"arbol_MiB": 26.5, "procesos": 1, "VmHWM": 39.0, "VmRSS": 26.5}, "mem_reposo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 48.1, "VmRSS": 17.6}, "fondo": 0.011, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 48.1, "VmRSS": 17.6}, "rep": 2}
{"herr": "gitui", "estado": 0.018, "mem_estado": {"arbol_MiB": 12.8, "procesos": 1, "VmHWM": 23.6, "VmRSS": 12.8}, "log": 0.048, "mem_log": {"arbol_MiB": 26.3, "procesos": 1, "VmHWM": 38.8, "VmRSS": 26.3}, "mem_reposo": {"arbol_MiB": 17.5, "procesos": 1, "VmHWM": 47.8, "VmRSS": 17.5}, "fondo": 0.01, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.5, "procesos": 1, "VmHWM": 47.8, "VmRSS": 17.5}, "rep": 3}
{"herr": "gitui", "estado": 0.018, "mem_estado": {"arbol_MiB": 13.2, "procesos": 1, "VmHWM": 24.1, "VmRSS": 13.2}, "log": 0.047, "mem_log": {"arbol_MiB": 26.5, "procesos": 1, "VmHWM": 38.8, "VmRSS": 26.5}, "mem_reposo": {"arbol_MiB": 17.6, "procesos": 1, "VmHWM": 48.1, "VmRSS": 17.6}, "fondo": 0.01, "pulsaciones": 1, "mem_fondo": {"arbol_MiB": 17.7, "procesos": 1, "VmHWM": 48.1, "VmRSS": 17.7}, "rep": 4}
{"herr": "lazygit", "estado": 0.025, "mem_estado": {"arbol_MiB": 105.5, "procesos": 6, "VmHWM": 22.9, "VmRSS": 22.9}, "log": 1.735, "mem_log": {"arbol_MiB": 24.5, "procesos": 1, "VmHWM": 24.5, "VmRSS": 24.5}, "mem_reposo": {"arbol_MiB": 24.6, "procesos": 1, "VmHWM": 24.6, "VmRSS": 24.6}, "fondo": 5.133, "pulsaciones": 50, "mem_fondo": {"arbol_MiB": 1089.1, "procesos": 2, "VmHWM": 1030.5, "VmRSS": 1021.7}, "rep": 0}
{"herr": "lazygit", "estado": 0.058, "mem_estado": {"arbol_MiB": 163.7, "procesos": 3, "VmHWM": 24.0, "VmRSS": 24.0}, "log": 1.716, "mem_log": {"arbol_MiB": 24.2, "procesos": 1, "VmHWM": 24.2, "VmRSS": 24.2}, "mem_reposo": {"arbol_MiB": 24.2, "procesos": 1, "VmHWM": 24.2, "VmRSS": 24.2}, "fondo": 3.645, "pulsaciones": 35, "mem_fondo": {"arbol_MiB": 1081.7, "procesos": 2, "VmHWM": 1029.6, "VmRSS": 1014.1}, "rep": 1}
{"herr": "lazygit", "estado": 0.038, "mem_estado": {"arbol_MiB": 92.2, "procesos": 4, "VmHWM": 22.8, "VmRSS": 22.8}, "log": 1.731, "mem_log": {"arbol_MiB": 24.1, "procesos": 2, "VmHWM": 24.1, "VmRSS": 24.1}, "mem_reposo": {"arbol_MiB": 24.6, "procesos": 1, "VmHWM": 24.6, "VmRSS": 24.6}, "fondo": 7.167, "pulsaciones": 70, "mem_fondo": {"arbol_MiB": 1094.1, "procesos": 2, "VmHWM": 1030.1, "VmRSS": 1028.8}, "rep": 2}
{"herr": "lazygit", "estado": 0.039, "mem_estado": {"arbol_MiB": 161.4, "procesos": 4, "VmHWM": 22.9, "VmRSS": 22.9}, "log": 1.739, "mem_log": {"arbol_MiB": 24.9, "procesos": 1, "VmHWM": 24.9, "VmRSS": 24.9}, "mem_reposo": {"arbol_MiB": 25.0, "procesos": 1, "VmHWM": 25.0, "VmRSS": 25.0}, "fondo": 4.376, "pulsaciones": 43, "mem_fondo": {"arbol_MiB": 1097.4, "procesos": 2, "VmHWM": 1032.1, "VmRSS": 1032.1}, "rep": 3}
{"herr": "lazygit", "estado": 0.038, "mem_estado": {"arbol_MiB": 91.6, "procesos": 3, "VmHWM": 22.3, "VmRSS": 22.3}, "log": 1.746, "mem_log": {"arbol_MiB": 24.4, "procesos": 2, "VmHWM": 24.4, "VmRSS": 24.4}, "mem_reposo": {"arbol_MiB": 24.5, "procesos": 1, "VmHWM": 24.5, "VmRSS": 24.5}, "fondo": 3.878, "pulsaciones": 38, "mem_fondo": {"arbol_MiB": 1090.3, "procesos": 2, "VmHWM": 1027.7, "VmRSS": 1022.9}, "rep": 4}
{"herr": "gitui", "estado": 0.015, "mem_estado": {"arbol_MiB": 12.9, "procesos": 1, "VmHWM": 31.8, "VmRSS": 12.9}, "log": 0.048, "mem_log": {"arbol_MiB": 72.0, "procesos": 1, "VmHWM": 128.9, "VmRSS": 72.0}, "mem_reposo": {"arbol_MiB": 170.0, "procesos": 1, "VmHWM": 170.0, "VmRSS": 170.0}, "fondo": 0.337, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.8, "procesos": 1, "VmHWM": 251.6, "VmRSS": 37.8}, "rep": 0}
{"herr": "gitui", "estado": 0.015, "mem_estado": {"arbol_MiB": 13.4, "procesos": 1, "VmHWM": 31.6, "VmRSS": 13.3}, "log": 0.048, "mem_log": {"arbol_MiB": 72.0, "procesos": 1, "VmHWM": 129.0, "VmRSS": 72.0}, "mem_reposo": {"arbol_MiB": 169.9, "procesos": 1, "VmHWM": 170.0, "VmRSS": 169.9}, "fondo": 0.338, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.8, "procesos": 1, "VmHWM": 251.4, "VmRSS": 37.8}, "rep": 1}
{"herr": "gitui", "estado": 0.018, "mem_estado": {"arbol_MiB": 13.0, "procesos": 1, "VmHWM": 31.7, "VmRSS": 13.0}, "log": 0.05, "mem_log": {"arbol_MiB": 72.2, "procesos": 1, "VmHWM": 129.1, "VmRSS": 72.2}, "mem_reposo": {"arbol_MiB": 170.3, "procesos": 1, "VmHWM": 170.3, "VmRSS": 170.3}, "fondo": 0.322, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.9, "procesos": 1, "VmHWM": 249.9, "VmRSS": 37.9}, "rep": 2}
{"herr": "gitui", "estado": 0.018, "mem_estado": {"arbol_MiB": 12.8, "procesos": 1, "VmHWM": 31.5, "VmRSS": 12.8}, "log": 0.05, "mem_log": {"arbol_MiB": 72.1, "procesos": 1, "VmHWM": 128.9, "VmRSS": 72.1}, "mem_reposo": {"arbol_MiB": 170.2, "procesos": 1, "VmHWM": 170.2, "VmRSS": 170.2}, "fondo": 0.324, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.8, "procesos": 1, "VmHWM": 249.9, "VmRSS": 37.8}, "rep": 3}
{"herr": "gitui", "estado": 0.019, "mem_estado": {"arbol_MiB": 12.9, "procesos": 1, "VmHWM": 32.0, "VmRSS": 12.9}, "log": 0.049, "mem_log": {"arbol_MiB": 72.2, "procesos": 1, "VmHWM": 128.9, "VmRSS": 72.2}, "mem_reposo": {"arbol_MiB": 170.2, "procesos": 1, "VmHWM": 170.2, "VmRSS": 170.2}, "fondo": 0.316, "pulsaciones": 4, "mem_fondo": {"arbol_MiB": 37.9, "procesos": 1, "VmHWM": 251.7, "VmRSS": 37.9}, "rep": 4}
```

FIN
