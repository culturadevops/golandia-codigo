// Qué puede hacer un programa al que solo le diste permiso de red en un puerto (--allow-net=0.0.0.0:8080).
async function probar(que: string, f: () => Promise<unknown>): Promise<void> {
  try {
    await f();
    console.log(`${que}: PERMITIDO`);
  } catch (e) {
    const err = e as Error;
    console.log(`${que}: ${err.name}: ${err.message}`);
  }
}

await probar("leer /etc/passwd", () => Deno.readTextFile("/etc/passwd"));
await probar("leer la variable HOME", async () => Deno.env.get("HOME"));
await probar("conectar a evil.example:443", () => Deno.connect({ hostname: "evil.example", port: 443 }));
await probar("ejecutar sh", () => new Deno.Command("sh", { args: ["-c", "id"] }).output());
const l = Deno.listen({ hostname: "0.0.0.0", port: 8080 });
console.log("escuchar en 0.0.0.0:8080: PERMITIDO");
l.close();
