require("./wasm_exec.js");
const fs = require("fs");

const go = new Go();
const bytes = fs.readFileSync("main.wasm");

WebAssembly.instantiate(bytes, go.importObject).then((result) => {
  go.run(result.instance);

  setTimeout(() => {
    console.log("sumarEnGo(15, 27) =", global.sumarEnGo(15, 27));
    console.log("esPrimoEnGo(17) =", global.esPrimoEnGo(17));
    console.log("esPrimoEnGo(18) =", global.esPrimoEnGo(18));
    process.exit(0);
  }, 100);
});
