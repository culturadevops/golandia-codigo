//> using scala 3.10.0
//> using jvm 25
// El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en Scala 3, para «¿Go o Scala?» (golandia 806).
// El mismo contrato: /suma (CPU) y /trabajo (objetos, Map, orden y JSON armado a mano). Servidor: jdk.httpserver, el mismo
// que usó Java en la revancha (506). EJECUTOR=virtual (por defecto) o EJECUTOR=plataforma (un hilo por núcleo).
// /trabajo está escrito como se escribe Scala: colecciones inmutables, case class, sortBy con tupla.
import com.sun.net.httpserver.{HttpExchange, HttpServer}
import java.net.InetSocketAddress
import java.util.concurrent.Executors

val arranque = System.nanoTime()

case class Item(id: Int, nombre: String, valor: Int)

def num(q: String | Null): Int =
  Option(q).flatMap(_.split("&").collectFirst { case s"n=$v" => v.toIntOption }.flatten).getOrElse(0)

def suma(n: Int): String =
  val t0 = System.nanoTime()
  var s = 0L
  var i = 1L
  while i <= n do
    s += i * i % 1000003
    i += 1
  s"""{"lang":"scala","n":$n,"suma":$s,"ns":${System.nanoTime() - t0}}"""

def trabajo(n: Int): String =
  val t0 = System.nanoTime()
  val m: Map[String, Item] = (0 until n).iterator.map(i => Item(i, s"item-$i", (i * 7919) % 1000)).map(it => it.nombre -> it).toMap
  val lista = m.values.toVector
  val total = lista.map(_.valor).sum
  val orden = lista.sortBy(it => (it.valor, it.id))
  val primeros = orden.take(50).map(it => s"""{"id":${it.id},"nombre":"${it.nombre}","valor":${it.valor}}""").mkString(",")
  s"""{"lang":"scala","n":$n,"total":$total,"primeros":[$primeros],"ns":${System.nanoTime() - t0}}"""

def responder(ex: HttpExchange, codigo: Int, cuerpo: String): Unit =
  val b = cuerpo.getBytes("UTF-8")
  ex.getResponseHeaders.set("Content-Type", "application/json")
  ex.sendResponseHeaders(codigo, b.length)
  val os = ex.getResponseBody
  os.write(b)
  os.close()

@main def servidor(): Unit =
  val srv = HttpServer.create(InetSocketAddress(8080), 1024)
  srv.createContext("/", ex =>
    val n = num(ex.getRequestURI.getRawQuery)
    ex.getRequestURI.getPath match
      case "/suma"    => responder(ex, 200, suma(n))
      case "/trabajo" => responder(ex, 200, trabajo(n))
      case _          => responder(ex, 404, "{}")
  )
  val ejecutor = sys.env.getOrElse("EJECUTOR", "virtual")
  if ejecutor == "plataforma" then srv.setExecutor(Executors.newFixedThreadPool(Runtime.getRuntime.availableProcessors))
  else srv.setExecutor(Executors.newVirtualThreadPerTaskExecutor())
  srv.start()
  System.err.printf("LISTO_EN_MS %.3f EJECUTOR %s NUCLEOS %d%n", (System.nanoTime() - arranque) / 1e6, ejecutor,
    Runtime.getRuntime.availableProcessors)
