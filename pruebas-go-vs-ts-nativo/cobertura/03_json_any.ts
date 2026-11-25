// JSON sin tipo: any y propiedades que no se conocen al compilar
const datos: any = JSON.parse(process.argv[2] ?? "{\"nombre\":\"go\",\"extra\":{\"v\":42}}");
const clave: string = "extra";
console.log("any:", datos.nombre, datos[clave].v);
