// La revancha de Go contra Java (2026-10-19): el mismo servidor del vídeo 148 más /trabajo, una carga que reserva
// memoria (objetos, mapa, orden y JSON armado a mano). SOLO el JDK (com.sun.net.httpserver).
// EJECUTOR=virtual (por defecto, hilos virtuales) o EJECUTOR=plataforma (un hilo de plataforma por núcleo):
// en los comentarios del 148 dijeron que los hilos virtuales no son para trabajo de CPU. Se miden los dos.
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.Executors;

public class Servidor {
    static final long ARRANQUE = System.nanoTime();

    record Item(int id, String nombre, int valor) {}

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
            responder(ex, "{\"lang\":\"java\",\"n\":" + n + ",\"suma\":" + s + ",\"ns\":" + (System.nanoTime() - t0) + "}");
        });

        server.createContext("/trabajo", ex -> {
            int n = parametro(ex, "n");
            long t0 = System.nanoTime();
            Map<String, Item> m = new HashMap<>(n * 4 / 3 + 1);
            for (int i = 0; i < n; i++) {
                Item it = new Item(i, "item-" + i, (int) ((long) i * 7919 % 1000));
                m.put(it.nombre(), it);
            }
            List<Item> lista = new ArrayList<>(m.size());
            int total = 0;
            for (Item it : m.values()) {
                lista.add(it);
                total += it.valor();
            }
            lista.sort((a, b) -> a.valor() != b.valor() ? Integer.compare(a.valor(), b.valor()) : Integer.compare(a.id(), b.id()));
            StringBuilder b = new StringBuilder();
            b.append("{\"lang\":\"java\",\"n\":").append(n).append(",\"total\":").append(total).append(",\"primeros\":[");
            for (int i = 0; i < 50 && i < lista.size(); i++) {
                if (i > 0) b.append(',');
                Item it = lista.get(i);
                b.append("{\"id\":").append(it.id()).append(",\"nombre\":\"").append(it.nombre())
                 .append("\",\"valor\":").append(it.valor()).append('}');
            }
            b.append("],\"ns\":").append(System.nanoTime() - t0).append('}');
            responder(ex, b.toString());
        });

        server.createContext("/espera", ex -> {
            int ms = parametro(ex, "ms");
            try { Thread.sleep(ms); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            responder(ex, "{\"lang\":\"java\",\"n\":" + ms + ",\"suma\":0,\"ns\":0}");
        });

        String ejecutor = System.getenv().getOrDefault("EJECUTOR", "virtual");
        if (ejecutor.equals("plataforma")) {
            server.setExecutor(Executors.newFixedThreadPool(Runtime.getRuntime().availableProcessors()));
        } else {
            server.setExecutor(Executors.newVirtualThreadPerTaskExecutor());
        }
        server.start();
        System.err.printf("LISTO_EN_MS %.3f EJECUTOR %s NUCLEOS %d%n", (System.nanoTime() - ARRANQUE) / 1_000_000.0, ejecutor,
                          Runtime.getRuntime().availableProcessors());
    }
}
