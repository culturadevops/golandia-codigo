// La variante justa de la revancha de Zig: pool de N hilos CON eventos. Cada hilo tiene su epoll y su socket de escucha
// (SO_REUSEPORT, el núcleo reparte las conexiones) y atiende miles de conexiones sin bloquearse: un hilo solo se ocupa de una
// conexión cuando hay algo que leer o escribir. Es lo que `std.Io.Evented` debería dar y en Zig 0.17.0 no da (evented/SALIDA-evented.md),
// así que va escrito a mano sobre std.os.linux. El trabajo (/suma, /trabajo) es EL MISMO código de zig/main.zig.
// /espera no duerme el hilo: la conexión sale de epoll y entra en una lista de plazos; epoll_wait despierta cuando vence el primero.
const std = @import("std");
const Io = std.Io;
const linux = std.os.linux;
const m = @import("main.zig");

const Conn = struct {
    fd: i32,
    n: usize = 0,
    salida: []const u8 = &.{},
    enviado: usize = 0,
    cerrar: bool = false,
    vence: i96 = 0,
    ms: usize = 0,
    propia: ?[]u8 = null, // copia de la salida si write no la pudo mandar entera
    entrada: [8192]u8 = undefined,
};

const Hilo = struct {
    io: Io,
    t0: Io.Timestamp,
    ep: i32,
    lfd: i32,
    plazos: std.ArrayList(*Conn) = .empty,
    primero: usize = 0,
    // Un arena por HILO, no por conexión: una petición corre entera sin ceder el hilo, así que 300 conexiones no retienen 300 arenas.
    arena: std.heap.ArenaAllocator,

    fn ahora(h: *Hilo) i96 {
        return h.t0.durationTo(Io.Timestamp.now(h.io, .awake)).nanoseconds;
    }

    fn cerrar(_: *Hilo, c: *Conn) void {
        _ = linux.close(c.fd);
        if (c.propia) |b| std.heap.smp_allocator.free(b);
        std.heap.smp_allocator.destroy(c);
    }

    fn vigilar(h: *Hilo, op: u32, c: *Conn, eventos: u32) void {
        var ev: linux.epoll_event = .{ .events = eventos, .data = .{ .ptr = @intFromPtr(c) } };
        _ = linux.epoll_ctl(h.ep, op, c.fd, &ev);
    }

    fn aceptar(h: *Hilo) void {
        while (true) {
            const rc = linux.accept4(h.lfd, null, null, linux.SOCK.NONBLOCK | linux.SOCK.CLOEXEC);
            if (linux.errno(rc) != .SUCCESS) return;
            const c = std.heap.smp_allocator.create(Conn) catch {
                _ = linux.close(@intCast(rc));
                continue;
            };
            c.* = .{ .fd = @intCast(rc) };
            h.vigilar(linux.EPOLL.CTL_ADD, c, linux.EPOLL.IN);
        }
    }

    fn leer(h: *Hilo, c: *Conn) void {
        while (true) {
            if (c.n == c.entrada.len) return h.cerrar(c);
            const rc = linux.read(c.fd, c.entrada[c.n..].ptr, c.entrada.len - c.n);
            switch (linux.errno(rc)) {
                .SUCCESS => {
                    if (rc == 0) return h.cerrar(c);
                    c.n += rc;
                },
                .AGAIN => break,
                .INTR => continue,
                else => return h.cerrar(c),
            }
        }
        h.procesar(c);
    }

    /// Atiende las peticiones completas que haya en la entrada.
    fn procesar(h: *Hilo, c: *Conn) void {
        while (true) {
            const fin = std.mem.indexOf(u8, c.entrada[0..c.n], "\r\n\r\n") orelse return;
            _ = h.arena.reset(.retain_capacity);
            const a = h.arena.allocator();
            var lineas = std.mem.splitSequence(u8, c.entrada[0..fin], "\r\n");
            var partes = std.mem.splitScalar(u8, lineas.next() orelse return h.cerrar(c), ' ');
            _ = partes.next();
            const ruta = a.dupe(u8, partes.next() orelse return h.cerrar(c)) catch return h.cerrar(c);
            c.cerrar = false;
            while (lineas.next()) |l| {
                if (std.ascii.startsWithIgnoreCase(l, "connection:") and std.ascii.endsWithIgnoreCase(l, "close")) c.cerrar = true;
            }
            const resto = c.n - (fin + 4);
            std.mem.copyForwards(u8, c.entrada[0..resto], c.entrada[fin + 4 .. c.n]);
            c.n = resto;
            if (std.mem.startsWith(u8, ruta, "/espera")) {
                c.ms = m.parametro(ruta, "ms");
                c.vence = h.ahora() + @as(i96, @intCast(c.ms)) * std.time.ns_per_ms;
                h.vigilar(linux.EPOLL.CTL_DEL, c, 0);
                h.plazos.append(std.heap.smp_allocator, c) catch return h.cerrar(c);
                var i = h.plazos.items.len - 1; // casi siempre ya va al final: todos esperan lo mismo
                while (i > h.primero and h.plazos.items[i - 1].vence > c.vence) : (i -= 1) {
                    std.mem.swap(*Conn, &h.plazos.items[i - 1], &h.plazos.items[i]);
                }
                return;
            }
            const json = (if (std.mem.startsWith(u8, ruta, "/suma"))
                m.suma(h.io, a, m.parametro(ruta, "n"))
            else if (std.mem.startsWith(u8, ruta, "/trabajo"))
                m.trabajo(h.io, a, m.parametro(ruta, "n"))
            else
                return h.cerrar(c)) catch return h.cerrar(c);
            if (!h.responder(c, json)) return;
        }
    }

    /// Devuelve true si la respuesta salió entera y la conexión sigue viva.
    fn responder(h: *Hilo, c: *Conn, json: []const u8) bool {
        c.salida = std.fmt.allocPrint(h.arena.allocator(), "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {d}\r\nConnection: {s}\r\n\r\n{s}", .{ json.len, if (c.cerrar) "close" else "keep-alive", json }) catch {
            h.cerrar(c);
            return false;
        };
        c.enviado = 0;
        return h.escribir(c);
    }

    fn escribir(h: *Hilo, c: *Conn) bool {
        while (c.enviado < c.salida.len) {
            const rc = linux.write(c.fd, c.salida[c.enviado..].ptr, c.salida.len - c.enviado);
            switch (linux.errno(rc)) {
                .SUCCESS => c.enviado += rc,
                .AGAIN => {
                    if (c.propia == null) { // lo que falta sobrevive al arena del hilo
                        const b = std.heap.smp_allocator.dupe(u8, c.salida[c.enviado..]) catch {
                            h.cerrar(c);
                            return false;
                        };
                        c.propia = b;
                        c.salida = b;
                        c.enviado = 0;
                    }
                    h.vigilar(linux.EPOLL.CTL_MOD, c, linux.EPOLL.OUT);
                    return false;
                },
                .INTR => continue,
                else => {
                    h.cerrar(c);
                    return false;
                },
            }
        }
        if (c.propia) |b| {
            std.heap.smp_allocator.free(b);
            c.propia = null;
        }
        if (c.cerrar) {
            h.cerrar(c);
            return false;
        }
        return true;
    }

    fn vencidos(h: *Hilo) void {
        const t = h.ahora();
        while (h.primero < h.plazos.items.len and h.plazos.items[h.primero].vence <= t) {
            const c = h.plazos.items[h.primero];
            h.primero += 1;
            h.vigilar(linux.EPOLL.CTL_ADD, c, linux.EPOLL.IN);
            _ = h.arena.reset(.retain_capacity);
            const json = std.fmt.allocPrint(h.arena.allocator(), "{{\"lang\":\"zig\",\"n\":{d},\"suma\":0,\"ns\":0}}", .{c.ms}) catch {
                h.cerrar(c);
                continue;
            };
            if (h.responder(c, json)) h.procesar(c);
        }
        if (h.primero == h.plazos.items.len) {
            h.plazos.clearRetainingCapacity();
            h.primero = 0;
        }
    }

    fn correr(h: *Hilo) void {
        var eventos: [256]linux.epoll_event = undefined;
        while (true) {
            var espera: i32 = -1;
            if (h.primero < h.plazos.items.len) {
                const falta = h.plazos.items[h.primero].vence - h.ahora();
                espera = if (falta <= 0) 0 else @intCast(@divTrunc(falta, std.time.ns_per_ms) + 1);
            }
            const rc = linux.epoll_wait(h.ep, &eventos, eventos.len, espera);
            if (linux.errno(rc) != .SUCCESS) continue;
            for (eventos[0..rc]) |ev| {
                if (ev.data.ptr == 0) {
                    h.aceptar();
                    continue;
                }
                const c: *Conn = @ptrFromInt(ev.data.ptr);
                if (ev.events & linux.EPOLL.OUT != 0) {
                    if (h.escribir(c)) {
                        h.vigilar(linux.EPOLL.CTL_MOD, c, linux.EPOLL.IN);
                        h.procesar(c);
                    }
                } else h.leer(c);
            }
            h.vencidos();
        }
    }
};

fn escucha() !i32 {
    const fd: i32 = @intCast(linux.socket(linux.AF.INET, linux.SOCK.STREAM | linux.SOCK.NONBLOCK | linux.SOCK.CLOEXEC, 0));
    const uno: u32 = 1;
    _ = linux.setsockopt(fd, linux.SOL.SOCKET, linux.SO.REUSEADDR, std.mem.asBytes(&uno), 4);
    _ = linux.setsockopt(fd, linux.SOL.SOCKET, linux.SO.REUSEPORT, std.mem.asBytes(&uno), 4);
    var dir: linux.sockaddr.in = .{ .port = std.mem.nativeToBig(u16, 8080), .addr = 0 };
    if (linux.errno(linux.bind(fd, @ptrCast(&dir), @sizeOf(linux.sockaddr.in))) != .SUCCESS) return error.Bind;
    if (linux.errno(linux.listen(fd, 4096)) != .SUCCESS) return error.Listen;
    return fd;
}

fn hilo(io: Io, t0: Io.Timestamp) void {
    const lfd = escucha() catch return;
    const ep: i32 = @intCast(linux.epoll_create1(0));
    var ev: linux.epoll_event = .{ .events = linux.EPOLL.IN, .data = .{ .ptr = 0 } };
    _ = linux.epoll_ctl(ep, linux.EPOLL.CTL_ADD, lfd, &ev);
    var h: Hilo = .{ .io = io, .t0 = t0, .ep = ep, .lfd = lfd, .arena = std.heap.ArenaAllocator.init(std.heap.smp_allocator) };
    h.correr();
}

pub fn main(init: std.process.Init) !void {
    const io = init.io;
    const t0 = Io.Timestamp.now(io, .awake);
    const hilos = std.fmt.parseInt(usize, init.environ_map.get("HILOS") orelse "2", 10) catch 2;
    for (1..hilos) |_| {
        const t = try std.Thread.spawn(.{}, hilo, .{ io, t0 });
        t.detach();
    }
    const listo = t0.durationTo(Io.Timestamp.now(io, .awake)).nanoseconds;
    const us: u64 = @intCast(@divTrunc(listo, 1000));
    std.debug.print("LISTO_EN_MS {d}.{d:0>3} modo=eventos hilos={d}\n", .{ us / 1000, us % 1000, hilos });
    hilo(io, t0);
}
