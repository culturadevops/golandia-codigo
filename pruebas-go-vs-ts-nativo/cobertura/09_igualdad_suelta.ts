// Comparar un objeto con == contra un número (conversión a primitivo)
const precio = { valueOf(): number { return 42; } };
// @ts-ignore
console.log("== objeto:", precio == 42);
