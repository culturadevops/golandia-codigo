//> using scala 3.10.0
// Los estados de un pedido. Alguien agregó «Devuelto» y nadie tocó la función que los describe.
enum Estado:
  case Pendiente, Pagado, Enviado, Devuelto

def describir(e: Estado): String = e match
  case Estado.Pendiente => "esperando el pago"
  case Estado.Pagado    => "listo para enviar"
  case Estado.Enviado   => "en camino"

case class Usuario(nombre: String)
val usuarios = Map("ana" -> Usuario("Ana"))

@main def demo(): Unit =
  println(describir(Estado.Enviado))
  val u: Option[Usuario] = usuarios.get("luis")
  println(u.map(_.nombre).getOrElse("no existe"))
