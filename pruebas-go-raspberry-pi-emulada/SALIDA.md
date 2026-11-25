# SALIDA — Go en una Raspberry Pi, EMULADA con QEMU (golandia)

> Corrida del 2026-11-23 07:37. Regenerable con `python correr.py`. ⛔ Sin placa: los contenedores ARM corren en un PC x86-64 con el QEMU que registra Docker Desktop (binfmt).

- `go version go1.27.2 linux/amd64` (imagen golang:1.27.2, en amd64) · periph.io/x/conn/v3 v3.7.3 · periph.io/x/host/v3 v3.8.5 · Docker 29.6.2 · Debian 13 (debian:trixie-slim)

## 1 · ¿Qué CPU ve cada contenedor? (`uname -m`)

```
--platform linux/amd64   (debian:trixie-slim) → x86_64
--platform linux/arm64   (debian:trixie-slim) → aarch64
--platform linux/arm/v7  (debian:trixie-slim) → armv7l
--platform linux/arm/v6  (alpine:3.22) → armv7l
```

## 2 · Compilación cruzada desde amd64

```
# sin variables:
linux amd64 1
# GOOS=linux GOARCH=arm64:
linux arm64 0
```

```
$ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /w/_bin/sensor-amd64 .
$ GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o /w/_bin/sensor-arm64 .
$ GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags='-s -w' -o /w/_bin/sensor-armv7 .
$ GOOS=linux GOARCH=arm GOARM=6 go build -trimpath -ldflags='-s -w' -o /w/_bin/sensor-armv6 .
```

`file` (recortado a arquitectura y enlazado) y tamaño:
```
ELF 64-bit LSB executable, x86-64, statically linked  6893728 bytes
ELF 64-bit LSB executable, ARM aarch64, statically linked  6750368 bytes
ELF 32-bit LSB executable, ARM, statically linked  7209120 bytes
ELF 32-bit LSB executable, ARM, statically linked  7209120 bytes
```

```
ELF 64-bit LSB executable
ARM aarch64
statically linked
```

Lo que el binario dice de sí mismo (`go version -m`, solo las líneas de la compilación):
```
# sensor-arm64
build	CGO_ENABLED=0
build	GOARCH=arm64
build	GOOS=linux
# sensor-armv7
build	CGO_ENABLED=0
build	GOARCH=arm
build	GOOS=linux
build	GOARM=7
# sensor-armv6
build	CGO_ENABLED=0
build	GOARCH=arm
build	GOOS=linux
build	GOARM=6
```

## 3 · Ejecutar los binarios

```
# sensor-arm64 en --platform linux/arm64 (debian:trixie-slim)
aarch64
calculo n=1000 suma=333833500 en 0s (GOARCH=arm64)
rc=0

# sensor-armv7 en --platform linux/arm/v7 (debian:trixie-slim)
armv7l
calculo n=1000 suma=333833500 en 0s (GOARCH=arm)
rc=0

# sensor-armv6 en --platform linux/arm/v6 (alpine:3.22)
armv7l
calculo n=1000 suma=333833500 en 0s (GOARCH=arm)
rc=0

# sensor-arm64 en --platform linux/amd64 (debian:trixie-slim)
x86_64
calculo n=1000 suma=333833500 en 0s (GOARCH=arm64)
rc=0
```

(La última: el binario ARM64 en un contenedor amd64 NO da «exec format error»: el kernel de Docker Desktop lo pasa a QEMU por binfmt, sin avisar.)

## 4 · La trampa de cgo

```
$ GOARCH=arm64 go build -o /w/_bin/concgo-arm64 .
rc=1
package concgo: build constraints exclude all Go files in /w/concgo
$ CGO_ENABLED=1 GOARCH=arm64 go build -o /w/_bin/concgo-arm64 .
rc=1
# runtime/cgo
gcc_arm64.S: Assembler messages:
gcc_arm64.S:30: Error: no such instruction: `stp x29,x30,[sp,'
gcc_arm64.S:34: Error: operand size mismatch for `mov'
gcc_arm64.S:36: Error: no such instruction: `stp x19,x20,[sp,'
gcc_arm64.S:39: Error: no such instruction: `stp x21,x22,[sp,'
$ CGO_ENABLED=1 GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o /w/_bin/concgo-arm64 .
rc=0
ARM aarch64
dynamically linked
interpreter /lib/ld-linux-aarch64.so.1
```

```
# concgo-arm64 en debian:trixie-slim (linux/arm64)
doble(21) desde C = 42
rc=0

# concgo-arm64 en alpine:3.22 (linux/arm64)
rc=255
qemu-aarch64: Could not open '/lib/ld-linux-aarch64.so.1': No such file or directory
```

## 5 · El servicio con systemd, en una «Pi» emulada (Debian 13 ARM64 con systemd como proceso 1)

La unidad (systemd/sensor.service):
```
[Unit]
Description=Sensor de Golandia (Go)
After=network-online.target

[Service]
ExecStart=/usr/local/bin/sensor -gpio=sim
Restart=always
RestartSec=2
DynamicUser=yes
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes

[Install]
WantedBy=multi-user.target
```

`systemctl is-system-running` → degraded. `systemctl status sensor` a los 8 s:
```
● sensor.service - Sensor de Golandia (Go)
     Loaded: loaded (/etc/systemd/system/sensor.service; enabled; preset: enabled)
     Active: activating (auto-restart) (Result: exit-code) since Mon 2026-11-23 12:38:21 UTC; 183ms ago
 Invocation: 095ee40a296a48c0b29773ec162dd563
    Process: 279 ExecStart=/usr/local/bin/sensor -gpio=sim (code=exited, status=226/NAMESPACE)
   Main PID: 279 (code=exited, status=226/NAMESPACE)
   Mem peak: 2.6M
        CPU: 57ms

NRestarts=4
```

`systemctl --failed` (por qué el sistema está «degraded») y `journalctl -u sensor`:
```
UNIT                                     LOAD   ACTIVE SUB    DESCRIPTION
console-getty.service                    loaded failed failed Console Getty
getty@tty1.service                       loaded failed failed Getty on tty1
getty@tty2.service                       loaded failed failed Getty on tty2
getty@tty3.service                       loaded failed failed Getty on tty3
getty@tty4.service                       loaded failed failed Getty on tty4
getty@tty5.service                       loaded failed failed Getty on tty5
getty@tty6.service                       loaded failed failed Getty on tty6

-- No entries --
No journal files were found.
```

La MISMA unidad y el MISMO binario ARM64, con el systemd nativo de un contenedor amd64 (el binario corre por QEMU):
```
● sensor.service - Sensor de Golandia (Go)
     Loaded: loaded (/etc/systemd/system/sensor.service; enabled; preset: enabled)
     Active: active (running) since Mon 2026-11-23 12:38:22 UTC; 6s ago
 Invocation: 20637e019a1b457aa277a7a7870670bf
   Main PID: 77 (sensor)
      Tasks: 7 (limit: 18855)
     Memory: 12.7M (peak: 13.2M)
        CPU: 92ms
     CGroup: /system.slice/sensor.service
             └─77 /usr/bin/qemu-aarch64 /usr/local/bin/sensor /usr/local/bin/sensor -gpio=sim
```

En la «Pi» emulada, con la unidad sin las líneas de endurecimiento (systemd/sensor-simple.service: solo ExecStart, Restart=always y RestartSec=2):
```
Removed '/etc/systemd/system/multi-user.target.wants/sensor.service'.
Created symlink '/etc/systemd/system/multi-user.target.wants/sensor-simple.service' → '/etc/systemd/system/sensor-simple.service'.
● sensor-simple.service - Sensor de Golandia (Go), sin el endurecimiento
     Loaded: loaded (/etc/systemd/system/sensor-simple.service; enabled; preset: enabled)
     Active: active (running) since Mon 2026-11-23 12:38:30 UTC; 4s ago
 Invocation: 55545a28e9374a81bf35b46c700390dd
   Main PID: 409 (sensor)
      Tasks: 8 (limit: 18855)
     Memory: 13.2M (peak: 13.7M)
        CPU: 135ms
     CGroup: /system.slice/sensor-simple.service
             └─409 /usr/bin/qemu-aarch64 /usr/local/bin/sensor /usr/local/bin/sensor -gpio=sim
```

`curl localhost:8080/estado` dentro de la «Pi»:
```
{
 "goarch": "arm64",
 "goarm": "",
 "goos": "linux",
 "led_encendido": true,
 "modelo": "(no existe: /proc/device-tree/model)",
 "nivel": "High",
 "pid": 409,
 "pin": "GPIO17(17)",
 "temperatura": "(no existe: /sys/class/thermal/thermal_zone0/temp)"
}
```

`kill -9 409` (el proceso del sensor) y, 4 s después:
```
     Active: active (running) since Mon 2026-11-23 12:38:36 UTC; 1s ago
      Tasks: 6 (limit: 18855)

MainPID=456
Result=success
NRestarts=1
```

## 6 · GPIO: lo que el emulador no tiene

```
ls: cannot access '/dev/gpiochip*': No such file or directory
ls: cannot access '/dev/gpiomem': No such file or directory
cat: /proc/device-tree/model: No such file or directory
cat: /sys/class/thermal/thermal_zone0/temp: No such file or directory
```

`sensor -gpio=real` (periph.io/x/host: host.Init() y gpioreg.ByName("GPIO17")) en la «Pi» emulada:
```
2026/11/23 12:38:39 drivers cargados: 0 · omitidos: 13 · con error: 0
2026/11/23 12:38:39 GPIO17 no existe en esta máquina: no hay driver de Raspberry Pi cargado
rc=0
```

## 7 · Una imagen de contenedor para la Pi

`docker buildx build --platform linux/amd64,linux/arm64 -t gl-prac-rpi-sensor:1` (FROM scratch + el binario de cada arquitectura): rc=0

`docker image inspect .Size`: 2.8 MiB
```
--platform linux/arm64 → goarch arm64, pin GPIO17(17)
--platform linux/amd64 → goarch amd64, pin GPIO17(17)
```

