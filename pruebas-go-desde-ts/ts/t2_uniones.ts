// 2 · Uniones de literales: el compilador conoce los valores validos.
type Estado = "activo" | "baja" | "moroso";

function describir(e: Estado): string {
  return "el cliente esta " + e;
}

console.log(describir("activo"));
console.log(describir("activa"));   // <- una letra. El compilador la caza.
