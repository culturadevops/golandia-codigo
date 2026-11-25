# SALIDA-demos — ¿Go o TypeScript compilado a nativo? (golandia 880)

> Corrida del 2026-11-18 09:41. Sin cronómetro. Cada programa se compila y se ejecuta en un contenedor nuevo; la salida es la literal (recortada a las líneas que importan con `head`).

## 1 · Cobertura: once programas, tres compiladores

### scriptc build + ejecutar (scriptc 0.4.14)

```
=== 01_eval.ts
/w/bin
[build: 0]
Uncaught EvalError: Code generation from strings disallowed for this context
[run: 1]
=== 02_new_function.ts
/w/02_new_function.ts:2:11 - error SC1090: 'number' values where '(number, number) => number' is expected is not supported yet

  1 | // new Function: igual, pero como función
  2 | const f = new Function("a", "b", "return a * b") as (a: number, b: number) => number;
    |           ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
  3 | console.log("new Function:", f(6, 7));
[build: 1]
=== 03_json_any.ts
/w/bin
[build: 0]
any: go 42
[run: 0]
=== 04_proxy.ts
/w/04_proxy.ts:2:51 - error SC1090: Proxy handlers whose traps have typed signatures are not supported yet

  1 | // Proxy: interceptar el acceso a propiedades
  2 | const p = new Proxy({} as Record<string, number>, { get: (_t, k) => String(k).length });
    |                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
  3 | console.log("proxy:", p["hola"], p["golandia"]);
[build: 1]
=== 05_prototipo.ts
/w/05_prototipo.ts:3:1 - error SC1092: replacing the built-in 'Array.prototype.push' here is not supported: the program calls it at 05_prototipo.ts:5:1, and compiled code does not read a built-in member from its object, so the call would not see the change

  2 | const original = Array.prototype.push;
  3 | Array.prototype.push = function (this: unknown[], ...xs: unknown[]): number { return original.apply(this, xs.map((x) => x)); };
    | ^~~~~~~~~~~~~~~~~~~~
  4 | const a: number[] = [];
[build: 1]
=== 06_generico_interno.ts
/w/06_generico_interno.ts:3:3 - error SC1090: generic nested functions (only top-level generic function declarations are supported) are not supported yet

  2 | function exterior(): string {
  3 |   function par<T>(a: T, b: T): T[] { return [a, b]; }
    |   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
  4 |   return par<number>(1, 2).join(",") + " " + par<string>("a", "b").join(",");
[build: 1]
=== 07_import_dinamico.ts
/w/07_import_dinamico.ts:3:17 - error SC1090: dynamic import() with a specifier that cannot be bounded at build time (at /w/07_import_dinamico.ts:2: a value read from `process.argv`) is not supported yet

  2 | const nombre: string = process.argv[2] ?? "node:os";
  3 | const m = await import(nombre);
    |                 ^~~~~~~~~~~~~~
  4 | console.log("import dinámico:", typeof m.cpus);
[build: 1]
=== 08_worker.ts
/w/bin
[build: 0]
worker: 42
[run: 0]
=== 09_igualdad_suelta.ts
/w/09_igualdad_suelta.ts:4:27 - error SC1040: loose equality (== and !=) is not supported yet

  3 | // @ts-ignore
  4 | console.log("== objeto:", precio == 42);
    |                           ^~~~~~~~~~~~
  5 | 
[build: 1]
=== 10_express.ts
/w/10_express.ts:1:1 - error SC8005: this program calls into Node.js (@scriptc/node), but no libnode pack is installed: install @scriptc/libnode-linux-x64-gnu, set SCRIPTC_LIBNODE or pass --libnode=<pack directory>

  1 | // Un paquete de npm muy usado: Express 5. Arranca, se pide a sí mismo una ruta y termina.
    | ^
  2 | import express from "express";

[build: 1]
=== 11_hono.ts
/w/bin
[build: 0]
scriptc: this program reached code that scriptc could not compile; each such operation throws where it runs and is reported here once ('scriptc coverage' lists them, SCRIPTC_DEFERRED_REPORT=0 silences these lines)
scriptc: not compiled: 'RegExpMatchArray | null.slice' is part of the standard library types but has no scriptc lowering yet [SC2020 at /w/node_modules/hono/dist/router/reg-exp-router/router.js:78]
hono: 
[run: 0]
```

### bun build --compile --aot + ejecutar (Bun 1.4.5-canary.1)

```
=== 01_eval.ts
[build: 0]
eval: 42
[run: 0]
=== 02_new_function.ts
[build: 0]
new Function: 42
[run: 0]
=== 03_json_any.ts
[build: 0]
any: go 42
[run: 0]
=== 04_proxy.ts
[build: 0]
proxy: 4 8
[run: 0]
=== 05_prototipo.ts
3 | Array.prototype.push = function (this: unknown[], ...xs: unknown[]): number { return original.apply(this, xs.map((x) => x)); };
          ^
error: TS100014: this assignment always throws: built-in objects are read-only in a compiled executable
    at /w/05_prototipo.ts:3:7
[build: 1]
=== 06_generico_interno.ts
[build: 0]
genérico interno: 1,2 a,b
[run: 0]
=== 07_import_dinamico.ts
3 | const m = await import(nombre);
                    ^
warn: TS100064: this import path isn't a constant, so it throws at run time if the file isn't in the executable. Add the file as an entry point, or build with --aot-allow-code-from-disk
   at /w/07_import_dinamico.ts:3:17
[build: 0]
import dinámico: function
[run: 0]
=== 08_worker.ts
[build: 0]
worker: 42
[run: 0]
=== 09_igualdad_suelta.ts
[build: 0]
== objeto: true
[run: 0]
=== 10_express.ts
32 | 	const supportsColor = require('supports-color');
                                    ^
warn: TS100151: Could not resolve: "supports-color". Maybe you need to "bun install"?
   at /w/node_modules/debug/src/node.js:32:32

81 |     var fn = require(mod).__express
[build: 0]
express: {"ok":true,"de":"express"}
[run: 0]
=== 11_hono.ts
             note  3 regular expressions in hono are compiled at run time (--explain lists them)
[build: 0]
hono: {"ok":true,"de":"hono"}
[run: 0]
```

### bun build --compile + ejecutar (Bun 1.4.4)

```
=== 01_eval.ts
[build: 0]
eval: 42
[run: 0]
=== 02_new_function.ts
[build: 0]
new Function: 42
[run: 0]
=== 03_json_any.ts
[build: 0]
any: go 42
[run: 0]
=== 04_proxy.ts
[build: 0]
proxy: 4 8
[run: 0]
=== 05_prototipo.ts
[build: 0]
prototipo: 2
[run: 0]
=== 06_generico_interno.ts
[build: 0]
genérico interno: 1,2 a,b
[run: 0]
=== 07_import_dinamico.ts
[build: 0]
import dinámico: function
[run: 0]
=== 08_worker.ts
[build: 0]
worker: 42
[run: 0]
=== 09_igualdad_suelta.ts
[build: 0]
== objeto: true
[run: 0]
=== 10_express.ts
[build: 0]
express: {"ok":true,"de":"express"}
[run: 0]
=== 11_hono.ts
[build: 0]
hono: {"ok":true,"de":"hono"}
[run: 0]
```

## 1b · Express con scriptc, instalando lo que pide el error (el paquete libnode de la misma versión)

### npm install @scriptc/libnode-linux-x64-gnu@0.4.14 + scriptc build 10_express.ts

```
126M
/w/10_express.ts:2:1 - note SC8001: package 'express@5.3.0' runs in a Node island: runtime: R1.2 dynamic module loading: require() with a computed specifier
/w/node_modules/express/lib/express.js:1:1 - note SC8001: package 'body-parser@2.4.0' runs in a Node island: gap: SC1010: the 'qs' module is not supported yet
/w/node_modules/express/lib/utils.js:1:1 - note SC8001: package 'qs@6.17.1' runs in a Node island: gap: SC1010: the 'side-channel' module is not supported yet
/w/node_modules/get-intrinsic/index.js:1:1 - note SC8001: package 'math-intrinsics@1.1.0' runs in a Node island: gap: SC1013: Node cannot load this module because '/w/node_modules/math-intrinsics/package.json' is an invalid package configuration (ERR_INVALID_PACKAGE_CONFIG)
[build: 0]
bin 3794880
Uncaught TypeError: expected Server at $, got Object from a Node island (an island object stays an island value: declare the island's type or 'unknown' for it, or copy it with structuredClone)
[run: 1]
```

## 2 · scriptc coverage del servidor de la serie

### scriptc coverage servidor.ts

```
scriptc coverage /src/servidor.ts

  statements analyzed   50
  compile statically    50  (100%)

  fully static — all analyzed statements compile to native code.
```

### scriptc coverage 08_worker.ts

```
scriptc coverage /cob/08_worker.ts

  statements analyzed   6
  compile statically    6  (100%)
    in dynamic tier     3  (operate on values typed only at runtime; --explain-dynamic lists them)

  fully static — all analyzed statements compile to native code.
```

## 3 · Bun 1.4.4 (estable) con --aot

### bun build --compile --aot en 1.4.4

```
1.4.4+38afee6c8
0
   [3ms]  bundle  1 modules
 [109ms] compile  a
[codigo de salida: 0]
a 109123760
b 109123760
```

### bun build --compile --aot en 1.4.5-canary.1

```
1.4.5-canary.1+4ce14ba8b
   [56ms]  bundle  1 module
  [326ms]   check  81 lines
  [592ms] compile  a  72.4 MB
         build id  b25f6baf69bdb3edee934f534339e32e
a 72404520
b 156769000
```

## 4 · ¿De qué depende cada binario? (ldd)

### ldd

```
$ ldd scriptc
	linux-vdso.so.1 (0x0000794e8a8f1000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x0000794e8a527000)
	/lib64/ld-linux-x86-64.so.2 (0x0000794e8a8f3000)
```

### ldd bun

```
$ ldd bun-aot
	linux-vdso.so.1 (0x00007453d5fbb000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007453d5dbe000)
	/lib64/ld-linux-x86-64.so.2 (0x00007453d5fbd000)
	libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007453d5db9000)
	libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007453d5db4000)
	libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6 (0x00007453d5cc4000)
```

