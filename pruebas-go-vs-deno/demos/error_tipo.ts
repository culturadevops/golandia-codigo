// Un error de tipo a propósito: searchParams.get devuelve string | null, no number.
const q = new URLSearchParams("n=5");
const n: number = q.get("n");
console.log("n + 1 =", n + 1);
