// ¿Se puede escribir el servidor con std.Io.Evented (io_uring + fibras) en Zig 0.17? Prueba mínima: escuchar en un puerto.
const std = @import("std");
const Io = std.Io;

pub fn main(init: std.process.Init) !void {
    var ev: Io.Evented = undefined;
    try ev.init(init.gpa, .{});
    defer ev.deinit();
    const io = ev.io();
    const direccion = try Io.net.IpAddress.parse("0.0.0.0", 8080);
    var server = direccion.listen(io, .{ .reuse_address = true }) catch |err| {
        std.debug.print("Evented ({s}): listen falló: error.{s}\n", .{ @typeName(Io.Evented), @errorName(err) });
        return err;
    };
    defer server.deinit(io);
    std.debug.print("Evented: escuchando\n", .{});
}
