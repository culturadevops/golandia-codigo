// Los estados de un pedido. Alguien agregó Devuelto y nadie tocó la función que los describe.
import gleam/int
import gleam/io

pub type Estado {
  Pendiente
  Pagado
  Enviado
  Devuelto
}

pub fn describir(e: Estado) -> String {
  case e {
    Pendiente -> "esperando el pago"
    Pagado -> "listo para enviar"
    Enviado -> "en camino"
  }
}

pub fn main() {
  io.println(describir(Enviado))
  io.println("2^64 + 1 = " <> int.to_string(18_446_744_073_709_551_615 + 2))
}
