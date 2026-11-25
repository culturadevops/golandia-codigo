// Un paquete de npm muy usado: Express 5. Arranca, se pide a sí mismo una ruta y termina.
import express from "express";
import { get } from "node:http";
const app = express();
app.get("/hola", (_req, res) => { res.json({ ok: true, de: "express" }); });
const srv = app.listen(8081, () => {
  get("http://127.0.0.1:8081/hola", (r) => {
    let b = "";
    r.on("data", (c) => (b += c));
    r.on("end", () => { console.log("express:", b); srv.close(); });
  });
});
