// Parchear un prototipo de la librería estándar
const original = Array.prototype.push;
Array.prototype.push = function (this: unknown[], ...xs: unknown[]): number { return original.apply(this, xs.map((x) => x)); };
const a: number[] = [];
a.push(1, 2);
console.log("prototipo:", a.length);
