// El servidor de la revancha (recursos/pruebas-go-java-revancha) en JavaScript, para Bun y para Node, con la API nativa de cada uno:
// Bun.serve en Bun y node:http en Node. El mismo contrato: /suma (CPU), /trabajo (objetos, mapa, orden y JSON a mano), /espera.
// Pedido en un comentario de «Go vs Node vs Python 2026» (2026-08-26): «ahora que TypeScript se reescribió en Go y salió Bun 1.4…
// podrías hacer una comparativa de rendimiento».
const ARRANQUE = performance.now();
const LANG = typeof Bun !== "undefined" ? "bun" : "node";

function num(url, k) {
  const v = parseInt(url.searchParams.get(k) || "0", 10);
  return Number.isFinite(v) ? v : 0;
}

function suma(n) {
  const t0 = process.hrtime.bigint();
  let s = 0;
  for (let i = 1; i <= n; i++) s += (i * i) % 1000003;
  return `{"lang":"${LANG}","n":${n},"suma":${s},"ns":${process.hrtime.bigint() - t0}}`;
}

function trabajo(n) {
  const t0 = process.hrtime.bigint();
  const m = new Map();
  for (let i = 0; i < n; i++) {
    const it = { id: i, nombre: "item-" + i, valor: (i * 7919) % 1000 };
    m.set(it.nombre, it);
  }
  const lista = [];
  let total = 0;
  for (const it of m.values()) {
    lista.push(it);
    total += it.valor;
  }
  lista.sort((a, b) => (a.valor !== b.valor ? a.valor - b.valor : a.id - b.id));
  let b = `{"lang":"${LANG}","n":${n},"total":${total},"primeros":[`;
  for (let i = 0; i < 50 && i < lista.length; i++) {
    if (i > 0) b += ",";
    const it = lista[i];
    b += `{"id":${it.id},"nombre":"${it.nombre}","valor":${it.valor}}`;
  }
  b += `],"ns":${process.hrtime.bigint() - t0}}`;
  return b;
}

async function responder(url) {
  if (url.pathname === "/suma") return suma(num(url, "n"));
  if (url.pathname === "/trabajo") return trabajo(num(url, "n"));
  if (url.pathname === "/espera") {
    const ms = num(url, "ms");
    await new Promise((r) => setTimeout(r, ms));
    return `{"lang":"${LANG}","n":${ms},"suma":0,"ns":0}`;
  }
  return null;
}

// WORKERS=2: un proceso por núcleo con node:cluster (Bun también lo soporta). Sin esto, JavaScript usa UN hilo para tu código,
// aunque el contenedor tenga 2 CPU (2026-10-19: medido aparte, para que la comparación con Go sea justa).
const WORKERS = parseInt(process.env.WORKERS || "1", 10);

function iniciar() {
  if (LANG === "bun") {
    Bun.serve({
      port: 8080,
      reusePort: WORKERS > 1,
      async fetch(req) {
        const cuerpo = await responder(new URL(req.url));
        return cuerpo === null ? new Response("no", { status: 404 })
                               : new Response(cuerpo, { headers: { "Content-Type": "application/json" } });
      },
    });
  } else {
    const http = require("node:http");
    http.createServer(async (req, res) => {
      const cuerpo = await responder(new URL(req.url, "http://x"));
      if (cuerpo === null) { res.writeHead(404); res.end("no"); return; }
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(cuerpo);
    }).listen(8080);
  }
}

const cluster = WORKERS > 1 ? require("node:cluster") : null;
if (cluster && cluster.isPrimary) {
  for (let i = 0; i < WORKERS; i++) cluster.fork();
} else {
  iniciar();
}
if (!cluster || cluster.isPrimary) {
  console.error(`LISTO_EN_MS ${(performance.now() - ARRANQUE).toFixed(3)} ${LANG} ${LANG === "bun" ? Bun.version : process.version} procesos=${WORKERS}`);
}
