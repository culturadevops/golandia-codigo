# SALIDA — Workflows duraderos con Temporal en Go (golandia)

> Corrida del 2026-11-23 06:56. Regenerable con `python correr.py`. Todo en local: el servidor de desarrollo de Temporal en un contenedor y los workers en otros. Los «efectos» de cada paso se apuntan en un archivo compartido (_datos/efectos.log) para contar cuántas veces se ejecutó de verdad cada uno.

- `temporal version 1.10.1 (Server 1.32.2, UI 2.56.4)` (imagen `temporalio/temporal:1.10.1`, `temporal server start-dev`) · `go version go1.27.2 linux/amd64` · SDK `go.temporal.io/sdk v1.51.0`
- Imagen de los binarios de Go (`gl-prac-temporal-go:1`, FROM scratch): 19.8 MiB
- El flujo: Reservar (1 s) → Cobrar (la pasarela falla los intentos 1 y 2) → Facturar (6 páginas de 1 s, con heartbeat por página) → Enviar. Reintentos: 1 s, ×2, máximo 10 s. HeartbeatTimeout 3 s, StartToCloseTimeout 30 s.

## 0 · El servidor

```
SERVING
```

## A · Sin Temporal: el mismo flujo como un programa normal, y se mata a mitad de la factura

Primera ejecución (SIGKILL tras la página 2):
```
11:56:43.019 [sin-temporal] reservo el stock de pedido-A
11:56:44.019 [sin-temporal] cobro pedido-A
11:56:45.023 [sin-temporal] página 1 de 6
11:56:46.024 [sin-temporal] página 2 de 6
```

Se vuelve a lanzar el mismo programa:
```
11:56:46.626 [sin-temporal] reservo el stock de pedido-A
11:56:47.627 [sin-temporal] cobro pedido-A
11:56:48.630 [sin-temporal] página 1 de 6
11:56:49.634 [sin-temporal] página 2 de 6
11:56:50.636 [sin-temporal] página 3 de 6
11:56:51.639 [sin-temporal] página 4 de 6
11:56:52.641 [sin-temporal] página 5 de 6
11:56:53.643 [sin-temporal] página 6 de 6
11:56:53.643 [sin-temporal] envío pedido-A
```

Efectos de pedido-A:
```
11:56:44.019 sin-temporal reservar pedido-A
11:56:44.020 sin-temporal cobrar pedido-A
11:56:45.022 sin-temporal factura pedido-A página 1
11:56:46.023 sin-temporal factura pedido-A página 2
11:56:47.626 sin-temporal reservar pedido-A
11:56:47.628 sin-temporal cobrar pedido-A
11:56:48.630 sin-temporal factura pedido-A página 1
11:56:49.633 sin-temporal factura pedido-A página 2
11:56:50.635 sin-temporal factura pedido-A página 3
11:56:51.638 sin-temporal factura pedido-A página 4
11:56:52.640 sin-temporal factura pedido-A página 5
11:56:53.642 sin-temporal factura pedido-A página 6
11:56:53.644 sin-temporal enviar pedido-A
```

**Resumen A:** reservar: 2 · cobrar: 2 · validar: 0 · páginas de factura: 8 (1, 2, 1, 2, 3, 4, 5, 6) · enviar: 1

## B · Con Temporal: reintentos, SIGKILL al worker a mitad de la factura, y otro worker que sigue

Se inicia el workflow:
```
workflow pedido-B iniciado (run 01a20844-2a63-740e-8006-a4ca6dde6f98)
```

Consulta del estado (query «estado») con worker1 vivo:
```
Query result:
  QueryResult  "facturando"
```

SIGKILL a worker1 a las 11:57:01.003 (UTC). Su salida completa:
```
[worker1] worker listo, versión 1 del workflow
11:56:54.771 [worker1] Reservar (intento 1) reservo el stock de pedido-B
11:56:55.787 [worker1] Cobrar (intento 1) la pasarela no responde
2026/11/23 11:56:55 ERROR Activity error. Namespace default TaskQueue pedidos WorkerID worker1 WorkflowID pedido-B RunID 01a20844-2a63-740e-8006-a4ca6dde6f98 ActivityType Cobrar Attempt 1 Error pasarela de pago: tiempo agotado
11:56:56.793 [worker1] Cobrar (intento 2) la pasarela no responde
2026/11/23 11:56:56 ERROR Activity error. Namespace default TaskQueue pedidos WorkerID worker1 WorkflowID pedido-B RunID 01a20844-2a63-740e-8006-a4ca6dde6f98 ActivityType Cobrar Attempt 2 Error pasarela de pago: tiempo agotado
11:56:58.800 [worker1] Cobrar (intento 3) cobro pedido-B
11:56:58.807 [worker1] Facturar (intento 1) empiezo en la página 1 de 6
11:56:59.809 [worker1] Facturar (intento 1) página 1 de 6
11:57:00.813 [worker1] Facturar (intento 1) página 2 de 6
```

5 s después, SIN ningún worker, `temporal workflow describe pedido-B` (campos):
```
status: WORKFLOW_EXECUTION_STATUS_RUNNING
historyLength: 17
pendingActivity: Facturar · state PENDING_ACTIVITY_STATE_SCHEDULED · attempt 2
heartbeatDetails: [1]
lastFailure: activity Heartbeat timeout
```

Y la query sin worker (la responde el código del workflow, que corre en un worker); tardó 29.2 s en responder:
```
Error: querying workflow failed: no poller seen for task queue recently, worker may be down
```

worker2 arranca 34.9 s después del SIGKILL. Su salida:
```
[worker2] worker listo, versión 1 del workflow
11:57:35.911 [worker2] Facturar (intento 2) empiezo en la página 2 de 6
11:57:36.914 [worker2] Facturar (intento 2) página 2 de 6
11:57:37.918 [worker2] Facturar (intento 2) página 3 de 6
11:57:38.920 [worker2] Facturar (intento 2) página 4 de 6
11:57:39.924 [worker2] Facturar (intento 2) página 5 de 6
11:57:40.926 [worker2] Facturar (intento 2) página 6 de 6
11:57:40.933 [worker2] Enviar (intento 1) envío pedido-B
```

Resultado:
```
Results:
  Status          COMPLETED
  Result          "pedido pedido-B enviado"
  ResultEncoding  json/plain
```

Efectos de pedido-B:
```
11:56:55.775 worker1 reservar pedido-B
11:56:58.801 worker1 cobrar pedido-B
11:56:59.808 worker1 factura pedido-B página 1
11:57:00.812 worker1 factura pedido-B página 2
11:57:36.913 worker2 factura pedido-B página 2
11:57:37.917 worker2 factura pedido-B página 3
11:57:38.920 worker2 factura pedido-B página 4
11:57:39.923 worker2 factura pedido-B página 5
11:57:40.925 worker2 factura pedido-B página 6
11:57:40.935 worker2 enviar pedido-B
```

**Resumen B:** reservar: 1 · cobrar: 1 · validar: 0 · páginas de factura: 7 (1, 2, 2, 3, 4, 5, 6) · enviar: 1

El historial de pedido-B: 29 eventos. Los de actividades, en orden:
```
  5 SCHEDULED Reservar
  6 STARTED   attempt 1 · worker1
  7 COMPLETED · worker1
 11 SCHEDULED Cobrar
 12 STARTED   attempt 3 · worker1
 13 COMPLETED · worker1
 17 SCHEDULED Facturar
 18 STARTED   attempt 2 · worker2
 19 COMPLETED · worker2
 23 SCHEDULED Enviar
 24 STARTED   attempt 1 · worker2
 25 COMPLETED · worker2
```

Conteo por tipo de evento: WORKFLOW_EXECUTION_STARTED 1, WORKFLOW_TASK_SCHEDULED 5, WORKFLOW_TASK_STARTED 5, WORKFLOW_TASK_COMPLETED 5, ACTIVITY_TASK_SCHEDULED 4, ACTIVITY_TASK_STARTED 4, ACTIVITY_TASK_COMPLETED 4, WORKFLOW_EXECUTION_COMPLETED 1

## C · Cambiar el código de un workflow que está a medias

(Antes se quita worker2: en C solo hay el worker que se dice en cada paso.)

pedido-C empezó con la versión 1 (worker3) y worker3 murió en la factura. Ahora arranca worker4 con la VERSIÓN 2: un paso nuevo, Validar, entre Reservar y Cobrar, agregado sin más.

worker4, 12 s después: 4 líneas con «nondeterministic» en su log. La primera (recortada a 600 caracteres):
```
2026/11/23 11:57:56 ERROR Workflow panic Namespace default TaskQueue pedidos WorkerID worker4 WorkflowType Pedido WorkflowID pedido-C RunID 01a20844-e371-798f-8e93-06fdb9b31c6f Attempt 1 Error [TMPRL1100] nondeterministic workflow: history event is ActivityTaskScheduled: (ActivityId:11, ActivityType:(Name:Cobrar), TaskQueue:(Name:pedidos, Kind:Normal), Header:(Fields:map[]), Input:(Payloads:[len=1]), ScheduleToCloseTimeout:(Seconds:0, Nanos:0), ScheduleToStartTimeout:(Seconds:0, Nanos:0), StartToCloseTimeout:(Seconds:30, Nanos:0), HeartbeatTimeout:(Seconds:3, Nanos:0), WorkflowTaskCompletedEve
```

Estado de pedido-C: WORKFLOW_EXECUTION_STATUS_RUNNING · efectos de pedido-C hasta aquí: reservar: 1 · cobrar: 1 · validar: 0 · páginas de factura: 6 (1, 2, 3, 4, 5, 6) · enviar: 0

Se quita worker4 y arranca worker5 con la VERSIÓN 3: el mismo paso, detrás de `workflow.GetVersion(ctx, "validar-direccion", workflow.DefaultVersion, 1)`.

worker5:
```
[worker5] worker listo, versión 3 del workflow
11:58:06.265 [worker5] Enviar (intento 1) envío pedido-C
11:58:06.814 [worker5] Reservar (intento 1) reservo el stock de pedido-D
11:58:07.822 [worker5] Validar (intento 1) valido la dirección de pedido-D
11:58:07.830 [worker5] Cobrar (intento 1) la pasarela no responde
2026/11/23 11:58:07 ERROR Activity error. Namespace default TaskQueue pedidos WorkerID worker5 WorkflowID pedido-D RunID 01a20845-43d6-77e7-858d-95fe19174546 ActivityType Cobrar Attempt 1 Error pasarela de pago: tiempo agotado
11:58:08.835 [worker5] Cobrar (intento 2) la pasarela no responde
2026/11/23 11:58:08 ERROR Activity error. Namespace default TaskQueue pedidos WorkerID worker5 WorkflowID pedido-D RunID 01a20845-43d6-77e7-858d-95fe19174546 ActivityType Cobrar Attempt 2 Error pasarela de pago: tiempo agotado
11:58:10.843 [worker5] Cobrar (intento 3) cobro pedido-D
11:58:10.854 [worker5] Facturar (intento 1) empiezo en la página 1 de 6
11:58:11.856 [worker5] Facturar (intento 1) página 1 de 6
11:58:12.862 [worker5] Facturar (intento 1) página 2 de 6
11:58:13.862 [worker5] Facturar (intento 1) página 3 de 6
11:58:14.865 [worker5] Facturar (intento 1) página 4 de 6
11:58:15.867 [worker5] Facturar (intento 1) página 5 de 6
11:58:16.869 [worker5] Facturar (intento 1) página 6 de 6
11:58:16.876 [worker5] Enviar (intento 1) envío pedido-D
```

- pedido-C (empezado con la v1, terminado con la v3): WORKFLOW_EXECUTION_STATUS_COMPLETED · reservar: 1 · cobrar: 1 · validar: 0 · páginas de factura: 6 (1, 2, 3, 4, 5, 6) · enviar: 1
- pedido-D (nuevo, entero con la v3): WORKFLOW_EXECUTION_STATUS_COMPLETED · reservar: 1 · cobrar: 1 · validar: 1 · páginas de factura: 6 (1, 2, 3, 4, 5, 6) · enviar: 1

Contenedores y red borrados al terminar: gl-prac-temporal, gl-prac-tmp-sin1, gl-prac-tmp-w1, gl-prac-tmp-w2, gl-prac-tmp-w3, gl-prac-tmp-w4, gl-prac-tmp-w5, red gl-prac-temporal.
