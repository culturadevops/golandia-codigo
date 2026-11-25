// El servidor de «¿Go o Kotlin Native?» (golandia 879). EL MISMO archivo se compila a la JVM (y de ahí a GraalVM
// native-image) y a Kotlin/Native (linuxX64). Mismo contrato que toda la serie: /suma (CPU) y /trabajo (20.000 objetos,
// mapa, orden y JSON armado a mano), con el mismo resultado que el servidor de Go.
//
// Por qué Ktor y no com.sun.net.httpserver como en el ¿Go o Kotlin? (486): Kotlin/Native no tiene el JDK, así que el
// servidor web de Java no existe ahí. Ktor con el motor CIO es Kotlin puro y compila a los dos lados: así el código es
// el mismo, byte a byte, y lo único que cambia es cómo se compila.
//
// El cálculo va en Dispatchers.Default (lo normal en Kotlin para trabajo de CPU). Con la variable HILOS se limita a
// ese número de hilos a la vez (limitedParallelism), para comparar con lo que el runtime cree que tiene.
import io.ktor.http.ContentType
import io.ktor.server.application.ServerReady
import io.ktor.server.cio.CIO
import io.ktor.server.engine.embeddedServer
import io.ktor.server.response.respondText
import io.ktor.server.routing.get
import io.ktor.server.routing.routing
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlin.time.TimeSource

data class Item(val id: Int, val nombre: String, val valor: Int)

fun suma(n: Int): String {
    val t0 = TimeSource.Monotonic.markNow()
    var s = 0L
    for (i in 1..n) {
        s += i.toLong() * i % 1000003
    }
    val ns = t0.elapsedNow().inWholeNanoseconds
    return """{"lang":"kotlin","n":$n,"suma":$s,"ns":$ns}"""
}

fun trabajo(n: Int): String {
    val t0 = TimeSource.Monotonic.markNow()
    val m = HashMap<String, Item>(n * 4 / 3 + 1)
    for (i in 0 until n) {
        val it = Item(i, "item-$i", ((i.toLong() * 7919) % 1000).toInt())
        m[it.nombre] = it
    }
    val lista = ArrayList<Item>(m.size)
    var total = 0
    for (it in m.values) {
        lista.add(it)
        total += it.valor
    }
    lista.sortWith(compareBy<Item>({ it.valor }, { it.id }))
    val b = StringBuilder()
    b.append("""{"lang":"kotlin","n":""").append(n).append(""","total":""").append(total).append(""","primeros":[""")
    for (i in 0 until minOf(50, lista.size)) {
        if (i > 0) b.append(',')
        val it = lista[i]
        b.append("""{"id":""").append(it.id).append(""","nombre":"""").append(it.nombre)
            .append("""","valor":""").append(it.valor).append('}')
    }
    b.append("""],"ns":""").append(t0.elapsedNow().inWholeNanoseconds).append('}')
    return b.toString()
}

fun servidor(plataforma: String, nucleos: Int, hilos: Int?, arranque: TimeSource.Monotonic.ValueTimeMark) {
    val calc: CoroutineDispatcher = if (hilos != null) Dispatchers.Default.limitedParallelism(hilos) else Dispatchers.Default
    val srv = embeddedServer(CIO, port = 8080) {
        monitor.subscribe(ServerReady) {
            val ms = arranque.elapsedNow().inWholeMicroseconds / 1000.0
            println("LISTO_EN_MS $ms PLATAFORMA $plataforma NUCLEOS $nucleos HILOS ${hilos ?: "Default"}")
        }
        routing {
            get("/suma") {
                val n = call.request.queryParameters["n"]?.toIntOrNull() ?: 0
                call.respondText(withContext(calc) { suma(n) }, ContentType.Application.Json)
            }
            get("/trabajo") {
                val n = call.request.queryParameters["n"]?.toIntOrNull() ?: 0
                call.respondText(withContext(calc) { trabajo(n) }, ContentType.Application.Json)
            }
        }
    }
    srv.start(wait = true)
}
