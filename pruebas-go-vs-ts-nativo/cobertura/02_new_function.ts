// new Function: igual, pero como función
const f = new Function("a", "b", "return a * b") as (a: number, b: number) => number;
console.log("new Function:", f(6, 7));
