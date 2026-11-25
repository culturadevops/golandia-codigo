// 1 · La excepcion que el tipo no cuenta.
// cobrar() lanza. La firma de cobrar() NO lo dice, y quien la llama no se entera
// hasta que revienta en produccion. TypeScript compila esto sin una sola queja.

function cobrar(saldo: number, importe: number): number {
  if (importe > saldo) {
    throw new Error("saldo insuficiente");
  }
  return saldo - importe;
}

// El que llama ve `(saldo: number, importe: number) => number`.
// Nada en ese tipo dice "esto puede explotar".
const restante = cobrar(10, 50);
console.log("quedan:", restante);
