// eval: código que llega como texto en tiempo de ejecución
const expr: string = process.argv[2] ?? "6 * 7";
console.log("eval:", eval(expr));
