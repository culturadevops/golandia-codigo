// El servidor de la comparativa, en Zig 0.17: solo la librería estándar (std.Io, la interfaz de E/S nueva).
// El mismo contrato que el de Go: /suma hace CPU de verdad, /espera simula E/S. Una tarea concurrente por conexión.
const std = @import("std");
const Io = std.Io;
const net = std.Io.net;

fn parametro(ruta: []const u8, clave: []const u8) u64 {
    const q = std.mem.indexOfScalar(u8, ruta, '?') orelse return 0;
    var it = std.mem.splitScalar(u8, ruta[q + 1 ..], '&');
    while (it.next()) |par| {
        const igual = std.mem.indexOfScalar(u8, par, '=') orelse continue;
        if (std.mem.eql(u8, par[0..igual], clave)) return std.fmt.parseInt(u64, par[igual + 1 ..], 10) catch 0;
    }
    return 0;
}

fn atender(io: Io, stream: net.Stream) Io.Cancelable!void {
    defer stream.close(io);
    var rbuf: [4096]u8 = undefined;
    var wbuf: [512]u8 = undefined;
    var lector = stream.reader(io, &rbuf);
    var escritor = stream.writer(io, &wbuf);
    const r = &lector.interface;
    const w = &escritor.interface;
    const linea = r.takeDelimiterExclusive('\n') catch return;
    // «GET /suma?n=200000 HTTP/1.1»: la ruta va entre los dos espacios
    var partes = std.mem.splitScalar(u8, linea, ' ');
    _ = partes.next();
    const ruta = partes.next() orelse return;
    var cuerpo: [160]u8 = undefined;
    var json: []const u8 = undefined;
    if (std.mem.startsWith(u8, ruta, "/suma")) {
        const n = parametro(ruta, "n");
        const t0 = Io.Timestamp.now(io, .awake);
        var s: u64 = 0;
        var i: u64 = 1;
        while (i <= n) : (i += 1) {
            s += i * i % 1000003;
        }
        const ns = t0.durationTo(Io.Timestamp.now(io, .awake)).nanoseconds;
        json = std.fmt.bufPrint(&cuerpo, "{{\"lang\":\"zig\",\"n\":{d},\"suma\":{d},\"ns\":{d}}}", .{ n, s, ns }) catch return;
    } else if (std.mem.startsWith(u8, ruta, "/espera")) {
        const ms = parametro(ruta, "ms");
        try io.sleep(.{ .nanoseconds = @as(i96, ms) * std.time.ns_per_ms }, .awake);
        json = std.fmt.bufPrint(&cuerpo, "{{\"lang\":\"zig\",\"n\":{d},\"suma\":0,\"ns\":0}}", .{ms}) catch return;
    } else return;
    w.print("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {d}\r\nConnection: close\r\n\r\n{s}", .{ json.len, json }) catch return;
    w.flush() catch return;
}

pub fn main(init: std.process.Init) !void {
    const io = init.io;
    const arranque = Io.Timestamp.now(io, .awake);
    const direccion = try net.IpAddress.parse("0.0.0.0", 8080);
    var server = try direccion.listen(io, .{ .reuse_address = true });
    // La marca de «listo» del propio proceso: main → escuchando, sin el contenedor por medio.
    const listo = arranque.durationTo(Io.Timestamp.now(io, .awake)).nanoseconds;
    const us: u64 = @intCast(@divTrunc(listo, 1000));
    std.debug.print("LISTO_EN_MS {d}.{d:0>3}\n", .{ us / 1000, us % 1000 });
    var grupo: Io.Group = .init;
    while (true) {
        const stream = server.accept(io) catch continue;
        grupo.concurrent(io, atender, .{ io, stream }) catch stream.close(io);
    }
}
