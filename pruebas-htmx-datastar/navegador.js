// Prueba en un navegador de verdad (Chromium de Playwright): la misma acción en las dos páginas.
// Abre la página, espera 5 s quieta, escribe «ru» y espera 2 s. Cuenta las peticiones a la app y lee lo que quedó en pantalla.
//   node navegador.js [puerto]   (la app corre en Docker con -p <puerto>:8080; lo lanza ejecutar.sh)
const PW = ["<proyecto>/canales/cartasdeldiablo/recursos/node_modules/playwright"];
const { chromium } = require(PW[0]);
const PUERTO = process.argv[2] || "8090";

(async () => {
  const b = await chromium.launch();
  for (const lib of ["htmx", "datastar"]) {
    const p = await b.newPage();
    const pet = [];
    p.on("request", (r) => { if (!r.url().includes("/static/")) pet.push(new URL(r.url()).pathname); });
    let bytesLib = 0;
    p.on("response", async (r) => { if (r.url().includes("/static/")) bytesLib += (await r.body()).length; });
    await p.goto(`http://127.0.0.1:${PUERTO}/${lib}`);
    await p.waitForTimeout(5000);
    const relojAntes = pet.filter((x) => x.endsWith("/hora")).length;
    await p.locator("input").pressSequentially("ru", { delay: 80 });
    await p.waitForTimeout(2000);
    const resultados = await p.locator("#resultados li").allTextContents();
    const hora = (await p.locator("p span").textContent()).trim();
    const busquedas = pet.filter((x) => x.endsWith("/buscar")).length;
    console.log(`${lib}: librería ${bytesLib} bytes · peticiones del reloj en 5 s: ${relojAntes} · búsquedas al escribir «ru»: ${busquedas}` +
                ` · resultados: ${JSON.stringify(resultados)} · hora en pantalla: ${hora}`);
    await p.close();
  }
  await b.close();
})();
