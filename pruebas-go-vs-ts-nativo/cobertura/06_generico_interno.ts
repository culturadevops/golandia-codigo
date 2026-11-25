// Genérico declarado dentro de otra función
function exterior(): string {
  function par<T>(a: T, b: T): T[] { return [a, b]; }
  return par<number>(1, 2).join(",") + " " + par<string>("a", "b").join(",");
}
console.log("genérico interno:", exterior());
