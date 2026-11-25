# -*- coding: utf-8 -*-
"""pruebas_nubecita.py — compila, prueba y DEMUESTRA `nubecita`, y guarda la salida LITERAL en
`recursos/pruebas-nubecita/SALIDA.md`.

Serie «Escribe tu propia nube en Go» · episodios 1 (S3), 2 (DynamoDB), 3 (SQS) y 4 (SigV4).

## Dónde vive cada cosa, y por qué están separadas (2026-08-25, lo pidió el dueño)

    <nubecita>/                               EL CODIGO. Es una herramienta, no material
                                                 de un video: crece por episodios, se puede
                                                 publicar, y se copia sin arrastrar el canal.
    canales/golandia/recursos/pruebas-nubecita/  LA EVIDENCIA (SALIDA.md). Es el respaldo de
                                                 lo que se afirma en pantalla y tiene que
                                                 SOBREVIVIR al video, asi que vive en el canal.

## Por qué la demostración es con la CLI de AWS y no con curl

Porque el argumento del vídeo es «esto es compatible», y eso no lo demuestra un `curl` que yo
mismo escribo a mi medida: lo demuestra el **cliente oficial de Amazon**, sin parchear, hablando
con nuestro servidor sin enterarse. Si `aws s3 sync` funciona, funciona.

## ⚠️ TODO corre en Docker, y no es una preferencia

El Application Control de esta máquina bloquea **a ratos** los binarios recién compilados
(«Una directiva de Control de aplicaciones bloqueó este archivo»). Primero fue el binario que
deja `go test` en el temporal; en el episodio 3 empezó también con el del servidor. Una prueba
que a veces corre no sirve para el canon, así que:

    · las pruebas corren dentro de `golang:1.26`   → y ahí SÍ hay cgo, o sea `-race`;
    · el servidor corre dentro de su propia imagen → y de paso se mide cuánto pesa.

Lo único que se ejecuta en Windows es la **CLI de AWS**, que está firmada y no la bloquea nadie.

⚠️ Se llama a los ejecutables por lista de argumentos, NUNCA por una shell. En Git Bash, un
argumento que sea `/` lo convierte MSYS en `C:/Program Files/Git/` y el `--delimiter /` llega
mal: el listado sale plano y parece un fallo del servidor. Costó un rato encontrarlo.

    python canales/golandia/investigacion/pruebas_nubecita.py
"""
import io
import json
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

CODIGO = Path("<nubecita>")                   # la herramienta, en su propia carpeta
EVID = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-nubecita"
DEMO = EVID / "_demo"          # ficheros de usar y tirar para el `cp` y el `sync`

IMAGEN_GO = "golang:1.26"      # entorno CITABLE, y a prueba del Application Control
IMAGEN = "nubecita:ep4"        # la imagen del servidor, construida aquí mismo
CACHE_MOD = "nubecita-gomod"   # volumen para que la segunda corrida no vuelva a bajar el SDK
CONTENEDOR = "nubecita-demo"
C_FIRMA = "nubecita-firma"     # la MISMA imagen, pero comprobando firmas

ENV_AWS = {
    "AWS_ACCESS_KEY_ID": "prueba",
    "AWS_SECRET_ACCESS_KEY": "prueba",
    "AWS_DEFAULT_REGION": "us-east-1",
    "AWS_EC2_METADATA_DISABLED": "true",
    "AWS_PAGER": "",
}


def correr(args, cwd=None, entorno=None, timeout=900):
    e = dict(os.environ)
    e.update(entorno or {})
    p = subprocess.run(args, cwd=cwd and str(cwd), env=e, capture_output=True, text=True,
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
            return True          # contesta: ya está viva
        except Exception:
            time.sleep(0.05)
    return False


def bloque(doc, r, nota=None, falla_a_proposito=False):
    marca = "   # ⚠️ falla A PROPÓSITO" if falla_a_proposito else ""
    doc += ["```console", f"$ {r['cmd']}{marca}", r["salida"],
            f"[codigo de salida: {r['codigo']}]", "```", ""]
    if nota:
        doc += ["> " + nota.replace("\n", "\n> "), ""]
    return doc


ARN_MUERTOS = "arn:aws:sqs:us-east-1:000000000000:muertos"
# ⚠️ `--attributes` en forma JSON y no en la abreviada `clave=valor`: el RedrivePolicy lleva
# JSON DENTRO de una cadena, y la forma abreviada se atraganta con las comillas.
ATRIBUTOS_TRABAJOS = json.dumps({
    "VisibilityTimeout": "0",
    "RedrivePolicy": json.dumps({"deadLetterTargetArn": ARN_MUERTOS, "maxReceiveCount": "2"}),
})


def main():
    if not shutil.which("docker"):
        raise SystemExit("[!] no hay docker")
    if not shutil.which("aws"):
        raise SystemExit("[!] no hay la CLI de aws")

    fallos = []
    ver_aws = correr(["aws", "--version"])
    ver_docker = correr(["docker", "--version"])

    print(f"1 · vet y pruebas del SDK, dentro de {IMAGEN_GO}")
    montaje = str(CODIGO).replace("\\", "/")

    def en_docker(args):
        return correr(["docker", "run", "--rm",
                       "-v", f"{montaje}:/app", "-v", f"{CACHE_MOD}:/go/pkg/mod",
                       "-w", "/app", "-e", "GOFLAGS=-buildvcs=false", IMAGEN_GO] + args)

    ver_go = en_docker(["go", "version"])
    vet = en_docker(["go", "vet", "./..."])
    if vet["codigo"] != 0:
        fallos.append("go vet")
    pruebas = en_docker(["go", "test", "./...", "-count=1", "-v"])
    if pruebas["codigo"] != 0:
        fallos.append("go test")
    # 🥇 El detector de carreras SI se puede correr aqui: en el contenedor hay cgo. En Windows no
    # (CGO_ENABLED=0), y sin el, los candados del almacen y de las colas no los valida nadie.
    carreras = en_docker(["go", "test", "./...", "-count=1", "-race"])
    if carreras["codigo"] != 0:
        fallos.append("go test -race")
    print(f"    vet {vet['codigo']} · test {pruebas['codigo']} · race {carreras['codigo']}")

    print("2 · construir la imagen del servidor")
    build = correr(["docker", "build", "-t", IMAGEN, str(CODIGO)])
    if build["codigo"] != 0:
        raise SystemExit(f"[!] no construye la imagen:\n{build['salida']}")
    # ⚠️ SON DOS NUMEROS DISTINTOS y confundirlos es publicar una cifra falsa:
    #   `docker image inspect .Size`  -> capas COMPRIMIDAS: lo que te DESCARGAS
    #   `docker images`               -> descomprimida: lo que OCUPA en tu disco
    # (Se comprobo en el video 121 resolviendo el manifiesto a mano: sale el mismo byte.)
    tam = correr(["docker", "image", "inspect", IMAGEN, "--format", "{{.Size}}"])
    bytes_imagen = int(tam["salida"].strip())
    lst = correr(["docker", "images", "--format", "{{.Repository}}:{{.Tag}}	{{.Size}}"])
    en_disco = "?"
    for linea in lst["salida"].splitlines():
        if linea.startswith(IMAGEN + "	"):
            en_disco = linea.split("	", 1)[1].strip()
    print(f"    imagen {IMAGEN}: descarga {bytes_imagen:,} bytes · en disco {en_disco}"
          .replace(",", "."))

    print("3 · levantar la nubecita y darle con la CLI de AWS de verdad")
    puerto = puerto_libre()
    correr(["docker", "rm", "-f", CONTENEDOR])
    arranque = correr(["docker", "run", "-d", "--name", CONTENEDOR,
                       "-p", f"{puerto}:4566", IMAGEN])
    if arranque["codigo"] != 0:
        raise SystemExit(f"[!] no arranca el contenedor:\n{arranque['salida']}")
    punta = f"http://127.0.0.1:{puerto}"
    cola = f"{punta}/000000000000"
    registro = ""
    try:
        if not esperar(punta + "/"):
            raise SystemExit("[!] la nubecita no arrancó")

        DEMO.mkdir(parents=True, exist_ok=True)
        (DEMO / "gato.txt").write_text("hola nubecita\n", encoding="utf-8")
        (DEMO / "2026").mkdir(exist_ok=True)
        (DEMO / "2026" / "enero.txt").write_text("uno\n", encoding="utf-8")
        (DEMO / "2026" / "febrero.txt").write_text("dos\n", encoding="utf-8")

        E = ["--endpoint-url", punta]

        # ── episodio 1: S3 ────────────────────────────────────────────────
        pasos_s3 = [
            (["aws", *E, "s3", "mb", "s3://fotos"],
             "`mb` es CreateBucket. La CLI oficial, sin parchear, sin saber que no es Amazon.",
             False),
            (["aws", *E, "s3", "cp", str(DEMO / "gato.txt"), "s3://fotos/gato.txt"],
             None, False),
            (["aws", *E, "s3", "sync", str(DEMO / "2026"), "s3://fotos/2026"],
             "`sync` es la prueba de fuego: por dentro son varios ListObjectsV2 y varios "
             "PutObject encadenados. Si la compatibilidad fuera de mentira, se cae aquí.", False),
            (["aws", *E, "s3", "ls", "s3://fotos", "--recursive"], None, False),
            (["aws", *E, "s3api", "list-objects-v2", "--bucket", "fotos", "--delimiter", "/"],
             "🥇 Aquí está lo que nadie cuenta: S3 **no tiene carpetas**. Con `--delimiter /` "
             "las dos claves que empiezan por `2026/` se colapsan en UN prefijo común. Eso es "
             "todo lo que hay detrás de lo que tú ves como una carpeta.", False),
            (["aws", *E, "s3api", "head-object", "--bucket", "fotos", "--key", "gato.txt"],
             "El ETag es el md5 del contenido entre comillas, calculado igual que en S3.", False),
            (["aws", *E, "s3api", "get-object", "--bucket", "fotos", "--key", "no-existe",
              str(DEMO / "x")],
             "El error llega a la CLI con su **código**, no como un 404 pelado. Es lo que "
             "convierte un fallo en algo que tu código puede distinguir.", True),
            (["aws", *E, "s3", "rb", "s3://fotos"],
             "S3 se niega a borrar un bucket con cosas dentro. La nubecita también.", True),
            (["aws", *E, "s3", "rm", "s3://fotos", "--recursive"], None, False),
            (["aws", *E, "s3", "rb", "s3://fotos"],
             "Y así se destruye la demo entera: vaciar y borrar. Sin factura, porque no hay "
             "factura.", False),
        ]

        # ── episodio 2: DynamoDB, en EL MISMO puerto ──────────────────────
        pasos_dyn = [
            (["aws", *E, "dynamodb", "create-table", "--table-name", "pedidos",
              "--attribute-definitions",
              "AttributeName=cliente,AttributeType=S", "AttributeName=fecha,AttributeType=S",
              "--key-schema",
              "AttributeName=cliente,KeyType=HASH", "AttributeName=fecha,KeyType=RANGE",
              "--billing-mode", "PAY_PER_REQUEST"],
             "Clave COMPUESTA: `cliente` particiona y `fecha` ordena. Esto se elige una vez y "
             "no se cambia nunca: es lo que decide qué consultas van a ser posibles.", False),
            (["aws", *E, "dynamodb", "put-item", "--table-name", "pedidos", "--item",
              '{"cliente":{"S":"ana"},"fecha":{"S":"2026-01-15"},"total":{"N":"42.5"}}'],
             "El número viaja como **cadena**. No es un descuido — así no pasa por el float de "
             "ningún lenguaje y no pierde precisión.", False),
            (["aws", *E, "dynamodb", "put-item", "--table-name", "pedidos", "--item",
              '{"cliente":{"S":"ana"},"fecha":{"S":"2026-02-03"},"total":{"N":"20"}}'],
             None, False),
            (["aws", *E, "dynamodb", "put-item", "--table-name", "pedidos", "--item",
              '{"cliente":{"S":"ana"},"fecha":{"S":"2026-03-21"},"total":{"N":"30"}}'],
             None, False),
            (["aws", *E, "dynamodb", "query", "--table-name", "pedidos",
              "--key-condition-expression", "cliente = :c AND fecha BETWEEN :a AND :b",
              "--expression-attribute-values",
              '{":c":{"S":"ana"},":a":{"S":"2026-02-01"},":b":{"S":"2026-12-31"}}'],
             "🥇 Un rango sobre la clave de orden. Y funciona porque las fechas van en ISO-8601: "
             "el orden alfabético coincide con el cronológico. Con `15/01/2026` esto daría "
             "basura, y sin un solo error.", False),
            (["aws", *E, "dynamodb", "query", "--table-name", "pedidos",
              "--key-condition-expression", "begins_with(cliente, :p)",
              "--expression-attribute-values", '{":p":{"S":"a"}}'],
             "⛔ Sobre la clave de **partición** solo hay igualdad: no existe «los clientes que "
             "empiecen por a». No es una limitación de nubecita, es la regla de DynamoDB — y por "
             "eso el error lo dice.", True),
            (["aws", *E, "dynamodb", "scan", "--table-name", "pedidos", "--select", "COUNT"],
             "`Scan` lee la tabla **entera**. Aquí son tres elementos; en Amazon es la "
             "diferencia entre unos centavos y una factura que duele.", False),
            (["aws", *E, "dynamodb", "delete-table", "--table-name", "pedidos"], None, False),
        ]

        # ── episodio 3: SQS, en EL MISMO puerto ───────────────────────────
        # ⚠️ La demo NO duerme. La visibilidad se pone a 0, asi que el resultado no depende de
        # lo que tarde la CLI en arrancar. Con una visibilidad de verdad esto seria inestable.
        pasos_sqs = [
            (["aws", *E, "sqs", "create-queue", "--queue-name", "muertos"],
             "Primero la cola de muertos, porque la otra la va a nombrar.", False),
            (["aws", *E, "sqs", "create-queue", "--queue-name", "trabajos",
              "--attributes", ATRIBUTOS_TRABAJOS],
             "`RedrivePolicy` lleva **JSON dentro de una cadena** de otro JSON. Sí, de verdad. "
             "Y dice: a la tercera entrega, a la cola de muertos.", False),
            (["aws", *E, "sqs", "send-message", "--queue-url", f"{cola}/trabajos",
              "--message-body", "procesa el pedido 7"],
             "El `MD5OfMessageBody` no es decorativo: el SDK lo **comprueba**. Si lo calculas "
             "mal, el cliente da un error de suma y parece que se corrompió la red.", False),
            (["aws", *E, "sqs", "receive-message", "--queue-url", f"{cola}/trabajos",
              "--attribute-names", "All"],
             "🥇 Recibir NO saca el mensaje de la cola: lo esconde. Y el `ReceiptHandle` es de "
             "ESTA entrega — no es el identificador del mensaje, y el de la entrega anterior "
             "deja de valer.", False),
            (["aws", *E, "sqs", "receive-message", "--queue-url", f"{cola}/trabajos",
              "--attribute-names", "All"],
             "Segunda entrega del MISMO mensaje, porque nadie lo borró. "
             "`ApproximateReceiveCount` va por dos: ese contador es lo que delata a un "
             "consumidor que se está atragantando.", False),
            (["aws", *E, "sqs", "receive-message", "--queue-url", f"{cola}/trabajos"],
             "Y a la tercera ya no llega: se pasó de `maxReceiveCount` y se apartó. Salida "
             "vacía, sin error — porque una cola sin trabajo tampoco es un error.", False),
            (["aws", *E, "sqs", "receive-message", "--queue-url", f"{cola}/muertos"],
             "🥇 Ahí está, entero, en la cola de muertos. Eso es todo lo que hay detrás del "
             "«dead letter queue»: un contador de entregas y un sitio donde apartar.", False),
            (["aws", *E, "sqs", "get-queue-attributes", "--queue-url", f"{cola}/trabajos",
              "--attribute-names", "All"],
             "Los dos números que SQS llama «aproximados», y aquí se ve por qué: dependen del "
             "instante exacto en que preguntas.", False),
            (["aws", *E, "sqs", "delete-queue", "--queue-url", f"{cola}/trabajos"], None, False),
            (["aws", *E, "sqs", "delete-queue", "--queue-url", f"{cola}/muertos"], None, False),
            (["aws", *E, "sqs", "list-queues"],
             "Sin salida: no queda nada. Los tres servicios se destruyen igual de fácil, porque "
             "no hay nada que destruir.", False),
        ]

        resultados = []
        for servicio, pasos in (("s3", pasos_s3), ("dynamodb", pasos_dyn), ("sqs", pasos_sqs)):
            for args, nota, falla in pasos:
                r = correr(args, entorno=ENV_AWS)
                # `aws s3 cp` y `sync` escriben la barra de progreso con \r: queda ilegible.
                r["salida"] = r["salida"].replace("\r", "\n")
                r["cmd"] = (r["cmd"].replace(punta, "http://127.0.0.1:4566")
                            .replace(str(DEMO), "."))
                if (r["codigo"] != 0) != falla:
                    fallos.append(r["cmd"])
                    print(f"    [!] {r['cmd']} -> {r['codigo']} (esperaba "
                          f"{'fallo' if falla else 'ok'})")
                resultados.append((servicio, r, nota, falla))
                print(f"    · {servicio:9s} {args[3]:22s} codigo {r['codigo']}")
        registro = correr(["docker", "logs", CONTENEDOR])["salida"]

        # ── episodio 4: la MISMA imagen, con -firmar ──────────────────────
        print("5 · la misma imagen, comprobando firmas SigV4")
        puerto_f = puerto_libre()
        correr(["docker", "rm", "-f", C_FIRMA])
        correr(["docker", "run", "-d", "--name", C_FIRMA, "-p", f"{puerto_f}:4566",
                IMAGEN, "-escuchar", "0.0.0.0", "-firmar"])
        punta_f = f"http://127.0.0.1:{puerto_f}"
        if not esperar(punta_f + "/"):
            raise SystemExit("[!] la nubecita firmada no arrancó")
        F = ["--endpoint-url", punta_f]
        pasos_firma = [
            (["aws", *F, "s3", "mb", "s3://fotos"], ENV_AWS,
             "La CLI oficial firma cada petición sin que se lo pidas, y nuestra comprobación "
             "la acepta. Si esto pasa, nuestra implementación de SigV4 entiende exactamente lo "
             "que produce Amazon.", False),
            (["aws", *F, "s3", "ls"], {**ENV_AWS, "AWS_SECRET_ACCESS_KEY": "el-secreto-malo"},
             "🥇 Con otro secreto, la firma sale distinta — y el secreto NO ha viajado por el "
             "cable ni una vez. Lo que no coincide es el resultado de una cadena de HMAC.", True),
            (["aws", *F, "s3", "ls"], {**ENV_AWS, "AWS_ACCESS_KEY_ID": "no-te-conozco"},
             "Y una clave de acceso desconocida se rechaza antes de calcular nada.", True),
        ]
        for args, entorno, nota, falla in pasos_firma:
            r = correr(args, entorno=entorno)
            r["salida"] = r["salida"].replace(chr(13), chr(10))
            r["cmd"] = r["cmd"].replace(punta_f, "http://127.0.0.1:4566")
            if entorno.get("AWS_SECRET_ACCESS_KEY") == "el-secreto-malo":
                r["cmd"] = "AWS_SECRET_ACCESS_KEY=el-secreto-malo " + r["cmd"]
            if entorno.get("AWS_ACCESS_KEY_ID") == "no-te-conozco":
                r["cmd"] = "AWS_ACCESS_KEY_ID=no-te-conozco " + r["cmd"]
            if (r["codigo"] != 0) != falla:
                fallos.append(r["cmd"])
                print(f"    [!] {r['cmd']} -> {r['codigo']}")
            resultados.append(("firma", r, nota, falla))
            print(f"    · firma     codigo {r['codigo']}")
        # Sin firma NINGUNA: ni se mira qué pedías.
        # ⚠️ `curl -o /dev/null -w` sale con código 23 (error de escritura) aunque la petición
        # haya ido bien, y eso queda confuso al lado de un paso que SÍ pasó. Con `-i` se ve la
        # línea de estado y el cuerpo del error, que es lo que interesa enseñar.
        sin = correr(["curl", "-s", "-i", punta_f + "/"])
        sin["salida"] = sin["salida"].replace(chr(13), "")
        sin["cmd"] = "curl -s -i http://127.0.0.1:4566/"
        resultados.append(("firma", sin,
                           "Sin cabecera `Authorization` ni se mira qué pedías: 403 y a otra "
                           "cosa.", False))
        registro_firma = correr(["docker", "logs", C_FIRMA])["salida"]
        registro = (registro + chr(10) + chr(10) +
                    "# ── y con -firmar, la misma imagen ──" + chr(10) +
                    registro_firma)
    finally:
        correr(["docker", "rm", "-f", CONTENEDOR])
        correr(["docker", "rm", "-f", C_FIRMA])

    registro = registro.replace(f":{puerto}", ":4566")

    # ── el informe ───────────────────────────────────────────────────────
    fuentes = {f.name: f.read_text(encoding="utf-8") for f in sorted(CODIGO.glob("*.go"))}
    D = []
    A = D.append
    A("# Pruebas ejecutadas — nubecita · S3, DynamoDB, SQS y SigV4")
    A("")
    A("> Salida **literal** de esta máquina. La genera `investigacion/pruebas_nubecita.py` y se")
    A("> vuelve a generar sola. ⛔ Lo que no esté aquí no puede salir en pantalla.")
    A("")
    A(f"**Entorno citable en el vídeo:** `{ver_go['salida']}` (dentro de `{IMAGEN_GO}`) · "
      f"`{ver_docker['salida']}` · `{ver_aws['salida']}` · `aws-sdk-go-v2` para "
      "`s3`, `dynamodb` y `sqs`")
    A("")
    A(f"**El código vive en `{CODIGO}`**, en su propia carpeta, fuera del canal: es una "
      "herramienta que crece por episodios, no material de un vídeo.")
    A("")
    A(f"**La imagen entera se descarga en {bytes_imagen:,} bytes**".replace(",", ".") +
      f" ({bytes_imagen/1048576:.2f} MiB comprimidos) y ocupa **{en_disco}** en disco. La "
      "segunda etapa es `scratch`: dentro no hay NADA más que el binario, porque con "
      "`CGO_ENABLED=0` el binario de Go es estático.")
    A("")
    A("⚠️ Son dos números distintos y confundirlos es publicar una cifra falsa: lo que te")
    A("descargas (capas comprimidas) no es lo que ocupa descomprimida en tu disco.")
    A("")
    A("⚠️ **Todo corre en Docker, y no es una preferencia.** El Application Control de esta")
    A("máquina bloquea a ratos los binarios recién compilados («Una directiva de Control de")
    A("aplicaciones bloqueó este archivo»): primero el de `go test`, y desde el episodio 3")
    A("también el del servidor. Una prueba que a veces corre no sirve. Lo único que se ejecuta")
    A("en Windows es la CLI de AWS, que está firmada. De regalo, en el contenedor hay cgo y por")
    A("tanto **detector de carreras**, que en este Windows no hay.")
    A("")
    A("---")
    A("")
    A("## 1 · `go vet` y las pruebas contra el SDK de AWS de verdad")
    A("")
    A("Las pruebas NO usan un cliente de mentira: usan `aws-sdk-go-v2/service/s3`,")
    A("`.../dynamodb` y `.../sqs`, los mismos que usarías contra Amazon. Es la única forma")
    A("honesta de decir «es compatible»: si el SDK oficial lo entiende, lo es.")
    A("")
    D[:] = bloque(D, vet)
    D[:] = bloque(D, pruebas)
    A("### Y con el detector de carreras")
    A("")
    A("El almacén, las tablas y las colas se tocan desde varias gorrutinas a la vez: sin esto,")
    A("los candados no los valida nadie.")
    A("")
    D[:] = bloque(D, carreras)
    A("---")
    A("")
    A("## 2 · S3 con la CLI oficial de AWS, sin parchear")
    A("")
    A("Lo único que cambia respecto a hablar con Amazon es `--endpoint-url`.")
    A("")
    for servicio, r, nota, falla in resultados:
        if servicio == "s3":
            D[:] = bloque(D, r, nota, falla)
    A("---")
    A("")
    A("## 3 · DynamoDB, en el MISMO puerto y el MISMO proceso")
    A("")
    A("S3 dice qué quieres con el método y la ruta; DynamoDB manda siempre un `POST /` con el")
    A("cuerpo en JSON y la operación en una cabecera, `X-Amz-Target: DynamoDB_20120810.PutItem`.")
    A("Dos servicios de la misma nube, dos diseños — y los dos caben en el mismo puerto porque")
    A("el enrutador mira esa cabecera.")
    A("")
    for servicio, r, nota, falla in resultados:
        if servicio == "dynamodb":
            D[:] = bloque(D, r, nota, falla)
    A("---")
    A("")
    A("## 4 · SQS, también en el mismo puerto")
    A("")
    A("SQS habla como DynamoDB —`POST /`, JSON, operación en la cabecera— pero con otro prefijo:")
    A("`AmazonSQS.SendMessage`. Y eso tiene una historia detrás: SQS es más viejo que S3 y")
    A("durante años habló el protocolo «Query» de AWS, con los parámetros en el cuerpo tipo")
    A("formulario. Amazon lo migró a JSON. O sea: **el protocolo de un servicio no es para")
    A("siempre**.")
    A("")
    for servicio, r, nota, falla in resultados:
        if servicio == "sqs":
            D[:] = bloque(D, r, nota, falla)
    A("---")
    A("")
    A("## 5 · SigV4: la misma imagen, comprobando firmas")
    A("")
    A("Hasta aquí el doble no autenticaba a nadie, que es lo correcto para un doble de pruebas.")
    A("Con `-firmar`, la misma imagen comprueba la firma **SigV4** de cada petición. Y aquí lo")
    A("importante: quien FIRMA es la CLI oficial de Amazon y quien COMPRUEBA es nuestro código.")
    A("")
    A("⚠️ **Firmar no es cifrar.** La petición viaja en claro. Lo que la firma demuestra es")
    A("quién la manda —sin que el secreto viaje nunca— y que nadie la ha tocado por el camino.")
    A("")
    for servicio, r, nota, falla in resultados:
        if servicio == "firma":
            D[:] = bloque(D, r, nota, falla)
    A("---")
    A("")
    A("## 6 · Lo que vio el servidor")
    A("")
    A("Esto es la ventaja de tener tu propio emulador y no uno de caja: **ves el HTTP crudo que")
    A("manda el cliente de Amazon**. Fíjate en `?list-type=2`, en `encoding-type=url` (que la")
    A("CLI manda SIEMPRE), en `delimiter=%2F` — y en cómo cambia el registro cuando entran")
    A("DynamoDB y SQS: de rutas a `POST / [Operacion]`.")
    A("")
    A("```console")
    A(registro)
    A("```")
    A("")
    A("---")
    A("")
    A("## 7 · El código completo")
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
