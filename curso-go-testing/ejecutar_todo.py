# -*- coding: utf-8 -*-
"""Corre las 12 lecciones del curso y regenera SALIDA.md. Nada de lo que sale en el vídeo se
escribe a mano: sale de aquí.

Todo corre DENTRO de un contenedor golang:1.27.2 (no hace falta Go en tu máquina, solo Docker y
Python 3.10+). Los tests de testcontainers arrancan su Postgres hablando con TU Docker a través
del socket montado (-v /var/run/docker.sock), y llegan a él por host.docker.internal.

    python ejecutar_todo.py                 # todo (unos 6-8 minutos)
    python ejecutar_todo.py 05 07           # solo esas lecciones
    python ejecutar_todo.py --sin-docker    # usa el Go de tu máquina (Linux o macOS)

Variables opcionales:
    CURSO_PREFIJO     prefijo del nombre del contenedor (por defecto curso-go-testing)
    CURSO_GOMODCACHE  carpeta de módulos de Go del host a montar (si no, se descargan dentro del
                      contenedor y se van con él)

Limpieza: no deja nada en tu Docker. El contenedor de Go va con --rm y sin volúmenes con nombre;
los Postgres de testcontainers los borra Terminate o Ryuk junto con su volumen; el de
WithReuseByName se borra al final con `docker rm -f -v`. Si cortas el guion a mitad:
    docker rm -f -v curso-go-testing-ejecutor curso-go-testing-pg-reuso

Las demos que ROMPEN algo (los golden con un fallo, el fuzzing, el renombrado de la lección 11)
se hacen siempre en una copia del repo dentro de un directorio temporal (mktemp -d).
"""
import os
import platform
import re
import subprocess
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN_GO = "golang:1.27.2"
IMAGEN_PG = "postgres:18.6-alpine"
PREFIJO = os.environ.get("CURSO_PREFIJO", "curso-go-testing")
EJECUTOR = f"{PREFIJO}-ejecutor"
REUSO = "curso-go-testing-pg-reuso"  # el nombre que usa 05_testcontainers/trampa_test.go
SIN_DOCKER = "--sin-docker" in sys.argv
# testcontainers imprime su diagnóstico en cada arranque: para SALIDA.md sobra
RUIDO = re.compile(r"(🐳|✅|🔔|⏳|🚫|Connected to docker|^\s+(Server Version|API Version|"
                   r"Operating System|Total Memory|Labels|com\.docker|Testcontainers for Go|"
                   r"Resolved Docker|Test SessionID|Test ProcessID)|Shell not found)")
# copia del repo en un temporal, para las demos que cambian archivos
COPIA = 'T=$(mktemp -d) && cp -r . "$T" && cd "$T" && '


def sh(cmd, timeout=1800, env=None):
    """Ejecuta una línea de shell en el contenedor (o en local con --sin-docker)."""
    if SIN_DOCKER:
        full = ["bash", "-o", "pipefail", "-c", cmd]
        cwd = AQUI
    else:
        e = []
        for k, v in (env or {}).items():
            e += ["-e", f"{k}={v}"]
        full = ["docker", "exec", *e, "-w", "/src", EJECUTOR, "bash", "-o", "pipefail", "-c", cmd]
        cwd = None
    t0 = time.time()
    r = subprocess.run(full, capture_output=True, text=True, encoding="utf-8", errors="replace",
                       timeout=timeout, cwd=cwd, env={**os.environ, **(env or {})} if SIN_DOCKER else None)
    salida = "\n".join(l for l in (r.stdout + r.stderr).splitlines() if not RUIDO.search(l))
    return r.returncode, salida.strip("\n"), time.time() - t0


def limpiar():
    """Borra el contenedor de Go y el reutilizable, con sus volúmenes anónimos (-v)."""
    subprocess.run(["docker", "rm", "-f", "-v", EJECUTOR, REUSO], capture_output=True)


def levantar():
    limpiar()
    for img in (IMAGEN_GO, IMAGEN_PG):  # descargadas ANTES de medir: los tiempos no incluyen el pull
        subprocess.run(["docker", "pull", "-q", img], check=True, capture_output=True)
    # --rm y ningún volumen con nombre: la caché de módulos y de compilación vive en la capa del
    # contenedor y se va con él (salvo CURSO_GOMODCACHE, que es una carpeta TUYA del host)
    mod = os.environ.get("CURSO_GOMODCACHE")
    cmd = ["docker", "run", "-d", "--rm", "--name", EJECUTOR, "--cpus", "4",
           "-v", f"{AQUI}:/src", *(["-v", f"{mod}:/go/pkg/mod"] if mod else []),
           "-v", "/var/run/docker.sock:/var/run/docker.sock",
           "--add-host", "host.docker.internal:host-gateway",
           "-e", "TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal",
           "-e", "GOTOOLCHAIN=local", "-e", "GOFLAGS=-mod=mod",
           IMAGEN_GO, "sleep", "infinity"]
    subprocess.run(cmd, check=True, capture_output=True,
                   env={**os.environ, "MSYS_NO_PATHCONV": "1"})


partes = []
fallos = []


# Si la máquina falla (disco lleno, Docker caído), una trampa «falla como se esperaba» por el
# motivo equivocado. Estas salidas invalidan el paso, dé el código que dé.
AVERIA = re.compile(r"read-only file system|no space left|mktemp: failed|Cannot connect to the Docker|"
                    r"Error response from daemon", re.I)


def paso(titulo, cmd, esperado=0, visible=None, cabeza=None, env=None, timeout=1800, debe=None):
    """Un paso: el comando, su código de salida y su salida literal.
    esperado: el código que tiene que dar (las trampas fallan A PROPÓSITO).
    debe: un texto que TIENE que salir (que el fallo sea el de la trampa y no otro)."""
    codigo, salida, seg = sh(cmd, timeout=timeout, env=env)
    averia = AVERIA.search(salida)
    falta = debe and debe not in salida
    if cabeza:
        lineas = salida.splitlines()
        salida = "\n".join(lineas[:cabeza] + ([f"... ({len(lineas) - cabeza} líneas más)"] if len(lineas) > cabeza else []))
    ok = codigo == esperado and not averia and not falta
    if not ok:
        motivo = ("avería de la máquina: " + averia.group(0)) if averia else (
            f"falta «{debe}» en la salida" if falta else f"código {codigo}, esperaba {esperado}")
        fallos.append(f"{titulo}: {motivo}")
    envs = " ".join(f"{k}={v}" for k, v in (env or {}).items())
    partes.extend([f"### {titulo}", "",
                   f"`$ {(envs + ' ') if envs else ''}{visible or cmd}`", "",
                   f"código de salida `{codigo}` (esperado `{esperado}`){'' if ok else ' ⚠️'} · {seg:.1f} s de reloj", "",
                   "```", salida, "```", ""])
    print(f"    {'ok ' if ok else 'MAL'} {titulo} · {seg:.1f} s", flush=True)
    return codigo, salida, seg


def leccion(n, titulo):
    partes.extend([f"## Lección {n} — {titulo}", ""])
    print(f"  [{n}] {titulo}", flush=True)


SIN_RUN = "grep -v '^=== '"


def l01():
    leccion("01", "Tests de tabla, subtests y t.Parallel")
    paso("Tabla con subtests paralelos", f"go test -count=1 -v -run TestTotal ./01_tabla | {SIN_RUN}")
    paso("La trampa: un fake compartido, 50 veces",
         "go test -tags trampa -count=50 -json -run TestCompartido ./01_tabla | go run ./12_ci/resumen -max-fallos 1 -lentos 0", 1, debe="se cobró")
    paso("El mismo test, de uno en uno (-parallel 1)",
         "go test -tags trampa -count=50 -parallel 1 -json -run TestCompartido ./01_tabla | go run ./12_ci/resumen -lentos 0")
    paso("Con -race", "go test -tags trampa -count=1 -race -run TestCompartido ./01_tabla", 1, cabeza=12, debe="DATA RACE")
    paso("El arreglo: un fake por subtest, 50 veces y con -race",
         "go test -tags trampa -count=50 -race -json -run TestPorSubtest ./01_tabla | go run ./12_ci/resumen -lentos 0")
    paso("go.mod con «go 1.21»: la variable del for compartida",
         f"cd 01_tabla/go121 && go test -count=1 -v . | {SIN_RUN}")
    paso("El mismo test con «go 1.22»",
         COPIA + f"sed -i 's/^go 1.21$/go 1.22/' 01_tabla/go121/go.mod && cd 01_tabla/go121 && go test -count=1 -v . | {SIN_RUN}",
         1, visible="(copia temporal) sed -i 's/^go 1.21$/go 1.22/' go.mod && go test -v .", debe="Abs(-3) = -3")


def l02():
    leccion("02", "t.Cleanup, t.TempDir, t.Setenv y t.Context")
    paso("Los ayudantes, bien usados", f"go test -count=1 -v ./02_ayudantes | {SIN_RUN}")
    paso("Las trampas: defer con subtests paralelos y t.Context en un Cleanup",
         f"go test -tags trampa -count=1 -v -run 'TestDefer|TestContextCancelado' ./02_ayudantes | {SIN_RUN}", 1, debe="file already closed")
    paso("t.Setenv con t.Parallel",
         "go test -tags trampa -count=1 -run TestSetenvParalelo ./02_ayudantes", 1, cabeza=3, debe="can not use t.Parallel")


def l03():
    leccion("03", "Fakes frente a mocks: lo que rompe un refactor")
    for etiqueta, desc, esperado in (("", "antes", 0), ("refactor", "refactor sin cambio de comportamiento", 1),
                                     ("conbug", "con un fallo de verdad", 1)):
        tags = f"-tags {etiqueta} " if etiqueta else ""
        paso(f"Las 12 pruebas contra la versión «{desc}»",
             f"go test {tags}-count=1 -json ./03_fakes_mocks | go run ./12_ci/resumen -lentos 0 -max-fallos 12", esperado,
             debe="12 tests")


def l04():
    leccion("04", "httptest para servidores y clientes")
    paso("El handler, con NewRecorder", f"go test -count=1 -v ./04_httptest | {SIN_RUN}")
    paso("El servidor entero, el cliente sin timeout y el Body sin cerrar",
         f"go test -tags trampa -count=1 -v ./04_httptest | {SIN_RUN}", 1, debe="404")
    paso("La API y el cliente de cobros de la tienda, con httptest.NewServer",
         "go test -count=1 -json ./tienda/api ./tienda/cobros | go run ./12_ci/resumen -lentos 3")


def esperar_que_desaparezca(nombre, limite=90):
    t0 = time.time()
    while time.time() - t0 < limite:
        r = subprocess.run(["docker", "ps", "-a", "-q", "--filter", f"name=^{nombre}$"],
                           capture_output=True, text=True)
        if not r.stdout.strip():
            return time.time() - t0
        time.sleep(2)
    return None


def l05():
    leccion("05", "testcontainers-go con Postgres de verdad")
    paso("Un contenedor para el paquete, Snapshot y Restore por test",
         f"go test -tags integracion -count=1 -v ./05_testcontainers | {SIN_RUN}")
    paso("La hora: fake contra Postgres, y lo que cuesta un contenedor por test",
         f"go test -tags 'integracion trampa' -count=1 -v -run 'TestHora|TestTresContenedores' ./05_testcontainers | {SIN_RUN}", 1, debe="iguales=false")
    reuso = f"go test -tags 'integracion trampa' -count=1 -v -run TestReuso ./05_testcontainers | grep -E 'Reuse|ok|FAIL'"
    paso("WithReuseByName, primera vez", reuso)
    paso("WithReuseByName, otra vez enseguida", reuso)
    seg = esperar_que_desaparezca(REUSO)
    partes.extend(["```", f"docker ps -a --filter name={REUSO}  ->  "
                   + (f"desaparecido a los {seg:.0f} s (lo borró Ryuk)" if seg is not None else "sigue ahí a los 90 s"),
                   "```", ""])
    paso("WithReuseByName, después de esperar", reuso)
    esperar_que_desaparezca(REUSO)
    sin_ryuk = {"TESTCONTAINERS_RYUK_DISABLED": "true"}
    paso("Sin Ryuk, primera vez", reuso, env=sin_ryuk)
    time.sleep(30)
    paso("Sin Ryuk, 30 s después", reuso, env=sin_ryuk)
    # sin Ryuk, lo borras tú, y con -v: sin él se queda su volumen anónimo de /var/lib/postgresql
    subprocess.run(["docker", "rm", "-f", "-v", REUSO], capture_output=True)
    queda = subprocess.run(["docker", "ps", "-a", "-q", "--filter", f"name=^{REUSO}$"],
                           capture_output=True, text=True).stdout.strip()
    partes.extend(["```", f"docker rm -f -v {REUSO}  ->  " + ("sigue ahí" if queda else "borrado, con su volumen"),
                   "```", ""])


def l06():
    leccion("06", "Golden files con -update")
    paso("Los recibos contra testdata/*.golden", f"go test -count=1 -v ./06_golden | {SIN_RUN}")
    bug = "sed -i 's/\"Descuento primer pedido\", -d)/\"Descuento primer pedido\", d)/' tienda/pedido.go && "
    paso("Un cambio en Recibo (el descuento pierde el signo)",
         COPIA + bug + "go test -count=1 ./06_golden", 1, debe="no coincide",
         visible="(copia temporal, con el signo del descuento quitado) go test ./06_golden")
    paso("El mismo cambio, con -update",
         COPIA + bug + "go test -count=1 ./06_golden -update && go test -count=1 ./06_golden && "
         "diff -u \"$OLDPWD/06_golden/testdata/primer_pedido.golden\" 06_golden/testdata/primer_pedido.golden",
         1, visible="(copia temporal) go test ./06_golden -update && go test ./06_golden && diff -u <antes> <después>",
         debe="+Descuento")
    paso("El golden con finales de línea CRLF (lo que hace git en Windows sin .gitattributes)",
         COPIA + "sed -i 's/$/\\r/' 06_golden/testdata/primer_pedido.golden && go test -count=1 -run 'TestRecibo/primer' ./06_golden",
         1, visible="(copia temporal) sed -i 's/$/\\r/' primer_pedido.golden && go test -run TestRecibo/primer ./06_golden",
         cabeza=12, debe="no coincide")


def l07():
    leccion("07", "Fuzzing que encuentra un fallo real")
    paso("Las pruebas de tabla de la versión original", f"go test -count=1 -v ./07_fuzzing | {SIN_RUN}")
    paso("Fuzzing de la propiedad «ida y vuelta»",
         COPIA + "go test -fuzz=FuzzIdaYVuelta -fuzztime=60s ./07_fuzzing; "
         "echo '--- testdata/fuzz/FuzzIdaYVuelta/'; cat 07_fuzzing/testdata/fuzz/FuzzIdaYVuelta/*; "
         "echo '--- y ahora, go test sin -fuzz:'; go test -count=1 ./07_fuzzing",
         1, visible="(copia temporal) go test -fuzz=FuzzIdaYVuelta -fuzztime=60s ./07_fuzzing; cat testdata/...; go test ./07_fuzzing",
         debe="go test fuzz v1")
    paso("Arreglado Formatear (-tags paso2), otra vez",
         COPIA + "go test -tags paso2 -fuzz=FuzzIdaYVuelta -fuzztime=60s ./07_fuzzing",
         1, visible="(copia temporal) go test -tags paso2 -fuzz=FuzzIdaYVuelta -fuzztime=60s ./07_fuzzing",
         debe="ida y vuelta")
    paso("La versión de tienda/, 30 segundos de fuzzing",
         COPIA + "go test -fuzz=FuzzMontoIdaYVuelta -fuzztime=30s ./tienda | tail -3",
         visible="(copia temporal) go test -fuzz=FuzzMontoIdaYVuelta -fuzztime=30s ./tienda", debe="PASS")


def l08():
    leccion("08", "testing/synctest para código con tiempo")
    paso("Reintentos dentro de la burbuja", f"go test -count=1 -v ./08_synctest | {SIN_RUN}")
    paso("Con reloj de verdad, y la versión con time.Sleep",
         f"go test -tags trampa -count=1 -v -run 'Verdad|Sleep' ./08_synctest | {SIN_RUN}", 1, debe="sale a los 3s")


def l09():
    leccion("09", "Cobertura: -coverprofile, -coverpkg y el 100 % que miente")
    paso("100 % de cobertura", "go test -count=1 -cover ./09_cobertura")
    paso("Por función", "go test -count=1 -coverprofile=/tmp/c09.out ./09_cobertura >/dev/null && go tool cover -func=/tmp/c09.out")
    paso("El caso de la frontera", "go test -tags trampa -count=1 ./09_cobertura", 1, debe="Envio(50000)")
    paso("Etiqueta rota, y el test sin aserciones sigue en verde",
         COPIA + "sed -i 's/return \"envío 49,00\"/return \"envío gratis\"/' 09_cobertura/envio.go && go test -count=1 -cover ./09_cobertura",
         visible="(copia temporal, Etiqueta siempre dice «envío gratis») go test -cover ./09_cobertura", debe="100.0%")
    paso("La cobertura de tienda/ con sus propios tests", "go test -count=1 -cover ./tienda")
    paso("La de tienda/ con TODOS los tests del repo (-coverpkg)",
         "go test -count=1 -coverpkg=./tienda/... -coverprofile=/tmp/todo.out ./... >/dev/null && "
         "go tool cover -func=/tmp/todo.out | grep -E 'tienda/(monto|pedido).go|total'")


def l10():
    leccion("10", "Tests flaky: -count, -race y -shuffle")
    paso("El orden de un map, 100 veces",
         "go test -tags trampa -count=100 -json -run 'TestSKUs$' ./10_flaky | go run ./12_ci/resumen -max-fallos 1 -lentos 0", 1)
    paso("Estado global, en el orden del archivo",
         "go test -tags trampa -count=1 -v -run 'TestMoneda|TestCambia' ./10_flaky | grep -E '^(---|ok|FAIL)'")
    paso("Estado global, barajado con 10 semillas",
         "for s in 1 2 3 4 5 6 7 8 9 10; do go test -tags trampa -count=1 -shuffle=$s -run 'TestMoneda|TestCambia' ./10_flaky >/dev/null "
         "&& echo \"-shuffle=$s  pasa\" || echo \"-shuffle=$s  FALLA\"; done")
    paso("Una semilla que falla, reproducida",
         "go test -tags trampa -count=1 -v -shuffle=2 -run 'TestMoneda|TestCambia' ./10_flaky | grep -E 'shuffle|^(---|ok|FAIL)|got'", 1, debe="got")
    paso("La carrera de datos, 100 veces sin -race",
         "go test -tags trampa -count=100 -json -run 'TestVisitas$' ./10_flaky | go run ./12_ci/resumen -max-fallos 1 -lentos 0", 1)
    paso("La carrera de datos, una vez con -race",
         "go test -tags trampa -count=1 -race -run 'TestVisitas$' ./10_flaky", 1, cabeza=8, debe="DATA RACE")
    paso("Las versiones arregladas: -race, -shuffle=on y -count=100",
         "go test -count=100 -race -shuffle=on -json ./10_flaky | go run ./12_ci/resumen -lentos 0")


def l11():
    leccion("11", "Build tags para separar los de integración")
    paso("go test ./... (sin etiquetas)", "go test -count=1 -json ./... | go run ./12_ci/resumen -lentos 3")
    paso("go test -tags integracion ./...", "go test -tags integracion -count=1 -json ./... | go run ./12_ci/resumen -lentos 3")
    renombre = "sed -i 's/PagadosDe(/ContarPagados(/g' 11_etiquetas/informe.go 11_etiquetas/informe_test.go && "
    paso("Renombrar una función sin la etiqueta: todo verde",
         COPIA + renombre + "go vet ./... && go test -count=1 ./11_etiquetas",
         visible="(copia temporal, PagadosDe renombrada en informe.go e informe_test.go) go vet ./... && go test ./11_etiquetas")
    paso("Y con la etiqueta",
         COPIA + renombre + "go vet -tags integracion ./11_etiquetas", 1,
         visible="(copia temporal, el mismo renombrado) go vet -tags integracion ./11_etiquetas", debe="undefined")


def l12():
    leccion("12", "go test -json en CI")
    paso("Toda la suite, con integración, -race y -shuffle, por el resumen",
         "go test -tags integracion -race -shuffle=on -count=1 -json ./... | go run ./12_ci/resumen -lentos 5")
    paso("La trampa de la tubería: un test que falla, y el código de salida de la tubería",
         "set +o pipefail; go test -tags trampa -count=1 -json ./09_cobertura | cat > /dev/null; echo \"código de la tubería: $?\"",
         visible="go test -json ./09_cobertura | cat > /dev/null; echo $?   (sh, sin pipefail)", debe="tubería: 0")
    paso("Lo mismo con set -o pipefail",
         "set -o pipefail; go test -tags trampa -count=1 -json ./09_cobertura | cat > /dev/null; "
         "echo \"código de la tubería: $?\"", visible="set -o pipefail; go test -json ./09_cobertura | cat > /dev/null; echo $?", debe="tubería: 1")


LECCIONES = {"01": l01, "02": l02, "03": l03, "04": l04, "05": l05, "06": l06,
             "07": l07, "08": l08, "09": l09, "10": l10, "11": l11, "12": l12}


def main():
    pedidas = [a for a in sys.argv[1:] if a in LECCIONES] or list(LECCIONES)
    if not SIN_DOCKER:
        print("  levantando el contenedor de Go ...", flush=True)
        levantar()
    _, gov, _ = sh("go version")
    _, mods, _ = sh("go mod download && go list -m github.com/testcontainers/testcontainers-go "
                    "github.com/jackc/pgx/v5 go.uber.org/mock")
    dk = subprocess.run(["docker", "version", "--format", "{{.Server.Version}}"],
                        capture_output=True, text=True).stdout.strip()
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    cab = ["# SALIDA — curso «Testing en Go para producción»", "",
           "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
           f"- fecha: `{sello}`", f"- {gov.strip()}" + ("" if SIN_DOCKER else f" (contenedor `{IMAGEN_GO}`, --cpus 4)"),
           f"- Docker `{dk}` · Postgres `{IMAGEN_PG}` (descargadas antes de medir)",
           *[f"- `{m}`" for m in mods.splitlines() if m.strip()],
           f"- anfitrión: `{platform.system()} {platform.release()}`",
           "- testcontainers desde DENTRO del contenedor: socket montado (`-v /var/run/docker.sock`) y "
           "`TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal`", "",
           "Las trampas fallan a propósito: cada paso dice el código de salida que se esperaba.", ""]
    try:
        for n in pedidas:
            LECCIONES[n]()
    finally:
        limpiar()
    nombre = "SALIDA.md" if len(pedidas) == len(LECCIONES) else "SALIDA_parcial.md"
    resumen = [f"**{len(fallos)} pasos con un código distinto del esperado.**", ""] + [f"- {f}" for f in fallos]
    (AQUI / nombre).write_text("\n".join(cab + resumen + [""] + partes), encoding="utf-8", newline="\n")
    print(f"\n  -> {nombre} ({len(pedidas)} lecciones, {len(fallos)} pasos inesperados)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
