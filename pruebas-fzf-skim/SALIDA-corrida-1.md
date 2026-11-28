# fzf (Go) contra skim (Rust) · corrida 1

> Corrida del 2026-11-26 21:41. Cada bloque, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_fzf_skim.py 1`.

- Máquina (docker info): 24 CPU · 16487464960 B · 6.18.33.2-microsoft-standard-WSL2 · 29.6.2
- Contenedor: `--cpus 2 --memory 1g`, imagen gl-grb-fzf:1 (debian:trixie-slim + binarios oficiales con su suma)

## 0 · Versiones, binarios y listas

```
0.75.0 (c7bf2b8e)
sk 5.9.0
-rwxr-xr-x 1  501 staff 5537952 Nov  7 12:00 /usr/local/bin/fzf
-rwxr-xr-x 1 root root  9179624 Nov 27 02:38 /usr/local/bin/sk
24
7756ae138e4ecd535625ce906fc9656f2e0d31a8c7ed886dc0bd9774d3a3faef  /datos/rutas-1000000.txt
f755076090316e085a291bd649f69176c817bed8a6820c14d98f148b0be7cdf4  /datos/rutas-3000000.txt
edf5e8e15a3869b3a8201c01852bd71d19703f0c139e136f847aaf5f847d666c  /datos/rutas-5000000.txt
  1000000 /datos/rutas-1000000.txt
  3000000 /datos/rutas-3000000.txt
  5000000 /datos/rutas-5000000.txt
  9000000 total
```

## A · ¿Devuelven lo mismo? (mismo conjunto de líneas, 1 millón)

- «authhandler»: fzf 5.991 líneas (conjunto f21e7dd22386c983) · sk 5.991 líneas (conjunto f21e7dd22386c983) · sk-skim_v2 5.991 líneas (conjunto f21e7dd22386c983)
  - top 3 fzf: ['auth/handlers/main.js', 'auth/handlers/main.js', 'auth/handlers/user.rs']
  - top 3 sk:  ['auth/handlers/users/repository_716.proto', 'auth/handlers/i18n/metrics.css', 'auth/handlers/notifications/notifications/utils/tsconfig.sql']
  - top 10 en común fzf∩sk: 0 de 10
- «src main.go»: fzf 13.880 líneas (conjunto 673583a6d3486663) · sk 13.880 líneas (conjunto 673583a6d3486663) · sk-skim_v2 13.880 líneas (conjunto 673583a6d3486663)
  - top 3 fzf: ['cli/src/main.go', 'backend/src/main.go', 'frontend/src/main.go']
  - top 3 sk:  ['cli/src/main.go', 'api/src/security/main.go', 'web/src/dashboard/server/handlers/main.go']
  - top 10 en común fzf∩sk: 4 de 10
- «usrctrl»: fzf 47.506 líneas (conjunto 96d82d7122533e81) · sk 47.506 líneas (conjunto 96d82d7122533e81) · sk-skim_v2 47.506 líneas (conjunto 96d82d7122533e81)
  - top 3 fzf: ['platform/users/src/test/order.lock', 'platform/utils/src/test/routes/helpers.sh', 'admin/config/users/src/templates/src/logging/eslint.lock']
  - top 3 sk:  ['sdk/users/controllers/cache/lib/generated/readme.py', 'sdk/users/controllers/static/storage/templates/invoice.jsx', 'api/users/controllers/build/storage/makefile.spec.ts']
  - top 10 en común fzf∩sk: 0 de 10
- «payments/invoices»: fzf 4.423 líneas (conjunto 6d1be1ea60e08952) · sk 4.423 líneas (conjunto 6d1be1ea60e08952) · sk-skim_v2 4.423 líneas (conjunto 6d1be1ea60e08952)
  - top 3 fzf: ['payments/invoices/go.tsx', 'payments/invoices/app.js', 'payments/invoices/form.js']
  - top 3 sk:  ['payments/invoices/client.go', 'payments/invoices/routes.tsx', 'payments/invoices/node_modules/events/logging/api_232.html']
  - top 10 en común fzf∩sk: 0 de 10

## B · Tiempo hasta los resultados, RAM y CPU (mediana de 5; --filter, salida a archivo)

| líneas | consulta | herramienta | coincidencias | s (mediana) | s (mín–máx) | RAM máx. MiB | CPU s (u+s) |
|---|---|---|---|---|---|---|---|
| 1.000.000 | authhandler | fzf | rc [0] | 0.09 | 0.09–0.10 | 112.6 | 0.20 |
| 1.000.000 | authhandler | sk | rc [0] | 0.23 | 0.19–0.25 | 139.7 | 0.39 |
| 1.000.000 | authhandler | sk-skim_v2 | rc [0] | 0.31 | 0.30–0.32 | 152.6 | 0.50 |
| 1.000.000 | src main.go | fzf | rc [0] | 0.25 | 0.20–0.29 | 119.0 | 0.53 |
| 1.000.000 | src main.go | sk | rc [0] | 0.82 | 0.78–0.86 | 152.4 | 1.01 |
| 1.000.000 | src main.go | sk-skim_v2 | rc [0] | 0.86 | 0.84–0.89 | 157.8 | 1.09 |
| 1.000.000 | usrctrl | fzf | rc [0] | 0.16 | 0.14–0.17 | 121.8 | 0.31 |
| 1.000.000 | usrctrl | sk | rc [0] | 0.36 | 0.34–0.38 | 146.8 | 0.57 |
| 1.000.000 | usrctrl | sk-skim_v2 | rc [0] | 0.46 | 0.43–0.50 | 159.3 | 0.68 |
| 1.000.000 | payments/invoices | fzf | rc [0] | 0.10 | 0.10–0.13 | 112.6 | 0.22 |
| 1.000.000 | payments/invoices | sk | rc [0] | 0.27 | 0.22–0.35 | 139.9 | 0.50 |
| 1.000.000 | payments/invoices | sk-skim_v2 | rc [0] | 0.34 | 0.33–0.37 | 149.2 | 0.60 |
| 3.000.000 | authhandler | fzf | rc [0] | 0.33 | 0.29–0.36 | 325.6 | 0.60 |
| 3.000.000 | authhandler | sk | rc [0] | 0.78 | 0.75–0.91 | 387.4 | 1.35 |
| 3.000.000 | authhandler | sk-skim_v2 | rc [0] | 1.06 | 0.99–1.26 | 418.9 | 1.68 |
| 3.000.000 | src main.go | fzf | rc [0] | 0.81 | 0.79–0.84 | 345.8 | 1.55 |
| 3.000.000 | src main.go | sk | rc [0] | 2.54 | 2.45–2.57 | 440.0 | 3.15 |
| 3.000.000 | src main.go | sk-skim_v2 | rc [0] | 2.76 | 2.72–2.89 | 438.4 | 3.41 |
| 3.000.000 | usrctrl | fzf | rc [0] | 0.53 | 0.51–0.54 | 352.8 | 0.83 |
| 3.000.000 | usrctrl | sk | rc [0] | 1.06 | 1.00–1.15 | 432.7 | 1.67 |
| 3.000.000 | usrctrl | sk-skim_v2 | rc [0] | 1.33 | 1.29–1.36 | 431.1 | 1.90 |
| 3.000.000 | payments/invoices | fzf | rc [0] | 0.29 | 0.27–0.39 | 324.8 | 0.57 |
| 3.000.000 | payments/invoices | sk | rc [0] | 0.77 | 0.70–0.80 | 387.4 | 1.32 |
| 3.000.000 | payments/invoices | sk-skim_v2 | rc [0] | 1.09 | 1.00–1.11 | 416.6 | 1.73 |
| 5.000.000 | authhandler | fzf | rc [0] | 0.59 | 0.59–0.65 | 539.1 | 1.00 |
| 5.000.000 | authhandler | sk | rc [0] | 1.29 | 1.19–1.32 | 638.2 | 2.30 |
| 5.000.000 | authhandler | sk-skim_v2 | rc [0] | 1.77 | 1.65–1.83 | 668.7 | 2.78 |
| 5.000.000 | src main.go | fzf | rc [0] | 1.72 | 1.40–2.39 | 572.0 | 3.33 |
| 5.000.000 | src main.go | sk | rc [0] | 3.97 | 3.88–4.18 | 704.5 | 5.02 |
| 5.000.000 | src main.go | sk-skim_v2 | rc [0] | 4.20 | 4.15–4.37 | 727.7 | 5.14 |
| 5.000.000 | usrctrl | fzf | rc [0] | 1.37 | 0.89–1.38 | 587.3 | 2.61 |
| 5.000.000 | usrctrl | sk | rc [0] | 1.69 | 1.63–1.72 | 684.9 | 2.60 |
| 5.000.000 | usrctrl | sk-skim_v2 | rc [0] | 2.05 | 2.03–2.13 | 713.0 | 3.08 |
| 5.000.000 | payments/invoices | fzf | rc [0] | 0.70 | 0.60–1.30 | 537.6 | 1.30 |
| 5.000.000 | payments/invoices | sk | rc [0] | 1.21 | 1.17–1.26 | 640.4 | 2.08 |
| 5.000.000 | payments/invoices | sk-skim_v2 | rc [0] | 1.66 | 1.63–1.71 | 692.8 | 2.61 |

## C · ¿En qué puesto sale la ruta que buscabas? (1 millón, 30 rutas al azar, semilla 7)

Consulta «fragmentos»: `proyecto + 4 letras de la última carpeta + nombre del archivo`. Consulta «abreviado»: `3 letras del proyecto + 3 de la carpeta + nombre sin extensión`, sin espacios.

### fragmentos (30 búsquedas)

| herramienta | 1.º | top 10 | top 100 | no aparece | mediana del puesto |
|---|---|---|---|---|---|
| fzf | 20 | 30 | 30 | 0 | 1.0 |
| sk | 19 | 30 | 30 | 0 | 1.0 |
| sk-skim_v2 | 19 | 30 | 30 | 0 | 1.0 |

### abreviado (30 búsquedas)

| herramienta | 1.º | top 10 | top 100 | no aparece | mediana del puesto |
|---|---|---|---|---|---|
| fzf | 7 | 13 | 22 | 0 | 25.5 |
| sk | 7 | 11 | 24 | 0 | 40.5 |
| sk-skim_v2 | 7 | 11 | 20 | 0 | 26.5 |

Detalle (puesto fzf · sk · sk-skim_v2, y coincidencias de cada uno):

```
frag 'infra metr router.go'                           → 1 · 1 · 1   (n 886 · 886 · 886)   objetivo infra/mocks/dist/metrics/router.go
abre 'infmetrouter'                                   → 4 · 12 · 11   (n 1466 · 1466 · 1466)   objetivo infra/mocks/dist/metrics/router.go
frag 'billing mock config.proto'                      → 1 · 1 · 1   (n 41 · 41 · 41)   objetivo billing/queue/mocks/config.proto
abre 'bilmocconfig'                                   → 35 · 79 · 44   (n 1988 · 1988 · 1988)   objetivo billing/queue/mocks/config.proto
frag 'sdk stor readme.json'                           → 1 · 1 · 1   (n 537 · 537 · 537)   objetivo sdk/cache/storage/readme.json
abre 'sdkstoreadme'                                   → 6 · 6 · 6   (n 1167 · 1167 · 1167)   objetivo sdk/cache/storage/readme.json
frag 'worker dist reducer.html'                       → 1 · 1 · 1   (n 167 · 167 · 167)   objetivo worker/scripts/hooks/vendor/mocks/dist/reducer.html
abre 'wordisreducer'                                  → 19 · 22 · 25   (n 583 · 583 · 583)   objetivo worker/scripts/hooks/vendor/mocks/dist/reducer.html
frag 'infra view modal.js'                            → 1 · 2 · 2   (n 146 · 146 · 146)   objetivo infra/controllers/dist/assets/server/views/modal.js
abre 'infviemodal'                                    → 39 · 69 · 70   (n 1051 · 1051 · 1051)   objetivo infra/controllers/dist/assets/server/views/modal.js
frag 'analytics migr utils_602.json'                  → 1 · 1 · 1   (n 2 · 2 · 2)   objetivo analytics/lib/logging/assets/events/test/server/migrations/utils_602.json
abre 'anamigutils_602'                                → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo analytics/lib/logging/assets/events/test/server/migrations/utils_602.json
frag 'payments metr schema_339_test.go'               → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo payments/i18n/metrics/orders/fixtures/metrics/schema_339_test.go
abre 'paymetschema_339_test'                          → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo payments/i18n/metrics/orders/fixtures/metrics/schema_339_test.go
frag 'analytics stor webpack_360.html'                → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo analytics/components/templates/storage/webpack_360.html
abre 'anastowebpack_360'                              → 2 · 2 · 2   (n 4 · 4 · 4)   objetivo analytics/components/templates/storage/webpack_360.html
frag 'mobile scri errors.js'                          → 5 · 5 · 5   (n 1579 · 1579 · 1579)   objetivo mobile/migrations/generated/src/tests/scripts/errors.js
abre 'mobscrerrors'                                   → 55 · 75 · 85   (n 2710 · 2710 · 2710)   objetivo mobile/migrations/generated/src/tests/scripts/errors.js
frag 'tools node controller.yaml'                     → 1 · 1 · 1   (n 1480 · 1480 · 1480)   objetivo tools/store/node_modules/controller.yaml
abre 'toonodcontroller'                               → 29 · 29 · 28   (n 2463 · 2463 · 2463)   objetivo tools/store/node_modules/controller.yaml
frag 'mobile migr header.tsx'                         → 1 · 1 · 1   (n 115 · 115 · 115)   objetivo mobile/internal/migrations/generated/middleware/assets/migrations/header.tsx
abre 'mobmigheader'                                   → 93 · 52 · 123   (n 236 · 236 · 236)   objetivo mobile/internal/migrations/generated/middleware/assets/migrations/header.tsx
frag 'payments node errors.sum'                       → 3 · 3 · 3   (n 412 · 412 · 412)   objetivo payments/node_modules/core/migrations/vendor/events/node_modules/errors.sum
abre 'paynoderrors'                                   → 143 · 254 · 613   (n 3373 · 3373 · 3373)   objetivo payments/node_modules/core/migrations/vendor/events/node_modules/errors.sum
frag 'api stor login_583.tsx'                         → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo api/handlers/build/cache/handlers/metrics/store/login_583.tsx
abre 'apistologin_583'                                → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo api/handlers/build/cache/handlers/metrics/store/login_583.tsx
frag 'api asse router.spec.ts'                        → 1 · 1 · 1   (n 1779 · 1779 · 1779)   objetivo api/cache/reports/handlers/core/assets/security/assets/router.spec.ts
abre 'apiassrouter'                                   → 160 · 97 · 220   (n 1613 · 1613 · 1613)   objetivo api/cache/reports/handlers/core/assets/security/assets/router.spec.ts
frag 'infra buil modal.js'                            → 7 · 7 · 7   (n 330 · 330 · 330)   objetivo infra/vendor/utils/lib/orders/settings/server/build/modal.js
abre 'infbuimodal'                                    → 62 · 56 · 92   (n 326 · 326 · 326)   objetivo infra/vendor/utils/lib/orders/settings/server/build/modal.js
frag 'cli hand utils.tsx'                             → 3 · 3 · 3   (n 915 · 915 · 915)   objetivo cli/reports/security/internal/test/lib/proto/handlers/utils.tsx
abre 'clihanutils'                                    → 413 · 232 · 485   (n 1369 · 1369 · 1369)   objetivo cli/reports/security/internal/test/lib/proto/handlers/utils.tsx
frag 'search buil profile.go'                         → 1 · 1 · 1   (n 43 · 43 · 43)   objetivo search/metrics/templates/test/build/profile.go
abre 'seabuiprofile'                                  → 22 · 72 · 23   (n 410 · 410 · 410)   objetivo search/metrics/templates/test/build/profile.go
frag 'worker vend payment.json'                       → 3 · 3 · 3   (n 14 · 14 · 14)   objetivo worker/build/middleware/orders/fixtures/views/models/vendor/payment.json
abre 'worvenpayment'                                  → 105 · 83 · 134   (n 403 · 403 · 403)   objetivo worker/build/middleware/orders/fixtures/views/models/vendor/payment.json
frag 'auth cach api.css'                              → 1 · 1 · 1   (n 447 · 447 · 447)   objetivo auth/orders/cache/api.css
abre 'autcacapi'                                      → 5 · 5 · 5   (n 1086 · 1086 · 1086)   objetivo auth/orders/cache/api.css
frag 'frontend vend types.json'                       → 2 · 2 · 2   (n 17 · 17 · 17)   objetivo frontend/settings/client/logging/config/store/vendor/types.json
abre 'froventypes'                                    → 70 · 58 · 104   (n 647 · 647 · 647)   objetivo frontend/settings/client/logging/config/store/vendor/types.json
frag 'frontend orde constants_test.go'                → 2 · 2 · 2   (n 109 · 109 · 109)   objetivo frontend/storage/cache/tests/orders/constants_test.go
abre 'froordconstants_test'                           → 2 · 2 · 2   (n 29 · 29 · 29)   objetivo frontend/storage/cache/tests/orders/constants_test.go
frag 'platform test prettier.jsx'                     → 3 · 3 · 3   (n 535 · 535 · 535)   objetivo platform/handlers/internal/dashboard/routes/build/tests/prettier.jsx
abre 'platesprettier'                                 → 1385 · 1065 · 1532   (n 2798 · 2798 · 2798)   objetivo platform/handlers/internal/dashboard/routes/build/tests/prettier.jsx
frag 'docs stat model_325.ts'                         → 1 · 1 · 1   (n 5 · 5 · 5)   objetivo docs/metrics/tests/events/lib/static/model_325.ts
abre 'docstamodel_325'                                → 1 · 1 · 1   (n 2 · 2 · 2)   objetivo docs/metrics/tests/events/lib/static/model_325.ts
frag 'cli lib order.txt'                              → 2 · 2 · 2   (n 1944 · 1944 · 1944)   objetivo cli/node_modules/generated/config/lib/order.txt
abre 'cliliborder'                                    → 229 · 260 · 244   (n 2347 · 2347 · 2347)   objetivo cli/node_modules/generated/config/lib/order.txt
frag 'scheduler dash layout_191.yml'                  → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo scheduler/config/middleware/users/test/models/dashboard/layout_191.yml
abre 'schdaslayout_191'                               → 1 · 1 · 1   (n 2 · 2 · 2)   objetivo scheduler/config/middleware/users/test/models/dashboard/layout_191.yml
frag 'tools orde table.yaml'                          → 1 · 1 · 1   (n 1080 · 1080 · 1080)   objetivo tools/queue/services/routes/fixtures/services/fixtures/orders/table.yaml
abre 'tooordtable'                                    → 121 · 123 · 263   (n 1470 · 1470 · 1470)   objetivo tools/queue/services/routes/fixtures/services/fixtures/orders/table.yaml
frag 'worker test client.txt'                         → 1 · 1 · 1   (n 357 · 357 · 357)   objetivo worker/test/client.txt
abre 'wortesclient'                                   → 2 · 11 · 11   (n 3920 · 3920 · 3920)   objetivo worker/test/client.txt
frag 'mobile fixt eslint_10.go'                       → 1 · 1 · 1   (n 8 · 8 · 8)   objetivo mobile/fixtures/eslint_10.go
abre 'mobfixeslint_10'                                → 1 · 1 · 1   (n 9 · 9 · 9)   objetivo mobile/fixtures/eslint_10.go
frag 'search hand auth_597.html'                      → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo search/tests/users/dashboard/cmd/client/fixtures/handlers/auth_597.html
abre 'seahanauth_597'                                 → 1 · 1 · 1   (n 1 · 1 · 1)   objetivo search/tests/users/dashboard/cmd/client/fixtures/handlers/auth_597.html
frag 'gateway test user.html'                         → 3 · 3 · 3   (n 696 · 696 · 696)   objetivo gateway/users/controllers/fixtures/test/user.html
abre 'gattesuser'                                     → 409 · 1263 · 714   (n 12404 · 12404 · 12404)   objetivo gateway/users/controllers/fixtures/test/user.html
```

FIN
