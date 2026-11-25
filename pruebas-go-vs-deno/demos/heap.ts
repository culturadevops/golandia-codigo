// ¿Con qué límite de memoria arranca V8? Lo mismo con `deno run` y con el binario de `deno compile`.
import v8 from "node:v8";

const h = v8.getHeapStatistics();
const mib = (n: number) => (n / 1048576).toFixed(0) + " MiB";
console.log(`heap_size_limit=${mib(h.heap_size_limit)} total_heap_size=${mib(h.total_heap_size)} ` +
  `nucleos=${navigator.hardwareConcurrency}`);
