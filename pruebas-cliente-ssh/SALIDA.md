# SALIDA — Cliente SSH en Go (golandia 682, 2026-10-26)

Ejecutado el 2026-10-26 con `python correr.py` (dos pasadas completas; los números de abajo son de la segunda y coinciden con la primera
en ±0,1 s salvo donde se indica). Las salidas sin editar están en `salidas/*.txt`.

## Versiones
- Docker 29.6.2 (24 CPU). Red propia `g682` (10.68.2.0/24), borrada al terminar junto con los 52 contenedores `g682-*`.
- Servidores: imagen `g682-sshd:1` = alpine:3.22 + openssh-server 10.0_p1-r10 (banner `SSH-2.0-OpenSSH_10.0`), 16,8 MB.
  Cada contenedor genera sus propias claves de host al arrancar. Solo clave pública (PasswordAuthentication no). Usuario `admin`.
- Máquina de control: `g682-cliente:1` = python:3.13.16-slim-trixie + openssh-client 1:10.0p1-7+deb13u4 + ansible-core 2.21.5.
- Cliente Go: `flota/main.go` (266 líneas), golang:1.27.2, golang.org/x/crypto v0.58.0 (única dependencia en `go version -m`).
  Binario `CGO_ENABLED=0 -trimpath -ldflags='-s -w'`: 4.612.256 bytes.

## El laboratorio
- 50 servidores s01…s50. La tarea `revisar` tarda 1 s y escribe `sNN disco 41%`.
- s13 «lleno»: escribe `disco 97%`, una línea en stderr y sale con 2.
- s48 congelado: `docker pause` → el kernel acepta la conexión TCP pero sshd nunca manda el banner.
- s49 lento: `revisar` tarda 30 s.
- s50 apagado: `docker stop` → nadie contesta en su IP (10.68.2.60).
- known_hosts: `ssh-keyscan` de los 50 con todos sanos → 150 claves (rsa, ecdsa y ed25519 de cada uno).
- «hosts-sanos.txt» = s01…s47 (47 hosts; incluye s13).

## 1 · Clave pública y código de salida (OpenSSH y flota)
```
$ ssh -o BatchMode=yes admin@s01 revisar; echo "código de salida: $?"
s01 disco 41%
código de salida: 0
$ ssh -o BatchMode=yes admin@s13 revisar; echo "código de salida: $?"
s13 disco 97%
revisar: disco casi lleno
código de salida: 2
$ flota -hosts s13.txt; echo "código de salida: $?"
s13      2     1.1s  s13 disco 97%
s13    stderr        revisar: disco casi lleno
1 hosts · 0 ok · 1 con código distinto de 0 · 0 sin respuesta · 1.1s en total
código de salida: 1
```
(flota sale con 0 si todo bien, 1 si algún host dio código distinto de 0, 2 si alguno no respondió.)

## 2 · En fila frente a paralelo (47 hosts sanos, semáforo de tamaño p)
```
-p 1    50.4s   (primera pasada 50.5s)
-p 5    10.7s   10.7s
-p 10    5.4s    5.4s
-p 25    2.2s    2.2s
-p 50    1.1s    1.1s   (primera pasada 1.1s y 1.2s)
47 hosts · 46 ok · 1 con código distinto de 0 · 0 sin respuesta
```

## 3 · Salida ordenada
`-llegada` (imprime al terminar cada uno): `s38 s15 s32 s23 s24 s36 s39 s44 …` (otra pasada: `s40 s12 s28 s02 s46 …`).
Por defecto (resultados en un slice por índice): `s01 s02 s03 s04 s05 s06 s07 s08 …`.

## 4 · Los 50, con los tres que fallan (-conectar 3s -timeout 10s)
```
$ flota -hosts hosts-50.txt -p 50; echo "código de salida: $?"
s01      0     1.1s  s01 disco 41%
…
s13      2     1.1s  s13 disco 97%
s13    stderr        revisar: disco casi lleno
…
s47      0     1.1s  s47 disco 41%
s48    ERR     3.0s  ssh: handshake failed: read tcp 10.68.2.5:54638->10.68.2.58:22: i/o timeout
s49    ERR    10.0s  el comando no terminó en 10s
s50    ERR     3.0s  dial tcp 10.68.2.60:22: i/o timeout
50 hosts · 46 ok · 1 con código distinto de 0 · 3 sin respuesta · 10.0s en total
código de salida: 2
$ flota -hosts hosts-50.txt -p 1 | tail -1
50 hosts · 46 ok · 1 con código distinto de 0 · 3 sin respuesta · 66.5s en total   (primera pasada 66.6s)
```

## 5 · La trampa: `ssh.Dial` con `ClientConfig.Timeout` y nada más (`-ingenuo`)
```
$ time timeout 60 flota -ingenuo -hosts s48.txt; echo "código de salida: $?"
real	0m59.999s
código de salida: 124          ← seguía colgado en el handshake; lo mató timeout a los 60 s
$ flota -ingenuo -hosts s49.txt
s49      0    30.1s  s49 disco 41%
$ flota -ingenuo -hosts s50.txt
s50    ERR     3.0s  dial tcp 10.68.2.60:22: i/o timeout
```
`ClientConfig.Timeout` cubre la conexión TCP (s50 cae a los 3 s), pero no el handshake (s48 se cuelga) ni el comando (s49 espera los 30 s).
El arreglo de `flota`: `net.Dialer` + `conn.SetDeadline` durante `ssh.NewClientConn`, y un `context.WithTimeout` que cierra el cliente.

## 6 · HostKeyCallback: un impostor en la IP de s07
Se apaga s07 y otra máquina (sus propias claves de host, acepta la clave del cliente) toma su IP 10.68.2.17.
```
$ flota -inseguro -hosts s07.txt; echo "código de salida: $?"
s07      0     1.1s  s07 disco 12% · todo bien (soy el impostor)
1 hosts · 1 ok · 0 con código distinto de 0 · 0 sin respuesta · 1.1s en total
código de salida: 0
$ flota -hosts s07.txt; echo "código de salida: $?"          # knownhosts.New
s07    ERR     0.0s  ssh: handshake failed: knownhosts: key mismatch
código de salida: 2
$ ssh -o BatchMode=yes admin@s07 revisar
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!
Host key verification failed.
código de salida: 255
$ flota -known known_hosts_vacio -hosts s01.txt
s01    ERR     0.0s  ssh: handshake failed: knownhosts: key is unknown
$ (impostor fuera, s07 de vuelta) flota -hosts s07.txt
s07      0     1.1s  s07 disco 41%
```

## 7 · Agente
```
$ eval $(ssh-agent -s); ssh-add /root/.ssh/id_ed25519
Identity added: /root/.ssh/id_ed25519 (flota-682)
$ flota -agente -clave /no/existe -hosts s01.txt
s01      0     1.1s  s01 disco 41%
```

## 8 · Comando frente a terminal (RequestPty)
```
$ flota -hosts s01.txt -cmd 'tty; echo al-stderr >&2'
s01      0     0.1s  not a tty
s01    stderr        al-stderr
$ flota -hosts s01.txt -pty -cmd 'tty; echo al-stderr >&2'
s01      0     0.1s  /dev/pts/0 | al-stderr        ← con pty, stderr llega mezclado en stdout
```

## 9 · Bash con ssh (`bucle.sh` en fila, `paralelo.sh` con xargs -P; ConnectTimeout=3 y timeout 10)
```
47 sanos, en fila           52.752s   (primera pasada 52.971s)
47 sanos, xargs -P 10        5.647s   (5.625s)
47 sanos, xargs -P 50        1.167s   (1.168s)
50, en fila                 68.791s   (68.755s)
50, xargs -P 50             10.046s   (10.042s)
s48 255 Connection timed out during banner exchange      ← ConnectTimeout de OpenSSH sí cubre el banner
s49 124                                                    ← el lento lo corta `timeout 10`, no ssh
s50 255 ssh: connect to host s50 port 22: Connection timed out
```
xargs -P imprime en orden de llegada (`s45 s47`, `s47 s46`, `s39 s41` como últimas líneas).

## 10 · Ansible (ansible-core 2.21.5, `-m raw -a revisar`, no necesita Python en el servidor)
```
47 sanos, -f 5 (el valor por defecto)   11.858s   (primera pasada 11.815s)   46 con rc=0
47 sanos, -f 10                           5.481s   (5.450s)
47 sanos, -f 50                           1.617s y 1.562s   (1.539s y 1.556s)
50, -f 50 -T 3                           30.700s   (30.722s)
s13 | FAILED | rc=2
s48 | UNREACHABLE!: … Connection timed out during banner exchange
s50 | UNREACHABLE!: … ssh: connect to host s50 port 22: Connection timed out
```
`-T 3` corta la conexión, pero el comando lento (s49) se esperó entero: 30,7 s. En 2.21.5, `-o` avisa que está obsoleto (se quita en 2.23).
