# SALIDA-extra — el porqué de tres cifras (golandia 866)

> Generado por `banco_extra.py`. **No se escribe a mano.** Mismo montaje que SALIDA.md.

## A · Conexiones nuevas a Postgres y CPU del servidor, 10 s de «leer una tarea por id» a tope (`hey -c 16`)

| variante | pet/s | conexiones nuevas en 10 s | CPU del servidor (100 % = 1 núcleo) |
|---|---|---|---|
| Gin + GORM (pool por defecto) | 1.382 | 3716 | 48 % (mediana de 4) |
| Gin + GORM (POOL=10) | 67.128 | 0 | 183 % (mediana de 4) |
| Serverpod | 10.507 | 0 | 112 % (mediana de 4) |

## B · Serverpod sin ninguna petición: lo que manda a Postgres en 65 s (log_statement=all)

```
2×  DISCARD ALL
2×  SELECT "serverpod_health_connection_info"."id" AS "serverpod_health_connection_info.id", "serverpod_health_connection_info"."serverId" AS "serverpod_h
2×  SELECT "serverpod_health_metric"."id" AS "serverpod_health_metric.id", "serverpod_health_metric"."name" AS "serverpod_health_metric.name", "serverpod_
1×  SELECT 1;
1×  INSERT INTO "serverpod_health_metric" ("name", "serverId", "timestamp", "isHealthy", "value", "granularity") VALUES ('serverpod_database', 'default', 
1×  INSERT INTO "serverpod_health_metric" ("name", "serverId", "timestamp", "isHealthy", "value", "granularity") VALUES ('serverpod_cpu', 'default', '2026
1×  INSERT INTO "serverpod_health_metric" ("name", "serverId", "timestamp", "isHealthy", "value", "granularity") VALUES ('serverpod_memory', 'default', '2
1×  INSERT INTO "serverpod_health_connection_info" ("serverId", "timestamp", "active", "closing", "idle", "granularity") VALUES ('default', '2026-11-17T14
1×  SELECT "serverpod_runtime_settings"."id" AS "serverpod_runtime_settings.id", "serverpod_runtime_settings"."logSettings" AS "serverpod_runtime_settings
1×  UPDATE serverpod_session_log SET touched = '"2026-11-17T14:40:00.004863Z"' WHERE "serverId" = 'default' AND "isOpen" = TRUE AND "time" > '"2026-11-17T
1×  UPDATE serverpod_session_log SET "isOpen" = FALSE WHERE "isOpen" = TRUE AND "touched" < '"2026-11-17T14:37:00.004870Z"'
total: 14 consultas en 65 s
```
