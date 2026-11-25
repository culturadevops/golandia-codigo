// Un proceso, un hilo para tu código: Deno.serve, la API nativa de Deno 2.
// Se ejecuta con `deno run --allow-net=0.0.0.0:8080 servidor.ts` o compilado con `deno compile`.
const ARRANQUE = performance.now();
import { atender } from "./logica.ts";

Deno.serve({
  port: 8080,
  hostname: "0.0.0.0",
  onListen() {
    console.error(
      `LISTO_EN_MS ${(performance.now() - ARRANQUE).toFixed(3)} deno ${Deno.version.deno} ` +
        `desde_inicio_proceso=${performance.now().toFixed(1)} nucleos=${navigator.hardwareConcurrency}`,
    );
  },
}, atender);
