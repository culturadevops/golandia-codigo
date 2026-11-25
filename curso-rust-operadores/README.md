# Curso de Rust para operadores Linux 2027 — Golandia (miembros)

Código de estos vídeos del canal Golandia:

- [Curso de Rust para operadores Linux 2027 · Parte 1 de 3: disco y archivos](https://youtu.be/kC_88J3rZEI) (miembros)
- [Curso de Rust para operadores Linux 2027 · Parte 2 de 3: memoria, CPU y procesos](https://youtu.be/3vWEQMicHP8) (miembros)
- [Curso de Rust para operadores Linux 2027 · Parte 3 de 3: logs y chequeo de guardia](https://youtu.be/BVbx6X652qk) (miembros)

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/kC_88J3rZEI**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/3vWEQMicHP8**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/BVbx6X652qk**

---

Una caja de herramientas de comandos propios para el día a día de un operador, escrita en Rust:
`opstool/`, con un binario por tarea. Cada lección es una carpeta con un `demo.sh` que los usa
contra directorios temporales (nada se toca fuera de `mktemp`).

| comando | qué hace | parte |
|---|---|---|
| `ops-du` | tamaño de un directorio: len(), sin repetir inode y bloques en disco | 1 |
| `ops-grandes` | los N archivos más grandes de un árbol | 1 |
| `ops-libre` | espacio con statvfs(2): libre, disponible y reservado, como df | 1 |
| `ops-borrados` | archivos borrados que siguen ocupando espacio porque alguien los tiene abiertos | 1 |
| `ops-mem` | memoria según /proc/meminfo y según el cgroup | 2 |
| `ops-comer` | reserva memoria de verdad (para probar límites) | 2 |
| `ops-cpu` / `ops-quemar` | uso de CPU de la máquina y de tu cgroup; carga de prueba | 2 |
| `ops-ps` | lista procesos leyendo /proc | 2 |
| `ops-matar` | SIGTERM, espera, y SIGKILL si no sale | 2 |
| `ops-seguir` | tail -f que aguanta la rotación (mv y copytruncate) | 3 |
| `ops-filtrar` | errores por minuto de un log con regex | 3 |
| `ops-capturar` | guarda stdout y stderr de un comando con rotación por tamaño | 3 |
| `ops-guardia` | chequeo de disco y memoria con código de salida para cron | 3 |

    python ejecutar_todo.py           # todo en Docker y regenera SALIDA.md
    python ejecutar_todo.py parte2    # solo una parte, por pantalla

El contenedor va con `--memory=1g --memory-swap=1g --cpus=2 --shm-size=256m` a propósito: varias
lecciones miden lo que pasa con esos límites. Crates fijados con `=` y `Cargo.lock`: libc 0.2.191,
regex 1.13.1, walkdir 2.5.0.

Para usarlos en tu máquina: `cd opstool && cargo build --release` y copia `target/release/ops-*`.
Solo Linux: leen /proc y /sys/fs/cgroup.
