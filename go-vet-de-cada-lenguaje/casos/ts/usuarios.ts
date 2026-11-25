interface Usuario {
  nombre: string;
  correo: string;
  roles: string[];
}

async function guardar(u: Usuario): Promise<void> {
  await new Promise((r) => setTimeout(r, 10));
  console.log("guardado", u.nombre);
}

const ana: Usuario = { nombre: "ana", correo: "ana@ejemplo.com", roles: ["admin"] };

if (ana.roles.contains("admin")) {
  console.log("escribo a", ana.email);
}

guardar(ana);
