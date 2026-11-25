# SALIDA — Las plantillas de Helm son de Go (golandia 676, 2026-10-26)

Ejecutado el 2026-10-26 con `python correr.py` en Docker 29.6.2. Imágenes fijas: `golang:1.27.2`, `alpine/helm:4.3.1` (compilado con go1.27.2), `ghcr.io/yannh/kubeconform:v0.7.0`.
Go: `github.com/Masterminds/sprig/v3 v3.3.0` y `sigs.k8s.io/yaml v1.6.0` (las mismas que fija el go.mod de helm.sh/helm/v4 v4.3.1).
Cada caso se renderiza en un chart temporal (tempfile) con `Chart.yaml` + la plantilla como `templates/caso.yaml` + sus values. `helm template demo …` (release «demo», namespace «default»).
Lo que sigue es el contenido de `salidas/*.txt`, sin editar.

## Resumen medido

- `go run . helm ../mini` y `helm template demo mini` dan los mismos 439 bytes (sin las 2 líneas de cabecera `---` / `# Source:` que agrega Helm).
- text/template: 19 funciones · Sprig v3.3.0: 211 · Helm v4.3.1 quita `env` y `expandenv` y declara 27 propias (24 nuevas: `toYaml`, `include`, `tpl`, `required`, `lookup`…) → 252 nombres.
- Sin Sprig, el primer error es `function "quote" not defined`; con Sprig, `function "toYaml" not defined`: toYaml es de Helm, no de Sprig.
- `default`: false, 0, "", lista vacía y nil cuentan como vacíos. `replicas: 0` → 3 y `montarToken: false` → true. Ni helm lint ni kubeconform lo detectan.
- 6 plantillas rotas: helm lint atrapa 1 (YAML que no parsea); kubeconform -strict atrapa 3 más (toYaml, quote, template sin indentar); 2 pasan las dos (default y la imagen vacía ":1.30.5-alpine" sin required).
- `version: 1.10` en values.yaml llega como 1.1 aunque pongas quote: el número ya se perdió al leer el YAML. Con `"1.10"` llega bien.
- `{{ template "x" . | nindent 4 }}` no indenta la salida: el tubo se aplica al argumento (el punto) → `wrong type for value; expected string; got common.Values`.
- range sobre un mapa sale ordenado por clave (DEBUG, PUERTO, VERSION en el chart final).
- Chart final: helm lint 0 fallos; kubeconform -strict: 2 recursos, Valid: 2, Invalid: 0.

## Versiones

```
$ versiones
docker 29.6.2
go version go1.27.2 linux/amd64
helm version.BuildInfo{Version:"v4.3.1", GitCommit:"c0d08d0d6d9424d4a14febb5e8a3e51b0dd00dff", GitTreeState:"clean", GoVersion:"go1.27.2", KubeClientVersion:"v1.37"}
kubeconform v0.7.0
[código de salida: 0]
```

## 1 · Go puro: text/template renderiza el manifiesto (mini/templates/deployment.yaml)

```
$ go run . solo-go ../mini
error: template: deployment.yaml:6: function "quote" not defined
exit status 1
[código de salida: 1]
```

```
$ go run . sprig ../mini
error: template: deployment.yaml:22: function "toYaml" not defined
exit status 1
[código de salida: 1]
```

```
$ go run . helm ../mini
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-web
  labels:
    app: "web"
spec:
  replicas: 2
  selector:
    matchLabels:
      app: "web"
  template:
    metadata:
      labels:
        app: "web"
    spec:
      containers:
        - name: web
          image: "nginx:1.30.5-alpine"
          resources:
            limits:
              cpu: 200m
              memory: 128Mi
            requests:
              cpu: 50m
[código de salida: 0]
```

## 2 · El mismo chart con helm template, y la comparación byte a byte

```
$ helm template demo mini
---
# Source: mini/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-web
  labels:
    app: "web"
spec:
  replicas: 2
  selector:
    matchLabels:
      app: "web"
  template:
    metadata:
      labels:
        app: "web"
    spec:
      containers:
        - name: web
          image: "nginx:1.30.5-alpine"
          resources:
            limits:
              cpu: 200m
              memory: 128Mi
            requests:
              cpu: 50m
[código de salida: 0]
```

```
$ comparar helm template con go run . helm
helm template (sin las 2 líneas de cabecera): 439 bytes
go run . helm: 439 bytes
¿idénticos? sí
[código de salida: 0]
```

## 3 · Cuántas funciones pone cada capa

```
$ go run . cuenta
text/template trae: 19 funciones
Sprig v3.3.0 agrega: 211
¿Sprig tiene env? true · ¿y toYaml? false · ¿required? false · ¿include? false · ¿tpl? false
Helm v4.3.1 quita env y expandenv, declara 27 propias ( 24 nuevas, 3 pisan a Sprig) → total 252
[código de salida: 0]
```

## 4 · {{ frente a {{- en text/template puro

```
$ go run . guiones
--- plantilla:
a:
  {{ if true }}
  b: 1
  {{ end }}
--- salida (· = espacio):
a:
··
··b:·1
··

--- plantilla:
a:
  {{- if true }}
  b: 1
  {{- end }}
--- salida (· = espacio):
a:
··b:·1

--- plantilla:
a:
  {{- if true -}}
  b: 1
  {{- end }}
--- salida (· = espacio):
a:b:·1
[código de salida: 0]
```

## 5 · Trampa 1: los guiones {{- y -}} (casos/01-guiones)

```
$ helm template demo 01-guiones  # sin-guion.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: guiones
data:
  modo: produccion
  
  debug: "true"
[código de salida: 0]
```

```
$ helm template demo 01-guiones  # mal.yaml + values.yaml
Error: YAML parse error on caso/templates/caso.yaml: error converting YAML to JSON: yaml: line 6: mapping values are not allowed in this context

Use --debug flag to render out invalid YAML
[código de salida: 1]
```

```
$ helm lint 01-guiones  # mal.yaml
==> Linting /chart
[INFO] Chart.yaml: icon is recommended
[ERROR] templates/caso.yaml: unable to parse YAML: error converting YAML to JSON: yaml: line 6: mapping values are not allowed in this context

Error: 1 chart(s) linted, 1 chart(s) failed
[código de salida: 1]
```

```
$ helm template demo 01-guiones  # bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: guiones
data:
  modo: produccion
  debug: "true"
[código de salida: 0]
```

## 6 · Trampa 2: indent frente a nindent (casos/02-nindent)

```
$ helm template demo 02-nindent  # mal.yaml + values.yaml
Error: YAML parse error on caso/templates/caso.yaml: error converting YAML to JSON: yaml: line 6: did not find expected key

Use --debug flag to render out invalid YAML
[código de salida: 1]
```

```
$ helm template demo 02-nindent  # bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: indent
  labels:
    app: web
    equipo: plataforma
data:
  a: "1"
[código de salida: 0]
```

## 7 · Trampa 3: toYaml (casos/03-toyaml)

```
$ helm template demo 03-toyaml  # mal.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Pod
metadata:
  name: toyaml
spec:
  containers:
    - name: web
      image: nginx:1.30.5-alpine
      resources: map[limits:map[cpu:200m memory:128Mi]]
[código de salida: 0]
```

```
$ helm lint 03-toyaml  # mal.yaml
==> Linting /chart
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
[código de salida: 0]
```

```
$ helm template … | kubeconform -strict -summary -
stdin - Pod toyaml is invalid: problem validating schema. Check JSON formatting: jsonschema validation failed with 'https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/master-standalone-strict/pod-v1.json#' - at '/spec/containers/0/resources': got string, want null or object
Summary: 1 resource found parsing stdin - Valid: 0, Invalid: 1, Errors: 0, Skipped: 0
[código de salida: 1]
```

```
$ helm template demo 03-toyaml  # bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Pod
metadata:
  name: toyaml
spec:
  containers:
    - name: web
      image: nginx:1.30.5-alpine
      resources:
        limits:
          cpu: 200m
          memory: 128Mi
[código de salida: 0]
```

```
$ helm template … | kubeconform -strict -summary -
Summary: 1 resource found parsing stdin - Valid: 1, Invalid: 0, Errors: 0, Skipped: 0
[código de salida: 0]
```

## 8 · Trampa 4: default con false y 0 (casos/04-default)

```
$ go run . vacios
a_false  → DEFAULT
b_cero   → DEFAULT
c_vacio  → DEFAULT
d_lista  → DEFAULT
e_nil    → DEFAULT
f_texto  → hola
g_true   → true
h_uno    → 1
[código de salida: 0]
```

```
$ helm template demo 04-default  # mal.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      automountServiceAccountToken: true
      containers:
        - name: web
          image: nginx:1.30.5-alpine
[código de salida: 0]
```

```
$ helm lint 04-default  # mal.yaml
==> Linting /chart
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
[código de salida: 0]
```

```
$ helm template … | kubeconform -strict -summary -
Summary: 1 resource found parsing stdin - Valid: 1, Invalid: 0, Errors: 0, Skipped: 0
[código de salida: 0]
```

```
$ helm template demo 04-default  # bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: default
spec:
  replicas: 0
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      automountServiceAccountToken: false
      containers:
        - name: web
          image: nginx:1.30.5-alpine
[código de salida: 0]
```

## 9 · Trampa 5: quote (casos/05-quote)

```
$ helm template demo 05-quote  # mal.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Pod
metadata:
  name: quote
spec:
  containers:
    - name: web
      image: nginx:1.30.5-alpine
      env:
        - name: PUERTO
          value: 8080
        - name: DEBUG
          value: true
        - name: VERSION
          value: 1.1
[código de salida: 0]
```

```
$ helm lint 05-quote  # mal.yaml
==> Linting /chart
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
[código de salida: 0]
```

```
$ helm template … | kubeconform -strict -summary -
stdin - Pod quote is invalid: problem validating schema. Check JSON formatting: jsonschema validation failed with 'https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/master-standalone-strict/pod-v1.json#' - at '/spec/containers/0/env/0/value': got number, want null or string - at '/spec/containers/0/env/1/value': got boolean, want null or string - at '/spec/containers/0/env/2/value': got number, want null or string
Summary: 1 resource found parsing stdin - Valid: 0, Invalid: 1, Errors: 0, Skipped: 0
[código de salida: 1]
```

```
$ helm template demo 05-quote  # bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Pod
metadata:
  name: quote
spec:
  containers:
    - name: web
      image: nginx:1.30.5-alpine
      env:
        - name: PUERTO
          value: "8080"
        - name: DEBUG
          value: "true"
        - name: VERSION
          value: "1.1"
[código de salida: 0]
```

```
$ helm template … | kubeconform -strict -summary -
Summary: 1 resource found parsing stdin - Valid: 1, Invalid: 0, Errors: 0, Skipped: 0
[código de salida: 0]
```

```
$ helm template demo 05-quote  # bien.yaml + values-texto.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Pod
metadata:
  name: quote
spec:
  containers:
    - name: web
      image: nginx:1.30.5-alpine
      env:
        - name: PUERTO
          value: "8080"
        - name: DEBUG
          value: "true"
        - name: VERSION
          value: "1.10"
[código de salida: 0]
```

## 10 · Trampa 6: required (casos/06-required)

```
$ helm template demo 06-required  # mal.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Pod
metadata:
  name: required
spec:
  containers:
    - name: web
      image: ":1.30.5-alpine"
[código de salida: 0]
```

```
$ helm lint 06-required  # mal.yaml
==> Linting /chart
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
[código de salida: 0]
```

```
$ helm template … | kubeconform -strict -summary -
Summary: 1 resource found parsing stdin - Valid: 1, Invalid: 0, Errors: 0, Skipped: 0
[código de salida: 0]
```

```
$ helm template demo 06-required  # bien.yaml + values.yaml
Error: execution error at (caso/templates/caso.yaml:8:17): falta imagen.repo en values.yaml

Use --debug flag to render out invalid YAML
[código de salida: 1]
```

```
$ helm template demo 06-required  # bien.yaml + values-con-repo.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Pod
metadata:
  name: required
spec:
  containers:
    - name: web
      image: "nginx:1.30.5-alpine"
[código de salida: 0]
```

```
$ helm template demo 06-required  # mal.yaml + values-sin-imagen.yaml
Error: caso/templates/caso.yaml:8:24
  executing "caso/templates/caso.yaml" at <.Values.imagen.repo>:
    nil pointer evaluating interface {}.repo

Use --debug flag to render out invalid YAML
[código de salida: 1]
```

## 11 · Trampa 7: el punto dentro de range y with, y $ (casos/07-scope)

```
$ helm template demo 07-scope  # mal.yaml + values.yaml
Error: caso/templates/caso.yaml:6:18
  executing "caso/templates/caso.yaml" at <.Values.nombre>:
    nil pointer evaluating interface {}.nombre

Use --debug flag to render out invalid YAML
[código de salida: 1]
```

```
$ helm template demo 07-scope  # bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Service
metadata:
  name: web-http
spec:
  selector:
    app: web
  ports:
    - port: 80

---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: Service
metadata:
  name: web-metricas
spec:
  selector:
    app: web
  ports:
    - port: 9090
[código de salida: 0]
```

```
$ helm template demo 07-scope  # with-mal.yaml + values.yaml
Error: caso/templates/caso.yaml:8:22
  executing "caso/templates/caso.yaml" at <.Release.Name>:
    nil pointer evaluating interface {}.Name

Use --debug flag to render out invalid YAML
[código de salida: 1]
```

```
$ helm template demo 07-scope  # with-bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: with
data:
  host: db.default.svc
  release: demo
[código de salida: 0]
```

## 12 · Trampa 8: tpl (casos/08-tpl)

```
$ helm template demo 08-tpl  # mal.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tpl
data:
  DB_HOST: "{{ .Release.Name }}-db.{{ .Release.Namespace }}.svc"
[código de salida: 0]
```

```
$ helm template demo 08-tpl  # bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tpl
data:
  DB_HOST: "demo-db.default.svc"
[código de salida: 0]
```

## 13 · Trampa 9: include frente a template (casos/09-include)

```
$ helm template demo 09-include  # template-tubo.yaml + values.yaml
Error: caso/templates/caso.yaml:6:41
  executing "caso/templates/caso.yaml" at <4>:
    wrong type for value; expected string; got common.Values

Use --debug flag to render out invalid YAML
[código de salida: 1]
```

```
$ helm template demo 09-include  # template-solo.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: include
  labels:
    app: web
equipo: plataforma
data:
  a: "1"
[código de salida: 0]
```

```
$ helm lint 09-include  # template-solo.yaml
==> Linting /chart
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
[código de salida: 0]
```

```
$ helm template … | kubeconform -strict -summary -
stdin - ConfigMap include is invalid: problem validating schema. Check JSON formatting: jsonschema validation failed with 'https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/master-standalone-strict/configmap-v1.json#' - at '': additional properties 'equipo' not allowed
Summary: 1 resource found parsing stdin - Valid: 0, Invalid: 1, Errors: 0, Skipped: 0
[código de salida: 1]
```

```
$ helm template demo 09-include  # bien.yaml + values.yaml
---
# Source: caso/templates/caso.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: include
  labels:
    app: web
    equipo: plataforma
data:
  a: "1"
[código de salida: 0]
```

## 14 · Extra: env existe en Sprig pero Helm la quita (casos/10-env)

```
$ helm template demo 10-env  # mal.yaml + values.yaml
Error: parse error at (caso/templates/caso.yaml:6): function "env" not defined

Use --debug flag to render out invalid YAML
[código de salida: 1]
```

## 15 · El chart final: helm lint + kubeconform (chart/)

```
$ helm lint chart
==> Linting /chart

1 chart(s) linted, 0 chart(s) failed
[código de salida: 0]
```

```
$ helm template demo chart
---
# Source: web/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: demo-web
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/instance: demo
spec:
  selector:
    app.kubernetes.io/name: web
    app.kubernetes.io/instance: demo
  ports:
    - name: http
      port: 80
      targetPort: http
      appProtocol: http
    - name: metricas
      port: 9090
      targetPort: metricas
      appProtocol: http

---
# Source: web/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-web
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/instance: demo
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: web
      app.kubernetes.io/instance: demo
  template:
    metadata:
      labels:
        app.kubernetes.io/name: web
        app.kubernetes.io/instance: demo
    spec:
      automountServiceAccountToken: false
      containers:
        - name: web
          image: "nginx:1.30.5-alpine"
          ports:
            - name: http
              containerPort: 80
            - name: metricas
              containerPort: 9090
          env:
            - name: DB_HOST
              value: "demo-db.default.svc"
            - name: DEBUG
              value: "true"
            - name: PUERTO
              value: "8080"
            - name: VERSION
              value: "1.10"
          resources:
            limits:
              cpu: 200m
              memory: 128Mi
            requests:
              cpu: 50m
              memory: 64Mi
[código de salida: 0]
```

```
$ helm template demo chart | kubeconform -strict -summary -
Summary: 2 resources found parsing stdin - Valid: 2, Invalid: 0, Errors: 0, Skipped: 0
[código de salida: 0]
```
