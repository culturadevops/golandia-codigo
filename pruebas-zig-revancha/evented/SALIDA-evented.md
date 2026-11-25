# SALIDA — ¿std.Io.Evented en Zig 0.17? (prueba ejecutada 2026-11-05 17:41)

Programa: `evented/prueba.zig` (iniciar Io.Evented y escuchar en el 8080). Compilado en `golandia-zig:0.17.0` (tarball oficial 0.17.0).

```
$ zig build-exe prueba.zig -O ReleaseFast -target x86_64-linux-musl
0.17.0
/opt/zig/lib/std/Io/Uring.zig:759:14: error: no field named 'processReplacePath' in struct 'Io.VTable'
            .processReplacePath = processReplacePath,
             ^~~~~~~~~~~~~~~~~~
/opt/zig/lib/std/Io.zig:51:20: note: struct declared here
pub const VTable = struct {
                   ^~~~~~
[código de salida: 1]
```

Y aunque compilara: lo que la librería estándar 0.17.0 declara en `lib/std/Io/Uring.zig` (la implementación Evented de Linux):

```
774:            .netListenIp = netListenIpUnavailable,
775:            .netAccept = netAcceptUnavailable,
2120:        .net_write => @panic("TODO implement net_write operation"),
2523:                .net_write => @panic("TODO"),
4995:    return error.NetworkDown;
5007:    return error.NetworkDown;
```
