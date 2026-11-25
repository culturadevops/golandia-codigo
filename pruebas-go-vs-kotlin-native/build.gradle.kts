// El mismo src/commonMain/kotlin/Servidor.kt para la JVM (jar con todo dentro) y para Kotlin/Native (linuxX64).
plugins {
    kotlin("multiplatform") version "2.4.21"
}

repositories { mavenCentral() }

kotlin {
    jvm()
    linuxX64 {
        binaries { executable { entryPoint = "main" } }
    }
    sourceSets {
        commonMain.dependencies {
            implementation("io.ktor:ktor-server-core:3.6.1")
            implementation("io.ktor:ktor-server-cio:3.6.1")
        }
    }
}

// El jar para la JVM, con Ktor y la librería de Kotlin dentro (java -jar servidor.jar).
tasks.register<Jar>("jarCompleto") {
    archiveFileName.set("servidor.jar")
    manifest { attributes["Main-Class"] = "MainKt" }
    duplicatesStrategy = DuplicatesStrategy.EXCLUDE
    val jvmJar = tasks.named<Jar>("jvmJar")
    dependsOn(jvmJar)
    from(jvmJar.map { zipTree(it.archiveFile) })
    from(configurations.named("jvmRuntimeClasspath").map { c -> c.map { if (it.isDirectory) it else zipTree(it) } })
    exclude("META-INF/*.SF", "META-INF/*.DSA", "META-INF/*.RSA")
}
