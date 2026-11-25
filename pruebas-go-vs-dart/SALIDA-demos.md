# SALIDA — demos de pruebas-go-vs-dart (golandia)

> Corrida del 2026-11-05 08:16. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py dart`. Cada bloque es la salida literal del guion de `demos/`.

## Dart: null safety contra el nil de Go

Imagen `dart:3.13.7`, `--cpus 1`, guion `demos/nulos.sh`.

```
### dart analyze
Analyzing n...

  error - bin/nulos.dart:13:20 - The property 'nombre' can't be unconditionally accessed because the receiver can be 'null'. Try making the access conditional (using '?.') or adding a null check to the target ('!'). - unchecked_use_of_nullable_value

1 issue found.
[código de salida: 3]
### dart compile exe bin/nulos.dart
bin/nulos.dart:13:20: Error: Property 'nombre' cannot be accessed on 'Usuario?' because it is potentially null.
 - 'Usuario' is from 'bin/nulos.dart'.
Try accessing using ?. instead.
  print('hola, ${u.nombre}');
                   ^^^^^^
Error: AOT compilation failed
Bad state: Generating AOT kernel dill failed!
[código de salida: 254]
```

## Go: el mismo código con nil

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/nulos_go.sh`.

```
### go vet .
[código de salida: 0]
### go build . && ./demo
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x499efb]
[código de salida: 2]
```

## Dart: un isolate no comparte memoria; Go sí

Imagen `dart:3.13.7`, `--cpus 1`, guion `demos/isolates.sh`.

```
### dart run bin/isolates.dart
cada isolate vio su propio contador: [1000, 1000]
el contador del isolate principal: 0
[código de salida: 0]
```

## Go: goroutines que comparten memoria

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/isolates_go.sh`.

```
### go run .
el contador compartido: 2000
[código de salida: 0]
### go run -race .
WARNING: DATA RACE
el contador compartido: 2000
Found 1 data race(s)
[código de salida: 1]
```

## Líneas de cada servidor

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/lineas.sh`.

```
90 líneas  go/main.go
67 líneas  bin/servidor.dart
```

