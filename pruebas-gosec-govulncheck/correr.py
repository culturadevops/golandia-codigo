# -*- coding: utf-8 -*-
"""Demo del vídeo 683 de golandia «gosec y govulncheck no son lo mismo».

    python correr.py            -> SALIDA.md + salidas/*

app/ es UNA app en Go vulnerable A PROPÓSITO, con dos tipos de fallo:
  · en el código (para gosec): SQL con fmt.Sprintf, path traversal, math/rand para un token, error ignorado, InsecureSkipVerify;
  · en las dependencias (para govulncheck): github.com/golang-jwt/jwt/v4 v4.5.0 (jwt.Parse SÍ se llama) y
    golang.org/x/net v0.20.0 (solo html.EscapeString, que no es vulnerable).
nosec/main.go.plantilla: las anotaciones #nosec. cobra@v1.10.1: un repo real con versión fija para medir el ruido.
app/.github/workflows/seguridad.yml: el pipeline, ejecutado en local con act dentro de Docker.
Todo corre en contenedores con la etiqueta demo=g683, que se borran al terminar; nada se escribe fuera de salidas/.
"""
import json
import re
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).parent
SAL = AQUI / "salidas"
OUT = []
GOSEC = "v2.30.0"
GOVULN = "v1.9.0"
LINT = "v2.15.0"
ACT = "v0.2.89"
COBRA = "github.com/spf13/cobra@v1.10.1"
VOL = ["-v", "g683-gomod:/go/pkg/mod", "-v", "g683-gocache:/root/.cache"]


def run(args, timeout=1800):
    p = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout)
    return (p.stdout + p.stderr).strip(), p.returncode


def win(p):
    return str(p).replace("\\", "/")


def arrancar(nombre, imagen, *extra):
    run(["docker", "rm", "-f", nombre])
    s, rc = run(["docker", "run", "-d", "--name", nombre, "--label", "demo=g683", *VOL,
                 "-v", f"{win(AQUI)}:/src:ro", "-v", f"{win(SAL)}:/out", *extra, "--entrypoint", "sh", imagen,
                 "-c", "cp -r /src/app /w && sleep infinity"])
    assert rc == 0, s
    time.sleep(1)


def ex(nombre, cmd, timeout=1800):
    return run(["docker", "exec", "-w", "/w", nombre, "sh", "-c", cmd], timeout)


def bloque(titulo, mostrado, salida, rc=None, nota=None):
    pie = f"(código de salida {rc})\n" if rc is not None else ""
    OUT.append(f"### {titulo}\n\n`$ {mostrado}`\n\n```\n{salida}\n```\n{pie}" + (f"\n> {nota}\n" if nota else ""))
    print(f"[{rc}] {titulo}")


def limpia(s):
    return "\n".join(x for x in s.splitlines() if not x.startswith(("go: downloading", "[gosec] 20")))


def resumen_gosec(fichero):
    d = json.loads((SAL / fichero).read_text(encoding="utf-8"))
    iss = sorted(d.get("Issues") or [], key=lambda i: (i["file"], int(i["line"].split("-")[0]), i["rule_id"]))
    filas = [f"{Path(i['file']).name}:{i['line']:<4} {i['rule_id']}  {i['severity']:<6} {i['details'][:70]}" for i in iss]
    st = d.get("Stats", {})
    filas.append(f"Files: {st.get('files')} · Lines: {st.get('lines')} · Nosec: {st.get('nosec')} · Issues: {st.get('found')}")
    return "\n".join(filas), iss


def por_regla(iss):
    c = {}
    for i in iss:
        c[i["rule_id"]] = c.get(i["rule_id"], 0) + 1
    return "\n".join(f"{k}  {v}" for k, v in sorted(c.items(), key=lambda kv: (-kv[1], kv[0])))


def govuln_json(fichero):
    """Clasifica cada OSV como llamado / importado / solo requerido, por módulo."""
    texto = (SAL / fichero).read_text(encoding="utf-8")
    dec, pos, objs = json.JSONDecoder(), 0, []
    while pos < len(texto):
        while pos < len(texto) and texto[pos].isspace():
            pos += 1
        if pos >= len(texto):
            break
        o, pos = dec.raw_decode(texto, pos)
        objs.append(o)
    nivel = {}
    for o in objs:
        f = o.get("finding")
        if not f:
            continue
        t = f["trace"][0]
        n = 3 if t.get("function") else 2 if t.get("package") else 1
        clave = (t.get("module"), f["osv"])
        nivel[clave] = max(nivel.get(clave, 0), n)
    tabla = {}
    for (mod, osv), n in nivel.items():
        tabla.setdefault(mod, {1: [], 2: [], 3: []})[n].append(osv)
    lineas = []
    for mod in sorted(tabla):
        t = tabla[mod]
        lineas.append(f"{mod}: llamadas {len(t[3])} {sorted(t[3])} · importadas sin llamar {len(t[2])} · solo en go.mod {len(t[1])}")
    return "\n".join(lineas)


def codigo_y_dependencias():
    OUT.append("## 1 · La misma app con dos tipos de fallo\n")
    n = "g683-go"
    arrancar(n, "golang:1.27.2")
    ex(n, f"go install github.com/securego/gosec/v2/cmd/gosec@{GOSEC} && go install golang.org/x/vuln/cmd/govulncheck@{GOVULN}")
    s, rc = ex(n, "go version; go version -m $(which gosec) | grep -E '^\\s+mod'; govulncheck -version | grep -E 'Scanner|DB'")
    bloque("versiones", "go version; go version -m gosec; govulncheck -version", s, rc)
    bloque("las dependencias", "cat go.mod", *ex(n, "cat go.mod"))
    bloque("lo que el código usa de cada una", "grep -n 'jwt\\.\\|html\\.' main.go", *ex(n, "grep -n 'jwt\\.\\|html\\.' main.go"))
    s, rc = ex(n, "gosec -quiet -fmt json -out /out/gosec-app.json ./... ; echo $?")
    vista, iss = resumen_gosec("gosec-app.json")
    bloque("gosec ./...", "gosec ./...", vista, int(s.strip().splitlines()[-1]),
           "gosec mira solo los paquetes de tu módulo: ninguna línea menciona jwt ni x/net.")
    s, rc = ex(n, "gosec ./... 2>&1 | sed -n '/^Summary/,$p'")
    bloque("gosec: el resumen", "gosec ./... | tail", s, rc)
    s, rc = ex(n, "govulncheck ./... 2>&1")
    bloque("govulncheck ./...", "govulncheck ./...", s, rc)
    ex(n, "govulncheck -format json ./... > /out/govulncheck-app.json 2>/dev/null")
    bloque("govulncheck por módulo (de -format json)", "govulncheck -format json ./... | clasificar", govuln_json("govulncheck-app.json"),
           nota="«llamadas» = tu código llega a la función vulnerable (las que hacen fallar con código 3). «importadas sin llamar» = "
                "importas el paquete pero no la función. «solo en go.mod» = el módulo está, el paquete ni se importa.")
    # el arreglo: subir jwt y la imagen de Go; x/net se queda vieja
    run(["docker", "rm", "-f", "g683-fix"])
    s, rc = run(["docker", "run", "--rm", "--name", "g683-fix", "--label", "demo=g683", *VOL, "-v", f"{win(AQUI)}:/src:ro",
                 "golang:1.27.3", "sh", "-c",
                 "cp -r /src/app /w && cd /w && go get github.com/golang-jwt/jwt/v4@v4.5.2 >/dev/null 2>&1 && "
                 f"go install golang.org/x/vuln/cmd/govulncheck@{GOVULN} >/dev/null 2>&1 && go version && grep -E 'jwt|x/net' go.mod && "
                 "govulncheck ./... ; echo \"(código de salida: $?)\""])
    bloque("el arreglo: jwt v4.5.2 y golang:1.27.3 (x/net sigue en v0.20.0)", "go get github.com/golang-jwt/jwt/v4@v4.5.2 && govulncheck ./...",
           limpia(s), None, "x/net sigue vieja, pero tu código no llama a nada vulnerable: govulncheck sale con 0.")
    return n


def ruido(n):
    OUT.append("## 2 · Falsos positivos, #nosec y -exclude/-include\n")
    ex(n, "mkdir -p /ns && cp /src/nosec/main.go.plantilla /ns/")
    variantes = {
        "a-sin-anotacion": ("", ""),
        "b-nosec-pelado": ("// #nosec", "// #nosec"),
        "c-solo-g304": ("// #nosec G304 -- la ruta la elige el operador", "// #nosec G306 -- estado.json es público: lo sirve nginx"),
        "d-justificado": ("// #nosec G304 G703 -- la ruta la elige el operador",
                          "// #nosec G306 G703 -- estado.json es público y la ruta es fija"),
    }
    for v, (leer, escribir) in variantes.items():
        ex(n, f"mkdir -p /ns/{v} && cd /ns/{v} && sed -e 's|MARCA_LEER|{leer}|' -e 's|MARCA_ESCRIBIR|{escribir}|' "
              f"../main.go.plantilla > main.go && printf 'module nosec\\n\\ngo 1.27\\n' > go.mod")
        s, rc = ex(n, f"cd /ns/{v} && grep -n -E 'ReadFile|WriteFile' main.go && echo && gosec ./... 2>&1 | grep -E '^\\[|Nosec|Issues'")
        bloque(f"#nosec · {v}", "gosec ./...", s.replace(f"/ns/{v}/", ""), rc)
    s, rc = ex(n, "cd /ns/b-nosec-pelado && gosec -nosec ./... 2>&1 | grep -E '^\\[|Nosec|Issues'")
    bloque("-nosec: ignorar todas las anotaciones (auditoría)", "gosec -nosec ./...", s.replace("/ns/b-nosec-pelado/", ""), rc)
    s, rc = ex(n, "cd /ns/d-justificado && gosec -quiet -track-suppressions -fmt json ./... 2>/dev/null")
    d = json.loads(s[s.index("{"):])
    vista = "\n".join(f"{i['rule_id']} línea {i['line']} · " + (", ".join(f"{x['kind']}: {x['justification']}" for x in i.get('suppressions') or [])
                                                              or "sin suprimir") for i in d["Issues"])
    bloque("-track-suppressions: el motivo queda en el informe", "gosec -track-suppressions -fmt json ./...", vista, rc)
    # -exclude / -include / -severity sobre la app
    pruebas = [("todo", ""), ("-exclude=G104", "-exclude=G104"), ("-exclude=G104,G304", "-exclude=G104,G304"),
               ("-include=G201,G304,G402,G404,G701,G703", "-include=G201,G304,G402,G404,G701,G703"),
               ("-severity high", "-severity high"), ("-severity high -confidence high", "-severity high -confidence high")]
    filas = []
    for nombre, flags in pruebas:
        s, _ = ex(n, f"gosec -quiet {flags} ./... 2>&1 | grep -E '^\\[|Issues'")
        reglas = sorted({m for m in re.findall(r"\] - (G\d+)", s)})
        tot = re.search(r"Issues\s*:\s*(\d+)", s)
        filas.append(f"{nombre:<42} {tot.group(1) if tot else '0':>3}  {' '.join(reglas)}")
    bloque("filtros sobre la app", "gosec <filtro> ./...", "\n".join(filas))
    # repo real con versión fija
    s, _ = ex(n, f"d=$(go mod download -json {COBRA} | grep '\"Dir\"' | cut -d'\"' -f4) && rm -rf /cobra && cp -r $d /cobra && "
                 "chmod -R u+w /cobra && cd /cobra && go mod download && gosec -quiet -fmt json -out /out/gosec-cobra.json ./... ; "
                 "gosec -quiet ./... 2>&1 | sed -n '/^Summary/,$p'")
    vista, iss = resumen_gosec("gosec-cobra.json")
    bloque(f"gosec sobre {COBRA}", "gosec ./...", vista + "\n\npor regla:\n" + por_regla(iss))
    ex(n, "cd /cobra && gosec -quiet -tests -fmt json -out /out/gosec-cobra-tests.json ./...")
    _, iss = resumen_gosec("gosec-cobra-tests.json")
    bloque(f"gosec -tests sobre {COBRA} (también los _test.go)", "gosec -tests ./...", f"Issues: {len(iss)}\npor regla:\n" + por_regla(iss))


def golangci(n):
    OUT.append("## 3 · gosec dentro de golangci-lint\n")
    m = "g683-lint"
    run(["docker", "rm", "-f", m])
    run(["docker", "run", "-d", "--name", m, "--label", "demo=g683", *VOL, "-v", f"{win(AQUI)}:/src:ro", "--entrypoint", "sh",
         f"golangci/golangci-lint:{LINT}", "-c", "cp -r /src/app /w && sleep infinity"])
    time.sleep(1)
    s, rc = ex(m, "golangci-lint version; go version -m $(which golangci-lint) | grep securego")
    bloque("versión y el gosec que trae dentro", "golangci-lint version; go version -m golangci-lint | grep gosec", s, rc)
    base = "golangci-lint run --no-config --default none --enable gosec"
    s, rc = ex(m, f"{base} ./...")
    bloque("solo gosec, opciones por defecto", f"{base} ./...", s, rc)
    s, rc = ex(m, f"{base} --uniq-by-line=false ./... | grep -E '^main.go|issues|gosec:'")
    bloque("lo mismo con --uniq-by-line=false", f"{base} --uniq-by-line=false ./...", s, rc)
    bloque(".golangci.yml", "cat .golangci.yml", *ex(m, "cat .golangci.yml"))
    s, rc = ex(m, "golangci-lint run ./...")
    bloque("con .golangci.yml", "golangci-lint run ./...", s, rc)
    s, rc = ex(m, f"d=$(go mod download -json {COBRA} | grep '\"Dir\"' | cut -d'\"' -f4) && "
                  "rm -rf /c && cp -r $d /c && chmod -R u+w /c && cd /c && "
                  f"for f in '' '--max-same-issues 0' '--max-same-issues 0 --tests=false'; do echo \"--- $f\"; {base} $f ./... | "
                  "grep -oE '^[^ ]+[.]go:[0-9]+:[0-9]+: G[0-9]+' | sed -E 's/.*: (G[0-9]+)/\\1/' | sort | uniq -c | sort -rn; done")
    bloque("cobra en golangci-lint: max-same-issues y los _test.go", f"{base} [opciones] ./... | contar por regla", limpia(s), rc,
           "Por defecto golangci-lint muestra como mucho 3 avisos con el mismo texto, y analiza también los _test.go "
           "(gosec suelto no, salvo con -tests).")


def con_act():
    OUT.append("## 5 · El pipeline, ejecutado en local con act\n")
    n = "g683-act"
    run(["docker", "rm", "-f", n])
    run(["docker", "run", "-d", "--name", n, "--label", "demo=g683", "-v", "//var/run/docker.sock:/var/run/docker.sock",
         "-v", f"{win(AQUI)}:/src:ro", "alpine:3.22", "sh", "-c", "cp -r /src/app /w && sleep infinity"])
    ex(n, "apk add -q curl git docker-cli >/dev/null && "
          f"curl -sL https://github.com/nektos/act/releases/download/{ACT}/act_Linux_x86_64.tar.gz | tar xz -C /usr/local/bin act && "
          "git init -q && git config user.email a@b.c && git config user.name lab && git add -A && git commit -qm inicio")
    bloque("act --version", "act --version", *ex(n, "act --version"))
    bloque("el workflow", "cat .github/workflows/seguridad.yml", *ex(n, "cat .github/workflows/seguridad.yml"))
    bloque("los trabajos", "act -l", *ex(n, "act -l"))
    t0 = time.perf_counter()
    ex(n, "act push --rm -P ubuntu-latest=node:22-bookworm-slim > /tmp/act.log 2>&1", timeout=3600)
    seg = time.perf_counter() - t0
    run(["docker", "cp", f"{n}:/tmp/act.log", str(SAL / "act.log")])
    reg = re.sub(r"\x1b\[[0-9;]*m", "", (SAL / "act.log").read_text(encoding="utf-8", errors="replace"))
    pasos = [re.sub(r"\s*\[[0-9.]+m?s\]", "", x).strip() for x in reg.splitlines()
             if re.search(r"(Success|Failure|Skipped?) - (Main|Post)", x) or "Job succeeded" in x or "Job failed" in x]
    bloque("act push", "act push", "\n".join(pasos) + f"\n(act tardó {seg:.1f} s)")
    lin = [x for x in reg.splitlines() if re.search(r"\|\s*(\[/|Vulnerability #|Your code|  Issues|  Files|.*exitcode)", x)]
    bloque("lo que dijo cada trabajo (del registro de act)", "grep … act.log", "\n".join(lin[:60]))


def sarif(n):
    OUT.append("## 4 · El SARIF de gosec\n")
    ex(n, "gosec -quiet -no-fail -fmt sarif -out /out/gosec.sarif ./...")
    d = json.loads((SAL / "gosec.sarif").read_text(encoding="utf-8"))
    r = d["runs"][0]
    reglas = sorted({x["ruleId"] for x in r["results"]})
    niveles = {}
    for x in r["results"]:
        niveles[x.get("level")] = niveles.get(x.get("level"), 0) + 1
    vista = (f"version: {d['version']} · tool: {r['tool']['driver']['name']} {r['tool']['driver'].get('version', '')}\n"
             f"results: {len(r['results'])} · niveles: {niveles}\nreglas: {' '.join(reglas)}")
    s, rc = ex(n, "gosec -quiet -no-fail -fmt sarif -out /tmp/x.sarif ./...; echo \"código con -no-fail: $?\"")
    bloque("gosec -no-fail -fmt sarif -out gosec.sarif ./...", "gosec -no-fail -fmt sarif -out gosec.sarif ./...", vista + "\n" + s)


def main():
    SAL.mkdir(exist_ok=True)
    OUT.append(f"# SALIDA — gosec y govulncheck no son lo mismo (golandia 683)\n\nGenerado {datetime.now(timezone.utc):%Y-%m-%d %H:%M} UTC "
               f"con `python correr.py`. Docker 29.6.2. Código vulnerable A PROPÓSITO (app/): nunca se despliega.\n")
    try:
        n = codigo_y_dependencias()
        ruido(n)
        golangci(n)
        sarif(n)
        con_act()
    finally:
        (AQUI / "SALIDA.md").write_text("\n".join(OUT), encoding="utf-8")
        for filtro in ("label=demo=g683", "name=act-seguridad"):
            ids, _ = run(["docker", "ps", "-aq", "--filter", filtro])
            if ids:
                run(["docker", "rm", "-f", *ids.split()])
        print("ok SALIDA.md")


if __name__ == "__main__":
    main()
