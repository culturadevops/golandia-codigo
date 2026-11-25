// Arranque en la JVM (y en GraalVM native-image): cuántos núcleos dice el runtime y la variable HILOS.
import kotlin.time.TimeSource

fun main() {
    val arranque = TimeSource.Monotonic.markNow()
    servidor("jvm", Runtime.getRuntime().availableProcessors(), System.getenv("HILOS")?.toIntOrNull(), arranque)
}
