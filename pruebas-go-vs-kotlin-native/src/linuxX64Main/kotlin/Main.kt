// Arranque en Kotlin/Native: núcleos según sysconf (lo que ve el proceso) y la variable HILOS.
@file:OptIn(kotlinx.cinterop.ExperimentalForeignApi::class)

import kotlinx.cinterop.toKString
import platform.posix._SC_NPROCESSORS_ONLN
import platform.posix.getenv
import platform.posix.sysconf
import kotlin.time.TimeSource

fun main() {
    val arranque = TimeSource.Monotonic.markNow()
    val hilos = getenv("HILOS")?.toKString()?.toIntOrNull()
    servidor("native", sysconf(_SC_NPROCESSORS_ONLN).toInt(), hilos, arranque)
}
