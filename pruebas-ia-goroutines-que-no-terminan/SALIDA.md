# SALIDA — 705 «Goroutines que no terminan: el código concurrente de la IA con el perfil goroutineleak»

> Regenerado el 2026-10-29 09:37 con `python ejecutar_todo.py`. Medido en `golang:1.27.2` (Docker, --cpus 2), go.mod `go 1.27`.
> No es un estudio: 12 tareas propias con su prueba fija, una sola máquina, un modelo local y una muestra pequeña de Claude Code.

## Quién generó
- Local: Ollama `qwen2.5-coder:7b`, digest `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364`, `/api/generate`, temperature 0.7, seeds 1-4, num_predict 2048.
- Claude Code 2.1.312, `claude -p` desde un directorio temporal vacío, flags `-p --safe-mode --tools  --strict-mcp-config --permission-prompts none --no-session-persistence --output-format json --max-budget-usd 0.50`. Modelos en modelUsage: {'claude-haiku-5-5': 24, 'claude-opus-5-5': 24}. Costo de las 24 llamadas: 0.55 USD. Sin seed: una muestra por tarea y variante.
- `--bare` no se pudo usar (solo ANTHROPIC_API_KEY; la cuenta entra por OAuth): se usó `--safe-mode` y `--tools ""`.
- Prompts: prompts.json (A sin aviso · B añade «Asegúrate de que no quede ninguna goroutine bloqueada cuando la función termine.»). Tareas: tareas.json. Pruebas: pruebas/.

## Cómo se activa el perfil (Go 1.27)
- go.dev/doc/go1.27 (fuente-go1.27.html): «previously available as an experiment in Go 1.26, is now generally available» · «supported in the runtime/pprof package» · «also available as the net/http/pprof endpoint /debug/pprof/goroutineleak» · «The goroutineleakprofile GOEXPERIMENT setting is now deleted». Límite: «may fail to identify leaks caused by blocking on concurrency primitives reachable through global variables or the local variables of runnable goroutines».
- En la prueba: `pprof.Lookup("goroutineleak").WriteTo(&b, 1)` tras el escenario (pruebas/ayuda_test.go). Sin flags ni GOEXPERIMENT.

## Controles del banco
- control-bien c01-worker-pool: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c02-con-timeout: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c03-con-contexto: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c04-el-mas-rapido: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c05-fan-out-fan-in: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c06-generador: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c07-etapa-pipeline: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c08-unir-canales: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c09-primer-error: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c10-cache-aciertos: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c11-contar-primos: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-bien c12-consumidores: compila=True termina=True correcto=True fugas=0 races=0 vet=True
- control-mal c02-con-timeout: compila=True termina=True correcto=True fugas=3 races=0 vet=True
- control-mal c04-el-mas-rapido: compila=True termina=True correcto=True fugas=2 races=0 vet=True
- control-mal c06-generador: compila=True termina=True correcto=True fugas=1 races=0 vet=True
- control-mal c09-primer-error: compila=True termina=True correcto=True fugas=9 races=0 vet=True
- control-mal c10-cache-aciertos: compila=True termina=True correcto=True fugas=0 races=2 vet=True
- control-mal c11-contar-primos: compila=True termina=True correcto=False fugas=0 races=2 vet=True
- caso-ia c04-el-mas-rapido: compila=True termina=True correcto=True fugas=2 races=0 vet=True
- caso-arreglado c04-el-mas-rapido: compila=True termina=True correcto=True fugas=0 races=0 vet=True

## Tabla principal (por modelo y prompt)
| modelo | prompt | muestras | compila | go vet limpio* | termina la prueba | correctas | con fuga | goroutines fugadas | con data race* | sin ningún problema |
|---|---|---|---|---|---|---|---|---|---|---|
| claude-code | A | 12 | 12 | 12/12 | 12 | 12 | 0 | 0 | 0/12 | 12/12 |
| claude-code | B | 12 | 12 | 12/12 | 12 | 12 | 0 | 0 | 0/12 | 12/12 |
| qwen2.5-coder_7b | A | 48 | 37 | 37/37 | 30 | 26 | 7 | 21 | 5/37 | 19/48 |
| qwen2.5-coder_7b | B | 48 | 33 | 33/33 | 27 | 22 | 4 | 9 | 6/33 | 17/48 |

*sobre las que compilan. «con fuga» y «goroutines fugadas» incluyen las que se cuelgan (su goroutine atascada también cuenta).

## Estado de cada muestra
- claude-code A: {'termina': 12}
- claude-code B: {'termina': 12}
- qwen2.5-coder_7b A: {'termina': 30, 'se cuelga': 3, 'no compila': 11, 'panic': 3, 'no cumple la firma': 1}
- qwen2.5-coder_7b B: {'no compila': 15, 'termina': 27, 'panic': 5, 'no cumple la firma': 1}

## Las que pasan su prueba y aun así fugan
- qwen2.5-coder_7b/A/c03-con-contexto/s2: fugas=3
- qwen2.5-coder_7b/A/c04-el-mas-rapido/s1: fugas=2
- qwen2.5-coder_7b/A/c05-fan-out-fan-in/s3: fugas=1
- qwen2.5-coder_7b/A/c12-consumidores/s4: fugas=2
- qwen2.5-coder_7b/B/c03-con-contexto/s1: fugas=3
- qwen2.5-coder_7b/B/c04-el-mas-rapido/s1: fugas=2
- qwen2.5-coder_7b/B/c05-fan-out-fan-in/s1: fugas=3
- qwen2.5-coder_7b/B/c05-fan-out-fan-in/s2: fugas=1

## Por tarea (local A+B: con fuga / con race / panic; Claude A+B: con problema)
| tarea | patrón | local con fuga | local con race | local panic | Claude con algún problema |
|---|---|---|---|---|---|
| c01-worker-pool | worker pool con cancelación | 1/8 | 0/8 | 0/8 | 0/2 |
| c02-con-timeout | timeout | 1/8 | 0/8 | 0/8 | 0/2 |
| c03-con-contexto | timeout con context | 2/8 | 1/8 | 0/8 | 0/2 |
| c04-el-mas-rapido | el primero que responde | 2/8 | 0/8 | 0/8 | 0/2 |
| c05-fan-out-fan-in | fan-out/fan-in | 4/8 | 2/8 | 2/8 | 0/2 |
| c06-generador | productor | 0/8 | 0/8 | 0/8 | 0/2 |
| c07-etapa-pipeline | pipeline | 0/8 | 0/8 | 0/8 | 0/2 |
| c08-unir-canales | fan-in de canales | 0/8 | 0/8 | 5/8 | 0/2 |
| c09-primer-error | primer error | 0/8 | 0/8 | 0/8 | 0/2 |
| c10-cache-aciertos | estado compartido | 0/8 | 6/8 | 0/8 | 0/2 |
| c11-contar-primos | contador compartido | 0/8 | 2/8 | 1/8 | 0/2 |
| c12-consumidores | productor/consumidor | 1/8 | 0/8 | 0/8 | 0/2 |

## El caso del vídeo: c04 ElMasRapido, qwen2.5-coder:7b, prompt A, seed 1
### caso-ia (fugas=2, correcto=True)
```
goroutineleak profile: total 2
1 @ 0x4866ea 0x41531c 0x414f17 0x545aac 0x48cce1
#	0x545aab	tarea.ElMasRapido.func1+0x6b	/w/m/m0138/tarea.go:15

1 @ 0x4866ea 0x461c32 0x461c09 0x487bae 0x491ca5 0x5459e5 0x48cce1
#	0x487bad	sync.runtime_SemacquireWaitGroup+0x2d	/usr/local/go/src/runtime/sema.go:114
#	0x491ca4	sync.(*WaitGroup).Wait+0x84		/usr/local/go/src/sync/waitgroup.go:206
#	0x5459e4	tarea.ElMasRapido.func2+0x24		/w/m/m0138/tarea.go:20
```
### caso-arreglado (fugas=0, correcto=True)
```
sin fugas
```
Código: demos/caso/ia/tarea.go (copia literal de respuestas/qwen2.5-coder_7b/A/c04-el-mas-rapido/s1) y demos/caso/arreglado/tarea.go.

## Pilas de las fugas (local)
### qwen2.5-coder_7b/A/c01-worker-pool/s2 (fugas=5, termina=False)
```
goroutineleak profile: total 5
4 @ 0x4866ea 0x41531c 0x414f17 0x545e5e 0x48cce1
#	0x545e5d	tarea.ProcesarConWorkers.func1+0xfd	/w/m/m0025/tarea.go:23

1 @ 0x4866ea 0x41531c 0x414f17 0x545431 0x545fdd 0x545f82 0x48cce1
#	0x545430	tarea.ProcesarConWorkers+0x130	/w/m/m0025/tarea.go:16
#	0x545fdc	tarea.escenarioC01.func3+0x3c	/w/m/m0025/c01-worker-pool_test.go:24
#	0x545f81	tarea.termina.func1+0x21	/w/m/m0025/ayuda_test.go:42
```
### qwen2.5-coder_7b/A/c02-con-timeout/s1 (fugas=4, termina=False)
```
goroutineleak profile: total 4
2 @ 0x4866ea 0x41531c 0x414f17 0x545b0f 0x48cce1
#	0x545b0e	tarea.ConTimeout.func1+0xee	/w/m/m0028/tarea.go:22

1 @ 0x4866ea 0x461c32 0x461c09 0x487bae 0x491ca5 0x54537c 0x545c26 0x545bc2 0x48cce1
#	0x487bad	sync.runtime_SemacquireWaitGroup+0x2d	/usr/local/go/src/runtime/sema.go:114
#	0x491ca4	sync.(*WaitGroup).Wait+0x84		/usr/local/go/src/sync/waitgroup.go:206
#	0x54537b	tarea.ConTimeout+0x11b			/w/m/m0028/tarea.go:26
#	0x545c25	tarea.escenarioC02.func2+0x45		/w/m/m0028/c02-con-timeout_test.go:18
#	0x545bc1	tarea.termina.func1+0x21		/w/m/m0028/ayuda_test.go:42

1 @ 0x4866ea 0x461c32 0x461c09 0x487bae 0x491ca5 0x54537c 0x545c91 0x545bc2 0x48cce1
#	0x487bad	sync.runtime_SemacquireWaitGroup+0x2d	/usr/local/go/src/runtime/sema.go:114
#	0x491ca4	sync.(*WaitGroup).Wait+0x84		/usr/local/go/src/sync/waitgroup.go:206
#	0x54537b	tarea.ConTimeout+0x11b			/w/m/m0028/tarea.go:26
#	0x545c90	tarea.escenarioC02.func1+0x30		/w/m/m0028/c02-con-timeout_test.go:11
#	0x545bc1	tarea.termina.func1+0x21		/w/m/m0028/ayuda_test.go:42
```
### qwen2.5-coder_7b/A/c03-con-contexto/s2 (fugas=3, termina=True)
```
goroutineleak profile: total 3
3 @ 0x4866ea 0x41531c 0x414f17 0x54661a 0x48cc21
#	0x546619	tarea.LlamarConContexto.func1+0x79	/w/m/m0033/tarea.go:16
```
### qwen2.5-coder_7b/A/c04-el-mas-rapido/s1 (fugas=2, termina=True)
```
goroutineleak profile: total 2
1 @ 0x4866ea 0x41531c 0x414f17 0x545aac 0x48cce1
#	0x545aab	tarea.ElMasRapido.func1+0x6b	/w/m/m0036/tarea.go:15

1 @ 0x4866ea 0x461c32 0x461c09 0x487bae 0x491ca5 0x5459e5 0x48cce1
#	0x487bad	sync.runtime_SemacquireWaitGroup+0x2d	/usr/local/go/src/runtime/sema.go:114
#	0x491ca4	sync.(*WaitGroup).Wait+0x84		/usr/local/go/src/sync/waitgroup.go:206
#	0x5459e4	tarea.ElMasRapido.func2+0x24		/w/m/m0036/tarea.go:20
```
### qwen2.5-coder_7b/A/c05-fan-out-fan-in/s3 (fugas=1, termina=True)
```
goroutineleak profile: total 1
1 @ 0x48670a 0x41531c 0x414f17 0x545c05 0x48cd01
#	0x545c04	tarea.FanOutFanIn.func3+0x44	/w/m/m0042/tarea.go:41
```
### qwen2.5-coder_7b/A/c05-fan-out-fan-in/s4 (fugas=4, termina=False)
```
goroutineleak profile: total 4
2 @ 0x48670a 0x41531c 0x414f17 0x545aff 0x48cd01
#	0x545afe	tarea.FanOutFanIn.func1+0xfe	/w/m/m0043/tarea.go:19

1 @ 0x48670a 0x41624e 0x415d72 0x545f52 0x545ee2 0x48cd01
#	0x545f51	tarea.escenarioC05.func2+0x51	/w/m/m0043/c05-fan-out-fan-in_test.go:24
#	0x545ee1	tarea.termina.func1+0x21	/w/m/m0043/ayuda_test.go:42

1 @ 0x48670a 0x41624e 0x415d92 0x546056 0x545ee2 0x48cd01
#	0x546055	tarea.escenarioC05.func1+0xd5	/w/m/m0043/c05-fan-out-fan-in_test.go:12
#	0x545ee1	tarea.termina.func1+0x21	/w/m/m0043/ayuda_test.go:42
```
### qwen2.5-coder_7b/A/c12-consumidores/s4 (fugas=2, termina=True)
```
goroutineleak profile: total 2
1 @ 0x4866ea 0x41531c 0x414f17 0x545cc5 0x48cce1
#	0x545cc4	tarea.Consumir.func1+0x144	/w/m/m0071/tarea.go:19

1 @ 0x4866ea 0x461c32 0x461c09 0x487bae 0x491ca5 0x545b65 0x48cce1
#	0x487bad	sync.runtime_SemacquireWaitGroup+0x2d	/usr/local/go/src/runtime/sema.go:114
#	0x491ca4	sync.(*WaitGroup).Wait+0x84		/usr/local/go/src/sync/waitgroup.go:206
#	0x545b64	tarea.Consumir.func2+0x24		/w/m/m0071/tarea.go:32
```
### qwen2.5-coder_7b/B/c03-con-contexto/s1 (fugas=3, termina=True)
```
goroutineleak profile: total 3
3 @ 0x4864aa 0x41531c 0x414f17 0x54616d 0x4906eb 0x5460fb 0x5460b3 0x48c941
#	0x54616c	tarea.LlamarConContexto.func1.1+0x4c	/w/m/m0080/tarea.go:19
#	0x4906ea	sync.(*Once).doSlow+0xaa		/usr/local/go/src/sync/once.go:78
#	0x5460fa	sync.(*Once).Do+0x7a			/usr/local/go/src/sync/once.go:69
#	0x5460b2	tarea.LlamarConContexto.func1+0x32	/w/m/m0080/tarea.go:15
```
### qwen2.5-coder_7b/B/c04-el-mas-rapido/s1 (fugas=2, termina=True)
```
goroutineleak profile: total 2
1 @ 0x4866ea 0x41531c 0x414f17 0x545aac 0x48cce1
#	0x545aab	tarea.ElMasRapido.func1+0x6b	/w/m/m0084/tarea.go:15

1 @ 0x4866ea 0x461c32 0x461c09 0x487bae 0x491ca5 0x5459e5 0x48cce1
#	0x487bad	sync.runtime_SemacquireWaitGroup+0x2d	/usr/local/go/src/runtime/sema.go:114
#	0x491ca4	sync.(*WaitGroup).Wait+0x84		/usr/local/go/src/sync/waitgroup.go:206
#	0x5459e4	tarea.ElMasRapido.func2+0x24		/w/m/m0084/tarea.go:20
```
### qwen2.5-coder_7b/B/c05-fan-out-fan-in/s1 (fugas=3, termina=True)
```
goroutineleak profile: total 3
2 @ 0x48670a 0x41531c 0x414f17 0x545e0c 0x48cd01
#	0x545e0b	tarea.FanOutFanIn.func1+0x8b	/w/m/m0088/tarea.go:21

1 @ 0x48670a 0x461c52 0x461c29 0x487bce 0x491cc5 0x545b65 0x48cd01
#	0x487bcd	sync.runtime_SemacquireWaitGroup+0x2d	/usr/local/go/src/runtime/sema.go:114
#	0x491cc4	sync.(*WaitGroup).Wait+0x84		/usr/local/go/src/sync/waitgroup.go:206
#	0x545b64	tarea.FanOutFanIn.func3+0x24		/w/m/m0088/tarea.go:46
```
### qwen2.5-coder_7b/B/c05-fan-out-fan-in/s2 (fugas=1, termina=True)
```
goroutineleak profile: total 1
1 @ 0x48670a 0x41531c 0x414f17 0x545d45 0x48cd01
#	0x545d44	tarea.FanOutFanIn.func1+0x44	/w/m/m0089/tarea.go:16
```

## Races (primera de cada muestra)
### qwen2.5-coder_7b/A/c05-fan-out-fan-in/s2 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Read at 0x00c0000822b0 by goroutine 12:
  runtime.chansend()
      /usr/local/go/src/runtime/chan.go:176 +0x0
  tarea.FanOutFanIn.func1.1()
      /w/m/m0041/tarea.go:24 +0x1b1
  tarea.FanOutFanIn.func1.gowrap2()
      /w/m/m0041/tarea.go:27 +0x57

Previous write at 0x00c0000822b0 by goroutine 10:
  runtime.closechan()
      /usr/local/go/src/runtime/chan.go:414 +0x0
  tarea.FanOutFanIn.func1.deferwrap1()
      /w/m/m0041/tarea.go:14 +0x2e
  runtime.deferreturn()
      /usr/local/go/src/runtime/panic.go:668 +0x53

Goroutine 12 (running) created at:
  tarea.FanOutFanIn.func1()
      /w/m/m0041/tarea.go:18 +0x33d

Goroutine 10 (finished) created at:
  tarea.FanOutFanIn()
      /w/m/m0041/tarea.go:13 +0x1fd
  tarea.escenarioC05.func1()
      /w/m/m0041/c05-fan-out-fan-in_test.go:12 +0x138
  tarea.termina.func1()
      /w/m/m0041/ayuda_test.go:42 +0x33
==
```
### qwen2.5-coder_7b/A/c10-cache-aciertos/s1 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Read at 0x00c000100038 by goroutine 12:
  tarea.(*Cache).Get()
      /w/m/m0060/tarea.go:31 +0x144
  tarea.escenarioC10.func1.1()
      /w/m/m0060/c10-cache-aciertos_test.go:20 +0xcd

Previous write at 0x00c000100038 by goroutine 16:
  tarea.(*Cache).Get()
      /w/m/m0060/tarea.go:31 +0x164
  tarea.escenarioC10.func1.1()
      /w/m/m0060/c10-cache-aciertos_test.go:20 +0xcd

Goroutine 12 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0060/c10-cache-aciertos_test.go:17 +0x13b
  tarea.termina.func1()
      /w/m/m0060/ayuda_test.go:42 +0x33

Goroutine 16 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0060/c10-cache-aciertos_test.go:17 +0x13b
  tarea.termina.func1()
      /w/m/m0060/ayuda_test.go:42 +0x33
==================
==================
WARNING: DATA RACE
Write at 0x00c000100038 by goroutine 17:
  tarea.(*Cache).Get(
```
### qwen2.5-coder_7b/A/c10-cache-aciertos/s3 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Read at 0x00c000082038 by goroutine 17:
  tarea.(*Cache).Get()
      /w/m/m0062/tarea.go:30 +0x144
  tarea.escenarioC10.func1.1()
      /w/m/m0062/c10-cache-aciertos_test.go:20 +0xcd

Previous write at 0x00c000082038 by goroutine 10:
  tarea.(*Cache).Get()
      /w/m/m0062/tarea.go:30 +0x164
  tarea.escenarioC10.func1.1()
      /w/m/m0062/c10-cache-aciertos_test.go:20 +0xcd

Goroutine 17 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0062/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0062/ayuda_test.go:42 +0x33

Goroutine 10 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0062/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0062/ayuda_test.go:42 +0x33
==================
==================
WARNING: DATA RACE
Write at 0x00c000082038 by goroutine 17:
  tarea.(*Cache).Get(
```
### qwen2.5-coder_7b/A/c10-cache-aciertos/s4 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Read at 0x00c00007a050 by goroutine 15:
  tarea.(*Cache).Get()
      /w/m/m0063/tarea.go:30 +0x144
  tarea.escenarioC10.func1.1()
      /w/m/m0063/c10-cache-aciertos_test.go:20 +0xcd

Previous write at 0x00c00007a050 by goroutine 10:
  tarea.(*Cache).Get()
      /w/m/m0063/tarea.go:30 +0x164
  tarea.escenarioC10.func1.1()
      /w/m/m0063/c10-cache-aciertos_test.go:20 +0xcd

Goroutine 15 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0063/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0063/ayuda_test.go:42 +0x33

Goroutine 10 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0063/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0063/ayuda_test.go:42 +0x33
==================
==================
WARNING: DATA RACE
Write at 0x00c00007a050 by goroutine 15:
  tarea.(*Cache).Get(
```
### qwen2.5-coder_7b/A/c11-contar-primos/s1 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Read at 0x00c00009e1b8 by goroutine 11:
  tarea.ContarPrimos.func1()
      /w/m/m0064/tarea.go:36 +0x10d

Previous write at 0x00c00009e1b8 by goroutine 10:
  tarea.ContarPrimos.func1()
      /w/m/m0064/tarea.go:36 +0x124

Goroutine 11 (running) created at:
  tarea.ContarPrimos()
      /w/m/m0064/tarea.go:32 +0xa4
  tarea.escenarioC11.func1()
      /w/m/m0064/c11-contar-primos_test.go:11 +0x54
  tarea.termina.func1()
      /w/m/m0064/ayuda_test.go:42 +0x33

Goroutine 10 (running) created at:
  tarea.ContarPrimos()
      /w/m/m0064/tarea.go:32 +0xa4
  tarea.escenarioC11.func1()
      /w/m/m0064/c11-contar-primos_test.go:11 +0x54
  tarea.termina.func1()
      /w/m/m0064/ayuda_test.go:42 +0x33
==================
==================
WARNING: DATA RACE
Write at 0x00c00009e1b8 by goroutine 17:
  tarea.ContarPrimos.func1()
      /w/m/m0064/tarea.go:36 +0x124
```
### qwen2.5-coder_7b/B/c03-con-contexto/s2 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Write at 0x00c000024080 by goroutine 12:
  tarea.LlamarConContexto.func1()
      /w/m/m0081/tarea.go:15 +0x6c

Previous read at 0x00c000024080 by goroutine 11:
  tarea.LlamarConContexto()
      /w/m/m0081/tarea.go:27 +0x352
  tarea.escenarioC03.func2()
      /w/m/m0081/c03-con-contexto_test.go:23 +0x75
  tarea.termina.func1()
      /w/m/m0081/ayuda_test.go:42 +0x33

Goroutine 12 (running) created at:
  tarea.LlamarConContexto()
      /w/m/m0081/tarea.go:14 +0x244
  tarea.escenarioC03.func2()
      /w/m/m0081/c03-con-contexto_test.go:23 +0x75
  tarea.termina.func1()
      /w/m/m0081/ayuda_test.go:42 +0x33

Goroutine 11 (finished) created at:
  tarea.termina()
      /w/m/m0081/ayuda_test.go:42 +0xe9
  tarea.escenarioC03()
      /w/m/m0081/c03-con-contexto_test.go:20 +0x309
  tarea.TestC03()
      /w/m/m0081/c03-con-contexto_test.go:32 +0x24
  testing.t
```
### qwen2.5-coder_7b/B/c05-fan-out-fan-in/s3 (WARNING: DATA RACE ×1)
```
==================
WARNING: DATA RACE
Read at 0x00c0000222b0 by goroutine 10:
  runtime.chansend()
      /usr/local/go/src/runtime/chan.go:176 +0x0
  tarea.FanOutFanIn.func1()
      /w/m/m0090/tarea.go:24 +0x1ee
  tarea.FanOutFanIn.gowrap1()
      /w/m/m0090/tarea.go:27 +0x38

Previous write at 0x00c0000222b0 by goroutine 13:
  runtime.closechan()
      /usr/local/go/src/runtime/chan.go:414 +0x0
  tarea.FanOutFanIn.func2.deferwrap1()
      /w/m/m0090/tarea.go:32 +0x2e
  runtime.deferreturn()
      /usr/local/go/src/runtime/panic.go:668 +0x53

Goroutine 10 (running) created at:
  tarea.FanOutFanIn()
      /w/m/m0090/tarea.go:18 +0xdb
  tarea.escenarioC05.func1()
      /w/m/m0090/c05-fan-out-fan-in_test.go:12 +0x138
  tarea.termina.func1()
      /w/m/m0090/ayuda_test.go:42 +0x33

Goroutine 13 (finished) created at:
  tarea.FanOutFanIn()
      /w/m/m0090/tarea.go:31 +0x453
  tarea.escenario
```
### qwen2.5-coder_7b/B/c10-cache-aciertos/s2 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Read at 0x00c000100038 by goroutine 13:
  tarea.(*Cache).Get()
      /w/m/m0109/tarea.go:30 +0x144
  tarea.escenarioC10.func1.1()
      /w/m/m0109/c10-cache-aciertos_test.go:20 +0xcd

Previous write at 0x00c000100038 by goroutine 14:
  tarea.(*Cache).Get()
      /w/m/m0109/tarea.go:30 +0x164
  tarea.escenarioC10.func1.1()
      /w/m/m0109/c10-cache-aciertos_test.go:20 +0xcd

Goroutine 13 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0109/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0109/ayuda_test.go:42 +0x33

Goroutine 14 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0109/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0109/ayuda_test.go:42 +0x33
==================
==================
WARNING: DATA RACE
Write at 0x00c000100038 by goroutine 13:
  tarea.(*Cache).Get(
```
### qwen2.5-coder_7b/B/c10-cache-aciertos/s3 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Read at 0x00c0000a0728 by goroutine 11:
  tarea.(*Cache).Get()
      /w/m/m0110/tarea.go:30 +0x144
  tarea.escenarioC10.func1.1()
      /w/m/m0110/c10-cache-aciertos_test.go:20 +0xcd

Previous write at 0x00c0000a0728 by goroutine 16:
  tarea.(*Cache).Get()
      /w/m/m0110/tarea.go:30 +0x164
  tarea.escenarioC10.func1.1()
      /w/m/m0110/c10-cache-aciertos_test.go:20 +0xcd

Goroutine 11 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0110/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0110/ayuda_test.go:42 +0x33

Goroutine 16 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0110/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0110/ayuda_test.go:42 +0x33
==================
==================
WARNING: DATA RACE
Write at 0x00c0000a0728 by goroutine 11:
  tarea.(*Cache).Get(
```
### qwen2.5-coder_7b/B/c10-cache-aciertos/s4 (WARNING: DATA RACE ×2)
```
==================
WARNING: DATA RACE
Read at 0x00c00007a050 by goroutine 15:
  tarea.(*Cache).Get()
      /w/m/m0111/tarea.go:30 +0x144
  tarea.escenarioC10.func1.1()
      /w/m/m0111/c10-cache-aciertos_test.go:20 +0xcd

Previous write at 0x00c00007a050 by goroutine 10:
  tarea.(*Cache).Get()
      /w/m/m0111/tarea.go:30 +0x164
  tarea.escenarioC10.func1.1()
      /w/m/m0111/c10-cache-aciertos_test.go:20 +0xcd

Goroutine 15 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0111/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0111/ayuda_test.go:42 +0x33

Goroutine 10 (running) created at:
  tarea.escenarioC10.func1()
      /w/m/m0111/c10-cache-aciertos_test.go:17 +0x11b
  tarea.termina.func1()
      /w/m/m0111/ayuda_test.go:42 +0x33
==================
==================
WARNING: DATA RACE
Write at 0x00c00007a050 by goroutine 15:
  tarea.(*Cache).Get(
```
### qwen2.5-coder_7b/B/c11-contar-primos/s1 (WARNING: DATA RACE ×1)
```
==================
WARNING: DATA RACE
Write at 0x00c000022320 by goroutine 9:
  runtime.closechan()
      /usr/local/go/src/runtime/chan.go:414 +0x0
  tarea.ContarPrimos()
      /w/m/m0112/tarea.go:46 +0x376
  tarea.escenarioC11.func1()
      /w/m/m0112/c11-contar-primos_test.go:11 +0x54
  tarea.termina.func1()
      /w/m/m0112/ayuda_test.go:42 +0x33

Previous read at 0x00c000022320 by goroutine 13:
  runtime.chansend()
      /usr/local/go/src/runtime/chan.go:176 +0x0
  tarea.ContarPrimos.func3()
      /w/m/m0112/tarea.go:40 +0xea
  tarea.ContarPrimos.gowrap1()
      /w/m/m0112/tarea.go:42 +0x46

Goroutine 9 (running) created at:
  tarea.termina()
      /w/m/m0112/ayuda_test.go:42 +0xe9
  tarea.escenarioC11()
      /w/m/m0112/c11-contar-primos_test.go:11 +0x144
  tarea.TestC11()
      /w/m/m0112/c11-contar-primos_test.go:16 +0x24
  testing.tRunner()
      /usr/local/go/src/testing/testin
```

## Repetibilidad: segunda corrida del mismo banco (`python medir.py resultados-corrida2.json`)
Muestras cuyo resultado (termina, correcto, fugas, race) cambió entre corridas:
- qwen2.5-coder_7b/A/c05-fan-out-fan-in/s2: corrida 1 (None, None, None, True) → corrida 2 (None, None, None, False)
- qwen2.5-coder_7b/A/c08-unir-canales/s1: corrida 1 (True, True, 0, False) → corrida 2 (True, True, 3, False)
- qwen2.5-coder_7b/A/c10-cache-aciertos/s3: corrida 1 (True, False, 0, True) → corrida 2 (True, True, 0, True)
- qwen2.5-coder_7b/B/c05-fan-out-fan-in/s1: corrida 1 (True, True, 3, False) → corrida 2 (True, True, 5, False)
- control-mal/-/c10-cache-aciertos/s1: corrida 1 (True, True, 0, True) → corrida 2 (True, False, 0, True)
- Total: 5 de 140 muestras cambiaron.
- corrida 2 · claude-code A: con fuga 0, goroutines fugadas 0, con race 0
- corrida 2 · claude-code B: con fuga 0, goroutines fugadas 0, con race 0
- corrida 2 · qwen2.5-coder_7b A: con fuga 8, goroutines fugadas 24, con race 4
- corrida 2 · qwen2.5-coder_7b B: con fuga 4, goroutines fugadas 11, con race 6

## Demos
```
$ go run ./global
canal local:   goroutineleak profile: total 1 | goroutines vivas: 2
canal global:  goroutineleak profile: total 1 | goroutines vivas: 3

$ go run ./http
goroutineleak profile: total 5
5 @ 0x48c1ca 0x418cdc 0x4188d7 0x6b5616 0x4930a1
#	0x6b5615	main.buscar.func1+0x35	/d/http/main.go:18
```
