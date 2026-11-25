// 4 · strictNullChecks: null y undefined son parte del tipo, y hay que descartarlos.
interface Cliente { nombre: string; }

function buscar(id: number): Cliente | null {
  return id === 1 ? { nombre: "Ada" } : null;
}

const c = buscar(2);
console.log(c.nombre);          // <- puede ser null, y el compilador no deja
console.log(c?.nombre ?? "-");  // <- asi si: el operador existe en el lenguaje
