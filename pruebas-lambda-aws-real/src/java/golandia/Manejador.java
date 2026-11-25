// Handler de prueba: EXACTAMENTE el mismo trabajo que en Go, Python y Node (vídeo 121).
//   entrada  {"n": 200000, "io_ms": 0}
//   salida   {"lang":"java","n":200000,"suma":...,"io_ms":0}
// Sin dependencias: el runtime gestionado de Java acepta un método cualquiera (Clase::metodo)
// que reciba la entrada; no hace falta implementar RequestHandler.
package golandia;

import java.util.LinkedHashMap;
import java.util.Map;

public class Manejador {
    static final String LANG = System.getProperty("golandia.lang", "java");

    public Map<String, Object> manejar(Map<String, Object> evento) throws InterruptedException {
        int n = numero(evento.get("n"));
        int ioMs = numero(evento.get("io_ms"));
        long suma = 0;
        for (long i = 1; i <= n; i++) {
            suma += i * i % 1000003;
        }
        if (ioMs > 0) {
            Thread.sleep(ioMs);
        }
        Map<String, Object> r = new LinkedHashMap<>();
        r.put("lang", LANG);
        r.put("n", n);
        r.put("suma", suma);
        r.put("io_ms", ioMs);
        return r;
    }

    static int numero(Object o) {
        return o == null ? 0 : ((Number) o).intValue();
    }
}
