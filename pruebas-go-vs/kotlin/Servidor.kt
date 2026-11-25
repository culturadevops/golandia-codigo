// El servidor de la comparativa, en Kotlin (JVM): solo el JDK (com.sun.net.httpserver) y la librería estándar de Kotlin.
// El mismo contrato que el de Go: /suma hace CPU de verdad, /espera simula E/S. Hilos virtuales de Java 21.
import com.sun.net.httpserver.HttpExchange
import com.sun.net.httpserver.HttpServer
import java.net.InetSocketAddress
import java.util.concurrent.Executors

val arranque = System.nanoTime()

fun parametro(ex: HttpExchange, clave: String): Int =
    ex.requestURI.query?.split("&")?.map { it.split("=", limit = 2) }
        ?.firstOrNull { it.size == 2 && it[0] == clave }?.get(1)?.toInt() ?: 0

fun responder(ex: HttpExchange, json: String) {
    val cuerpo = json.toByteArray()
    ex.responseHeaders.set("Content-Type", "application/json")
    ex.sendResponseHeaders(200, cuerpo.size.toLong())
    ex.responseBody.use { it.write(cuerpo) }
}

fun main() {
    val t = arranque
    val server = HttpServer.create(InetSocketAddress(8080), 0)
    server.createContext("/suma") { ex ->
        val n = parametro(ex, "n")
        val t0 = System.nanoTime()
        var s = 0L
        for (i in 1..n) {
            s += i.toLong() * i % 1000003
        }
        val ns = System.nanoTime() - t0
        responder(ex, """{"lang":"kotlin","n":$n,"suma":$s,"ns":$ns}""")
    }
    server.createContext("/espera") { ex ->
        val ms = parametro(ex, "ms")
        Thread.sleep(ms.toLong())
        responder(ex, """{"lang":"kotlin","n":$ms,"suma":0,"ns":0}""")
    }
    server.executor = Executors.newVirtualThreadPerTaskExecutor()
    server.start()
    // La marca de «listo» del propio proceso: main → escuchando, sin el contenedor por medio.
    System.err.println("LISTO_EN_MS %.3f".format(java.util.Locale.ROOT, (System.nanoTime() - t) / 1_000_000.0))
}
