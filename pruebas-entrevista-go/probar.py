# -*- coding: utf-8 -*-
"""probar.py — escribe y EJECUTA en Docker las pruebas Go de cada episodio, y guarda la salida
LITERAL en `_trabajo/pruebas/SALIDA.md` del propio vídeo.

Por qué Docker y no el Go local: el Application Control de este Windows bloquea a ratos los
binarios que `go run` deja en el temporal, y una prueba que a veces corre no sirve para el canon.
Además el contenedor da un entorno CITABLE en el vídeo, que es lo que hizo el canal en 2026-08-01
con `golang:1.23-alpine`.

⚠️ Por qué vive en `recursos/` y NO en el `_trabajo/` del vídeo: `_trabajo/` es DESECHABLE — al
publicar, `fab.py limpiar` se lo lleva a la papelera. Estas pruebas son el respaldo de lo que se
afirma en pantalla y tienen que sobrevivir al vídeo. Es donde las puso el canal viejo
(`recursos/pruebas-*/`).

    python investigacion/pruebas_ep01.py    # regenera SALIDA.md del episodio 1
"""
import io
import subprocess
import sys
from pathlib import Path

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")
RAIZ = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-entrevista-go"
IMAGEN = "golang:1.26"


def correr(base, args, espera_fallo=False, timeout=600):
    """Ejecuta `args` dentro del contenedor, con la carpeta de pruebas montada en /app."""
    montaje = str(base).replace("\\", "/")
    docker = ["docker", "run", "--rm", "-v", f"{montaje}:/app", "-w", "/app",
              "-e", "GOFLAGS=-buildvcs=false", IMAGEN] + args
    p = subprocess.run(docker, capture_output=True, text=True,
                       encoding="utf-8", errors="replace", timeout=timeout)
    salida = ((p.stdout or "") + (p.stderr or "")).rstrip()
    cmd = " ".join(args)
    if p.returncode != 0 and not espera_fallo:
        print(f"    [!] {cmd} -> codigo {p.returncode}")
        print("        " + salida.replace("\n", "\n        ")[:900])
    elif p.returncode == 0 and espera_fallo:
        print(f"    [!] {cmd} -> se esperaba que FALLARA y salio 0")
    return {"cmd": cmd, "codigo": p.returncode, "salida": salida,
            "esperado": "fallo" if espera_fallo else "ok"}


def episodio(carpeta, titulo, pruebas, gomod="module pruebas\n\ngo 1.24\n"):
    base = RAIZ / carpeta
    base.mkdir(parents=True, exist_ok=True)
    (base / "go.mod").write_text(gomod, encoding="utf-8")
    ver = correr(base, ["go", "version"])
    doc = [f"# Pruebas ejecutadas — {titulo}", "",
           "> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.",
           "> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,",
           "> y lo que se muestra como código se ejecuta antes).", "",
           f"**Entorno citable en el vídeo:** contenedor `{IMAGEN}` · `{ver['salida']}`", ""]
    fallos = []
    for pr in pruebas:
        print(f"    · {pr['nombre']}")
        for ruta, contenido in pr["archivos"].items():
            f = base / ruta
            f.parent.mkdir(parents=True, exist_ok=True)
            f.write_text(contenido, encoding="utf-8")
        doc += ["---", "", f"## {pr['nombre']}", "", f"**La pregunta:** {pr['pregunta']}", ""]
        for ruta in pr["archivos"]:
            doc += [f"`{ruta}`", "", "```go",
                    (base / ruta).read_text(encoding="utf-8").rstrip(), "```", ""]
        for args, falla in pr["correr"]:
            r = correr(base, args, espera_fallo=falla)
            if (r["codigo"] != 0) != falla:
                fallos.append(f"{pr['nombre']} :: {r['cmd']}")
            marca = "   # ⚠️ falla A PROPÓSITO" if falla else ""
            doc += ["```console", f"$ {r['cmd']}{marca}", r["salida"],
                    f"[codigo de salida: {r['codigo']}]", "```", ""]
        if pr.get("nota"):
            doc += ["> " + pr["nota"].replace("\n", "\n> "), ""]
    # ⛔ Una corrida fallida NO puede llevarse por delante la prueba buena que ya está en disco.
    # Pasó el 2026-08-21: se cayó el demonio de Docker a mitad y la corrida vacía sobrescribió el
    # SALIDA.md ya verificado del episodio 1. Si algo no salió como se esperaba, se escribe al lado
    # con el nombre `SALIDA.FALLIDA.md` y el bueno se queda donde está.
    destino = base / ("SALIDA.md" if not fallos else "SALIDA.FALLIDA.md")
    destino.write_text("\n".join(doc), encoding="utf-8")
    rel = str(destino.relative_to(Path(r"<proyecto>"))).replace("\\", "/")
    print(f"  [{'ok' if not fallos else 'REVISAR'}] {rel}")
    for f in fallos:
        print(f"     ⚠️ no salio como se esperaba: {f}")
    if fallos:
        print("     ⛔ SALIDA.md intacto: sigue estando la última corrida buena.")
    return not fallos
