export const handler = async (evento) => {
  const n = Number(evento.n || 0);
  const ioMs = Number(evento.io_ms || 0);
  let suma = 0n;
  for (let i = 1n; i <= BigInt(n); i++) {
    suma += (i * i) % 1000003n;
  }
  if (ioMs > 0) {
    await new Promise((r) => setTimeout(r, ioMs));
  }
  return { lang: "node", n, suma: Number(suma), io_ms: ioMs };
};
