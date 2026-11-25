// La misma lógica para `deno serve --parallel`: Deno arranca un hilo de JavaScript por núcleo y reparte las conexiones.
// `deno serve` pide un export por defecto con `fetch`, no una llamada a Deno.serve.
import { atender } from "./logica.ts";

export default { fetch: atender } satisfies Deno.ServeDefaultExport;
