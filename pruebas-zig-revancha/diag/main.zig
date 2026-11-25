// GENERADO de zig/main.zig: el mismo servidor con contadores de tareas, SOLO para diagnosticar (no se mide con este).
// La revancha de Zig (golandia, pedido de un comentario del vídeo 487): el MISMO contrato que el servidor de Go del banco
// (recursos/pruebas-go-java-revancha/go/main.go): /suma (CPU), /trabajo (objetos, mapa, orden y JSON a mano) y /espera.
// Zig 0.17, solo la librería estándar (std.Io con su implementación por defecto, Threaded).
//
// Dos formas de repartir las conexiones, elegidas con la variable MODO:
//   MODO=hilos  → la del vídeo 487: una tarea concurrente por conexión (Io.Group.concurrent). Con Threaded, cada tarea
//                 ocupa un hilo del sistema mientras dura la conexión.
//   MODO=pool   → un pool fijo de HILOS hilos (por defecto 2, uno por CPU del contenedor): el hilo principal acepta y mete
//                 la conexión en una cola (Io.Queue); los hilos del pool la sacan y la atienden de principio a fin.
// HTTP/1.1 con keep-alive: se atienden peticiones en la misma conexión hasta que el cliente pide «Connection: close» o cierra.
const std = @import("std");
const Io = std.Io;
const net = std.Io.net;

const Item = struct { id: usize, nombre: []const u8, valor: usize };

pub fn parametro(ruta: []const u8, clave: []const u8) usize {
    const q = std.mem.indexOfScalar(u8, ruta, '?') orelse return 0;
    var it = std.mem.splitScalar(u8, ruta[q + 1 ..], '&');
    while (it.next()) |par| {
        const igual = std.mem.indexOfScalar(u8, par, '=') orelse continue;
        if (std.mem.eql(u8, par[0..igual], clave)) return std.fmt.parseInt(usize, par[igual + 1 ..], 10) catch 0;
    }
    return 0;
}

fn menor(_: void, a: *const Item, b: *const Item) bool {
    if (a.valor != b.valor) return a.valor < b.valor;
    return a.id < b.id;
}

pub fn suma(io: Io, a: std.mem.Allocator, n: usize) ![]const u8 {
    const t0 = Io.Timestamp.now(io, .awake);
    var s: u64 = 0;
    var i: u64 = 1;
    while (i <= n) : (i += 1) s += i * i % 1000003;
    const ns = t0.durationTo(Io.Timestamp.now(io, .awake)).nanoseconds;
    return std.fmt.allocPrint(a, "{{\"lang\":\"zig\",\"n\":{d},\"suma\":{d},\"ns\":{d}}}", .{ n, s, ns });
}

pub fn trabajo(io: Io, a: std.mem.Allocator, n: usize) ![]const u8 {
    const t0 = Io.Timestamp.now(io, .awake);
    var m = std.StringHashMap(*Item).init(a);
    try m.ensureTotalCapacity(@intCast(n));
    for (0..n) |i| {
        const it = try a.create(Item);
        it.* = .{ .id = i, .nombre = try std.fmt.allocPrint(a, "item-{d}", .{i}), .valor = (i * 7919) % 1000 };
        try m.put(it.nombre, it);
    }
    const lista = try a.alloc(*const Item, m.count());
    var total: usize = 0;
    var k: usize = 0;
    var iter = m.valueIterator();
    while (iter.next()) |v| : (k += 1) {
        lista[k] = v.*;
        total += v.*.valor;
    }
    std.mem.sortUnstable(*const Item, lista, {}, menor);
    var b: std.ArrayList(u8) = .empty;
    try b.print(a, "{{\"lang\":\"zig\",\"n\":{d},\"total\":{d},\"primeros\":[", .{ n, total });
    for (lista[0..@min(50, lista.len)], 0..) |it, i| {
        if (i > 0) try b.append(a, ',');
        try b.print(a, "{{\"id\":{d},\"nombre\":\"{s}\",\"valor\":{d}}}", .{ it.id, it.nombre, it.valor });
    }
    const ns = t0.durationTo(Io.Timestamp.now(io, .awake)).nanoseconds;
    try b.print(a, "],\"ns\":{d}}}", .{ns});
    return b.items;
}

/// Atiende UNA conexión: peticiones una tras otra mientras haya keep-alive.
fn atender(io: Io, stream: net.Stream) void {
    defer stream.close(io);
    var rbuf: [8192]u8 = undefined;
    var wbuf: [8192]u8 = undefined;
    var lector = stream.reader(io, &rbuf);
    var escritor = stream.writer(io, &wbuf);
    const r = &lector.interface;
    const w = &escritor.interface;
    // Un arena por conexión, vaciado tras cada petición: lo idiomático en Zig para memoria «de la petición».
    var arena = std.heap.ArenaAllocator.init(std.heap.smp_allocator);
    defer arena.deinit();
    while (true) {
        _ = arena.reset(.retain_capacity);
        const a = arena.allocator();
        const primera = r.takeDelimiterInclusive('\n') catch return;
        // «GET /suma?n=200000 HTTP/1.1»: la ruta va entre los dos espacios
        var partes = std.mem.splitScalar(u8, std.mem.trimEnd(u8, primera, "\r\n"), ' ');
        _ = partes.next();
        const ruta = a.dupe(u8, partes.next() orelse return) catch return;
        // Las cabeceras: solo interesa «Connection: close»
        var cerrar = false;
        while (true) {
            const h = std.mem.trimEnd(u8, r.takeDelimiterInclusive('\n') catch return, "\r\n");
            if (h.len == 0) break;
            if (std.ascii.startsWithIgnoreCase(h, "connection:") and std.ascii.endsWithIgnoreCase(h, "close")) cerrar = true;
        }
        const json = (if (std.mem.startsWith(u8, ruta, "/suma"))
            suma(io, a, parametro(ruta, "n"))
        else if (std.mem.startsWith(u8, ruta, "/trabajo"))
            trabajo(io, a, parametro(ruta, "n"))
        else if (std.mem.startsWith(u8, ruta, "/espera")) blk: {
            const ms = parametro(ruta, "ms");
            io.sleep(.{ .nanoseconds = @as(i96, @intCast(ms)) * std.time.ns_per_ms }, .awake) catch return;
            break :blk std.fmt.allocPrint(a, "{{\"lang\":\"zig\",\"n\":{d},\"suma\":0,\"ns\":0}}", .{ms});
        } else return) catch return;
        w.print("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {d}\r\nConnection: {s}\r\n\r\n{s}", .{
            json.len, if (cerrar) "close" else "keep-alive", json,
        }) catch return;
        w.flush() catch return;
        if (cerrar) return;
    }
}

var aceptadas = std.atomic.Value(i64).init(0);
var terminadas = std.atomic.Value(i64).init(0);
var empezadas = std.atomic.Value(i64).init(0);
var vivas_max = std.atomic.Value(i64).init(0);
var corriendo_max = std.atomic.Value(i64).init(0);

fn atenderTarea(io: Io, stream: net.Stream) Io.Cancelable!void {
    const e = empezadas.fetchAdd(1, .monotonic) + 1;
    _ = corriendo_max.fetchMax(e - terminadas.load(.monotonic), .monotonic);
    atender(io, stream);
    _ = terminadas.fetchAdd(1, .monotonic);
}

fn informar(io: Io) void {
    while (true) {
        io.sleep(.{ .nanoseconds = std.time.ns_per_s }, .awake) catch return;
        const a = aceptadas.load(.monotonic);
        const t = terminadas.load(.monotonic);
        std.debug.print("DIAG aceptadas={d} terminadas={d} vivas={d} vivas_max={d} corriendo_max={d}\n", .{ a, t, a - t, vivas_max.load(.monotonic), corriendo_max.load(.monotonic) });
    }
}

fn trabajador(io: Io, cola: *Io.Queue(net.Stream)) void {
    while (true) {
        const stream = cola.getOneUncancelable(io) catch return;
        atender(io, stream);
    }
}

pub fn main(init: std.process.Init) !void {
    const io = init.io;
    const arranque = Io.Timestamp.now(io, .awake);
    const modo = init.environ_map.get("MODO") orelse "hilos";
    const hilos = std.fmt.parseInt(usize, init.environ_map.get("HILOS") orelse "2", 10) catch 2;
    const direccion = try net.IpAddress.parse("0.0.0.0", 8080);
    // backlog 4096, el mismo que usa Go (lee /proc/sys/net/core/somaxconn); el valor por defecto de Zig es 128.
    var server = try direccion.listen(io, .{ .reuse_address = true, .kernel_backlog = 4096 });
    const listo = arranque.durationTo(Io.Timestamp.now(io, .awake)).nanoseconds;
    const us: u64 = @intCast(@divTrunc(listo, 1000));
    std.debug.print("LISTO_EN_MS {d}.{d:0>3} modo={s} hilos={d}\n", .{ us / 1000, us % 1000, modo, hilos });
    if (std.mem.eql(u8, modo, "pool")) {
        var buf: [4096]net.Stream = undefined;
        var cola: Io.Queue(net.Stream) = .init(&buf);
        for (0..hilos) |_| {
            const t = try std.Thread.spawn(.{}, trabajador, .{ io, &cola });
            t.detach();
        }
        while (true) {
            const stream = server.accept(io) catch continue;
            cola.putOneUncancelable(io, stream) catch stream.close(io);
        }
    } else {
        var grupo: Io.Group = .init;
        (try std.Thread.spawn(.{}, informar, .{io})).detach();
        while (true) {
            const stream = server.accept(io) catch continue;
            const ac = aceptadas.fetchAdd(1, .monotonic) + 1;
            _ = vivas_max.fetchMax(ac - terminadas.load(.monotonic), .monotonic);
            grupo.concurrent(io, atenderTarea, .{ io, stream }) catch stream.close(io);
        }
    }
}
