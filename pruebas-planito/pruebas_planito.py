# -*- coding: utf-8 -*-
"""pruebas_planito.py — compila, prueba y DEMUESTRA `planito`, y guarda la salida LITERAL en
`recursos/pruebas-planito/SALIDA.md`.

`planito` es la segunda herramienta de la casa: un Terraform de bolsillo en Go, que corre contra
`nubecita` (la primera). Entre las dos cierran el arco del encargo original del dueño:
*«servicios que se puedan destruir al final de la demo»* — aquí la demo se destruye sola, gratis,
y se demuestra que **no toca lo que no creó**.

    c:/p/planito/                                EL CODIGO
    canales/golandia/recursos/pruebas-planito/   LA EVIDENCIA

## Cómo se monta la demo, y por qué así

Las dos herramientas viven en contenedores en una **red de Docker propia**, para que planito
llame a nubecita por su nombre. Y se publica un puente con `socat` para que la CLI de AWS, que
corre en Windows, pueda mirar la misma nube desde fuera. Ese puente es lo que permite la escena
importante: **alguien crea un bucket a mano y planito no lo toca**.

⚠️ Todo en contenedor por el mismo motivo que en nubecita: el Application Control de esta máquina
bloquea a ratos los binarios recién compilados. Lo único que corre en Windows es la CLI de AWS,
que está firmada.

    python canales/golandia/investigacion/pruebas_planito.py
"""
import io
import os
import shutil
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")

CODIGO = Path("c:/p/planito")
NUBECITA = Path("<nubecita>")
EVID = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-planito"

IMAGEN_GO = "golang:1.26"
IMAGEN_NUBE = "nubecita:ep3"
IMAGEN = "planito:1"
CACHE_MOD = "nubecita-gomod"
RED = "planito-demo"
C_NUBE, C_PUENTE = "planito-nubecita", "planito-puente"

ENV_AWS = {
    "AWS_ACCESS_KEY_ID": "prueba",
    "AWS_SECRET_ACCESS_KEY": "prueba",
    "AWS_DEFAULT_REGION": "us-east-1",
    "AWS_EC2_METADATA_DISABLED": "true",
    "AWS_PAGER": "",
}

INFRA = """{
  "buckets": ["fotos", "copias"],
  "tablas": [
    {"nombre": "pedidos", "particion": "cliente", "orden": "fecha"}
  ],
  "colas": [
    {"nombre": "trabajos", "muertos": "muertos", "maxRecibos": 3},
    {"nombre": "muertos"}
  ]
}
"""

# El mismo fichero, con la clave de la tabla cambiada. Sirve para enseñar el `~` del plan.
INFRA_CLAVE_CAMBIADA = INFRA.replace(
    '{"nombre": "pedidos", "particion": "cliente", "orden": "fecha"}',
    '{"nombre": "pedidos", "particion": "id"}')


def correr(args, entorno=None, timeout=900):
    e = dict(os.environ)
    e.update(entorno or {})
    p = subprocess.run(args, env=e, capture_output=True, text=True,
                       encoding="utf-8", errors="replace", timeout=timeout)
    return {"cmd": " ".join(args), "codigo": p.returncode,
            "salida": ((p.stdout or "") + (p.stderr or "")).rstrip()}


def puerto_libre():
    s = socket.socket()
    s.bind(("127.0.0.1", 0))
    n = s.getsockname()[1]
    s.close()
    return n


def esperar(url, intentos=200):
    for _ in range(intentos):
        try:
            urllib.request.urlopen(url, timeout=1).read()
            return True
        except urllib.error.HTTPError:
            return True
        except Exception:
            time.sleep(0.05)
    return False


def bloque(doc, r, nota=None, falla=False):
    marca = "   # ⚠️ falla A PROPÓSITO" if falla else ""
    doc += ["```console", f"$ {r['cmd']}{marca}", r["salida"],
            f"[codigo de salida: {r['codigo']}]", "```", ""]
    if nota:
        doc += ["> " + nota.replace("\n", "\n> "), ""]
    return doc


def main():
    for x in ("docker", "aws"):
        if not shutil.which(x):
            raise SystemExit(f"[!] no hay {x}")
    EVID.mkdir(parents=True, exist_ok=True)
    fallos = []
    montaje = str(CODIGO).replace("\\", "/")

    def en_go(args):
        return correr(["docker", "run", "--rm", "-v", f"{montaje}:/app",
                       "-v", f"{CACHE_MOD}:/go/pkg/mod", "-w", "/app",
                       "-e", "GOFLAGS=-buildvcs=false", IMAGEN_GO] + args)

    print(f"1 · vet y pruebas, dentro de {IMAGEN_GO}")
    ver_go = en_go(["go", "version"])
    vet = en_go(["go", "vet", "./..."])
    pruebas = en_go(["go", "test", "./...", "-count=1", "-v"])
    carreras = en_go(["go", "test", "./...", "-count=1", "-race"])
    for nombre, r in (("go vet", vet), ("go test", pruebas), ("go test -race", carreras)):
        if r["codigo"] != 0:
            fallos.append(nombre)
    print(f"    vet {vet['codigo']} · test {pruebas['codigo']} · race {carreras['codigo']}")

    print("2 · construir las dos imágenes")
    build_nube = correr(["docker", "build", "-t", IMAGEN_NUBE, str(NUBECITA)])
    if build_nube["codigo"] != 0:
        raise SystemExit(f"[!] no construye nubecita:\n{build_nube['salida']}")
    build = correr(["docker", "build", "-t", IMAGEN, str(CODIGO)])
    if build["codigo"] != 0:
        raise SystemExit(f"[!] no construye planito:\n{build['salida']}")
    tam = int(correr(["docker", "image", "inspect", IMAGEN,
                      "--format", "{{.Size}}"])["salida"].strip())
    print(f"    imagen {IMAGEN}: descarga {tam:,} bytes".replace(",", "."))

    print("3 · levantar la nube y el puente")
    puerto = puerto_libre()
    correr(["docker", "rm", "-f", C_NUBE])
    correr(["docker", "rm", "-f", C_PUENTE])
    correr(["docker", "network", "create", RED])
    correr(["docker", "run", "-d", "--name", C_NUBE, "--network", RED, IMAGEN_NUBE])
    # El puente deja que la CLI de AWS, que corre en Windows, mire la MISMA nube que planito.
    correr(["docker", "run", "-d", "--name", C_PUENTE, "--network", RED,
            "-p", f"{puerto}:4566", "alpine/socat",
            f"TCP-LISTEN:4566,fork,reuseaddr", f"TCP:{C_NUBE}:4566"])
    punta = f"http://127.0.0.1:{puerto}"
    registro = ""
    try:
        if not esperar(punta + "/"):
            raise SystemExit("[!] la nube no arrancó")

        (CODIGO / "infra.json").write_text(INFRA, encoding="utf-8")
        for basura in ("planito.estado.json", "planito.estado.json.tmp"):
            (CODIGO / basura).unlink(missing_ok=True)

        def planito(*args):
            return correr(["docker", "run", "--rm", "--network", RED,
                           "-v", f"{montaje}:/w", "-w", "/w", IMAGEN, *args,
                           "-endpoint", f"http://{C_NUBE}:4566"])

        def cli(*args):
            return correr(["aws", "--endpoint-url", punta, *args], entorno=ENV_AWS)

        print("4 · plan / aplicar / plan / destruir")
        pasos = []

        def paso(r, nota=None, falla=False, limpiar=True):
            if limpiar:
                # Los nombres internos de los contenedores no aportan nada en pantalla: se
                # sustituyen por los que usaría cualquiera. Se hace en el comando Y en la
                # salida, o queda un `nubecita` en uno y un `planito-nubecita` en la otra.
                for texto in ("cmd", "salida"):
                    r[texto] = (r[texto]
                                .replace("docker run --rm --network " + RED + " -v " + montaje +
                                         ":/w -w /w " + IMAGEN, "planito")
                                .replace(f"http://{C_NUBE}:4566", "http://nubecita:4566")
                                .replace(punta, "http://127.0.0.1:4566"))
            if (r["codigo"] != 0) != falla:
                fallos.append(r["cmd"])
                print(f"    [!] {r['cmd']} -> {r['codigo']}")
            pasos.append((r, nota, falla))
            print(f"    · codigo {r['codigo']}  {r['cmd'][:70]}")
            return r

        paso(planito("plan", "-f", "infra.json"),
             "🥇 `plan` no toca nada. Y fíjate en el orden: la cola de muertos se crea **antes** "
             "que la cola que la nombra. Es la única dependencia del modelo, y se resuelve "
             "ordenando.")
        paso(planito("aplicar", "-f", "infra.json"),
             "⚠️ Sin `-si` no hace nada. Una herramienta que toca infraestructura sin "
             "confirmación acaba borrando lo que no debía.")
        paso(planito("aplicar", "-f", "infra.json", "-si"))
        paso(planito("plan", "-f", "infra.json"),
             "🥇 Idempotencia: la segunda pasada no propone nada. Es la propiedad que define a "
             "una herramienta declarativa.")
        paso(cli("s3", "ls"), "Y ahí están, vistos por el cliente oficial de Amazon.")

        # el cambio de clave: el `~` que avisa de que se pierden los datos
        (CODIGO / "infra-clave.json").write_text(INFRA_CLAVE_CAMBIADA, encoding="utf-8")
        paso(planito("plan", "-f", "infra-clave.json"),
             "🥇 La clave de una tabla de DynamoDB no se puede modificar. Cambiarla significa "
             "recrearla, o sea **perder los datos** — y la herramienta lo dice con todas las "
             "letras en vez de hacerlo callando.")

        # alguien crea algo A MANO
        paso(cli("s3", "mb", "s3://de-otro-equipo"),
             "Ahora otra persona crea un bucket **sin usar planito**. Esto es lo que pasa en "
             "cualquier cuenta real.")
        paso(planito("destruir", "-si"),
             "🥇 `destruir` se lleva solo lo que planito creó, y las colas en orden **inverso** "
             "al de creación. Lo de la otra persona ni aparece en el plan.")
        paso(cli("s3", "ls"),
             "🥇 Ahí sigue `de-otro-equipo`. Eso es lo que compra el fichero de estado: sin él, "
             "«lo que sobra» habría incluido lo ajeno.")

        estado = (CODIGO / "planito.estado.json").read_text(encoding="utf-8")
        registro = correr(["docker", "logs", C_NUBE])["salida"]
    finally:
        for c in (C_PUENTE, C_NUBE):
            correr(["docker", "rm", "-f", c])
        correr(["docker", "network", "rm", RED])

    # ── el informe ───────────────────────────────────────────────────────
    fuentes = {f.name: f.read_text(encoding="utf-8") for f in sorted(CODIGO.glob("*.go"))}
    D = []
    A = D.append
    A("# Pruebas ejecutadas — planito, un Terraform de bolsillo en Go")
    A("")
    A("> Salida **literal** de esta máquina. La genera `investigacion/pruebas_planito.py` y se")
    A("> vuelve a generar sola. ⛔ Lo que no esté aquí no puede salir en pantalla.")
    A("")
    A(f"**Entorno citable:** `{ver_go['salida']}` (dentro de `{IMAGEN_GO}`) · "
      "`aws-sdk-go-v2` para `s3`, `dynamodb` y `sqs`")
    A("")
    A(f"**El código vive en `{CODIGO}`** y habla con `nubecita` (`{NUBECITA}`) por una red de")
    A("Docker. La demo entera es gratis y se destruye sola.")
    A("")
    A(f"**La imagen de planito se descarga en {tam:,} bytes**".replace(",", ".") +
      f" ({tam/1048576:.2f} MiB). Segunda etapa `scratch`, igual que nubecita.")
    A("")
    A("---")
    A("")
    A("## 1 · `go vet` y las pruebas del corazón")
    A("")
    A("Fíjate en lo que **no** hay en estas pruebas: red, contenedores, credenciales ni esperas.")
    A("`Calcular` recibe tres estructuras y devuelve una cuarta, porque toda la red vive en")
    A("`nube.go`. Por eso la lógica que de verdad se puede equivocar se prueba en milisegundos.")
    A("")
    D[:] = bloque(D, vet)
    D[:] = bloque(D, pruebas)
    A("### Con el detector de carreras")
    A("")
    D[:] = bloque(D, carreras)
    A("---")
    A("")
    A("## 2 · La herramienta, de principio a fin")
    A("")
    A("El fichero de entrada:")
    A("")
    A("```json")
    A(INFRA.rstrip())
    A("```")
    A("")
    for r, nota, falla in pasos:
        D[:] = bloque(D, r, nota, falla)
    A("---")
    A("")
    A("## 3 · El fichero de estado que quedó")
    A("")
    A("Es la lista de «esto lo hice yo», y es lo **único** que autoriza a borrar. Se escribe de")
    A("forma atómica (temporal + renombrado) y después de CADA paso, no al final: si el proceso")
    A("muere a mitad, lo ya creado tiene que estar apuntado o nadie lo borraría nunca.")
    A("")
    A("```json")
    A(estado.rstrip())
    A("```")
    A("")
    A("---")
    A("")
    A("## 4 · Lo que vio la nube")
    A("")
    A("El registro de `nubecita` durante toda la demo: se ve a planito leyendo el estado real")
    A("(los tres `List...`) antes de cada plan, y luego creando y borrando.")
    A("")
    A("```console")
    A(registro)
    A("```")
    A("")
    A("---")
    A("")
    A("## 5 · El código completo")
    A("")
    for nombre, texto in fuentes.items():
        A(f"### `{nombre}`")
        A("")
        A("```go")
        A(texto.rstrip())
        A("```")
        A("")
    A("### `Dockerfile`")
    A("")
    A("```dockerfile")
    A((CODIGO / "Dockerfile").read_text(encoding="utf-8").rstrip())
    A("```")
    A("")

    destino = EVID / ("SALIDA.md" if not fallos else "SALIDA.FALLIDA.md")
    destino.write_text("\n".join(D), encoding="utf-8")
    if fallos:
        print("\n[!] hubo fallos: " + ", ".join(fallos))
        print(f"    -> {destino} (el SALIDA.md bueno NO se ha tocado)")
        sys.exit(1)
    print(f"\n✅ {destino}")


if __name__ == "__main__":
    main()
