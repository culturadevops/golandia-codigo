// El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en Swift 6.4.1 con Vapor 4.122.3.
// El mismo contrato: /suma (CPU), /trabajo (objetos, diccionario, orden y JSON a mano), /espera.
// Para «¿Go o Swift?» (2026-10-24): «golang vs swift». Swift en el servidor = Vapor sobre SwiftNIO.
import Foundation
import Vapor

let arranque = DispatchTime.now()

final class Item {
    let id: Int
    let nombre: String
    let valor: Int
    init(id: Int, nombre: String, valor: Int) { self.id = id; self.nombre = nombre; self.valor = valor }
}

func ns(desde t0: DispatchTime) -> UInt64 { DispatchTime.now().uptimeNanoseconds - t0.uptimeNanoseconds }

func json(_ s: String) -> Response {
    var h = HTTPHeaders()
    h.add(name: .contentType, value: "application/json")
    return Response(status: .ok, headers: h, body: .init(string: s))
}

let app = try await Application.make(.production)
app.http.server.configuration.hostname = "0.0.0.0"
app.http.server.configuration.port = 8080
app.logger.logLevel = .warning

app.get("suma") { req -> Response in
    let n = req.query[Int.self, at: "n"] ?? 0
    let t0 = DispatchTime.now()
    var s: UInt64 = 0
    if n >= 1 {
        for i in 1...n { s &+= UInt64(i) &* UInt64(i) % 1000003 }
    }
    return json("{\"lang\":\"swift\",\"n\":\(n),\"suma\":\(s),\"ns\":\(ns(desde: t0))}")
}

app.get("trabajo") { req -> Response in
    let n = req.query[Int.self, at: "n"] ?? 0
    let t0 = DispatchTime.now()
    var m = [String: Item](minimumCapacity: n)
    for i in 0..<n {
        let it = Item(id: i, nombre: "item-\(i)", valor: (i * 7919) % 1000)
        m[it.nombre] = it
    }
    var lista = [Item]()
    lista.reserveCapacity(m.count)
    var total = 0
    for it in m.values {
        lista.append(it)
        total += it.valor
    }
    lista.sort { $0.valor != $1.valor ? $0.valor < $1.valor : $0.id < $1.id }
    var b = "{\"lang\":\"swift\",\"n\":\(n),\"total\":\(total),\"primeros\":["
    for i in 0..<min(50, lista.count) {
        if i > 0 { b += "," }
        let it = lista[i]
        b += "{\"id\":\(it.id),\"nombre\":\"\(it.nombre)\",\"valor\":\(it.valor)}"
    }
    b += "],\"ns\":\(ns(desde: t0))}"
    return json(b)
}

app.get("espera") { req async throws -> Response in
    let ms = req.query[Int.self, at: "ms"] ?? 0
    try await Task.sleep(nanoseconds: UInt64(ms) * 1_000_000)
    return json("{\"lang\":\"swift\",\"n\":\(ms),\"suma\":0,\"ns\":0}")
}

FileHandle.standardError.write("LISTO_EN_MS \(String(format: "%.3f", Double(ns(desde: arranque)) / 1e6)) swift hilos=\(System.coreCount)\n".data(using: .utf8)!)
try await app.execute()
try await app.asyncShutdown()
