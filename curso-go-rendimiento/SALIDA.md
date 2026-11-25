# SALIDA — curso «Rendimiento en Go medido: pprof, trace, GC y race»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-10-28 18:23:19Z`
- go version go1.27.2 linux/amd64 (imagen `golang:1.27.2`)
- CPU: `Intel(R) Core(TM) Ultra 7 270K Plus` · el anfitrion tiene 24 nucleos; el contenedor, `--cpus=2`
- Docker 29.6.2 · anfitrion `Windows 10`
- **Los numeros son de esta maquina.** Lo que vale es la comparacion entre versiones en la misma maquina.

## Leccion 01 — Benchmarks bien hechos

`sh 01_benchmarks/comandos.sh` · codigo de salida `0`

```
## La trampa: un benchmark que no mide nada
$ go test ./01_benchmarks -bench . -benchmem -count 5
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/01_benchmarks
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkEsError_bN-2     	1000000000	         0.1956 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_bN-2     	1000000000	         0.1931 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_bN-2     	1000000000	         0.1385 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_bN-2     	1000000000	         0.1537 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_bN-2     	1000000000	         0.1389 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_Loop-2   	1000000000	         0.3762 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_Loop-2   	1000000000	         0.3933 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_Loop-2   	1000000000	         0.3876 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_Loop-2   	1000000000	         0.3925 ns/op	       0 B/op	       0 allocs/op
BenchmarkEsError_Loop-2   	1000000000	         0.3598 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	curso-go-rendimiento/01_benchmarks	2.859s

## Antes de medir: todas las versiones dan la misma respuesta
$ go test ./resumen -run TestTodasIguales -v -count 1
=== RUN   TestTodasIguales
--- PASS: TestTodasIguales (0.05s)
PASS
ok  	curso-go-rendimiento/resumen	0.058s

## Pocas repeticiones: benchstat no puede afirmar nada
$ go test ./resumen -run '^$' -bench Resumen -benchmem -count 3 > $T/n3.txt

$ go tool benchstat -col /version $T/n3.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/resumen
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
          │      V1       │               V2                │               V3                │
          │    sec/op     │    sec/op     vs base           │    sec/op     vs base           │
Resumen-2   16.007m ± ∞ ¹   9.733m ± ∞ ¹  ~ (p=0.100 n=3) ²   3.018m ± ∞ ¹  ~ (p=0.100 n=3) ²
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

          │      V1       │                V2                │                V3                 │
          │      B/s      │      B/s       vs base           │      B/s        vs base           │
Resumen-2   56.03Mi ± ∞ ¹   92.15Mi ± ∞ ¹  ~ (p=0.100 n=3) ²   297.18Mi ± ∞ ¹  ~ (p=0.100 n=3) ²
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

          │       V1        │                V2                 │                V3                │
          │      B/op       │      B/op       vs base           │     B/op       vs base           │
Resumen-2   13066.8Ki ± ∞ ¹   5884.4Ki ± ∞ ¹  ~ (p=0.100 n=3) ²   208.7Ki ± ∞ ¹  ~ (p=0.100 n=3) ²
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

          │       V1       │                V2                 │               V3                │
          │   allocs/op    │   allocs/op     vs base           │  allocs/op    vs base           │
Resumen-2   231.658k ± ∞ ¹   146.607k ± ∞ ¹  ~ (p=0.100 n=3) ²   3.555k ± ∞ ¹  ~ (p=0.100 n=3) ²
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

## Diez repeticiones
$ grep 'version=V1' $T/resumen.txt | head -3
BenchmarkResumen/version=V1-2         	      86	  13998640 ns/op	  67.18 MB/s	13379980 B/op	  231660 allocs/op
BenchmarkResumen/version=V1-2         	      88	  15470926 ns/op	  60.79 MB/s	13384746 B/op	  231665 allocs/op
BenchmarkResumen/version=V1-2         	      78	  15231666 ns/op	  61.75 MB/s	13381255 B/op	  231663 allocs/op

$ go tool benchstat -col /version $T/resumen.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/resumen
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
          │      V1      │                  V2                  │                 V3                  │
          │    sec/op    │    sec/op     vs base                │   sec/op     vs base                │
Resumen-2   15.183m ± 2%   10.727m ± 5%  -29.35% (p=0.000 n=10)   3.074m ± 6%  -79.75% (p=0.000 n=10)

          │      V1      │                  V2                  │                   V3                   │
          │     B/s      │     B/s       vs base                │      B/s       vs base                 │
Resumen-2   59.08Mi ± 2%   83.62Mi ± 5%  +41.54% (p=0.000 n=10)   291.78Mi ± 6%  +393.91% (p=0.000 n=10)

          │       V1       │                  V2                   │                  V3                  │
          │      B/op      │     B/op       vs base                │     B/op      vs base                │
Resumen-2   13067.0Ki ± 0%   5889.9Ki ± 0%  -54.93% (p=0.000 n=10)   208.7Ki ± 0%  -98.40% (p=0.000 n=10)

          │      V1       │                  V2                   │                 V3                  │
          │   allocs/op   │   allocs/op    vs base                │  allocs/op   vs base                │
Resumen-2   231.663k ± 0%   146.613k ± 0%  -36.71% (p=0.000 n=10)   3.555k ± 0%  -98.47% (p=0.000 n=10)
```

## Leccion 02 — pprof de CPU

`sh 02_pprof_cpu/comandos.sh` · codigo de salida `0`

```
## Perfil de CPU de la versión 1
$ go test ./resumen -run '^$' -bench 'Resumen/version=V1$' -benchtime 3s -cpuprofile $T/cpu_v1.prof -o $T/resumen.test > /dev/null

$ go tool pprof -top -nodecount=10 $T/resumen.test $T/cpu_v1.prof
File: resumen.test
Build ID: 07242d903273d752c151c4363a4ad146bff16327
Type: cpu
Time: 2026-10-28 18:24:22 UTC
Duration: 3.63s, Total samples = 4020ms (110.75%)
Showing nodes accounting for 1120ms, 27.86% of 4020ms total
Dropped 173 nodes (cum <= 20.10ms)
Showing top 10 nodes out of 175
      flat  flat%   sum%        cum   cum%
     230ms  5.72%  5.72%      230ms  5.72%  runtime.memmove
     210ms  5.22% 10.95%      210ms  5.22%  encoding/json/internal/jsonwire.ConsumeSimpleString
     120ms  2.99% 13.93%      220ms  5.47%  runtime.mallocgcSmallScanNoHeaderSC2
     110ms  2.74% 16.67%      160ms  3.98%  runtime.scanObject
      90ms  2.24% 18.91%       90ms  2.24%  internal/runtime/maps.memHashAES
      80ms  1.99% 20.90%      100ms  2.49%  runtime.scanblock
      70ms  1.74% 22.64%      170ms  4.23%  encoding/json/v2.makeString
      70ms  1.74% 24.38%      140ms  3.48%  regexp.(*Regexp).tryBacktrack
      70ms  1.74% 26.12%      110ms  2.74%  runtime.mallocgcTinySC2
      70ms  1.74% 27.86%       70ms  1.74%  runtime.memclrNoHeapPointers

$ go tool pprof -top -cum -nodecount=14 $T/resumen.test $T/cpu_v1.prof
File: resumen.test
Build ID: 07242d903273d752c151c4363a4ad146bff16327
Type: cpu
Time: 2026-10-28 18:24:22 UTC
Duration: 3.63s, Total samples = 4.02s (110.75%)
Showing nodes accounting for 0.13s, 3.23% of 4.02s total
Dropped 173 nodes (cum <= 0.02s)
Showing top 14 nodes out of 175
      flat  flat%   sum%        cum   cum%
         0     0%     0%      3.50s 87.06%  curso-go-rendimiento/resumen.BenchmarkResumen.func1
     0.01s  0.25%  0.25%      3.50s 87.06%  curso-go-rendimiento/resumen.ResumenV1
         0     0%  0.25%      3.50s 87.06%  testing.(*B).run1.func1
         0     0%  0.25%      3.50s 87.06%  testing.(*B).runN
         0     0%  0.25%      1.94s 48.26%  encoding/json.Unmarshal (inline)
         0     0%  0.25%      1.94s 48.26%  encoding/json/v2.Unmarshal
         0     0%  0.25%      1.89s 47.01%  encoding/json/v2.unmarshalDecode
     0.03s  0.75%     1%      1.62s 40.30%  encoding/json/v2.makeMapArshaler.func3
     0.01s  0.25%  1.24%      0.90s 22.39%  encoding/json/v2.makeInterfaceArshaler.func2
     0.06s  1.49%  2.74%      0.83s 20.65%  runtime.mallocgc
         0     0%  2.74%      0.65s 16.17%  regexp.Compile (inline)
         0     0%  2.74%      0.65s 16.17%  regexp.MustCompile
     0.02s   0.5%  3.23%      0.65s 16.17%  regexp.compile
         0     0%  3.23%      0.54s 13.43%  runtime.systemstack

$ go tool pprof -list 'resumen.ResumenV1$' $T/resumen.test $T/cpu_v1.prof
Total: 4.02s
ROUTINE ======================== curso-go-rendimiento/resumen.ResumenV1 in /curso/resumen/v1.go
      10ms      3.50s (flat, cum) 87.06% of Total
         .          .     13:func ResumenV1(cuerpo []byte) ([]byte, error) {
         .          .     14:	latencias := map[string][]float64{}
         .          .     15:	errores := map[string]int{}
         .      130ms     16:	for _, linea := range strings.Split(string(cuerpo), "\n") {
         .          .     17:		if strings.TrimSpace(linea) == "" {
         .          .     18:			continue
         .          .     19:		}
         .       30ms     20:		var m map[string]any
         .      2.09s     21:		if err := json.Unmarshal([]byte(linea), &m); err != nil {
         .          .     22:			return nil, fmt.Errorf("línea mal formada: %w", err)
         .          .     23:		}
         .          .     24:		metodo, _ := m["metodo"].(string)
         .       10ms     25:		ruta, _ := m["ruta"].(string)
         .       10ms     26:		ms, _ := m["ms"].(float64)
         .       40ms     27:		estado, _ := m["estado"].(float64)
         .          .     28:
         .          .     29:		// /api/v1/pedidos/48213 -> /api/v1/pedidos/:id
         .      650ms     30:		re := regexp.MustCompile(`/[0-9]+`)
         .      390ms     31:		ruta = re.ReplaceAllString(ruta, "/:id")
         .          .     32:
      10ms       80ms     33:		clave := fmt.Sprintf("%s %s", metodo, ruta)
         .       40ms     34:		latencias[clave] = append(latencias[clave], ms)
         .          .     35:		if estado >= 500 {
         .       10ms     36:			errores[clave]++
         .          .     37:		}
         .          .     38:	}
         .          .     39:
         .          .     40:	var filas []Fila
         .          .     41:	for clave, ms := range latencias {
         .          .     42:		partes := strings.SplitN(clave, " ", 2)
         .          .     43:		orden := append([]float64(nil), ms...)
         .       10ms     44:		sort.Float64s(orden)
         .          .     45:		filas = append(filas, Fila{
         .          .     46:			Metodo: partes[0], Ruta: partes[1], Total: len(ms), Errores: errores[clave],
         .          .     47:			P50: percentil(orden, 50), P95: percentil(orden, 95),
         .          .     48:		})
         .          .     49:	}
         .          .     50:	ordenar(filas)
         .       10ms     51:	return json.Marshal(filas)
         .          .     52:}

$ go tool pprof -peek 'regexp.MustCompile$' $T/resumen.test $T/cpu_v1.prof
File: resumen.test
Build ID: 07242d903273d752c151c4363a4ad146bff16327
Type: cpu
Time: 2026-10-28 18:24:22 UTC
Duration: 3.63s, Total samples = 4.02s (110.75%)
Showing nodes accounting for 4.02s, 100% of 4.02s total
----------------------------------------------------------+-------------
      flat  flat%   sum%        cum   cum%   calls calls% + context 	 	 
----------------------------------------------------------+-------------
                                             0.65s   100% |   curso-go-rendimiento/resumen.ResumenV1
         0     0%     0%      0.65s 16.17%                | regexp.MustCompile
                                             0.65s   100% |   regexp.Compile (inline)
----------------------------------------------------------+-------------

## Versión 2: la expresión regular, compilada una vez
$ go test ./resumen -run '^$' -bench 'Resumen/version=V2$' -benchtime 3s -cpuprofile $T/cpu_v2.prof -o $T/resumen.test > /dev/null

$ go tool pprof -top -cum -nodecount=10 $T/resumen.test $T/cpu_v2.prof
File: resumen.test
Build ID: 07242d903273d752c151c4363a4ad146bff16327
Type: cpu
Time: 2026-10-28 18:24:29 UTC
Duration: 3.62s, Total samples = 3.89s (107.52%)
Showing nodes accounting for 0.14s, 3.60% of 3.89s total
Dropped 97 nodes (cum <= 0.02s)
Showing top 10 nodes out of 198
      flat  flat%   sum%        cum   cum%
         0     0%     0%      3.67s 94.34%  curso-go-rendimiento/resumen.BenchmarkResumen.func1
     0.01s  0.26%  0.26%      3.67s 94.34%  curso-go-rendimiento/resumen.ResumenV2
         0     0%  0.26%      3.67s 94.34%  testing.(*B).run1.func1
         0     0%  0.26%      3.67s 94.34%  testing.(*B).runN
         0     0%  0.26%      2.66s 68.38%  encoding/json.Unmarshal (inline)
         0     0%  0.26%      2.66s 68.38%  encoding/json/v2.Unmarshal
     0.02s  0.51%  0.77%      2.61s 67.10%  encoding/json/v2.unmarshalDecode
     0.01s  0.26%  1.03%      2.31s 59.38%  encoding/json/v2.makeMapArshaler.func3
     0.02s  0.51%  1.54%      1.33s 34.19%  encoding/json/v2.makeInterfaceArshaler.func2
     0.08s  2.06%  3.60%      0.78s 20.05%  runtime.mallocgc

$ go tool benchstat -filter '/version:(V1 OR V2)' -col /version $T/resumen.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/resumen
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
          │     V1      │                 V2                  │
          │   sec/op    │   sec/op     vs base                │
Resumen-2   15.18m ± 2%   10.73m ± 5%  -29.35% (p=0.000 n=10)

          │      V1      │                  V2                  │
          │     B/s      │     B/s       vs base                │
Resumen-2   59.08Mi ± 2%   83.62Mi ± 5%  +41.54% (p=0.000 n=10)

          │      V1       │                  V2                  │
          │     B/op      │     B/op      vs base                │
Resumen-2   12.761Mi ± 0%   5.752Mi ± 0%  -54.93% (p=0.000 n=10)

          │     V1      │                 V2                  │
          │  allocs/op  │  allocs/op   vs base                │
Resumen-2   231.7k ± 0%   146.6k ± 0%  -36.71% (p=0.000 n=10)
```

## Leccion 03 — pprof de memoria: alloc_space contra inuse_space

`sh 03_memoria/comandos.sh` · codigo de salida `0`

```
## alloc_space en el benchmark de la versión 2
$ go test ./resumen -run '^$' -bench 'Resumen/version=V2$' -benchtime 2s -memprofile $T/mem_v2.prof -o $T/resumen.test > /dev/null

$ go tool pprof -sample_index=alloc_space -top -nodecount=8 $T/resumen.test $T/mem_v2.prof
File: resumen.test
Build ID: 07242d903273d752c151c4363a4ad146bff16327
Type: alloc_space
Time: 2026-10-28 18:24:36 UTC
Showing nodes accounting for 1342.10MB, 96.07% of 1397.05MB total
Dropped 24 nodes (cum <= 6.99MB)
Showing top 8 nodes out of 32
      flat  flat%   sum%        cum   cum%
  537.49MB 38.47% 38.47%  1394.98MB 99.85%  curso-go-rendimiento/resumen.ResumenV2
  329.59MB 23.59% 62.07%   329.59MB 23.59%  reflect.mapassign_faststr0
     242MB 17.32% 79.39%      242MB 17.32%  reflect.unsafe_New
   72.50MB  5.19% 84.58%    72.50MB  5.19%  encoding/json/v2.makeString
      58MB  4.15% 88.73%       58MB  4.15%  reflect.makemap
   41.50MB  2.97% 91.70%    65.64MB  4.70%  regexp.(*Regexp).replaceAll
   35.50MB  2.54% 94.24%    35.50MB  2.54%  fmt.Sprintf
   25.50MB  1.83% 96.07%    91.14MB  6.52%  regexp.(*Regexp).ReplaceAllString

$ go tool pprof -sample_index=alloc_space -list 'resumen.ResumenV2$' $T/resumen.test $T/mem_v2.prof
Total: 1.36GB
ROUTINE ======================== curso-go-rendimiento/resumen.ResumenV2 in /curso/resumen/v2.go
  537.49MB     1.36GB (flat, cum) 99.85% of Total
         .          .     15:func ResumenV2(cuerpo []byte) ([]byte, error) {
         .          .     16:	latencias := map[string][]float64{}
         .          .     17:	errores := map[string]int{}
  221.87MB   245.09MB     18:	for _, linea := range strings.Split(string(cuerpo), "\n") {
         .          .     19:		if strings.TrimSpace(linea) == "" {
         .          .     20:			continue
         .          .     21:		}
       5MB        5MB     22:		var m map[string]any
  227.55MB   930.15MB     23:		if err := json.Unmarshal([]byte(linea), &m); err != nil {
         .          .     24:			return nil, fmt.Errorf("línea mal formada: %w", err)
         .          .     25:		}
         .          .     26:		metodo, _ := m["metodo"].(string)
         .          .     27:		ruta, _ := m["ruta"].(string)
         .          .     28:		ms, _ := m["ms"].(float64)
         .          .     29:		estado, _ := m["estado"].(float64)
         .          .     30:
         .    91.14MB     31:		ruta = reID.ReplaceAllString(ruta, "/:id")
         .          .     32:
   43.50MB       79MB     33:		clave := fmt.Sprintf("%s %s", metodo, ruta)
   29.56MB    29.56MB     34:		latencias[clave] = append(latencias[clave], ms)
         .          .     35:		if estado >= 500 {
         .          .     36:			errores[clave]++
         .          .     37:		}
         .          .     38:	}
         .          .     39:
         .          .     40:	var filas []Fila
         .          .     41:	for clave, ms := range latencias {
         .          .     42:		partes := strings.SplitN(clave, " ", 2)
    9.52MB     9.52MB     43:		orden := append([]float64(nil), ms...)
         .          .     44:		sort.Float64s(orden)
  513.12kB   513.12kB     45:		filas = append(filas, Fila{
         .          .     46:			Metodo: partes[0], Ruta: partes[1], Total: len(ms), Errores: errores[clave],
         .          .     47:			P50: percentil(orden, 50), P95: percentil(orden, 95),
         .          .     48:		})
         .          .     49:	}
         .          .     50:	ordenar(filas)
         .     5.01MB     51:	return json.Marshal(filas)
         .          .     52:}

## El servicio con el panel de errores: 300 peticiones
$ go build -o $T/memoria ./03_memoria && $T/memoria -perfil $T/heap.prof
300 peticiones · 2000 clientes con error guardado · heap en uso tras el GC: 47.4 MB
asignado en total durante la prueba: 2833 MB

$ go tool pprof -sample_index=alloc_space -top -nodecount=6 $T/memoria $T/heap.prof
File: memoria
Build ID: b0e98405577ce2642521cf06b5fbd3eebf0dacb9
Type: alloc_space
Time: 2026-10-28 18:24:40 UTC
Showing nodes accounting for 2372.33MB, 83.68% of 2834.92MB total
Dropped 68 nodes (cum <= 14.17MB)
Showing top 6 nodes out of 43
      flat  flat%   sum%        cum   cum%
  655.90MB 23.14% 23.14%  1716.80MB 60.56%  curso-go-rendimiento/resumen.ResumenV2
  412.99MB 14.57% 37.70%   412.99MB 14.57%  reflect.mapassign_faststr0
  367.36MB 12.96% 50.66%   807.58MB 28.49%  curso-go-rendimiento/datos.GenerarConSemilla
  330.47MB 11.66% 62.32%   330.47MB 11.66%  internal/bytealg.MakeNoZero
  319.62MB 11.27% 73.59%   319.62MB 11.27%  reflect.unsafe_New
  285.99MB 10.09% 83.68%   310.55MB 10.95%  main.guardarErrores

$ go tool pprof -sample_index=inuse_space -top -nodecount=5 $T/memoria $T/heap.prof
File: memoria
Build ID: b0e98405577ce2642521cf06b5fbd3eebf0dacb9
Type: inuse_space
Time: 2026-10-28 18:24:40 UTC
Showing nodes accounting for 46.09MB, 99.73% of 46.22MB total
Dropped 39 nodes (cum <= 0.23MB)
      flat  flat%   sum%        cum   cum%
   46.09MB 99.73% 99.73%    46.15MB 99.84%  main.guardarErrores
         0     0% 99.73%    46.22MB   100%  main.main
         0     0% 99.73%    46.22MB   100%  runtime.main

$ go tool pprof -sample_index=inuse_space -list 'main.guardarErrores' $T/memoria $T/heap.prof
Total: 46.22MB
ROUTINE ======================== main.guardarErrores in /curso/03_memoria/main.go
   46.09MB    46.15MB (flat, cum) 99.84% of Total
         .          .     30:func guardarErrores(cuerpo []byte) {
   45.94MB    45.94MB     31:	texto := string(cuerpo) // una copia del cuerpo entero: 1,1 MB
         .          .     32:	for _, linea := range strings.Split(texto, "\n") {
         .          .     33:		if !strings.Contains(linea, `"estado":5`) {
         .          .     34:			continue
         .          .     35:		}
         .          .     36:		var e struct {
         .          .     37:			IP string `json:"ip"`
         .          .     38:		}
         .    54.93kB     39:		if json.Unmarshal([]byte(linea), &e) != nil {
         .          .     40:			continue
         .          .     41:		}
         .          .     42:		if *arreglado {
         .          .     43:			linea = strings.Clone(linea) // 230 bytes propios
         .          .     44:		}
         .          .     45:		// Sin Clone, linea es un trozo de texto: apunta DENTRO de la copia de 1,1 MB.
         .          .     46:		// Mientras el mapa guarde ese trozo, el recolector no puede soltar el resto.
  160.01kB   160.01kB     47:		ultimoError[e.IP] = linea
         .          .     48:	}
         .          .     49:}
         .          .     50:
         .          .     51:func main() {
         .          .     52:	perfil := flag.String("perfil", "heap.prof", "dónde escribir el perfil del heap")

## Con strings.Clone
$ $T/memoria -perfil $T/heap2.prof -arreglado
300 peticiones · 2000 clientes con error guardado · heap en uso tras el GC: 2.2 MB
asignado en total durante la prueba: 2872 MB

$ go tool pprof -sample_index=inuse_space -top -nodecount=5 $T/memoria $T/heap2.prof
File: memoria
Build ID: b0e98405577ce2642521cf06b5fbd3eebf0dacb9
Type: inuse_space
Time: 2026-10-28 18:24:45 UTC
Showing nodes accounting for 684.59kB, 95.87% of 714.06kB total
Showing top 5 nodes out of 46
      flat  flat%   sum%        cum   cum%
  402.09kB 56.31% 56.31%   402.09kB 56.31%  internal/stringslite.Clone (inline)
  160.01kB 22.41% 78.72%   594.16kB 83.21%  main.guardarErrores
   76.68kB 10.74% 89.46%    76.68kB 10.74%  regexp.(*bitState).reset
   32.06kB  4.49% 93.95%    32.06kB  4.49%  encoding/json/v2.makeString
   13.75kB  1.93% 95.87%    13.75kB  1.93%  reflect.compiledTypelinks
```

## Leccion 04 — Analisis de escape y asignaciones

`sh 04_escape/comandos.sh` · codigo de salida `0`

```
## Lo que escapa en la versión 2
$ go build -gcflags=-m ./resumen 2>&1 | grep 'v2.go' | grep -vE 'inlin|does not escape|leaking'
resumen/v2.go:22:7: moved to heap: m
resumen/v2.go:18:45: string(cuerpo) escapes to heap
resumen/v2.go:23:35: ([]byte)(linea) escapes to heap
resumen/v2.go:24:26: &errors.errorString{...} escapes to heap
resumen/v2.go:33:33: metodo escapes to heap
resumen/v2.go:33:41: ruta escapes to heap
resumen/v2.go:34:28: append escapes to heap
resumen/v2.go:45:17: append escapes to heap
resumen/v2.go:50:9: filas escapes to heap
resumen/v2.go:51:22: filas escapes to heap

## Tres casos aislados
$ go test -gcflags=-m ./04_escape 2>&1 | grep 'escape_test.go' | grep -E 'moved to heap|metodo escapes|ruta escapes'
04_escape/escape_test.go:28:7: moved to heap: e
04_escape/escape_test.go:38:6: moved to heap: e
04_escape/escape_test.go:64:81: metodo escapes to heap
04_escape/escape_test.go:64:89: ruta escapes to heap
04_escape/escape_test.go:79:16: metodo escapes to heap
04_escape/escape_test.go:79:16: ruta escapes to heap

$ go test ./04_escape -run '^$' -bench . -benchmem -count 10 > $T/escape.txt

$ go tool benchstat -filter '.name:Decodificar' -col /e $T/escape.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/04_escape
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
              │   dentro    │               fuera                │
              │   sec/op    │   sec/op     vs base               │
Decodificar-2   520.8µ ± 1%   508.1µ ± 1%  -2.44% (p=0.002 n=10)

              │    dentro    │                fuera                 │
              │     B/op     │     B/op      vs base                │
Decodificar-2   62.39Ki ± 0%   15.55Ki ± 0%  -75.08% (p=0.000 n=10)

              │   dentro    │               fuera                │
              │  allocs/op  │ allocs/op   vs base                │
Decodificar-2   1588.0 ± 0%   589.0 ± 0%  -62.91% (p=0.000 n=10)

$ go tool benchstat -filter '.name:Clave' -col /forma $T/escape.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/04_escape
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
        │   Sprintf    │             concatenar              │            bufer_en_pila            │
        │    sec/op    │   sec/op     vs base                │   sec/op     vs base                │
Clave-2   55.32n ± 10%   14.48n ± 2%  -73.83% (p=0.000 n=10)   12.46n ± 0%  -77.48% (p=0.000 n=10)

        │  Sprintf   │             concatenar             │           bufer_en_pila            │
        │    B/op    │   B/op     vs base                 │   B/op     vs base                 │
Clave-2   56.00 ± 0%   0.00 ± 0%  -100.00% (p=0.000 n=10)   0.00 ± 0%  -100.00% (p=0.000 n=10)

        │  Sprintf   │             concatenar              │            bufer_en_pila            │
        │ allocs/op  │ allocs/op   vs base                 │ allocs/op   vs base                 │
Clave-2   3.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)   0.000 ± 0%  -100.00% (p=0.000 n=10)

## La versión 3
$ go build -gcflags=-m ./resumen 2>&1 | grep 'v3.go' | grep -E 'moved to heap|string\(clave\)|make\(map'
resumen/v3.go:40:6: moved to heap: e
resumen/v3.go:36:13: make(map[string]*acumulado, 64) does not escape
resumen/v3.go:59:19: string(clave) does not escape
resumen/v3.go:62:15: string(clave) escapes to heap

$ go tool benchstat -filter '/version:(V2 OR V3)' -col /version $T/resumen.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/resumen
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
          │      V2      │                 V3                  │
          │    sec/op    │   sec/op     vs base                │
Resumen-2   10.727m ± 5%   3.074m ± 6%  -71.34% (p=0.000 n=10)

          │      V2      │                   V3                   │
          │     B/s      │      B/s       vs base                 │
Resumen-2   83.62Mi ± 5%   291.78Mi ± 6%  +248.94% (p=0.000 n=10)

          │      V2       │                  V3                  │
          │     B/op      │     B/op      vs base                │
Resumen-2   5889.9Ki ± 0%   208.7Ki ± 0%  -96.46% (p=0.000 n=10)

          │      V2       │                 V3                  │
          │   allocs/op   │  allocs/op   vs base                │
Resumen-2   146.613k ± 0%   3.555k ± 0%  -97.58% (p=0.000 n=10)
```

## Leccion 05 — sync.Pool, y cuando no sirve

`sh 05_pool/comandos.sh` · codigo de salida `0`

```
## El servicio entero: io.ReadAll (V3) contra búferes del pool (V5)
$ go test ./resumen -run '^$' -bench Handler -benchmem -count 10 > $T/handler.txt

$ go tool benchstat -col /version $T/handler.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/resumen
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
          │     V3      │                 V5                  │
          │   sec/op    │   sec/op     vs base                │
Handler-2   3.614m ± 2%   3.018m ± 1%  -16.49% (p=0.000 n=10)

          │      V3      │                  V5                  │
          │     B/s      │     B/s       vs base                │
Handler-2   248.1Mi ± 2%   297.2Mi ± 1%  +19.75% (p=0.000 n=10)

          │      V3       │                  V5                  │
          │     B/op      │     B/op      vs base                │
Handler-2   2296.7Ki ± 0%   220.2Ki ± 2%  -90.41% (p=0.000 n=10)

          │     V3      │                 V5                 │
          │  allocs/op  │  allocs/op   vs base               │
Handler-2   3.614k ± 0%   3.575k ± 0%  -1.08% (p=0.000 n=10)

## Un objeto pequeño
$ go test ./05_pool -run '^$' -bench . -benchmem -count 10 > $T/pool.txt

$ go tool benchstat -filter '.name:Pequeno' -col /con $T/pool.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/05_pool
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
          │    pila     │                 new                  │                 Pool                 │
          │   sec/op    │   sec/op     vs base                 │   sec/op     vs base                 │
Pequeno-2   1.198n ± 0%   8.793n ± 3%  +633.71% (p=0.000 n=10)   4.492n ± 1%  +274.80% (p=0.000 n=10)

          │   pila    │             new              │             Pool              │
          │   B/op    │    B/op     vs base          │   B/op     vs base            │
Pequeno-2   0.00 ± 0%   32.00 ± 0%  ? (p=0.000 n=10)   0.00 ± 0%  ~ (p=1.000 n=10) ¹
¹ all samples are equal

          │    pila    │             new              │              Pool              │
          │ allocs/op  │ allocs/op   vs base          │ allocs/op   vs base            │
Pequeno-2   0.000 ± 0%   1.000 ± 0%  ? (p=0.000 n=10)   0.000 ± 0%  ~ (p=1.000 n=10) ¹
¹ all samples are equal

$ go tool benchstat -filter '.name:Grande' -col /con $T/pool.txt
goos: linux
goarch: amd64
pkg: curso-go-rendimiento/05_pool
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
         │      make       │                Pool                 │
         │     sec/op      │   sec/op     vs base                │
Grande-2   5933.500n ± 19%   4.733n ± 1%  -99.92% (p=0.000 n=10)

         │     make     │                 Pool                 │
         │     B/op     │    B/op      vs base                 │
Grande-2   64.00Ki ± 0%   0.00Ki ± 0%  -100.00% (p=0.000 n=10)

         │    make    │                Pool                 │
         │ allocs/op  │ allocs/op   vs base                 │
Grande-2   1.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)

## El pool no es una caché
$ go test ./05_pool -run TestSobreviveAlGC -v -count 1
=== RUN   TestSobreviveAlGC
    pool_test.go:82: Put, 0 GC, Get -> ¿el mismo objeto? true
    pool_test.go:82: Put, 1 GC, Get -> ¿el mismo objeto? false
    pool_test.go:82: Put, 2 GC, Get -> ¿el mismo objeto? false
--- PASS: TestSobreviveAlGC (0.00s)
PASS
ok  	curso-go-rendimiento/05_pool	0.004s
```

## Leccion 06 — go tool trace: goroutines bloqueadas

`sh 06_trace/comandos.sh` · codigo de salida `0`

```
## El mismo trabajo, con el candado en tres sitios
$ go build -o $T/traza ./06_trace

$ $T/traza -modo ancho -repetir 7
modo ancho     80000 líneas · 49299 claves · 7 repeticiones: mediana  60.4 ms, mínimo  56.5 ms  (GOMAXPROCS=2)

$ $T/traza -modo estrecho -repetir 7
modo estrecho  80000 líneas · 49299 claves · 7 repeticiones: mediana  39.0 ms, mínimo  34.8 ms  (GOMAXPROCS=2)

$ $T/traza -modo local -repetir 7
modo local     80000 líneas · 49299 claves · 7 repeticiones: mediana  33.0 ms, mínimo  31.6 ms  (GOMAXPROCS=2)

## La traza del candado ancho
$ $T/traza -modo ancho -trace $T/ancho.out
modo ancho     80000 líneas · 49299 claves · 1 repeticiones: mediana  63.5 ms, mínimo  63.5 ms  (GOMAXPROCS=2)

$ go tool trace -pprof=sync $T/ancho.out > $T/sync.prof && go tool pprof -top -nodecount=6 $T/sync.prof
Main binary filename not available.
Type: delay
Showing nodes accounting for 370.19ms, 99.82% of 370.85ms total
Dropped 1 node (cum <= 1.85ms)
Showing top 6 nodes out of 9
      flat  flat%   sum%        cum   cum%
  308.20ms 83.11% 83.11%   308.20ms 83.11%  sync.(*Mutex).Lock
   42.21ms 11.38% 94.49%    42.21ms 11.38%  runtime.chansend1
   19.78ms  5.33% 99.82%    19.78ms  5.33%  sync.(*WaitGroup).Wait
         0     0% 99.82%   308.20ms 83.11%  main.lineas
         0     0% 99.82%    61.98ms 16.71%  main.main
         0     0% 99.82%    61.98ms 16.71%  main.procesar

$ go tool trace -pprof=sched $T/ancho.out > $T/sched.prof && go tool pprof -top -nodecount=5 $T/sched.prof
Main binary filename not available.
Type: delay
Showing nodes accounting for 78.51ms, 98.74% of 79.51ms total
Dropped 18 nodes (cum <= 0.40ms)
Showing top 5 nodes out of 14
      flat  flat%   sum%        cum   cum%
   65.10ms 81.88% 81.88%    72.39ms 91.04%  sync.(*Mutex).Unlock
    7.29ms  9.16% 91.04%     7.29ms  9.16%  runtime.goyield
    4.58ms  5.76% 96.80%     4.58ms  5.76%  runtime.chansend1
    1.54ms  1.94% 98.74%     1.54ms  1.94%  runtime.chanrecv2
         0     0% 98.74%     7.29ms  9.16%  internal/sync.(*Mutex).Unlock

$ python3 06_trace/desglose.py $T/ancho.out
go tool trace · 8 goroutines de trabajo (wg.Go), en ms
  goroutine    total     ejecutando  en un candado  esperando CPU
         28     63.5           12.3           37.5           13.7
         27     63.1            8.6           45.6            8.8
         26     61.8            8.0           41.0           12.8
         31     60.9            8.0           45.6            7.3
         24     53.9            8.8           36.2            8.8
         30     51.1            8.0           34.2            8.8
         25     50.8            8.2           34.8            7.6
         29     49.8            5.8           33.4           10.6
       suma                    67.6          308.2           78.5

## La traza con un mapa por goroutine
$ $T/traza -modo local -trace $T/local.out
modo local     80000 líneas · 49299 claves · 1 repeticiones: mediana  35.7 ms, mínimo  35.7 ms  (GOMAXPROCS=2)

$ python3 06_trace/desglose.py $T/local.out
go tool trace · 8 goroutines de trabajo (wg.Go), en ms
  goroutine    total     ejecutando  en un candado  esperando CPU
         22     35.7            8.3            0.0           27.2
         27     33.1           13.1            1.5           15.1
         26     32.0           22.1            0.2            6.4
         21     26.7            3.9            0.0           22.7
         24     24.6            4.2            0.0           20.3
         25     22.9            3.6            0.0           19.1
         23     20.4            4.0            0.5           15.8
         20     19.8            7.3            0.0           12.4
       suma                    66.5            2.2          138.8
```

## Leccion 07 — El recolector: GOGC y GOMEMLIMIT

`sh 07_gc/comandos.sh` · codigo de salida `0`

```
$ go build -o $T/gc ./07_gc

## GOGC: tres ejecuciones de cada valor
$ GOGC=25 $T/gc; GOGC=25 $T/gc; GOGC=25 $T/gc
GOGC=25   GOMEMLIMIT=sin límite  GCs   48  memoria pico  113 MB  CPU  2.60 s  (pared 1.73 s)
GOGC=25   GOMEMLIMIT=sin límite  GCs   48  memoria pico  113 MB  CPU  2.64 s  (pared 1.76 s)
GOGC=25   GOMEMLIMIT=sin límite  GCs   48  memoria pico  113 MB  CPU  2.66 s  (pared 1.77 s)

$ GOGC=50 $T/gc; GOGC=50 $T/gc; GOGC=50 $T/gc
GOGC=50   GOMEMLIMIT=sin límite  GCs   25  memoria pico  132 MB  CPU  2.07 s  (pared 1.53 s)
GOGC=50   GOMEMLIMIT=sin límite  GCs   25  memoria pico  135 MB  CPU  2.04 s  (pared 1.51 s)
GOGC=50   GOMEMLIMIT=sin límite  GCs   25  memoria pico  133 MB  CPU  2.06 s  (pared 1.53 s)

$ GOGC=100 $T/gc; GOGC=100 $T/gc; GOGC=100 $T/gc
GOGC=100  GOMEMLIMIT=sin límite  GCs   13  memoria pico  164 MB  CPU  1.77 s  (pared 1.45 s)
GOGC=100  GOMEMLIMIT=sin límite  GCs   14  memoria pico  163 MB  CPU  1.79 s  (pared 1.47 s)
GOGC=100  GOMEMLIMIT=sin límite  GCs   13  memoria pico  165 MB  CPU  1.76 s  (pared 1.46 s)

$ GOGC=200 $T/gc; GOGC=200 $T/gc; GOGC=200 $T/gc
GOGC=200  GOMEMLIMIT=sin límite  GCs    6  memoria pico  244 MB  CPU  1.59 s  (pared 1.43 s)
GOGC=200  GOMEMLIMIT=sin límite  GCs    6  memoria pico  245 MB  CPU  1.60 s  (pared 1.43 s)
GOGC=200  GOMEMLIMIT=sin límite  GCs    7  memoria pico  244 MB  CPU  1.63 s  (pared 1.44 s)

$ GOGC=400 $T/gc; GOGC=400 $T/gc; GOGC=400 $T/gc
GOGC=400  GOMEMLIMIT=sin límite  GCs    3  memoria pico  360 MB  CPU  1.56 s  (pared 1.45 s)
GOGC=400  GOMEMLIMIT=sin límite  GCs    3  memoria pico  362 MB  CPU  1.56 s  (pared 1.45 s)
GOGC=400  GOMEMLIMIT=sin límite  GCs    3  memoria pico  366 MB  CPU  1.56 s  (pared 1.45 s)

## GOMEMLIMIT
$ GOGC=off GOMEMLIMIT=200MiB $T/gc; GOGC=off GOMEMLIMIT=200MiB $T/gc; GOGC=off GOMEMLIMIT=200MiB $T/gc
GOGC=off  GOMEMLIMIT=200 MiB     GCs   10  memoria pico  194 MB  CPU  1.69 s  (pared 1.44 s)
GOGC=off  GOMEMLIMIT=200 MiB     GCs   10  memoria pico  194 MB  CPU  1.67 s  (pared 1.44 s)
GOGC=off  GOMEMLIMIT=200 MiB     GCs   10  memoria pico  194 MB  CPU  1.69 s  (pared 1.44 s)

$ GOMEMLIMIT=100MiB $T/gc; GOMEMLIMIT=100MiB $T/gc; GOMEMLIMIT=100MiB $T/gc
GOGC=100  GOMEMLIMIT=100 MiB     GCs   91  memoria pico  100 MB  CPU  3.65 s  (pared 2.25 s)
GOGC=100  GOMEMLIMIT=100 MiB     GCs   91  memoria pico  100 MB  CPU  3.62 s  (pared 2.23 s)
GOGC=100  GOMEMLIMIT=100 MiB     GCs   91  memoria pico  100 MB  CPU  3.62 s  (pared 2.22 s)

$ GOMEMLIMIT=60MiB $T/gc; GOMEMLIMIT=60MiB $T/gc; GOMEMLIMIT=60MiB $T/gc
GOGC=100  GOMEMLIMIT=60 MiB      GCs  193  memoria pico   99 MB  CPU  6.85 s  (pared 3.48 s)
GOGC=100  GOMEMLIMIT=60 MiB      GCs  208  memoria pico   99 MB  CPU  7.30 s  (pared 3.70 s)
GOGC=100  GOMEMLIMIT=60 MiB      GCs  193  memoria pico   99 MB  CPU  6.88 s  (pared 3.48 s)

## gctrace: una línea por ciclo del recolector
$ GODEBUG=gctrace=1 $T/gc 2>&1 | grep '^gc ' | sed -n '9,12p'
gc 9 @0.580s 3%: 0.036+14+0.041 ms clock, 0.072+0.038/7.2/10+0.083 ms cpu, 138->142->78 MB, 162 MB goal, 0 MB stacks, 0 MB globals, 2 P
gc 10 @0.676s 3%: 0.031+16+0.047 ms clock, 0.063+0.032/4.7/13+0.095 ms cpu, 133->140->80 MB, 156 MB goal, 0 MB stacks, 0 MB globals, 2 P
gc 11 @0.777s 2%: 0.030+15+0.050 ms clock, 0.060+0.029/3.1/14+0.10 ms cpu, 137->144->80 MB, 160 MB goal, 0 MB stacks, 0 MB globals, 2 P
gc 12 @0.882s 3%: 0.043+15+0.040 ms clock, 0.086+0.33/9.0/11+0.080 ms cpu, 137->142->79 MB, 161 MB goal, 0 MB stacks, 0 MB globals, 2 P
```

### En un contenedor con 200 MB (`docker run --memory=200m`)

`sh 07_gc/oom.sh` · codigo de salida `0`

```
$ cat /sys/fs/cgroup/memory.max
209715200

$ go build -o $T/gc ./07_gc

$ GOGC=400 $T/gc; echo "código de salida: $?"
Killed
código de salida: 137

$ GOGC=400 GOMEMLIMIT=180MiB $T/gc; echo "código de salida: $?"
GOGC=400  GOMEMLIMIT=180 MiB     GCs   13  memoria pico  164 MB  CPU  1.75 s  (pared 1.45 s)
código de salida: 0
```

## Leccion 08 — La condicion de carrera que los tests no ven

`sh 08_carrera/comandos.sh` · codigo de salida `0`

```
## Las pruebas de siempre
$ go test ./resumen -count 1
ok  	curso-go-rendimiento/resumen	0.067s

$ go test ./08_carrera -run TestHandlerV4Concurrente -count 200
ok  	curso-go-rendimiento/08_carrera	16.666s

## Con el detector de carreras
$ go test -race ./08_carrera -run TestHandlerV4Concurrente -count 1 2>&1 | head -45
==================
WARNING: DATA RACE
Write at 0x00c000264c80 by goroutine 16:
  runtime.slicecopy()
      /usr/local/go/src/runtime/slice.go:392 +0x0
  bytes.(*Buffer).Write()
      /usr/local/go/src/bytes/buffer.go:199 +0x118
  encoding/json.(*Encoder).Encode()
      /usr/local/go/src/encoding/json/v2_stream.go:156 +0x441
  curso-go-rendimiento/resumen.ResumenV4()
      /curso/resumen/servicio.go:82 +0x27c
  curso-go-rendimiento/resumen.HandlerV4()
      /curso/resumen/servicio.go:63 +0x22d
  curso-go-rendimiento/08_carrera.TestHandlerV4Concurrente.func1()
      /curso/08_carrera/carrera_test.go:44 +0xf8
  curso-go-rendimiento/08_carrera.probarConcurrente.func1()
      /curso/08_carrera/carrera_test.go:28 +0x182

Previous read at 0x00c000264c80 by goroutine 45:
  runtime.slicecopy()
      /usr/local/go/src/runtime/slice.go:392 +0x0
  bytes.(*Buffer).Write()
      /usr/local/go/src/bytes/buffer.go:199 +0x118
  net/http/httptest.(*ResponseRecorder).Write()
      /usr/local/go/src/net/http/httptest/recorder.go:111 +0x97
  curso-go-rendimiento/resumen.HandlerV4()
      /curso/resumen/servicio.go:69 +0x363
  curso-go-rendimiento/08_carrera.TestHandlerV4Concurrente.func1()
      /curso/08_carrera/carrera_test.go:44 +0xf8
  curso-go-rendimiento/08_carrera.probarConcurrente.func1()
      /curso/08_carrera/carrera_test.go:28 +0x182

Goroutine 16 (running) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x72
  curso-go-rendimiento/08_carrera.TestHandlerV4Concurrente()
      /curso/08_carrera/carrera_test.go:43 +0x2d
  testing.tRunner()
      /usr/local/go/src/testing/testing.go:2195 +0x21c
  testing.(*T).Run.gowrap1()
      /usr/local/go/src/testing/testing.go:2260 +0x38

Goroutine 45 (finished) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x72

$ n=0; for i in 1 2 3 4 5 6 7 8 9 10; do go test -race ./08_carrera -run TestHandlerV4Concurrente -count 1 > /dev/null 2>&1 || n=$((n+1)); done; echo "carrera detectada en $n de 10 ejecuciones"
carrera detectada en 8 de 10 ejecuciones

## Y con tráfico de verdad, sin -race
$ go run ./10_carga -version V4 -clientes 32 -duracion 5s
V4  32 clientes · 5s:  2819 peticiones   563.8 pet/s  p50   55.4 ms  p99  115.2 ms  respuestas mal 0  errores 0

## El arreglo: escribir antes de devolver el búfer
$ go test -race ./08_carrera -run TestHandlerV5Concurrente -count 10
ok  	curso-go-rendimiento/08_carrera	25.307s
```

## Leccion 09 — pprof en un servicio vivo

`sh 09_vivo/comandos.sh` · codigo de salida `0`

```
## 1. Expuesto: pprof en el puerto público
$ curl http://172.17.0.3:8080/debug/pprof/          -> HTTP 200
$ curl http://172.17.0.3:8080/debug/pprof/cmdline
/tmp/curso-go-rendimiento/svc -modo expuesto -token sk_demo_no_es_real

## 2. Privado: el servicio en 0.0.0.0:8080, pprof en 127.0.0.1:6060
$ curl http://172.17.0.3:8080/debug/pprof/          -> HTTP 404
$ curl http://172.17.0.3:6060/debug/pprof/          -> conexión rechazada
$ curl http://127.0.0.1:6060/debug/pprof/    -> HTTP 200

## 3. Perfil de CPU de 5 s con el servicio atendiendo tráfico (versión 1)
$ go tool pprof -top -cum -seconds 5 http://127.0.0.1:6060/debug/pprof/profile
Type: cpu
Duration: 5.01s, Total samples = 9.46s (189.01%)
Showing nodes accounting for 0.22s, 2.33% of 9.46s total
Dropped 251 nodes (cum <= 0.05s)
Showing top 14 nodes out of 235
      flat  flat%   sum%        cum   cum%
         0     0%     0%      8.85s 93.55%  net/http.(*conn).serve
         0     0%     0%      8.80s 93.02%  curso-go-rendimiento/resumen.HandlerV1
         0     0%     0%      8.80s 93.02%  net/http.(*ServeMux).ServeHTTP
         0     0%     0%      8.80s 93.02%  net/http.HandlerFunc.ServeHTTP
         0     0%     0%      8.80s 93.02%  net/http.serverHandler.ServeHTTP
         0     0%     0%      8.36s 88.37%  curso-go-rendimiento/resumen.ResumenV1
         0     0%     0%      4.22s 44.61%  encoding/json.Unmarshal (inline)
         0     0%     0%      4.22s 44.61%  encoding/json/v2.Unmarshal
     0.01s  0.11%  0.11%      4.10s 43.34%  encoding/json/v2.unmarshalDecode
     0.03s  0.32%  0.42%      3.68s 38.90%  encoding/json/v2.makeMapArshaler.func3
     0.12s  1.27%  1.69%      2.88s 30.44%  runtime.mallocgc
         0     0%  1.69%      2.35s 24.84%  regexp.Compile (inline)
         0     0%  1.69%      2.35s 24.84%  regexp.MustCompile
     0.06s  0.63%  2.33%      2.35s 24.84%  regexp.compile

$ curl http://127.0.0.1:6060/debug/pprof/goroutine?debug=1 | head -1
goroutine profile: total 11

$ go tool pprof -sample_index=inuse_space -top http://127.0.0.1:6060/debug/pprof/heap
Type: inuse_space
Showing nodes accounting for 5868.93kB, 100% of 5868.93kB total
Showing top 5 nodes out of 27
      flat  flat%   sum%        cum   cum%
 3308.62kB 56.38% 56.38%  5356.88kB 91.28%  curso-go-rendimiento/resumen.ResumenV1
 1024.11kB 17.45% 73.82%  1024.11kB 17.45%  regexp/syntax.(*parser).newRegexp (inline)
  512.14kB  8.73% 82.55%   512.14kB  8.73%  reflect.mapassign_faststr0
  512.05kB  8.72% 91.28%   512.05kB  8.72%  runtime.mallocgc
  512.01kB  8.72%   100%   512.01kB  8.72%  regexp.(*Regexp).replaceAll

$ curl -o $T/traza.out 'http://127.0.0.1:6060/debug/pprof/trace?seconds=2'
$ go tool trace -pprof=sched $T/traza.out > $T/sched.prof && go tool pprof -top -nodecount=3 $T/sched.prof
Type: delay
Showing nodes accounting for 4.26s, 56.74% of 7.50s total
Dropped 227 nodes (cum <= 0.04s)
Showing top 3 nodes out of 121
      flat  flat%   sum%        cum   cum%
     1.95s 26.03% 26.03%      1.95s 26.03%  runtime.systemstack_switch
     1.25s 16.64% 42.67%      1.25s 16.64%  runtime.goschedIfBusy
     1.06s 14.06% 56.74%      1.73s 23.03%  runtime.gcMarkDone

## La carga mientras se tomaba el perfil
http://127.0.0.1:8080/resumen  4 clientes · 14s:  1453 peticiones   103.8 pet/s  p50   36.0 ms  p99   87.0 ms  respuestas mal 0  errores 0
```

## Leccion 10 — El antes y el despues, con trafico

`sh 10_carga/comandos.sh` · codigo de salida `0`

```
$ go build -o $T/carga ./10_carga

$ $T/carga -version V1 -clientes 8 -duracion 5s
V1  8 clientes · 5s:   558 peticiones   111.6 pet/s  p50   66.9 ms  p99  167.4 ms  respuestas mal 0  errores 0

$ $T/carga -version V3 -clientes 8 -duracion 5s
V3  8 clientes · 5s:  2162 peticiones   432.4 pet/s  p50   16.7 ms  p99   49.3 ms  respuestas mal 0  errores 0

$ $T/carga -version V5 -clientes 8 -duracion 5s
V5  8 clientes · 5s:  2829 peticiones   565.8 pet/s  p50   13.3 ms  p99   30.9 ms  respuestas mal 0  errores 0
```

