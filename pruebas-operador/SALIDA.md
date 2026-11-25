# SALIDA — Un operador de Kubernetes en Go (golandia)

> Corrida del 2026-10-20 12:59. golang:1.26 + kubebuilder v4.16.0 + kubectl v1.37.1, contra k3s (rancher/k3s:v1.37.1-k3s1).

```
go version go1.26.9 linux/amd64
### 1 · El esqueleto con kubebuilder
$ kubebuilder version
KubeBuilder:          v4.16.0
$ kubebuilder init --domain golandia.dev --repo paginas
(esqueleto creado)
$ kubebuilder create api --group web --version v1 --kind PaginaWeb --resource --controller
(API y controlador creados)
$ wc -l api/v1/paginaweb_types.go internal/controller/paginaweb_controller.go   # lo que escribo yo
  54 api/v1/paginaweb_types.go
  99 internal/controller/paginaweb_controller.go
 153 total
$ find . -name '*.go' | wc -l ; find . -type f | wc -l   # lo que genera kubebuilder
9
53
$ make manifests generate
(CRD, RBAC y DeepCopy generados)
$ make install   # la CRD en el clúster
customresourcedefinition.apiextensions.k8s.io/paginawebs.web.golandia.dev created
$ kubectl get crd | grep paginaweb
customresourcedefinition.apiextensions.k8s.io/paginawebs.web.golandia.dev

### 2 · El operador en marcha
$ go build -o bin/operador ./cmd/main.go
76965516 bytes
$ cat pagina.yaml
apiVersion: web.golandia.dev/v1
kind: PaginaWeb
metadata:
  name: hola
spec:
  replicas: 2
  mensaje: Hola desde un operador en Go
$ kubectl apply -f pagina.yaml
paginaweb.web.golandia.dev/hola created
$ kubectl get paginaweb,deployment,service,configmap -l app=hola -o name ; kubectl get paginaweb
deployment.apps/hola
service/hola
configmap/hola
NAME   REPLICAS   LISTAS   MENSAJE
hola   2          2        Hola desde un operador en Go
$ kubectl run prueba --rm -i --image=busybox:1.37 --restart=Never -- wget -qO- http://hola
<h1>Hola desde un operador en Go</h1>

### 3 · Cambiar la PaginaWeb: el operador ajusta el Deployment
$ kubectl patch paginaweb hola --type merge -p '{"spec":{"replicas":4}}'
paginaweb.web.golandia.dev/hola patched
$ kubectl get deployment hola ; kubectl get paginaweb
NAME   READY   UP-TO-DATE   AVAILABLE   AGE
hola   4/4     4            4           17s
NAME   REPLICAS   LISTAS   MENSAJE
hola   4          4        Hola desde un operador en Go

### 4 · Borrar el Deployment a mano: el operador lo vuelve a crear
$ kubectl delete deployment hola
deployment.apps "hola" deleted from default namespace
$ kubectl get deployment hola
NAME   READY   UP-TO-DATE   AVAILABLE   AGE
hola   4/4     4            4           6s

### 5 · La validación de la CRD
$ kubectl patch paginaweb hola --type merge -p '{"spec":{"replicas":50}}'
The PaginaWeb "hola" is invalid: spec.replicas: Invalid value: 50: spec.replicas in body should be less than or equal to 10

### 6 · Borrar la PaginaWeb: se va todo lo suyo
$ kubectl delete paginaweb hola
paginaweb.web.golandia.dev "hola" deleted from default namespace
$ kubectl get deployment,service,configmap hola
Error from server (NotFound): deployments.apps "hola" not found
Error from server (NotFound): services "hola" not found
Error from server (NotFound): configmaps "hola" not found
```
