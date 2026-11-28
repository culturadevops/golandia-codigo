# -*- coding: utf-8 -*-
"""Se ejecuta DENTRO del contenedor. Arranca una TUI en una pseudoterminal de 200×50, le pasa cada byte que pinta a pyte
(un emulador de terminal) y mira la PANTALLA, no el flujo de bytes: el cronómetro para cuando un texto aparece de verdad
en la pantalla emulada. Contesta las preguntas que la TUI le hace a la terminal (posición del cursor, etc.).

    python3 medir_tui.py <lazygit|gitui> <repo> <reps>

Por cada repetición imprime una línea JSON con:
  · estado: segundos hasta ver «nuevo-00.txt» (el estado del árbol de trabajo pintado);
  · log: segundos hasta ver el asunto del último commit en la lista de commits
    (lazygit lo enseña en la primera pantalla; gitui en la pestaña Log, tecla «2», que se pulsa al ver el estado);
  · fondo: segundos hasta ver el PRIMER commit del repositorio tras ir al final de la lista
    (lazygit: «4» panel de commits y «>»; gitui: tecla Fin; la tecla se repite cada décima hasta verlo), contados desde la primera pulsación;
  · RAM (VmRSS y VmHWM de /proc) en cada momento.
"""
import fcntl
import json
import os
import pty
import select
import signal
import struct
import sys
import termios
import time

import pyte

COLS, FILAS = 200, 50


class Pantalla(pyte.Screen):
    def __init__(self, fd):
        super().__init__(COLS, FILAS)
        self.fd = fd

    def write_process_input(self, data):
        os.write(self.fd, data.encode())

    def report_device_status(self, mode, **kw):     # pyte no acepta la forma privada (CSI ? n): se contesta igual
        return super().report_device_status(mode)

    def set_mode(self, *modes, **kw):
        return super().set_mode(*modes, **kw)


def arbol(pid):
    """RSS de la TUI más la de todos sus hijos vivos (lazygit lanza procesos git; gitui lee el repo en su propio proceso)."""
    hijos = {}
    for p in os.listdir("/proc"):
        if p.isdigit():
            try:
                for l in open(f"/proc/{p}/status"):
                    if l.startswith("PPid:"):
                        hijos.setdefault(int(l.split()[1]), []).append(int(p))
                        break
            except OSError:
                pass
    total, pila, n = 0, [pid], 0
    while pila:
        q = pila.pop()
        try:
            for l in open(f"/proc/{q}/status"):
                if l.startswith("VmRSS"):
                    total += int(l.split()[1])
            n += 1
        except OSError:
            pass
        pila += hijos.get(q, [])
    return round(total / 1024, 1), n


def mem(pid):
    d = {}
    d["arbol_MiB"], d["procesos"] = arbol(pid)
    try:
        for l in open(f"/proc/{pid}/status"):
            if l.startswith(("VmRSS", "VmHWM")):
                k, v = l.split(":")
                d[k] = round(int(v.split()[0]) / 1024, 1)
    except OSError:
        pass
    return d


def una(herr, repo, ultimo, primero):
    pid, fd = pty.fork()
    if pid == 0:
        fcntl.ioctl(0, termios.TIOCSWINSZ, struct.pack("HHHH", FILAS, COLS, 0, 0))
        os.chdir(repo)
        os.environ.update(TERM="xterm-256color")
        os.execvp(herr, [herr])
    scr = Pantalla(fd)
    # el pyte de Debian pasa private=True (secuencias CSI ? …) a manejadores que no lo aceptan: se le quita el argumento
    # a esos, en la instancia, antes de crear el flujo
    import inspect
    for nombre in set(pyte.Stream.csi.values()):
        f = getattr(scr, nombre, None)
        if f is None:
            continue
        ps = inspect.signature(f).parameters
        if "private" not in ps and not any(x.kind == x.VAR_KEYWORD for x in ps.values()):
            setattr(scr, nombre, (lambda g: (lambda *a, **kw: g(*a)))(f))
    st = pyte.ByteStream(scr)
    t0 = time.perf_counter()
    res = dict(herr=herr)

    def leer_hasta(texto, tope):
        t = time.perf_counter()
        while time.perf_counter() - t < tope:
            r, _, _ = select.select([fd], [], [], 0.02)
            if r:
                try:
                    data = os.read(fd, 1 << 16)
                except OSError:
                    return None
                if not data:
                    return None
                st.feed(data)
                if any(texto in l for l in scr.display):
                    return time.perf_counter()
        return None

    try:
        t = leer_hasta("nuevo-00.txt", 120)
        res["estado"] = None if t is None else round(t - t0, 3)
        res["mem_estado"] = mem(pid)
        if herr == "gitui":
            os.write(fd, b"2")
        t = leer_hasta(ultimo, 120)
        res["log"] = None if t is None else round(t - t0, 3)
        res["mem_log"] = mem(pid)
        leer_hasta("\x00nunca", 2.0)            # deja que termine lo que cargue en segundo plano
        res["mem_reposo"] = mem(pid)
        # ir al final: la tecla se repite cada décima de segundo hasta ver el primer commit (lazygit carga los commits por tandas
        # y su «>» lleva al final de lo ya cargado; a las dos se les trata igual)
        if herr == "lazygit":
            os.write(fd, b"4")
            leer_hasta("\x00nunca", 0.3)
        t1 = time.perf_counter()
        t, pulsaciones = None, 0
        while t is None and time.perf_counter() - t1 < 180:
            os.write(fd, b">" if herr == "lazygit" else b"\x1b[F")     # «>» en lazygit, Fin en gitui
            pulsaciones += 1
            t = leer_hasta(primero, 0.1)
        res["fondo"] = None if t is None else round(t - t1, 3)
        res["pulsaciones"] = pulsaciones
        leer_hasta("\x00nunca", 2.0)
        res["mem_fondo"] = mem(pid)
        res["pantalla_fondo"] = [l.rstrip() for l in scr.display if l.strip()][:6]
    finally:
        try:
            os.kill(pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
        os.waitpid(pid, 0)
        os.close(fd)
    return res


def main():
    herr, repo, reps = sys.argv[1], sys.argv[2], int(sys.argv[3])
    import subprocess
    n = int(subprocess.run(["git", "rev-list", "--count", "HEAD"], cwd=repo, capture_output=True, text=True).stdout)
    ultimo, primero = f"commit {n:06d}:", "commit 000001:"
    for i in range(reps):
        d = una(herr, repo, ultimo, primero)
        d["rep"] = i
        print(json.dumps(d, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
