# SALIDA — client-go: controla Kubernetes desde Go (golandia 677, 2026-10-26)

Ejecutado el 2026-10-26 en un clúster kind propio `g677` (kind v0.30.0, nodo Kubernetes v1.34.0, containerd 2.1.3),
Docker 29.6.2, compilado con golang:1.27.2 contra k8s.io/client-go v0.37.1 (api y apimachinery v0.37.1).
kgo corre en busybox dentro de la red `kind` con el kubeconfig interno (`kind get kubeconfig --internal`);
en el escenario 9 corre DENTRO de un pod (hostPath /opt/kgo) con `serviceAccountName: sin-permisos`.
Cada petición al API server se cuenta envolviendo el transporte (`cfg.Wrap`, config.go); KGO_URLS=1 imprime la URL de cada WATCH.
Los tiempos de la izquierda son desde que arranca kgo. Todo sale de `python correr.py` (salidas/*.txt, sin editar).
Antes de cada escenario de watch/contar/conflicto, correr.py deja web en 1 réplica lista.

Notas de lectura:
- El corte del watch se fuerza con TimeoutSeconds=5. En un clúster real lo corta el propio API server (min-request-timeout,
  1800 s por defecto → los watch duran entre 30 y 60 min) o cualquier balanceador/red; el informer de la escena 7 pidió timeoutSeconds=451/520 (aleatorio).
- El informer de client-go v0.37.1 contra un API server 1.34 hace UNA sola petición: un WATCH con sendInitialEvents=true
  (watch-list en streaming), no el clásico LIST + WATCH.
- El escenario A de «contar» tarda 18,0 s por el limitador del cliente: QPS=0/Burst=0 en rest.Config = 5 QPS y ráfaga de 10 por defecto.
- Los bytes de B y C son parecidos aquí porque en demo hay 1 pod; el sondeo crece con el número de objetos × cada segundo,
  el informer con el número de cambios.

## 0 · Versiones

```
kind v0.30.0 go1.26.5 windows/amd64
Client Version: v1.36.1
Kustomize Version: v5.8.1
Server Version: v1.34.0
Warning: version difference between client (1.36) and server (1.34) exceeds the supported minor version skew of +/-1
Docker 29.6.2
```

## 1 · Compilar kgo (golang:1.27.2, CGO_ENABLED=0, -s -w)

```
go version go1.27.2 linux/amd64
kgo 29851808 bytes
78 módulos en el grafo · 734 paquetes
k8s.io/client-go v0.37.1
k8s.io/api v0.37.1
k8s.io/apimachinery v0.37.1
```

## 2 · Cargar la config fuera del clúster (kubeconfig)

```
$ kgo config
   0.0s  InClusterConfig(): unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined
   0.0s  origen: kubeconfig /k/kubeconfig
   0.0s  host: https://g677-control-plane:6443
   0.0s  QPS=0 Burst=0 (0 = valores por defecto de client-go)
```

## 3 · Crear el Deployment web y escalarlo por /scale

```
$ kgo desplegar
   0.0s  creado demo/web · resourceVersion 2340
   1.3s  1/1 listas en 1.3s
   1.3s  escalado a 3
   3.3s  3/3 listas en 2.0s
   3.3s  peticiones: GET=16 POST=1 PUT=1
```

## 4 · Listar pods (todos los namespaces)

```
$ kgo pods
   0.0s  demo                 web-78c5486d8f-g4f46                          Running
   0.0s  demo                 web-78c5486d8f-jppvl                          Running
   0.0s  demo                 web-78c5486d8f-mp8mn                          Running
   0.0s  kube-system          coredns-66bc5c9577-6lb5c                      Running
   0.0s  kube-system          coredns-66bc5c9577-fmgh5                      Running
   0.0s  kube-system          etcd-g677-control-plane                       Running
   0.0s  kube-system          kindnet-22blb                                 Running
   0.0s  kube-system          kube-apiserver-g677-control-plane             Running
   0.0s  kube-system          kube-controller-manager-g677-control-plane    Running
   0.0s  kube-system          kube-proxy-w5lpc                              Running
   0.0s  kube-system          kube-scheduler-g677-control-plane             Running
   0.0s  local-path-storage   local-path-provisioner-7b8c8ddbd6-k6s26       Running
   0.0s  12 pods · resourceVersion de la lista: 2395 · peticiones: GET=1
```

## 5 · Watch ingenuo: el canal se cierra y el bucle termina (TimeoutSeconds=5 para forzar el corte)

```
$ kgo watch-ingenuo
   0.0s  → WATCH /api/v1/namespaces/demo/pods?labelSelector=app%3Dweb&timeout=5s&timeoutSeconds=5&watch=true
   0.0s  ADDED    web-78c5486d8f-jppvl
   1.0s  [escalador] replicas=3 ok
   1.1s  ADDED    web-78c5486d8f-z2grj
   1.1s  ADDED    web-78c5486d8f-mxhdr
   5.0s  el canal se cerró (6 MODIFIED) · el bucle terminó, el programa ya no ve nada
   8.0s  [escalador] replicas=1 ok
  19.0s  realidad al final: 1 pods de web
```

## 6 · Watch que se reabre desde el último resourceVersion · y el 410 de un resourceVersion viejo

```
$ kgo watch-reintento
   0.0s  lista inicial: 1 pods · resourceVersion 2497
   0.0s  → WATCH /api/v1/namespaces/demo/pods?labelSelector=app%3Dweb&resourceVersion=2497&timeout=5s&timeoutSeconds=5&watch=true
   1.0s  [escalador] replicas=3 ok
   1.0s  ADDED    web-78c5486d8f-nlpcw
   1.0s  ADDED    web-78c5486d8f-fthdv
   5.0s  canal cerrado · reabro desde resourceVersion 2526
   5.0s  → WATCH /api/v1/namespaces/demo/pods?labelSelector=app%3Dweb&resourceVersion=2526&timeout=5s&timeoutSeconds=5&watch=true
   8.0s  [escalador] replicas=1 ok
   9.0s  DELETED  web-78c5486d8f-nlpcw
   9.0s  DELETED  web-78c5486d8f-fthdv
  10.0s  canal cerrado · reabro desde resourceVersion 2560
  10.0s  → WATCH /api/v1/namespaces/demo/pods?labelSelector=app%3Dweb&resourceVersion=2560&timeout=5s&timeoutSeconds=5&watch=true
  15.0s  canal cerrado · reabro desde resourceVersion 2560
  15.0s  → WATCH /api/v1/namespaces/demo/pods?labelSelector=app%3Dweb&resourceVersion=2560&timeout=5s&timeoutSeconds=5&watch=true
  20.0s  canal cerrado · reabro desde resourceVersion 2560
  20.0s  → WATCH /api/v1/namespaces/demo/pods?labelSelector=app%3Dweb&resourceVersion=2560&timeout=5s&timeoutSeconds=5&watch=true
  25.0s  canal cerrado · reabro desde resourceVersion 2560
  25.0s  5 watch abiertos · 14 MODIFIED · peticiones: GET=3 PUT=2 WATCH=5
  25.0s  → WATCH /api/v1/namespaces/demo/pods?labelSelector=app%3Dweb&resourceVersion=1&timeout=5s&timeoutSeconds=5&watch=true
  25.0s  watch desde resourceVersion 1: too old resource version: 1 (2146) · código 410 · IsResourceExpired=true
```

## 7 · Listar en bucle frente a informer + lister (contador de peticiones envolviendo el transporte con rest.Config.Wrap)

```
$ kgo contar
  18.0s  A · 100 List seguidos: 18.0s · 100 peticiones · 291.6 KB
  38.3s  B · sondeo 1/s durante 20 s: 20 peticiones · 58.3 KB
  38.3s  → WATCH /api/v1/namespaces/demo/pods?allowWatchBookmarks=true&resourceVersionMatch=NotOlderThan&sendInitialEvents=true&timeout=8m40s&timeoutSeconds=520&watch=true
  38.3s  C · caché sincronizada en 4ms · peticiones: WATCH=1
  38.3s  C · 100 List del lister: 19.709µs · 1 pods · peticiones nuevas: 0
  58.3s  C · 20 s más, con escalado 1→3→1 de por medio: 19 eventos recibidos · lister ve 1 pods · peticiones totales: WATCH=1 · 50.5 KB
```

## 8 · Update frente a patch: el conflicto de resourceVersion

```
$ kgo conflicto
   0.0s  a y b leídos · replicas=1 · resourceVersion 2687
   0.0s  Update(a) replicas=2: ok · resourceVersion ahora 2710
   0.0s  Update(b) con la copia vieja: IsConflict=true
   0.0s    Operation cannot be fulfilled on deployments.apps "web": the object has been modified; please apply your changes to the latest version and try again
   0.0s  Update(b) sin resourceVersion: ok · replicas=1 (a había puesto 2)
   0.0s  Patch(labels): ok · replicas=2 · equipo=datos
   0.4s  RetryOnConflict: ok en 2 intentos
```

## 9 · RBAC desde dentro de un pod (config in-cluster, ServiceAccount sin-permisos)

```
$ kubectl auth can-i list pods -n demo --as system:serviceaccount:demo:sin-permisos
no
--- pod kgo-rbac con serviceAccountName: sin-permisos ---
   0.0s  InClusterConfig(): <nil>
   0.0s  origen: in-cluster
   0.0s  host: https://10.96.0.1:443
   0.0s  QPS=0 Burst=0 (0 = valores por defecto de client-go)
   0.0s  list pods en demo          403 · IsForbidden=true
   0.0s    pods is forbidden: User "system:serviceaccount:demo:sin-permisos" cannot list resource "pods" in API group "" in the namespace "demo"
   0.0s  get deployment web         403 · IsForbidden=true
   0.0s    deployments.apps "web" is forbidden: User "system:serviceaccount:demo:sin-permisos" cannot get resource "deployments" in API group "apps" in the namespace "demo"
   0.0s  list pods en kube-system   403 · IsForbidden=true
   0.0s    pods is forbidden: User "system:serviceaccount:demo:sin-permisos" cannot list resource "pods" in API group "" in the namespace "kube-system"
$ kubectl apply -f rol.yaml   (Role leer-pods: get, list, watch en pods · RoleBinding a sin-permisos)
role.rbac.authorization.k8s.io/leer-pods created
rolebinding.rbac.authorization.k8s.io/sin-permisos-lee-pods created
$ kubectl auth can-i list pods -n demo --as system:serviceaccount:demo:sin-permisos
yes
--- pod kgo-rbac otra vez ---
   0.0s  InClusterConfig(): <nil>
   0.0s  origen: in-cluster
   0.0s  host: https://10.96.0.1:443
   0.0s  QPS=0 Burst=0 (0 = valores por defecto de client-go)
   0.0s  list pods en demo          ok
   0.0s  get deployment web         403 · IsForbidden=true
   0.0s    deployments.apps "web" is forbidden: User "system:serviceaccount:demo:sin-permisos" cannot get resource "deployments" in API group "apps" in the namespace "demo"
   0.0s  list pods en kube-system   403 · IsForbidden=true
   0.0s    pods is forbidden: User "system:serviceaccount:demo:sin-permisos" cannot list resource "pods" in API group "" in the namespace "kube-system"
```
