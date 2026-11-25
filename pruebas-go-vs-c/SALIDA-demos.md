# SALIDA — demos de pruebas-go-vs-c (golandia)

> Corrida del 2026-11-05 08:16. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py c`. Cada bloque es la salida literal del guion de `demos/`.

## C: desborde, memoria liberada y dirección de una local (sin y con AddressSanitizer)

Imagen `gcc:16.2.0`, `--cpus 1`, guion `demos/correr.sh`.

```
### gcc -O2 -Wall -Wextra desborde.c
[código de salida de gcc: 0]
### gcc -O2 -Wall -Wextra liberado.c
/d/demos/liberado.c: In function 'main':
/d/demos/liberado.c:12:5: warning: pointer 'saldo' used after 'free' [-Wuse-after-free=]
   12 |     printf("leo lo liberado: %s\n", saldo);
      |     ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
/d/demos/liberado.c:9:5: note: call to 'free' here
    9 |     free(saldo);
      |     ^~~~~~~~~~~
[código de salida de gcc: 0]
### gcc -O2 -Wall -Wextra local.c
/d/demos/local.c: In function 'nuevo':
/d/demos/local.c:6:12: warning: function returns address of local variable [-Wreturn-local-addr]
    6 |     return &x;
      |            ^~
[código de salida de gcc: 0]
### ./desborde invitado
nombre=invitado admin=0
[código de salida: 0]
### ./desborde invitado-xyz
nombre=invitado-xyz admin=2054780973
[código de salida: 0]
### ./liberado
leo lo liberado: ±
[código de salida: 0]
### ./local
[código de salida: 139]
### gcc -fsanitize=address … && ./desborde-asan invitado-xyz
=================================================================
==36==ERROR: AddressSanitizer: stack-buffer-overflow on address 0x6dec096f002c at pc 0x71ec0bf8e6e2 bp 0x7ffd97c95fb0 sp 0x7ffd97c95770
WRITE of size 13 at 0x6dec096f002c thread T0
    #1 0x000000401271 in main /d/demos/desborde.c:11
[código de salida: 1]
### gcc -fsanitize=address … && ./liberado-asan
=================================================================
==39==ERROR: AddressSanitizer: heap-use-after-free on address 0x7259f01e0010 at pc 0x7639f1480560 bp 0x7ffc329fc1a0 sp 0x7ffc329fb960
READ of size 2 at 0x7259f01e0010 thread T0
[código de salida: 1]
Segmentation fault (core dumped)
```

## Go: los mismos tres errores

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/correr_go.sh`.

```
### go vet ./...
[código de salida: 0]
### go run .
a=42 b=7
panic: runtime error: index out of range [8] with length 8
exit status 2
[código de salida: 1]
```

## Líneas de cada servidor

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/lineas.sh`.

```
90 líneas  go/main.go
125 líneas  src/servidor.c
```

