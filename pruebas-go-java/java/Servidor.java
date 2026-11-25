// El servidor de la comparativa, en Java 21: SOLO el JDK (com.sun.net.httpserver).
// El mismo contrato que el de Go: /suma hace CPU de verdad, /espera simula E/S.
// Executor de hilos virtuales (Java 21): es el Java de HOY, no el de 2015.
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.Executors;

public class Servidor {
    static final long ARRANQUE = System.nanoTime();

    static int parametro(HttpExchange ex, String clave) {
        String q = ex.getRequestURI().getQuery();
        if (q == null) return 0;
        for (String par : q.split("&")) {
            String[] kv = par.split("=", 2);
            if (kv.length == 2 && kv[0].equals(clave)) return Integer.parseInt(kv[1]);
        }
        return 0;
    }

    static void responder(HttpExchange ex, String json) throws java.io.IOException {
        byte[] cuerpo = json.getBytes(StandardCharsets.UTF_8);
        ex.getResponseHeaders().set("Content-Type", "application/json");
        ex.sendResponseHeaders(200, cuerpo.length);
        try (OutputStream os = ex.getResponseBody()) { os.write(cuerpo); }
    }

    public static void main(String[] args) throws Exception {
        HttpServer server = HttpServer.create(new InetSocketAddress(8080), 0);

        server.createContext("/suma", ex -> {
            int n = parametro(ex, "n");
            long t0 = System.nanoTime();
            long s = 0;
            for (int i = 1; i <= n; i++) {
                s += (long) i * i % 1000003;
            }
            long ns = System.nanoTime() - t0;
            responder(ex, "{\"lang\":\"java\",\"n\":" + n + ",\"suma\":" + s + ",\"ns\":" + ns + "}");
        });

        server.createContext("/espera", ex -> {
            int ms = parametro(ex, "ms");
            try { Thread.sleep(ms); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            responder(ex, "{\"lang\":\"java\",\"n\":" + ms + ",\"suma\":0,\"ns\":0}");
        });

        server.setExecutor(Executors.newVirtualThreadPerTaskExecutor());
        server.start();
        // La marca de «listo» del propio proceso: main → escuchando, sin el contenedor por medio.
        System.err.printf("LISTO_EN_MS %.3f%n", (System.nanoTime() - ARRANQUE) / 1_000_000.0);
    }
}
