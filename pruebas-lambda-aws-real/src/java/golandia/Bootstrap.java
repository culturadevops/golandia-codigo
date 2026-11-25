// Runtime propio para la versión nativa (GraalVM native-image, runtime provided.al2023).
// El trabajo es el MISMO Manejador.java del runtime gestionado; aquí solo está el bucle del
// Runtime API de Lambda (next -> response), hecho a mano y sin librerías, y un lector mínimo
// del JSON de entrada (solo "n" e "io_ms").
package golandia;

import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class Bootstrap {
    public static void main(String[] args) throws Exception {
        System.setProperty("golandia.lang", "java-native");
        String api = "http://" + System.getenv("AWS_LAMBDA_RUNTIME_API") + "/2018-06-01/runtime/invocation/";
        Manejador m = new Manejador();
        while (true) {
            HttpURLConnection sig = (HttpURLConnection) URI.create(api + "next").toURL().openConnection();
            sig.setReadTimeout(0);
            String id = sig.getHeaderField("Lambda-Runtime-Aws-Request-Id");
            String cuerpo;
            try (InputStream in = sig.getInputStream()) {
                cuerpo = new String(in.readAllBytes(), StandardCharsets.UTF_8);
            }
            Map<String, Object> ev = new HashMap<>();
            ev.put("n", campo(cuerpo, "n"));
            ev.put("io_ms", campo(cuerpo, "io_ms"));
            Map<String, Object> r = m.manejar(ev);
            String json = "{\"lang\":\"" + r.get("lang") + "\",\"n\":" + r.get("n") + ",\"suma\":" + r.get("suma")
                    + ",\"io_ms\":" + r.get("io_ms") + "}";
            HttpURLConnection resp = (HttpURLConnection) URI.create(api + id + "/response").toURL().openConnection();
            resp.setRequestMethod("POST");
            resp.setDoOutput(true);
            try (OutputStream out = resp.getOutputStream()) {
                out.write(json.getBytes(StandardCharsets.UTF_8));
            }
            resp.getResponseCode();
            resp.disconnect();
        }
    }

    static Integer campo(String json, String nombre) {
        Matcher mt = Pattern.compile("\"" + nombre + "\"\\s*:\\s*(\\d+)").matcher(json);
        return mt.find() ? Integer.valueOf(mt.group(1)) : 0;
    }
}
