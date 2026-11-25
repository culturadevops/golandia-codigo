// El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en TypeScript para Deno 2.9.8.
// El mismo contrato: /suma (CPU), /trabajo (objetos, mapa, orden y JSON a mano), /espera. El mismo código que
// recursos/pruebas-go-bun-node/servidor.js (vídeo 514), con tipos. Para «¿Go o TypeScript con Deno?» (2026-10-29).
import process from "node:process";

interface Item {
  readonly id: number;
  readonly nombre: string;
  readonly valor: number;
}

const LANG = "deno";
const ahora = (): bigint => process.hrtime.bigint();

function num(url: URL, k: string): number {
  const v = parseInt(url.searchParams.get(k) ?? "0", 10);
  return Number.isFinite(v) ? v : 0;
}

function suma(n: number): string {
  const t0 = ahora();
  let s = 0;
  for (let i = 1; i <= n; i++) s += (i * i) % 1000003;
  return `{"lang":"${LANG}","n":${n},"suma":${s},"ns":${ahora() - t0}}`;
}

function trabajo(n: number): string {
  const t0 = ahora();
  const m = new Map<string, Item>();
  for (let i = 0; i < n; i++) {
    const it: Item = { id: i, nombre: "item-" + i, valor: (i * 7919) % 1000 };
    m.set(it.nombre, it);
  }
  const lista: Item[] = [];
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
  b += `],"ns":${ahora() - t0}}`;
  return b;
}

function json(s: string): Response {
  return new Response(s, { headers: { "Content-Type": "application/json" } });
}

export async function atender(req: Request): Promise<Response> {
  const url = new URL(req.url);
  if (url.pathname === "/suma") return json(suma(num(url, "n")));
  if (url.pathname === "/trabajo") return json(trabajo(num(url, "n")));
  if (url.pathname === "/espera") {
    const ms = num(url, "ms");
    await new Promise((r) => setTimeout(r, ms));
    return json(`{"lang":"${LANG}","n":${ms},"suma":0,"ns":0}`);
  }
  return new Response("no", { status: 404 });
}
