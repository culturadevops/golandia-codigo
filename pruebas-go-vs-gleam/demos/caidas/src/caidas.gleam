// Tres procesos trabajan; uno se cae a la mitad. ¿Qué les pasa a los otros dos?
import gleam/erlang/process
import gleam/int
import gleam/io

fn trabajar(nombre: String, falla: Bool) {
  process.sleep(50)
  case falla {
    True -> panic as "división entre cero en el pedido 42"
    False -> io.println(nombre <> ": terminé")
  }
}

pub fn main() {
  process.spawn_unlinked(fn() { trabajar("proceso 1", False) })
  process.spawn_unlinked(fn() { trabajar("proceso 2", True) })
  process.spawn_unlinked(fn() { trabajar("proceso 3", False) })
  process.sleep(300)
  io.println("main: sigo vivo, " <> int.to_string(3) <> " procesos lanzados")
}
