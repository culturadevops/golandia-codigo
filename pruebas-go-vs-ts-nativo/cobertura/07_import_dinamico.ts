// import() con un nombre que se arma en tiempo de ejecución
const nombre: string = process.argv[2] ?? "node:os";
const m = await import(nombre);
console.log("import dinámico:", typeof m.cpus);
