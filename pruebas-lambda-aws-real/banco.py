# -*- coding: utf-8 -*-
"""Banco del vídeo 882 (golandia): la revancha del 121, ahora en AWS Lambda DE VERDAD.

El 121 midió con el emulador RIE en local. Aquí se mide en la cuenta del dueño (us-east-1),
con el mismo handler, en x86_64 y 512 MB, y TODO lo que se crea lleva el prefijo `gl-lambda-`.

Uso (en este orden; cada paso deja su JSON en _datos/ y es reanudable):
    python banco.py construir [variante ...]   # compila en Docker (candado _logs/banco.lock)
    python banco.py medir [variante ...]       # crea funciones, comprueba, mide en frío y caliente
    python banco.py memoria [variante ...]     # el mismo trabajo con 128 y 1024 MB (tras medir)
    python banco.py limpiar                    # borra TODO lo gl-lambda-* y comprueba que no queda
    python banco.py informe                    # escribe SALIDA.md con todo lo anterior (literal)

⛔ Ni credenciales ni el número de cuenta salen a ningún archivo: se tachan antes de escribir.
"""
import base64
import concurrent.futures as cf
import datetime as dt
import io
import json
import math
import os
import re
import shutil
import statistics as st
import subprocess
import sys
import time
import zipfile
from pathlib import Path

AQUI = Path(__file__).resolve().parent
SRC = AQUI / "src"
PAQ = AQUI / "_paquetes"
DATOS = AQUI / "_datos"
RAIZ_FABRICA = Path(r"<proyecto>")
CANDADO = RAIZ_FABRICA / "_logs" / "banco.lock"

REGION = "us-east-1"
PREFIJO = "gl-lambda-"
ROL = "gl-lambda-rol"
POLITICA = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
MEMORIA = 512
TIMEOUT = 15
ARQ = "x86_64"

N_FRIOS = int(os.environ.get("N_FRIOS", "25"))
N_CALIENTE = int(os.environ.get("N_CALIENTE", "200"))
N_IO = int(os.environ.get("N_IO", "100"))
N_CPU = 200000
REFERENCIA = sum(i * i % 1000003 for i in range(1, N_CPU + 1))

# variante -> cómo se construye y cómo se despliega
VARIANTES = {
    "go":          dict(nombre="Go",                 runtime="provided.al2023", handler="bootstrap"),
    "rust":        dict(nombre="Rust",               runtime="provided.al2023", handler="bootstrap"),
    "dotnet-aot":  dict(nombre=".NET Native AOT",    runtime="provided.al2023", handler="bootstrap"),
    "java-native": dict(nombre="Java GraalVM native", runtime="provided.al2023", handler="bootstrap"),
    "java":        dict(nombre="Java (JVM)",         runtime="java25",          handler="golandia.Manejador::manejar"),
    "java-snap":   dict(nombre="Java + SnapStart",   runtime="java25",          handler="golandia.Manejador::manejar", paquete="java"),
    "python":      dict(nombre="Python",             runtime="python3.14",      handler="app.handler"),
    "node":        dict(nombre="Node",               runtime="nodejs24.x",      handler="app.handler"),
    # comprobación aparte: el MISMO binario de Rust con un archivo de relleno aleatorio hasta pesar lo que el .zip de Go
    "rust-relleno": dict(nombre="Rust + relleno", runtime="provided.al2023", handler="bootstrap"),
}
EXTRAS = ["rust-relleno"]
ORDEN = [v for v in VARIANTES if v not in EXTRAS]

IMAGENES = {
    "go": "golang:1.27.2",
    "rust": "rust:1.99.1-trixie",
    "dotnet-aot": "public.ecr.aws/sam/build-dotnet10:latest-x86_64",
    "java-native": "ghcr.io/graalvm/native-image-community:25",
    "java": "eclipse-temurin:25-jdk",
}

CUENTA = None  # se rellena al conectar; solo para tacharla


def tachar(texto):
    s = str(texto)
    if CUENTA:
        s = s.replace(CUENTA, "************")
    return re.sub(r"(?<!\d)\d{12}(?!\d)", "************", s)


def guardar(nombre, obj):
    DATOS.mkdir(exist_ok=True)
    (DATOS / nombre).write_text(tachar(json.dumps(obj, ensure_ascii=False, indent=1)), encoding="utf-8")


def cargar(nombre, defecto=None):
    p = DATOS / nombre
    return json.loads(p.read_text(encoding="utf-8")) if p.exists() else defecto


def log(*a):
    print(f"[{dt.datetime.now():%H:%M:%S}]", *a, flush=True)


# ───────────────────────────── construir ─────────────────────────────

def tomar_candado():
    """Espera a que nadie mida en local y toma el candado (las compilaciones se cronometran)."""
    avisado = False
    while True:
        try:
            CANDADO.mkdir()
            (CANDADO / "dueño.txt").write_text(f"banco 882 gl-lambda compilaciones pid {os.getpid()}", encoding="utf-8")
            return
        except FileExistsError:
            if not avisado:
                log("candado ocupado; espero a que se libere")
                avisado = True
            time.sleep(5)


def soltar_candado():
    try:
        (CANDADO / "dueño.txt").unlink(missing_ok=True)
        CANDADO.rmdir()
    except OSError as e:
        log("no pude soltar el candado:", e)


def docker(var, imagen, script, montar_src):
    nombre = f"gl-lambda-build-{var}"
    out = PAQ / var
    if out.exists():
        shutil.rmtree(out)
    out.mkdir(parents=True)
    cmd = ["docker", "run", "--rm", "--ulimit", "core=1", "--name", nombre, "--platform", "linux/amd64",
           "-v", f"{montar_src}:/src", "-v", f"{out}:/out", "--entrypoint", "sh", imagen, "-c", script]
    t0 = time.perf_counter()
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")
    seg = time.perf_counter() - t0
    (PAQ / f"{var}.log").write_text(r.stdout + "\n--- stderr ---\n" + r.stderr, encoding="utf-8")
    if r.returncode != 0:
        raise SystemExit(f"falló la compilación de {var} (ver _paquetes/{var}.log):\n{r.stderr[-3000:]}")
    return seg, r.stdout


def hacer_zip(var, archivos):
    """archivos: lista de (ruta_en_zip, Path, ejecutable)"""
    z = PAQ / f"{var}.zip"
    with zipfile.ZipFile(z, "w", zipfile.ZIP_DEFLATED) as f:
        for nombre, ruta, ejecutable in archivos:
            zi = zipfile.ZipInfo(nombre, date_time=(2026, 1, 1, 0, 0, 0))
            zi.compress_type = zipfile.ZIP_DEFLATED
            zi.create_system = 3  # Unix: si no, Lambda ignora el bit de ejecución de bootstrap
            zi.external_attr = ((0o755 if ejecutable else 0o644) | 0o100000) << 16
            f.writestr(zi, ruta.read_bytes())
    return z


def construir(vars_, archivo="construir.json"):
    PAQ.mkdir(parents=True, exist_ok=True)
    res = cargar(archivo, {})
    pesados = [v for v in vars_ if v in IMAGENES]
    if pesados:
        libre = shutil.disk_usage("C:\\").free / 1e9
        if libre < 50:
            raise SystemExit(f"C: tiene {libre:.0f} GB libres (< 50): paro")
        tomar_candado()
    try:
        for var in vars_:
            log("construyendo", var)
            r = {"imagen": IMAGENES.get(var)}
            if var == "go":
                seg, salida = docker(var, IMAGENES[var],
                                     "set -e; cp -r /src/go /tmp/w; cd /tmp/w; go version; "
                                     "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -trimpath "
                                     "-ldflags='-s -w' -o /out/bootstrap .; grep aws-lambda-go go.mod", SRC)
                archivos = [("bootstrap", PAQ / var / "bootstrap", True)]
            elif var == "rust":
                seg, salida = docker(var, IMAGENES[var],
                                     "set -e; cp -r /src/rust /tmp/w; cd /tmp/w; "
                                     "rustup target add x86_64-unknown-linux-musl >/dev/null 2>&1; "
                                     "if ! grep -q lambda_runtime Cargo.toml; then "
                                     "cargo add lambda_runtime >/dev/null 2>&1; cargo add serde --features derive >/dev/null 2>&1; "
                                     "cargo add tokio --features macros,rt-multi-thread,time >/dev/null 2>&1; fi; "
                                     "rustc --version; cargo --version; "
                                     "cargo build --release --target x86_64-unknown-linux-musl 2>&1 | tail -n 3; "
                                     "cp target/x86_64-unknown-linux-musl/release/bootstrap /out/; "
                                     "cp Cargo.toml Cargo.lock /src/rust/; "
                                     "grep -A1 -E '^name = \"(lambda_runtime|tokio|serde)\"$' Cargo.lock", SRC)
                archivos = [("bootstrap", PAQ / var / "bootstrap", True)]
            elif var == "dotnet-aot":
                seg, salida = docker(var, IMAGENES[var],
                                     "set -e; cp -r /src/dotnet /tmp/w; cd /tmp/w; dotnet --version; "
                                     "cat /etc/os-release | grep PRETTY; "
                                     "dotnet publish -c Release -r linux-x64 -o /tmp/pub 2>&1 | tail -n 3; "
                                     "cp /tmp/pub/bootstrap /out/; "
                                     "dotnet list package | grep Amazon", SRC)
                archivos = [("bootstrap", PAQ / var / "bootstrap", True)]
            elif var == "java-native":
                seg, salida = docker(var, IMAGENES[var],
                                     "set -e; mkdir -p /tmp/c; java -version 2>&1; native-image --version; "
                                     "javac --release 25 -d /tmp/c /src/java/golandia/*.java; "
                                     "cd /tmp && native-image --no-fallback --enable-url-protocols=http -cp /tmp/c "
                                     "golandia.Bootstrap -o /out/bootstrap 2>&1 | tail -n 4", SRC)
                archivos = [("bootstrap", PAQ / var / "bootstrap", True)]
            elif var == "java":
                seg, salida = docker(var, IMAGENES[var],
                                     "set -e; mkdir -p /out/c; java -version 2>&1; "
                                     "javac --release 25 -d /out/c /src/java/golandia/Manejador.java", SRC)
                archivos = [(p.relative_to(PAQ / var / "c").as_posix(), p, False)
                            for p in sorted((PAQ / var / "c").rglob("*.class"))]
            elif var == "python":
                seg, salida = 0.0, ""
                archivos = [("app.py", SRC / "python" / "app.py", False)]
            elif var == "node":
                seg, salida = 0.0, ""
                archivos = [("app.mjs", SRC / "node" / "app.mjs", False)]
            else:
                continue
            z = hacer_zip(var, archivos)
            r.update(segundos=round(seg, 1), salida=salida.strip(),
                     sin_comprimir=sum(a[1].stat().st_size for a in archivos), zip=z.stat().st_size)
            res[var] = r
            guardar(archivo, res)
            log(var, f"{seg:.1f} s", r["sin_comprimir"], "B ->", r["zip"], "B")
    finally:
        if pesados:
            soltar_candado()


# ───────────────────────────── AWS ─────────────────────────────

def sesion():
    global CUENTA
    import boto3
    s = boto3.Session(region_name=REGION)
    CUENTA = s.client("sts").get_caller_identity()["Account"]
    return s


def asegurar_rol(s):
    iam = s.client("iam")
    confianza = {"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"},
                                                          "Action": "sts:AssumeRole"}]}
    try:
        arn = iam.get_role(RoleName=ROL)["Role"]["Arn"]
    except iam.exceptions.NoSuchEntityException:
        arn = iam.create_role(RoleName=ROL, AssumeRolePolicyDocument=json.dumps(confianza),
                              Description="banco golandia 882; se borra al terminar",
                              Tags=[{"Key": "proyecto", "Value": "gl-lambda"}])["Role"]["Arn"]
        iam.attach_role_policy(RoleName=ROL, PolicyArn=POLITICA)
        log("rol creado; espero 12 s a que IAM lo propague")
        time.sleep(12)
    return arn


def nombre_fn(var):
    return PREFIJO + var


def crear_funcion(lam, var, rol):
    cfg = VARIANTES[var]
    z = (PAQ / f"{cfg.get('paquete', var)}.zip").read_bytes()
    kw = dict(FunctionName=nombre_fn(var), Runtime=cfg["runtime"], Role=rol, Handler=cfg["handler"],
              Code={"ZipFile": z}, MemorySize=MEMORIA, Timeout=TIMEOUT, Architectures=[ARQ],
              Environment={"Variables": {"FRIO": "0"}}, Tags={"proyecto": "gl-lambda"})
    if var == "java-snap":
        kw["SnapStart"] = {"ApplyOn": "PublishedVersions"}
    try:
        lam.get_function(FunctionName=nombre_fn(var))
        lam.delete_function(FunctionName=nombre_fn(var))  # empezar siempre de cero
        time.sleep(2)
    except lam.exceptions.ResourceNotFoundException:
        pass
    for intento in range(10):
        try:
            r = lam.create_function(**kw)
            break
        except lam.exceptions.InvalidParameterValueException as e:
            if "cannot be assumed" in str(e) or "role" in str(e).lower():
                time.sleep(5)
                continue
            raise
    lam.get_waiter("function_active_v2").wait(FunctionName=nombre_fn(var))
    return {"CodeSize": r["CodeSize"], "Runtime": r["Runtime"]}


CAMPOS = {"Duration": "duracion", "Billed Duration": "facturada", "Max Memory Used": "memoria",
          "Init Duration": "init", "Restore Duration": "restore", "Billed Restore Duration": "restore_facturada"}


def leer_report(texto):
    linea = next((l for l in texto.splitlines() if l.startswith("REPORT")), None)
    if not linea:
        return None
    d = {"linea": linea.strip()}
    for parte in linea.split("\t"):
        m = re.match(r"\s*([A-Za-z ]+): ([\d.]+) (ms|MB)\s*$", parte)
        if m and m.group(1).strip() in CAMPOS:
            d[CAMPOS[m.group(1).strip()]] = float(m.group(2))
    return d


def invocar(lam, var, carga, version=None):
    kw = dict(FunctionName=nombre_fn(var), Payload=json.dumps(carga).encode(), LogType="Tail")
    if version:
        kw["Qualifier"] = version
    t0 = time.perf_counter()
    r = lam.invoke(**kw)
    cuerpo = r["Payload"].read()
    e2e = (time.perf_counter() - t0) * 1000
    logs = base64.b64decode(r.get("LogResult", "")).decode("utf-8", "replace")
    if r.get("FunctionError"):
        raise RuntimeError(f"{var}: {r['FunctionError']} {cuerpo[:300]!r} {logs[-500:]}")
    rep = leer_report(logs) or {}
    rep["e2e"] = round(e2e, 2)
    rep["respuesta"] = json.loads(cuerpo)
    init_start = next((l for l in logs.splitlines() if l.startswith("INIT_START")), None)
    if init_start:
        rep["init_start"] = init_start.strip()
    return rep


def forzar_frio(lam, var, i):
    lam.update_function_configuration(FunctionName=nombre_fn(var), Environment={"Variables": {"FRIO": str(i)}})
    lam.get_waiter("function_updated_v2").wait(FunctionName=nombre_fn(var))


def medir_variante(s, var, rol):
    import botocore.config
    lam = s.client("lambda", config=botocore.config.Config(read_timeout=60, retries={"max_attempts": 4, "mode": "standard"}))
    d = cargar(f"medida-{var}.json", {}) or {}
    log(var, "creando la función")
    d["funcion"] = crear_funcion(lam, var, rol)
    d["memoria_mb"], d["arquitectura"] = MEMORIA, ARQ
    snap = var == "java-snap"

    # 1 · ¿devuelve lo mismo? (en $LATEST: en SnapStart, también la versión publicada abajo)
    comprobacion = []
    for carga in ({"n": N_CPU, "io_ms": 0}, {"n": 1000, "io_ms": 0}, {"n": 0, "io_ms": 100}):
        r = invocar(lam, var, carga)
        comprobacion.append({"carga": carga, "respuesta": r["respuesta"]})
    d["comprobacion"] = comprobacion
    resp = comprobacion[0]["respuesta"]
    if resp.get("suma") != REFERENCIA or resp.get("n") != N_CPU:
        guardar(f"medida-{var}.json", d)
        raise SystemExit(f"{var} NO devuelve lo mismo: {resp} (referencia {REFERENCIA})")

    # 2 · arranques en frío
    frios = []
    for i in range(1, N_FRIOS + 1):
        forzar_frio(lam, var, i)
        version = None
        if snap:
            version = lam.publish_version(FunctionName=nombre_fn(var), Description=f"frio {i}")["Version"]
            lam.get_waiter("published_version_active").wait(FunctionName=nombre_fn(var), Qualifier=version,
                                                             WaiterConfig={"Delay": 5, "MaxAttempts": 120})
        r = invocar(lam, var, {"n": N_CPU, "io_ms": 0}, version)
        r["version"] = version
        frios.append(r)
        log(var, f"frío {i}/{N_FRIOS}", "init", r.get("init"), "restore", r.get("restore"), "e2e", r["e2e"])
    d["frios"] = frios
    ultima = frios[-1]["version"]

    # 3 · caliente: CPU y E/S, en el entorno que ya existe
    for _ in range(3):
        invocar(lam, var, {"n": N_CPU, "io_ms": 0}, ultima)
    d["caliente_cpu"] = [invocar(lam, var, {"n": N_CPU, "io_ms": 0}, ultima) for _ in range(N_CALIENTE)]
    d["caliente_io"] = [invocar(lam, var, {"n": 0, "io_ms": 100}, ultima) for _ in range(N_IO)]
    for lista in ("caliente_cpu", "caliente_io"):
        frias = sum(1 for r in d[lista] if "init" in r or "restore" in r)
        d[lista + "_con_arranque"] = frias
    d["fin"] = dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds")
    try:  # la versión del runtime está en INIT_START, que el Tail de invoke no trae: se lee del log group propio
        d["init_start"] = leer_init_start(s, var)
    except Exception as e:  # noqa: BLE001
        d["init_start"] = [f"no leído: {type(e).__name__}"]
    guardar(f"medida-{var}.json", d)
    log(var, "terminado")
    return var


def leer_init_start(s, var):
    """INIT_START de los flujos de registro de la propia función (filter_log_events devuelve páginas vacías)."""
    logs = s.client("logs")
    g = "/aws/lambda/" + nombre_fn(var)
    vistos = set()
    for st_ in logs.describe_log_streams(logGroupName=g, orderBy="LastEventTime", descending=True, limit=10)["logStreams"]:
        for e in logs.get_log_events(logGroupName=g, logStreamName=st_["logStreamName"], startFromHead=True, limit=3)["events"]:
            if e["message"].startswith("INIT_START"):
                vistos.add(re.sub(r"\s+", " ", e["message"]).strip())
    return sorted(vistos)


def init_start(vars_):
    s = sesion()
    for v in vars_:
        d = cargar(f"medida-{v}.json")
        if d:
            d["init_start"] = leer_init_start(s, v)
            guardar(f"medida-{v}.json", d)
            print(v, d["init_start"])


def medir(vars_):
    s = sesion()
    rol = asegurar_rol(s)
    with cf.ThreadPoolExecutor(max_workers=len(vars_)) as ex:
        futs = {ex.submit(medir_variante, s, v, rol): v for v in vars_}
        for f in cf.as_completed(futs):
            try:
                f.result()
            except BaseException as e:  # noqa: BLE001
                log(futs[f], "FALLÓ:", tachar(e))


N_RAFAGA = 20


def rafaga_variante(s, var):
    """20 invocaciones A LA VEZ: cada una que no encuentra entorno libre crea uno nuevo. En SnapStart se usa la
    última versión publicada, cuya instantánea ya se restauró antes (no recién publicada)."""
    import botocore.config
    cfg = botocore.config.Config(max_pool_connections=N_RAFAGA + 5, read_timeout=60)
    lam = s.client("lambda", config=cfg)
    d = cargar(f"medida-{var}.json")
    version = d["frios"][-1].get("version")
    if not version:
        forzar_frio(lam, var, "rafaga")
    with cf.ThreadPoolExecutor(max_workers=N_RAFAGA) as ex:
        res = list(ex.map(lambda _: invocar(lam, var, {"n": N_CPU, "io_ms": 1500}, version), range(N_RAFAGA)))  # 1,5 s ocupado: nadie reutiliza
    guardar(f"rafaga-{var}.json", {"version": version, "invocaciones": res})
    log(var, "ráfaga:", sum(1 for r in res if "init" in r or "restore" in r), "arranques de", len(res))


def rafaga(vars_):
    s = sesion()
    with cf.ThreadPoolExecutor(max_workers=len(vars_)) as ex:
        futs = {ex.submit(rafaga_variante, s, v): v for v in vars_}
        for f in cf.as_completed(futs):
            try:
                f.result()
            except BaseException as e:  # noqa: BLE001
                log(futs[f], "FALLÓ:", tachar(e))


MEMORIAS = (128, 1024)
N_MEMORIA = 50


def memoria_variante(s, var):
    """Mismo trabajo de CPU con 128 y 1024 MB: Lambda reparte la CPU según la memoria."""
    lam = s.client("lambda")
    out = {}
    for mb in MEMORIAS:
        lam.update_function_configuration(FunctionName=nombre_fn(var), MemorySize=mb)
        lam.get_waiter("function_updated_v2").wait(FunctionName=nombre_fn(var))
        frio = invocar(lam, var, {"n": N_CPU, "io_ms": 0})
        for _ in range(2):
            invocar(lam, var, {"n": N_CPU, "io_ms": 0})
        out[str(mb)] = {"frio": frio, "caliente": [invocar(lam, var, {"n": N_CPU, "io_ms": 0}) for _ in range(N_MEMORIA)]}
        log(var, mb, "MB hecho")
    lam.update_function_configuration(FunctionName=nombre_fn(var), MemorySize=MEMORIA)
    lam.get_waiter("function_updated_v2").wait(FunctionName=nombre_fn(var))
    guardar(f"memoria-{var}.json", out)


def memoria(vars_):
    s = sesion()
    with cf.ThreadPoolExecutor(max_workers=len(vars_)) as ex:
        futs = {ex.submit(memoria_variante, s, v): v for v in vars_}
        for f in cf.as_completed(futs):
            try:
                f.result()
            except BaseException as e:  # noqa: BLE001
                log(futs[f], "FALLÓ:", tachar(e))


# ───────────────────────────── limpiar ─────────────────────────────

def limpiar():
    s = sesion()
    lam, iam, logs = s.client("lambda"), s.client("iam"), s.client("logs")
    hecho, fallos = [], []
    for var in VARIANTES:
        fn = nombre_fn(var)
        try:
            lam.delete_function(FunctionName=fn)  # borra también todas sus versiones y alias
            hecho.append(f"lambda delete-function {fn}")
        except lam.exceptions.ResourceNotFoundException:
            pass
        except Exception as e:  # noqa: BLE001
            fallos.append(f"{fn}: {tachar(e)}")
    try:
        for p in iam.list_attached_role_policies(RoleName=ROL)["AttachedPolicies"]:
            iam.detach_role_policy(RoleName=ROL, PolicyArn=p["PolicyArn"])
            hecho.append(f"iam detach-role-policy {ROL} {p['PolicyName']}")
        for p in iam.list_role_policies(RoleName=ROL)["PolicyNames"]:
            iam.delete_role_policy(RoleName=ROL, PolicyName=p)
            hecho.append(f"iam delete-role-policy {ROL} {p}")
        iam.delete_role(RoleName=ROL)
        hecho.append(f"iam delete-role {ROL}")
    except iam.exceptions.NoSuchEntityException:
        pass
    except Exception as e:  # noqa: BLE001
        fallos.append(f"{ROL}: {tachar(e)}")
    time.sleep(5)  # Lambda puede escribir el último registro justo después de borrar
    for g in logs.describe_log_groups(logGroupNamePrefix="/aws/lambda/" + PREFIJO).get("logGroups", []):
        try:
            logs.delete_log_group(logGroupName=g["logGroupName"])
            hecho.append(f"logs delete-log-group {g['logGroupName']}")
        except Exception as e:  # noqa: BLE001
            fallos.append(f"{g['logGroupName']}: {tachar(e)}")

    # comprobación: lo que queda con el prefijo, mirado donde vive
    quedan_fn = []
    for pag in lam.get_paginator("list_functions").paginate():
        quedan_fn += [f["FunctionName"] for f in pag["Functions"] if f["FunctionName"].startswith(PREFIJO)]
    try:
        iam.get_role(RoleName=ROL)
        rol_queda = True
    except iam.exceptions.NoSuchEntityException:
        rol_queda = False
    quedan_roles = []
    for pag in iam.get_paginator("list_roles").paginate():
        quedan_roles += [r["RoleName"] for r in pag["Roles"] if r["RoleName"].startswith(PREFIJO)]
    quedan_logs = [g["logGroupName"] for g in
                   logs.describe_log_groups(logGroupNamePrefix="/aws/lambda/" + PREFIJO).get("logGroups", [])]
    v = {"cuando": dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds"), "hecho": hecho, "fallos": fallos,
         "quedan_funciones": quedan_fn, "rol_get_role": "existe" if rol_queda else "NoSuchEntity",
         "quedan_roles": quedan_roles, "quedan_log_groups": quedan_logs}
    guardar("limpieza.json", v)
    print(json.dumps(v, ensure_ascii=False, indent=1))
    if fallos or quedan_fn or rol_queda or quedan_roles or quedan_logs:
        raise SystemExit("⚠ queda algo gl-lambda-*")


def verificar():
    """Repite la comprobación de limpiar SIN borrar nada y la añade a limpieza.json."""
    s = sesion()
    lam, iam, logs = s.client("lambda"), s.client("iam"), s.client("logs")
    fn = []
    for pag in lam.get_paginator("list_functions").paginate():
        fn += [f["FunctionName"] for f in pag["Functions"] if f["FunctionName"].startswith(PREFIJO)]
    try:
        iam.get_role(RoleName=ROL)
        rol = "existe"
    except iam.exceptions.NoSuchEntityException:
        rol = "NoSuchEntity"
    roles = []
    for pag in iam.get_paginator("list_roles").paginate():
        roles += [r["RoleName"] for r in pag["Roles"] if r["RoleName"].startswith(PREFIJO)]
    grupos = [g["logGroupName"] for g in
              logs.describe_log_groups(logGroupNamePrefix="/aws/lambda/" + PREFIJO).get("logGroups", [])]
    v = {"cuando": dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds"), "quedan_funciones": fn,
         "rol_get_role": rol, "quedan_roles": roles, "quedan_log_groups": grupos}
    limp = cargar("limpieza.json", {})
    limp.setdefault("verificaciones", []).append(v)
    guardar("limpieza.json", limp)
    print(json.dumps(v, ensure_ascii=False))
    if fn or rol != "NoSuchEntity" or roles or grupos:
        raise SystemExit("⚠ queda algo gl-lambda-*")


# ───────────────────────────── informe ─────────────────────────────

def pct(vals, p):
    v = sorted(vals)
    return v[max(0, math.ceil(p / 100 * len(v)) - 1)]


def f1(x, dec=1):
    return f"{x:,.{dec}f}".replace(",", "X").replace(".", ",").replace("X", ".")


def ms(x, dec=1):
    return f"{f1(x, dec)} ms"


if __name__ == "__main__":
    if len(sys.argv) < 2:
        raise SystemExit(__doc__)
    orden, resto = sys.argv[1], sys.argv[2:] or ORDEN
    if orden == "construir":
        construir([v for v in resto if v != "java-snap"])
    elif orden == "medir":
        medir(resto)
    elif orden == "cronometrar":  # compila otra vez, con las imágenes ya bajadas, en _paquetes/_crono (no toca lo medido)
        PAQ = AQUI / "_paquetes" / "_crono"
        for ronda in (1, 2):
            construir([v for v in resto if v in IMAGENES], f"cronometrar-{ronda}.json")
    elif orden == "relleno":
        import random
        c = cargar("construir.json", {})
        falta = c["go"]["zip"] - c["rust"]["zip"]
        rel = PAQ / "relleno.bin"
        rel.write_bytes(random.Random(882).randbytes(falta))
        z = hacer_zip("rust-relleno", [("bootstrap", PAQ / "rust" / "bootstrap", True), ("relleno.bin", rel, False)])
        c["rust-relleno"] = {"imagen": None, "segundos": 0, "salida": "", "relleno": falta,
                             "sin_comprimir": (PAQ / "rust" / "bootstrap").stat().st_size + falta, "zip": z.stat().st_size}
        guardar("construir.json", c)
        print("relleno", falta, "zip", z.stat().st_size, "(go", c["go"]["zip"], ")")
    elif orden == "rezip":  # rehace los .zip de binarios ya compilados (sin volver a compilar)
        for v in resto:
            if (PAQ / v / "bootstrap").exists():
                z = hacer_zip(v, [("bootstrap", PAQ / v / "bootstrap", True)])
                c = cargar("construir.json", {})
                c[v]["zip"] = z.stat().st_size
                guardar("construir.json", c)
                print(v, z.stat().st_size)
    elif orden == "init_start":
        init_start(resto)
    elif orden == "rafaga":
        rafaga(resto)
    elif orden == "memoria":
        memoria([v for v in resto if v != "java-snap"])
    elif orden == "verificar":
        verificar()
    elif orden == "limpiar":
        limpiar()
    elif orden == "informe":
        import informe
        informe.escribir()
    else:
        raise SystemExit(__doc__)
