# SALIDA — Admission webhook en Go (golandia 678, 2026-10-26)

Ejecutado el 2026-10-26 con `python correr.py preparar` y luego `python correr.py` (todos los escenarios seguidos);
salidas/*.txt sin editar. Aquí van copiadas; los tiempos entre paréntesis son de reloj del cliente (kubectl).

Versiones: kind v0.30.0 · nodo kindest/node:v1.34.0 (servidor Kubernetes v1.34.0) · kubectl v1.36.1 · Docker 29.6.2 ·
golang:1.27.2 · alpine/openssl:3.5.4 · CoreDNS registry.k8s.io/coredns/coredns:v1.12.1 (el de kind).
Clúster propio `g678`, contexto `kind-g678`.

El webhook: `webhook/main.go`, 176 líneas, solo biblioteca estándar (`go list -m all` = 1 módulo, el suyo).
Binario `CGO_ENABLED=0 -trimpath -ldflags="-s -w"`: 7.114.912 bytes. Imagen `FROM scratch`: 10,1 MB en disco, 3,01 MB de contenido.
Rutas: `/validar` (ValidatingWebhookConfiguration), `/mutar` (MutatingWebhookConfiguration), `/healthz`.
Reglas de /validar: imagen sin etiqueta o con `:latest` → rechazo; falta `limits.cpu` o `limits.memory` → rechazo.
Configuración base: `failurePolicy: Fail`, `sideEffects: None`, `operations: [CREATE, UPDATE]` sobre `pods`,
`namespaceSelector` con `kubernetes.io/metadata.name NotIn [kube-system, admision]`.

## 1 · TLS: certificado autofirmado (01-preparar.txt)

```
$ openssl x509 -in tls.crt -noout -subject -issuer -ext subjectAltName
subject=CN=admision.admision.svc
issuer=CN=admision-ca
X509v3 Subject Alternative Name:
    DNS:admision.admision.svc
```
ca.crt → `caBundle` (base64) en la configuración del webhook; tls.crt + tls.key → Secret montado en /certs.

## 2 · Lo que ve el usuario en kubectl (02-rechazos.txt)

```
$ kubectl apply -f pod-latest.yaml --dry-run=server   # nginx:latest, con limits
Error from server: error when creating "pod-latest.yaml": admission webhook "pods.admision.golandia" denied the request: admision: contenedor "web": la imagen "nginx:latest" usa :latest; fija una versión
  (código 1, 0.15 s)

$ kubectl apply -f pod-sin-tag.yaml --dry-run=server   # nginx, sin etiqueta
Error from server: error when creating "pod-sin-tag.yaml": admission webhook "pods.admision.golandia" denied the request: admision: contenedor "web": la imagen "nginx" usa :latest; fija una versión

$ kubectl apply -f pod-sin-limites.yaml --dry-run=server   # nginx:1.30.5-alpine, sin limits
Error from server: error when creating "pod-sin-limites.yaml": admission webhook "pods.admision.golandia" denied the request: admision: contenedor "web": falta limits.cpu; contenedor "web": falta limits.memory

$ kubectl apply -f pod-registro.yaml --dry-run=server   # registro con puerto: localhost:5000/app
Error from server: error when creating "pod-registro.yaml": admission webhook "pods.admision.golandia" denied the request: admision: contenedor "web": la imagen "localhost:5000/app" usa :latest; fija una versión

$ kubectl apply -f pod-bien.yaml --dry-run=server   # nginx:1.30.5-alpine, con limits
pod/web created (server dry run)

$ kubectl run prueba --image=nginx:latest --dry-run=server
Error from server: admission webhook "pods.admision.golandia" denied the request: admision: contenedor "prueba": la imagen "nginx:latest" usa :latest; fija una versión; contenedor "prueba": falta limits.cpu; contenedor "prueba": falta limits.memory
  (código 1, 0.06 s)
```
`--dry-run=server` pasa por toda la admisión (webhooks incluidos) sin guardar nada; por eso la configuración declara `sideEffects: None`.

## 3 · El AdmissionReview de entrada y el de salida (03-review.txt)

Entrada (recortada; el cuerpo completo midió 3.093 bytes):
```
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "request": {
    "uid": "…",
    "kind": {"group": "", "version": "v1", "kind": "Pod"},
    "resource": {"group": "", "version": "v1", "resource": "pods"},
    "namespace": "default",
    "operation": "CREATE",
    "userInfo": {"username": "kubernetes-admin", "groups": ["kubeadm:cluster-admins", "system:authenticated"], …},
    "dryRun": true
  }
}
object.spec.containers:
  "name": "web", "image": "nginx:latest", "resources": {},
  "volumeMounts": [ kube-api-access-… ], "imagePullPolicy": "Always", …
```
El objeto llega YA con los valores por defecto puestos: `imagePullPolicy: Always` (por ser :latest) y el volumen del service account.

Salida:
```
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "response": {
    "uid": "<el mismo uid de la entrada>",
    "allowed": false,
    "status": {
      "code": 403,
      "message": "admision: contenedor \"web\": la imagen \"nginx:latest\" usa :latest; fija una versión; contenedor \"web\": falta limits.cpu; contenedor \"web\": falta limits.memory"
    }
  }
}
log: validar ns=default op=CREATE allowed=false en 92.219µs
```

## 4 · Sin caBundle (04-tls.txt)

```
$ kubectl apply -f pod-bien.yaml --dry-run=server
Error from server (InternalError): error when creating "pod-bien.yaml": Internal error occurred: failed calling webhook "pods.admision.golandia": failed to call webhook: Post "https://admision.admision.svc:443/validar?timeout=10s": tls: failed to verify certificate: x509: certificate signed by unknown authority
```
Con `caBundle` = ca.crt en base64: `pod/web created (server dry run)`.

## 5 · Webhook caído (0 réplicas): Fail frente a Ignore (05-caido.txt)

```
# failurePolicy: Fail
$ kubectl apply -f pod-bien.yaml
Error from server (InternalError): error when creating "pod-bien.yaml": Internal error occurred: failed calling webhook "pods.admision.golandia": failed to call webhook: Post "https://admision.admision.svc:443/validar?timeout=10s": dial tcp 10.96.243.184:443: connect: connection refused
  (código 1, 0.10 s)
$ kubectl apply -f pod-latest.yaml
  → el mismo error (código 1, 0.10 s)
POD   IMAGEN                     ← no se creó ninguno

# failurePolicy: Ignore
$ kubectl apply -f pod-bien.yaml
pod/bien created   (0.09 s)
$ kubectl apply -f pod-latest.yaml
pod/latest created (0.09 s)
POD      IMAGEN
bien     nginx:1.30.5-alpine
latest   nginx:latest              ← se coló
```
Fail: el clúster no crea ningún pod, ni los buenos. Ignore: crea todos, también el :latest. La respuesta es inmediata (0,1 s) porque la conexión se rechaza.

## 6 · Webhook colgado: tarda 8 s en contestar (06-colgado.txt)

```
# Fail · timeoutSeconds: 3
Error from server (InternalError): … Post "https://admision.admision.svc:443/validar?timeout=3s": context deadline exceeded
  (código 1, 3.14 s)
# Ignore · timeoutSeconds: 3
pod/bien created (server dry run)
  (código 0, 3.11 s)
# Fail · timeoutSeconds sin poner (10 por defecto)
pod/bien created (server dry run)
  (código 0, 8.15 s)
```
Colgado es peor que caído: cada pod espera el timeout entero (Ignore también: 3,1 s por pod). Con el valor por defecto (10 s), un webhook de 8 s pasa, pero cada pod tarda 8 s.

## 7 · namespaceSelector y kube-system (07-selector.txt)

```
# los limits de CoreDNS, tal como los trae kind:
registry.k8s.io/coredns/coredns:v1.12.1 {"limits":{"memory":"170Mi"},"requests":{"cpu":"100m","memory":"70Mi"}}

# namespaceSelector que excluye kube-system y admision · borro un pod de CoreDNS y espero 12 s
coredns   2/2     2            2
# namespaceSelector QUITADO · borro un pod de CoreDNS y espero 12 s
coredns   1/2     1            1
eventos FailedCreate:
Error creating: admission webhook "pods.admision.golandia" denied the request: admision: contenedor "coredns": falta limits.cpu

# sin selector + Fail · bajo el webhook a 0 y lo vuelvo a subir a 1 · a los 15 s:
admision   0/1     0            0
eventos FailedCreate:
Error creating: Internal error occurred: failed calling webhook "pods.admision.golandia": failed to call webhook: Post "https://admision.admision.svc:443/validar?timeout=10s": dial tcp 10.96.243.184:443: connect: connection refused

# salida: borrar la ValidatingWebhookConfiguration · el webhook vuelve a estar listo en 7.9 s
admision   1/1     1            1
coredns   2/2     2            2
```
Sin selector, la propia regla tumba a CoreDNS (trae limits de memoria, no de CPU), y el webhook caído no puede volver a arrancar porque su pod también necesita al webhook.

## 8 · Webhook mutante: añade los limits (08-mutante.txt)

```
$ kubectl apply -f pod-sin-limites.yaml   # nginx:1.30.5-alpine, sin limits
pod/sinlim created
$ kubectl get pod sinlim -o jsonpath='{.spec.containers[0].resources}'
{"limits":{"cpu":"500m","memory":"256Mi"},"requests":{"cpu":"500m","memory":"256Mi"}}
$ kubectl apply -f pod-latest-sin-limites.yaml --dry-run=server
Error from server: … denied the request: admision: contenedor "web": la imagen "nginx:latest" usa :latest; fija una versión
patch (base64 → texto): [{"op":"add","path":"/spec/containers/0/resources","value":{"limits":{"cpu":"500m","memory":"256Mi"}}}]
```
Los mutantes corren antes que los validadores: el pod sin limits entra con los de por defecto; el :latest sigue rechazado, ya solo por la imagen.
Kubernetes copió los limits a requests (no los traía).

## 9 · ValidatingAdmissionPolicy (CEL, sin Go) (09-vap.txt)

Dos expresiones CEL (en correr.py, `VAP`) + un binding con `validationActions: [Deny]` y el mismo namespaceSelector:
```
object.spec.containers.all(c, c.image.split('/')[size(c.image.split('/')) - 1].contains(':') && !c.image.endsWith(':latest'))
object.spec.containers.all(c, has(c.resources.limits) && 'cpu' in c.resources.limits && 'memory' in c.resources.limits)
```
```
$ kubectl apply -f pod-latest.yaml --dry-run=server
The pods "web" is invalid: : ValidatingAdmissionPolicy 'admision-cel' with binding 'admision-cel' denied request: fija una versión: nada de :latest ni imágenes sin etiqueta
$ kubectl apply -f pod-sin-tag.yaml      → el mismo rechazo
$ kubectl apply -f pod-sin-limites.yaml
The pods "web" is invalid: : ValidatingAdmissionPolicy 'admision-cel' with binding 'admision-cel' denied request: cada contenedor necesita limits.cpu y limits.memory
$ kubectl apply -f pod-registro.yaml     → el mismo rechazo de la imagen
$ kubectl apply -f pod-bien.yaml
pod/web created (server dry run)
# con el Deployment del webhook a 0 réplicas, la política sigue funcionando:
→ pod-latest.yaml rechazado igual
```
Mismos 5 casos, mismo resultado que el webhook. Sin certificado, sin Deployment, sin nada que se caiga: corre dentro del apiserver.
El mensaje es fijo por regla (el webhook dice qué contenedor y qué falta).

## 10 · Latencia (10-latencia.txt)

300 POST de un pod válido con dryRun=All por kubectl proxy; 5 rondas con el orden rotado; antes de cada tanda, un pod :latest
comprueba que la regla está puesta (201 sin nada, 403 webhook, 422 política). Del apiserver: delta de
`apiserver_admission_step_admission_duration_seconds` (paso validate, CREATE admitidos) y de
`apiserver_admission_webhook_admission_duration_seconds` alrededor de cada tanda.

```
  sin nada                    cliente: mediana de las 5 rondas  2.21 ms (+0.00 ms)  · apiserver, paso validate:    37 µs/pod
  webhook en Go               cliente: mediana de las 5 rondas  2.82 ms (+0.61 ms)  · apiserver, paso validate:   528 µs/pod
  ValidatingAdmissionPolicy   cliente: mediana de las 5 rondas  2.16 ms (-0.04 ms)  · apiserver, paso validate:   197 µs/pod

  llamada al webhook según el apiserver, por ronda: 465 · 765 · 518 · 434 · 386 µs
# lo que tarda el handler de Go por dentro (últimas 5 líneas del log):
  62.777µs  48.7µs  41.398µs  50.781µs  51.271µs
```
Lectura: el webhook añade ~0,5 ms por pod en el paso de validación (la política, ~0,2 ms). El handler de Go se lleva ~50 µs;
el resto es la ida y vuelta HTTPS. Del lado del cliente la diferencia se pierde en el ruido (2-4 ms por petición en esta máquina,
que comparte CPU con otros trabajos): rondas sueltas varían más que la diferencia entre variantes.

## Fuentes de reglas
- Dynamic Admission Control: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
  (failurePolicy por defecto Fail, timeoutSeconds por defecto 10, sideEffects, mutantes antes que validadores, evitar kube-system)
- Validating Admission Policy: https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/ (estable desde v1.30)
