# -*- coding: utf-8 -*-
"""Demo del vídeo 682 de Golandia, «Cliente SSH en Go»: 50 servidores sshd en Docker y un cliente que los recorre.

    python correr.py [escenario ...] [--dejar]

Escenarios: openssh uno paralelo orden codigos malos ingenuo impostor agente pty bash ansible  (por defecto, todos)
Requiere las imágenes g682-sshd:1 (servidor/) y g682-cliente:1 (cliente/) y el binario flota/flota (Linux, estático).

Solo crea y borra contenedores g682-* y la red g682 (10.68.2.0/24); nunca toca otros. Al terminar los borra (salvo --dejar).
Roles: s13 «lleno» (sale con 2), s48 congelado (docker pause), s49 lento (tarda 30 s), s50 apagado (docker stop: su IP no contesta).
Cada escenario deja su salida en salidas/<escenario>.txt, sin editar; SALIDA.md se escribe a mano a partir de ellas.
"""
import os
import subprocess
import sys
import tempfile
import time
from pathlib import Path

AQUI = Path(__file__).parent
SAL = AQUI / "salidas"
SAL.mkdir(exist_ok=True)
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
RED = "g682"
CLI = "g682-cliente"
N = 50
HOSTS = [f"s{i:02d}" for i in range(1, N + 1)]
ROL = {"s13": "lleno", "s49": "lento"}
KH = "/root/.ssh/known_hosts"


def ip(h):
    return f"10.68.2.{10 + int(h[1:])}"


def sh(*a, check=True, tope=None):
    r = subprocess.run(list(a), capture_output=True, text=True, env=ENV, encoding="utf-8", errors="replace", timeout=tope)
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(a)}:\n{r.stdout}{r.stderr}")
    return r.stdout.strip()


def ex(cmd, tope=None):
    """Ejecuta en la máquina de control y devuelve «$ cmd» + la salida (stdout y stderr juntos)."""
    r = subprocess.run(["docker", "exec", CLI, "bash", "-c", "{ " + cmd + "\n} 2>&1"], capture_output=True, text=True, env=ENV,
                       encoding="utf-8", errors="replace", timeout=tope)
    return f"$ {cmd}\n{r.stdout.rstrip()}\n"


def guardar(nombre, texto):
    (SAL / f"{nombre}.txt").write_text(texto, encoding="utf-8")
    print(f"--- {nombre}\n{texto}")


def limpiar():
    nombres = sh("docker", "ps", "-aq", "--filter", "name=^g682-").split()
    if nombres:
        sh("docker", "rm", "-f", *nombres, check=False)
    sh("docker", "network", "rm", RED, check=False)


def servidor(h, rol="normal", nombre=None, pub=""):
    sh("docker", "run", "-d", "--name", nombre or f"g682-{h}", "--hostname", h, "--network", RED, "--ip", ip(h),
       "-e", f"CLAVE_PUB={pub}", "-e", f"ROL={rol}", "g682-sshd:1")


def esperar_ssh(h):
    for _ in range(60):
        if "ssh-ed25519" in sh("docker", "exec", CLI, "ssh-keyscan", "-T", "2", "-t", "ed25519", h, check=False):
            return
        time.sleep(0.5)
    raise SystemExit(f"{h} no levantó sshd")


def levantar():
    limpiar()
    sh("docker", "network", "create", "--subnet", "10.68.2.0/24", RED)
    hosts = [a for h in HOSTS for a in ("--add-host", f"{h}:{ip(h)}")]
    sh("docker", "run", "-d", "--name", CLI, "--network", RED, "--ip", "10.68.2.5", *hosts, "g682-cliente:1")
    sh("docker", "exec", CLI, "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-C", "flota-682", "-f", "/root/.ssh/id_ed25519")
    pub = sh("docker", "exec", CLI, "cat", "/root/.ssh/id_ed25519.pub")
    for h in HOSTS:
        servidor(h, ROL.get(h, "normal"), pub=pub)
    for h in HOSTS:
        esperar_ssh(h)
    # known_hosts: la huella de cada servidor, tomada mientras todos están sanos (en la vida real: al darlos de alta).
    sh("docker", "exec", CLI, "bash", "-c", f"ssh-keyscan -T 3 {' '.join(HOSTS)} > {KH} 2>/dev/null")
    sh("docker", "exec", CLI, "bash", "-c", "touch /root/.ssh/known_hosts_vacio")
    tmp = Path(tempfile.mkdtemp(prefix="g682-"))
    listas = {"hosts-50.txt": HOSTS, "hosts-sanos.txt": HOSTS[:47], "s01.txt": ["s01"], "s07.txt": ["s07"],
              "s13.txt": ["s13"], "s48.txt": ["s48"], "s49.txt": ["s49"], "s50.txt": ["s50"]}
    for nombre, hs in listas.items():
        (tmp / nombre).write_text("\n".join(hs) + "\n", encoding="utf-8", newline="\n")
        sh("docker", "cp", str(tmp / nombre), f"{CLI}:/root/{nombre}")
    for f in ("flota/flota", "bucle.sh", "paralelo.sh"):
        destino = "/usr/local/bin/flota" if f.startswith("flota") else f"/root/{f}"
        sh("docker", "cp", str(AQUI / f), f"{CLI}:{destino}")
    sh("docker", "exec", CLI, "chmod", "755", "/usr/local/bin/flota", "/root/bucle.sh", "/root/paralelo.sh")
    # Los tres que fallan: s48 congelado (acepta TCP y no contesta), s50 apagado (nadie en su IP). s49 ya es «lento».
    sh("docker", "pause", "g682-s48")
    sh("docker", "stop", "-t", "1", "g682-s50")


F = "cd /root && flota -clave /root/.ssh/id_ed25519 -known " + KH


def e_openssh():
    t = ex(f"head -c 120 {KH}; echo; grep -vc '^#' {KH}")
    t += ex("ssh -o BatchMode=yes admin@s01 revisar; echo \"código de salida: $?\"")
    t += ex("ssh -o BatchMode=yes admin@s13 revisar; echo \"código de salida: $?\"")
    guardar("openssh", t)


def e_uno():
    guardar("uno", ex(f"{F} -hosts s01.txt -p 1; echo \"código de salida: $?\""))


def e_paralelo():
    t = ex("wc -l < /root/hosts-sanos.txt")
    for p in (1, 5, 10, 25, 50):
        for _ in range(1 if p == 1 else 2):
            t += ex(f"{F} -hosts hosts-sanos.txt -p {p} | tail -1")
    guardar("paralelo", t)


def e_orden():
    t = ex(f"{F} -hosts hosts-sanos.txt -p 50 -llegada | head -8")
    t += ex(f"{F} -hosts hosts-sanos.txt -p 50 | head -8")
    guardar("orden", t)


def e_codigos():
    t = ex(f"{F} -hosts hosts-50.txt -p 50; echo \"código de salida: $?\"")
    t += ex(f"{F} -hosts s13.txt; echo \"código de salida: $?\"")
    guardar("codigos", t)


def e_malos():
    t = ""
    for h in ("s48", "s49", "s50"):
        t += ex(f"{F} -hosts {h}.txt")
    t += ex(f"{F} -hosts hosts-50.txt -p 1 | tail -4")
    guardar("malos", t)


def e_ingenuo():
    t = ex(f"cd /root && time timeout 60 {F[11:]} -ingenuo -hosts s48.txt; echo \"código de salida: $? (124 = lo mató timeout a los 60 s)\"", tope=90)
    t += ex(f"{F} -ingenuo -hosts s49.txt", tope=90)
    t += ex(f"{F} -ingenuo -hosts s50.txt", tope=90)
    guardar("ingenuo", t)


def e_impostor():
    pub = sh("docker", "exec", CLI, "cat", "/root/.ssh/id_ed25519.pub")
    sh("docker", "stop", "-t", "1", "g682-s07")
    servidor("s07", "impostor", nombre="g682-impostor", pub=pub)  # otra máquina en la IP de s07, con sus propias claves
    time.sleep(2)
    t = ex(f"{F} -inseguro -hosts s07.txt; echo \"código de salida: $?\"")
    t += ex(f"{F} -hosts s07.txt; echo \"código de salida: $?\"")
    t += ex("ssh -o BatchMode=yes admin@s07 revisar | head -3; echo \"código de salida: ${PIPESTATUS[0]}\"")
    t += ex("cd /root && flota -clave /root/.ssh/id_ed25519 -known /root/.ssh/known_hosts_vacio -hosts s01.txt")
    sh("docker", "rm", "-f", "g682-impostor")
    sh("docker", "start", "g682-s07")
    esperar_ssh("s07")
    t += ex(f"{F} -hosts s07.txt")
    guardar("impostor", t)


def e_agente():
    t = ex("eval $(ssh-agent -s) >/dev/null; ssh-add /root/.ssh/id_ed25519; ssh-add -l | cut -c1-30; "
           "cd /root && flota -agente -clave /no/existe -known " + KH + " -hosts s01.txt; ssh-agent -k >/dev/null")
    guardar("agente", t)


def e_pty():
    t = ex(f"{F} -hosts s01.txt -cmd 'tty; echo al-stderr >&2'")
    t += ex(f"{F} -hosts s01.txt -pty -cmd 'tty; echo al-stderr >&2'")
    guardar("pty", t)


def e_bash():
    tf = "TIMEFORMAT='tiempo: %Rs'; cd /root; "
    t = ex(tf + "time (./bucle.sh hosts-sanos.txt | tail -2)", tope=300)
    for p in (10, 50):
        t += ex(tf + f"time (./paralelo.sh hosts-sanos.txt {p} | tail -2)", tope=300)
    t += ex(tf + "time (./bucle.sh hosts-50.txt | tail -4)", tope=300)
    t += ex(tf + "time (./paralelo.sh hosts-50.txt 50 | sort | tail -4)", tope=300)
    guardar("bash", t)


def e_ansible():
    inv = "printf '[flota]\\n' > inv-sanos; cat hosts-sanos.txt >> inv-sanos; printf '[flota]\\n' > inv-50; cat hosts-50.txt >> inv-50; "
    tf = "TIMEFORMAT='tiempo: %Rs'; cd /root; "
    t = ex(f"cd /root; {inv} ansible --version | head -1")
    for f in (5, 10, 50, 50):
        t += ex(tf + f"time (ansible flota -i inv-sanos -u admin -m raw -a revisar -f {f} -o | grep -c 'rc=0')", tope=600)
    t += ex(tf + "time (timeout 120 ansible flota -i inv-50 -u admin -m raw -a revisar -f 50 -T 3 -o | grep -v 'rc=0' | cut -c1-120)", tope=300)
    guardar("ansible", t)


ESC = {"openssh": e_openssh, "uno": e_uno, "paralelo": e_paralelo, "orden": e_orden, "codigos": e_codigos, "malos": e_malos,
       "ingenuo": e_ingenuo, "impostor": e_impostor, "agente": e_agente, "pty": e_pty, "bash": e_bash, "ansible": e_ansible}


def main():
    pedidos = [a for a in sys.argv[1:] if not a.startswith("--")] or list(ESC)
    try:
        levantar()
        for n in pedidos:
            ESC[n]()
    finally:
        if "--dejar" not in sys.argv:
            limpiar()


if __name__ == "__main__":
    main()
