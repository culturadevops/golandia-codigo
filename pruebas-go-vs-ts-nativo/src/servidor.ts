// El servidor de «¿Go o TypeScript compilado a nativo?» (golandia 880). EL MISMO archivo para los tres compiladores:
// bun build --compile (el de siempre: Bun + tu código empaquetado), bun build --compile --aot (código de máquina, sin JIT) y
// scriptc build (Vercel Labs: TypeScript a binario nativo con los tipos). Mismo contrato que toda la serie: /suma (CPU) y
// /trabajo (20.000 objetos, mapa, orden y JSON armado a mano), con el mismo resultado que el servidor de Go.
// node:http y no Bun.serve: es la API de servidor que dicen soportar los dos compiladores nativos.
import { createServer } from "node:http";

interface Item {
  readonly id: number;
  readonly nombre: string;
  readonly valor: number;
}

const ARRANQUE: number = performance.now();
const LANG: string = "ts";

function ahoraNs(): number {
  return Math.round(performance.now() * 1e6);
}

function num(q: string, k: string): number {
  for (const par of q.split("&")) {
    const i = par.indexOf("=");
    if (i > 0 && par.substring(0, i) === k) {
      const v = parseInt(par.substring(i + 1), 10);
      return Number.isFinite(v) ? v : 0;
    }
  }
  return 0;
}

function suma(n: number): string {
  const t0 = ahoraNs();
  let s = 0;
  for (let i = 1; i <= n; i++) s += (i * i) % 1000003;
  return `{"lang":"${LANG}","n":${n},"suma":${s},"ns":${ahoraNs() - t0}}`;
}

function trabajo(n: number): string {
  const t0 = ahoraNs();
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
  lista.sort((a: Item, b: Item): number => (a.valor !== b.valor ? a.valor - b.valor : a.id - b.id));
  let b = `{"lang":"${LANG}","n":${n},"total":${total},"primeros":[`;
  for (let i = 0; i < 50 && i < lista.length; i++) {
    if (i > 0) b += ",";
    const it = lista[i];
    b += `{"id":${it.id},"nombre":"${it.nombre}","valor":${it.valor}}`;
  }
  b += `],"ns":${ahoraNs() - t0}}`;
  return b;
}

const srv = createServer((req, res) => {
  const url: string = req.url ?? "/";
  const i = url.indexOf("?");
  const ruta = i < 0 ? url : url.substring(0, i);
  const q = i < 0 ? "" : url.substring(i + 1);
  let cuerpo: string;
  if (ruta === "/suma") cuerpo = suma(num(q, "n"));
  else if (ruta === "/trabajo") cuerpo = trabajo(num(q, "n"));
  else {
    res.writeHead(404);
    res.end("no");
    return;
  }
  res.writeHead(200, { "Content-Type": "application/json" });
  res.end(cuerpo);
});
srv.listen(8080, () => {
  console.error(`LISTO_EN_MS ${(performance.now() - ARRANQUE).toFixed(3)}`);
});
