// Proxy: interceptar el acceso a propiedades
const p = new Proxy({} as Record<string, number>, { get: (_t, k) => String(k).length });
console.log("proxy:", p["hola"], p["golandia"]);
