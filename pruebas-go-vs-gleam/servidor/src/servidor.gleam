// El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en Gleam, para «¿Go o Gleam?» (golandia 807).
// El mismo contrato: /suma (CPU) y /trabajo (objetos, Dict, orden y JSON armado a mano). HTTP con mist, sobre la BEAM.
import gleam/bytes_tree
import gleam/dict
import gleam/erlang/process
import gleam/http/request.{type Request}
import gleam/http/response.{type Response}
import gleam/int
import gleam/io
import gleam/list
import gleam/order
import gleam/result
import gleam/string
import mist

pub type Item {
  Item(id: Int, nombre: String, valor: Int)
}

@external(erlang, "tiempo_ffi", "ahora_ns")
fn ahora_ns() -> Int

@external(erlang, "tiempo_ffi", "programadores")
fn programadores() -> Int

fn sumar(i: Int, n: Int, s: Int) -> Int {
  case i > n {
    True -> s
    False -> sumar(i + 1, n, s + { i * i } % 1_000_003)
  }
}

fn suma(n: Int) -> String {
  let t0 = ahora_ns()
  let s = sumar(1, n, 0)
  "{\"lang\":\"gleam\",\"n\":" <> int.to_string(n) <> ",\"suma\":" <> int.to_string(s) <> ",\"ns\":"
  <> int.to_string(ahora_ns() - t0) <> "}"
}

fn llenar(i: Int, n: Int, m: dict.Dict(String, Item)) -> dict.Dict(String, Item) {
  case i >= n {
    True -> m
    False -> {
      let it = Item(i, "item-" <> int.to_string(i), { i * 7919 } % 1000)
      llenar(i + 1, n, dict.insert(m, it.nombre, it))
    }
  }
}

fn trabajo(n: Int) -> String {
  let t0 = ahora_ns()
  let lista = llenar(0, n, dict.new()) |> dict.values
  let total = list.fold(lista, 0, fn(acc, it) { acc + it.valor })
  let orden =
    list.sort(lista, fn(a, b) {
      case int.compare(a.valor, b.valor) {
        order.Eq -> int.compare(a.id, b.id)
        otro -> otro
      }
    })
  let primeros =
    orden
    |> list.take(50)
    |> list.map(fn(it) {
      "{\"id\":" <> int.to_string(it.id) <> ",\"nombre\":\"" <> it.nombre <> "\",\"valor\":"
      <> int.to_string(it.valor) <> "}"
    })
    |> string.join(",")
  "{\"lang\":\"gleam\",\"n\":" <> int.to_string(n) <> ",\"total\":" <> int.to_string(total) <> ",\"primeros\":["
  <> primeros <> "],\"ns\":" <> int.to_string(ahora_ns() - t0) <> "}"
}

fn num(req: Request(a)) -> Int {
  request.get_query(req)
  |> result.unwrap([])
  |> list.key_find("n")
  |> result.try(int.parse)
  |> result.unwrap(0)
}

fn json(codigo: Int, cuerpo: String) -> Response(mist.ResponseData) {
  response.new(codigo)
  |> response.set_header("content-type", "application/json")
  |> response.set_body(mist.Bytes(bytes_tree.from_string(cuerpo)))
}

pub fn main() {
  let assert Ok(_) =
    fn(req: Request(mist.Connection)) -> Response(mist.ResponseData) {
      case request.path_segments(req) {
        ["suma"] -> json(200, suma(num(req)))
        ["trabajo"] -> json(200, trabajo(num(req)))
        _ -> json(404, "{}")
      }
    }
    |> mist.new
    |> mist.bind("0.0.0.0")
    |> mist.port(8080)
    |> mist.start
  io.println_error("INFO gleam programadores=" <> int.to_string(programadores()))
  process.sleep_forever()
}
