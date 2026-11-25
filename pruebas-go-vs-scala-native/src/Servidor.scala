// El servidor de «¿Go o Scala Native?» (golandia 865): EL MISMO código fuente se compila a la JVM y a Scala Native.
// Mismo contrato que la revancha y que el 806: /suma (CPU) y /trabajo (objetos, Map, orden y JSON armado a mano), con las mismas
// funciones suma() y trabajo() del 806, escritas como se escribe Scala (colecciones inmutables, case class, sortBy con tupla).
//
// Por qué no jdk.httpserver como en el 806: Scala Native no trae com.sun.net.httpserver (no es parte de la API de Java que
// reimplementa). Así que aquí el HTTP va a mano sobre java.net.ServerSocket, que sí existe en los dos lados: HTTP/1.1 con
// keep-alive, un hilo por conexión. El mismo servidor, byte a byte, en la JVM y en nativo: lo único que cambia es el backend.
//
// Como el ejecutor de hilos de plataforma del 806 (un hilo por núcleo), el CÁLCULO de cada petición pasa por un semáforo con
// tantos permisos como núcleos dice el runtime (Runtime.availableProcessors), o los que diga la variable HILOS. Sin ese límite,
// las 16 conexiones de hey calculan a la vez, con 16 montones de objetos vivos, y el recolector se ahoga.
import java.io.{BufferedInputStream, OutputStream}
import java.net.{ServerSocket, Socket}
import java.nio.charset.StandardCharsets.UTF_8

val arranque = System.nanoTime()
val permisos = sys.env.get("HILOS").flatMap(_.toIntOption).getOrElse(Runtime.getRuntime.availableProcessors)
val turno = java.util.concurrent.Semaphore(permisos, true)

def enTurno(f: => String): String =
  turno.acquire()
  try f
  finally turno.release()

case class Item(id: Int, nombre: String, valor: Int)

def num(q: String): Int =
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

/** Lee una línea terminada en \r\n (o \n). null si la conexión se cerró antes de empezar. */
def linea(in: BufferedInputStream): String =
  val sb = new java.lang.StringBuilder
  var c = in.read()
  if c == -1 then return null
  while c != -1 && c != '\n' do
    if c != '\r' then sb.append(c.toChar)
    c = in.read()
  sb.toString

def responder(out: OutputStream, codigo: Int, cuerpo: String, cerrar: Boolean): Unit =
  val b = cuerpo.getBytes(UTF_8)
  val estado = if codigo == 200 then "200 OK" else "404 Not Found"
  val cab = s"HTTP/1.1 $estado\r\nContent-Type: application/json\r\nContent-Length: ${b.length}\r\n" +
    (if cerrar then "Connection: close\r\n" else "") + "\r\n"
  out.write(cab.getBytes(UTF_8))
  out.write(b)
  out.flush()

def atender(s: Socket): Unit =
  try
    s.setTcpNoDelay(true)
    val in = BufferedInputStream(s.getInputStream, 8192)
    val out = s.getOutputStream
    var seguir = true
    while seguir do
      val pedido = linea(in)
      if pedido == null || pedido.isEmpty then seguir = false
      else
        var cerrar = false
        var h = linea(in)
        while h != null && h.nonEmpty do
          if h.toLowerCase.startsWith("connection:") && h.toLowerCase.contains("close") then cerrar = true
          h = linea(in)
        val objetivo = pedido.split(" ")(1)
        val (ruta, q) = objetivo.indexOf('?') match
          case -1 => (objetivo, null)
          case i  => (objetivo.substring(0, i), objetivo.substring(i + 1))
        val n = num(q)
        ruta match
          case "/suma"    => responder(out, 200, enTurno(suma(n)), cerrar)
          case "/trabajo" => responder(out, 200, enTurno(trabajo(n)), cerrar)
          case _          => responder(out, 404, "{}", cerrar)
        if cerrar then seguir = false
  catch case _: java.io.IOException => ()
  finally s.close()

@main def servidor(): Unit =
  val srv = ServerSocket(8080, 1024)
  System.err.println(f"LISTO_EN_MS ${(System.nanoTime() - arranque) / 1e6}%.3f NUCLEOS ${Runtime.getRuntime.availableProcessors} HILOS $permisos")
  while true do
    val s = srv.accept()
    val t = Thread(() => atender(s))
    t.setDaemon(true)
    t.start()
