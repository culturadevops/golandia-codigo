// Otro paquete de npm: Hono con su adaptador para Node. Arranca, se pide a sí mismo una ruta y termina.
import { Hono } from "hono";
import { serve } from "@hono/node-server";
import { get } from "node:http";
const app = new Hono();
app.get("/hola", (c) => c.json({ ok: true, de: "hono" }));
const srv = serve({ fetch: app.fetch, port: 8082 }, () => {
  get("http://127.0.0.1:8082/hola", (r) => {
    let b = "";
    r.on("data", (c) => (b += c));
    r.on("end", () => { console.log("hono:", b); srv.close(); });
  });
});
