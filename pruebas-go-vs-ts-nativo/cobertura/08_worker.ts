// Hilos: worker_threads
import { Worker, isMainThread, parentPort } from "node:worker_threads";
if (isMainThread) {
  const w = new Worker(new URL(import.meta.url));
  w.on("message", (m: number) => { console.log("worker:", m); w.terminate(); });
} else {
  parentPort!.postMessage(6 * 7);
}
