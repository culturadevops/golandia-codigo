# SALIDA — Curso básico de Rust 2027

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Los errores de compilación son a propósito:
> cada lección enseña qué dice el compilador. «codigo de salida 101» = cargo no compiló o el programa entró en pánico.

- fecha: `2026-11-13 11:52:32Z`
- `rustc 1.99.1 (58cf8f9c3 2026-10-13)`
- `cargo 1.99.1 (5f94df478 2026-08-27)`
- `Linux 6.18.33.2-microsoft-standard-WSL2 x86_64`

## parte1/01_hola

`bash parte1/01_hola/demo.sh` · código de salida `0`

```
$ cargo new hola
    Creating binary (application) `hola` package
note: see more `Cargo.toml` keys and their definitions at https://doc.rust-lang.org/cargo/reference/manifest.html
$ find . -type f -not -path "./.git/*"
./.gitignore
./Cargo.toml
./src/main.rs
$ cat Cargo.toml
[package]
name = "hola"
version = "0.1.0"
edition = "2024"

[dependencies]
$ cat src/main.rs
fn main() {
    println!("Hello, world!");
}
$ cargo run
   Compiling hola v0.1.0 (/tmp/hola)
    Finished `dev` profile [unoptimized + debuginfo] target(s) in X.XXs
     Running `target/debug/hola`
Hello, world!
$ cargo build --release
   Compiling hola v0.1.0 (/tmp/hola)
    Finished `release` profile [optimized] target(s) in X.XXs
$ ls -l target/debug/hola target/release/hola
4545624 target/debug/hola
444760 target/release/hola
$ ./target/release/hola
Hello, world!
```

## parte1/02_inmutable

`bash parte1/02_inmutable/demo.sh` · código de salida `0`

```
--- a_error.rs
fn main() {
    let vidas = 3;
    println!("vidas: {vidas}");
    vidas = 2;
    println!("vidas: {vidas}");
}
---
$ cargo run 
error[E0384]: cannot assign twice to immutable variable `vidas`
 --> src/main.rs:4:5
  |
2 |     let vidas = 3;
  |         ----- first assignment to `vidas`
3 |     println!("vidas: {vidas}");
4 |     vidas = 2;
  |     ^^^^^^^^^ cannot assign twice to immutable variable
  |
help: consider making this binding mutable
  |
2 |     let mut vidas = 3;
  |         +++

For more information about this error, try `rustc --explain E0384`.
error: could not compile `vidas` (bin "vidas") due to 1 previous error
[codigo de salida: 101]
--- b_mut.rs
fn main() {
    let mut vidas = 3;
    println!("vidas: {vidas}");
    vidas = 2;
    println!("vidas: {vidas}");
}
---
$ cargo run 
vidas: 3
vidas: 2
[codigo de salida: 0]
--- c_const.rs
const VIDAS_MAXIMAS: u32 = 3;

fn main() {
    let vidas = VIDAS_MAXIMAS - 1;
    println!("te quedan {vidas} de {VIDAS_MAXIMAS}");
}
---
$ cargo run 
te quedan 2 de 3
[codigo de salida: 0]
```

## parte1/03_shadowing

`bash parte1/03_shadowing/demo.sh` · código de salida `0`

```
--- a_sombra.rs
fn main() {
    let entrada = "   ";
    let entrada = entrada.len();
    println!("espacios: {entrada}");
}
---
$ cargo run 
espacios: 3
[codigo de salida: 0]
--- b_error.rs
fn main() {
    let mut entrada = "   ";
    entrada = entrada.len();
    println!("espacios: {entrada}");
}
---
$ cargo run 
error[E0308]: mismatched types
 --> src/main.rs:3:15
  |
2 |     let mut entrada = "   ";
  |                       ----- expected due to this value
3 |     entrada = entrada.len();
  |               ^^^^^^^^^^^^^ expected `&str`, found `usize`

For more information about this error, try `rustc --explain E0308`.
error: could not compile `sombra` (bin "sombra") due to 1 previous error
[codigo de salida: 101]
--- c_bloque.rs
fn main() {
    let x = 5;
    {
        let x = x * 2;
        println!("dentro del bloque: {x}");
    }
    println!("fuera del bloque: {x}");
}
---
$ cargo run 
dentro del bloque: 10
fuera del bloque: 5
[codigo de salida: 0]
```

## parte1/04_tipos

`bash parte1/04_tipos/demo.sh` · código de salida `0`

```
--- a_error.rs
fn main() {
    let edad = "42".parse().unwrap();
    println!("edad: {edad}");
}
---
$ cargo run 
error[E0284]: type annotations needed
 --> src/main.rs:2:9
  |
2 |     let edad = "42".parse().unwrap();
  |         ^^^^        ----- type must be known at this point
  |
  = note: cannot satisfy `<_ as FromStr>::Err == _`
  = note: the type must also implement `FromStr`
note: required by a bound in `core::str::<impl str>::parse`
 --> /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/core/src/str/mod.rs:2782:4
help: consider giving `edad` an explicit type
  |
2 |     let edad: /* Type */ = "42".parse().unwrap();
  |             ++++++++++++

For more information about this error, try `rustc --explain E0284`.
error: could not compile `tipos` (bin "tipos") due to 1 previous error
[codigo de salida: 101]
--- b_tipos.rs
fn main() {
    let edad: u32 = "42".parse().unwrap();
    let temperatura: i32 = -7;
    let precio: f64 = 19.99;
    let activo: bool = true;
    let cangrejo: char = '🦀';
    println!("{edad} {temperatura} {precio} {activo} {cangrejo}");
    println!("u8:  {} a {}", u8::MIN, u8::MAX);
    println!("i32: {} a {}", i32::MIN, i32::MAX);
    println!("u64: {} a {}", u64::MIN, u64::MAX);
    println!("0.1 + 0.2 = {}", 0.1 + 0.2);
}
---
$ cargo run 
42 -7 19.99 true 🦀
u8:  0 a 255
i32: -2147483648 a 2147483647
u64: 0 a 18446744073709551615
0.1 + 0.2 = 0.30000000000000004
[codigo de salida: 0]
--- c_negativo.rs
fn main() {
    let edad: u32 = "-5".parse().unwrap();
    println!("edad: {edad}");
}
---
$ cargo run 

thread 'main' (2251) panicked at src/main.rs:2:34:
called `Result::unwrap()` on an `Err` value: ParseIntError { kind: InvalidDigit }
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
```

## parte1/05_overflow

`bash parte1/05_overflow/demo.sh` · código de salida `0`

```
--- a_nivel.rs
fn main() {
    let mut nivel: u8 = 250;
    for _ in 0..6 {
        nivel += 1;
    }
    println!("nivel = {nivel}");
}
---
$ cargo run 

thread 'main' (2490) panicked at src/main.rs:4:9:
attempt to add with overflow
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
$ cargo run --release 
nivel = 0
[codigo de salida: 0]
--- b_explicito.rs
fn main() {
    let nivel: u8 = 255;
    println!("wrapping_add:   {}", nivel.wrapping_add(1));
    println!("checked_add:    {:?}", nivel.checked_add(1));
    println!("saturating_add: {}", nivel.saturating_add(1));
    println!("checked_add(0): {:?}", nivel.checked_add(0));
}
---
$ cargo run 
wrapping_add:   0
checked_add:    None
saturating_add: 255
checked_add(0): Some(255)
[codigo de salida: 0]
$ tail -2 Cargo.toml
[profile.release]
overflow-checks = true
$ cargo run --release 

thread 'main' (2995) panicked at src/main.rs:4:9:
attempt to add with overflow
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
```

## parte1/06_funciones

`bash parte1/06_funciones/demo.sh` · código de salida `0`

```
--- a_error.rs
fn sumar(a: i32, b: i32) -> i32 {
    a + b;
}

fn main() {
    println!("{}", sumar(2, 3));
}
---
$ cargo run 
error[E0308]: mismatched types
 --> src/main.rs:1:29
  |
1 | fn sumar(a: i32, b: i32) -> i32 {
  |    -----                    ^^^ expected `i32`, found `()`
  |    |
  |    implicitly returns `()` as its body has no tail or `return` expression
2 |     a + b;
  |          - help: remove this semicolon to return this value

For more information about this error, try `rustc --explain E0308`.
error: could not compile `funciones` (bin "funciones") due to 1 previous error
[codigo de salida: 101]
--- b_ok.rs
fn sumar(a: i32, b: i32) -> i32 {
    a + b
}

fn saludar(nombre: &str) {
    println!("hola, {nombre}");
}

fn main() {
    saludar("Golandia");
    let total = sumar(2, 3);
    let doble = {
        let t = total * 2;
        t + 0
    };
    println!("total {total}, doble {doble}");
}
---
$ cargo run 
hola, Golandia
total 5, doble 10
[codigo de salida: 0]
```

## parte1/07_arrays

`bash parte1/07_arrays/demo.sh` · código de salida `0`

```
--- a_tuplas.rs
fn main() {
    let jugador: (&str, u32, f64) = ("Ana", 3, 1.75);
    let (nombre, vidas, altura) = jugador;
    println!("{nombre} tiene {vidas} vidas y mide {altura}");
    println!("campo 1 de la tupla: {}", jugador.1);

    let puntos = [10, 20, 30, 40, 50];
    let ceros = [0; 4];
    println!("primero {}, largo {}, ceros {:?}", puntos[0], puntos.len(), ceros);
}
---
$ cargo run 
Ana tiene 3 vidas y mide 1.75
campo 1 de la tupla: 3
primero 10, largo 5, ceros [0, 0, 0, 0]
[codigo de salida: 0]
--- b_fuera.rs
fn main() {
    let puntos = [10, 20, 30, 40, 50];
    let i: usize = std::env::args().nth(1).unwrap().parse().unwrap();
    println!("pido la posición {i}");
    println!("valor: {}", puntos[i]);
    println!("fin del programa");
}
---
$ cargo run 2
pido la posición 2
valor: 30
fin del programa
[codigo de salida: 0]
$ cargo run 7
pido la posición 7

thread 'main' (3960) panicked at src/main.rs:5:27:
index out of bounds: the len is 5 but the index is 7
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
--- c_fuera_fijo.rs
fn main() {
    let puntos = [10, 20, 30, 40, 50];
    println!("valor: {}", puntos[7]);
}
---
$ cargo run 
error: this operation will panic at runtime
 --> src/main.rs:3:27
  |
3 |     println!("valor: {}", puntos[7]);
  |                           ^^^^^^^^^ index out of bounds: the length is 5 but the index is 7
  |
  = note: `#[deny(unconditional_panic)]` on by default

error: could not compile `arrays` (bin "arrays") due to 1 previous error
[codigo de salida: 101]
```

## parte2/01_if

`bash parte2/01_if/demo.sh` · código de salida `0`

```
--- a_if.rs
fn main() {
    let temperatura = 31;
    let estado = if temperatura > 30 { "calor" } else { "normal" };
    println!("hoy: {estado}");

    let nota = 72;
    if nota >= 90 {
        println!("excelente");
    } else if nota >= 70 {
        println!("aprobado");
    } else {
        println!("a repasar");
    }
}
---
$ cargo run 
hoy: calor
aprobado
[codigo de salida: 0]
--- b_no_bool.rs
fn main() {
    let mensajes = 3;
    if mensajes {
        println!("tienes mensajes");
    }
}
---
$ cargo run 
error[E0308]: mismatched types
 --> src/main.rs:3:8
  |
3 |     if mensajes {
  |        ^^^^^^^^ expected `bool`, found integer

For more information about this error, try `rustc --explain E0308`.
error: could not compile `si` (bin "si") due to 1 previous error
[codigo de salida: 101]
--- c_tipos.rs
fn main() {
    let temperatura = 31;
    let estado = if temperatura > 30 { "calor" } else { 0 };
    println!("hoy: {estado}");
}
---
$ cargo run 
error[E0308]: `if` and `else` have incompatible types
 --> src/main.rs:3:57
  |
3 |     let estado = if temperatura > 30 { "calor" } else { 0 };
  |                                        -------          ^ expected `&str`, found integer
  |                                        |
  |                                        expected because of this

For more information about this error, try `rustc --explain E0308`.
error: could not compile `si` (bin "si") due to 1 previous error
[codigo de salida: 101]
```

## parte2/02_bucles

`bash parte2/02_bucles/demo.sh` · código de salida `0`

```
--- a_bucles.rs
fn main() {
    // loop devuelve un valor con break
    let mut intentos = 0;
    let clave = loop {
        intentos += 1;
        if intentos == 3 {
            break intentos * 100;
        }
    };
    println!("clave {clave} tras {intentos} intentos");

    // while
    let mut cuenta = 3;
    while cuenta > 0 {
        println!("{cuenta}...");
        cuenta -= 1;
    }
    println!("¡despegue!");

    // for sobre rangos
    for i in 1..4 {
        print!("{i} ");
    }
    println!("<- 1..4");
    for i in 1..=4 {
        print!("{i} ");
    }
    println!("<- 1..=4");
    for i in (1..=3).rev() {
        print!("{i} ");
    }
    println!("<- rev");

    // for sobre un array
    let frutas = ["mango", "guayaba", "parchita"];
    for fruta in frutas {
        println!("fruta: {fruta}");
    }
}
---
$ cargo run 
clave 300 tras 3 intentos
3...
2...
1...
¡despegue!
1 2 3 <- 1..4
1 2 3 4 <- 1..=4
3 2 1 <- rev
fruta: mango
fruta: guayaba
fruta: parchita
[codigo de salida: 0]
--- b_etiquetas.rs
fn main() {
    'filas: for fila in 1..=3 {
        for col in 1..=3 {
            if fila * col == 4 {
                println!("encontré el 4 en ({fila}, {col})");
                break 'filas;
            }
        }
    }
    println!("fin");
}
---
$ cargo run 
encontré el 4 en (2, 2)
fin
[codigo de salida: 0]
```

## parte2/03_mover

`bash parte2/03_mover/demo.sh` · código de salida `0`

```
--- a_error.rs
fn main() {
    let s1 = String::from("hola");
    let s2 = s1;
    println!("{s1}, mundo");
    println!("{s2}");
}
---
$ cargo run 
error[E0382]: borrow of moved value: `s1`
 --> src/main.rs:4:16
  |
2 |     let s1 = String::from("hola");
  |         -- move occurs because `s1` has type `String`, which does not implement the `Copy` trait
3 |     let s2 = s1;
  |              -- value moved here
4 |     println!("{s1}, mundo");
  |                ^^ value borrowed here after move
  |
help: consider cloning the value if the performance cost is acceptable
  |
3 |     let s2 = s1.clone();
  |                ++++++++

For more information about this error, try `rustc --explain E0382`.
error: could not compile `mover` (bin "mover") due to 1 previous error
[codigo de salida: 101]
--- b_clone.rs
fn main() {
    let s1 = String::from("hola");
    let s2 = s1.clone();
    println!("s1 = {s1}, s2 = {s2}");
    println!("s1 en {:p}", s1.as_ptr());
    println!("s2 en {:p}", s2.as_ptr());
}
---
$ cargo run 
s1 = hola, s2 = hola
s1 en 0x61dcaa71dd60
s2 en 0x61dcaa71dd80
[codigo de salida: 0]
--- c_tamano.rs
fn main() {
    let s = String::from("hola");
    println!("String en la pila: {} bytes", std::mem::size_of::<String>());
    println!("texto en el montón: {} bytes, capacidad {}", s.len(), s.capacity());
    let literal: &str = "hola";
    println!("un &str ocupa {} bytes", std::mem::size_of::<&str>());
    println!("{literal}");
}
---
$ cargo run 
String en la pila: 24 bytes
texto en el montón: 4 bytes, capacidad 4
un &str ocupa 16 bytes
hola
[codigo de salida: 0]
```

## parte2/04_copiar

`bash parte2/04_copiar/demo.sh` · código de salida `0`

```
--- a_copy.rs
fn main() {
    let x = 5;
    let y = x;
    println!("x = {x}, y = {y}");

    let punto = (3, 4);
    let otro = punto;
    println!("{punto:?} {otro:?}");
}
---
$ cargo run 
x = 5, y = 5
(3, 4) (3, 4)
[codigo de salida: 0]
--- b_tupla_string.rs
fn main() {
    let jugador = (String::from("Ana"), 3);
    let copia = jugador;
    println!("{:?}", jugador);
    println!("{:?}", copia);
}
---
$ cargo run 
error[E0382]: borrow of moved value: `jugador`
 --> src/main.rs:4:22
  |
2 |     let jugador = (String::from("Ana"), 3);
  |         ------- move occurs because `jugador` has type `(String, i32)`, which does not implement the `Copy` trait
3 |     let copia = jugador;
  |                 ------- value moved here
4 |     println!("{:?}", jugador);
  |                      ^^^^^^^ value borrowed here after move
  |
help: consider cloning the value if the performance cost is acceptable
  |
3 |     let copia = jugador.clone();
  |                        ++++++++

For more information about this error, try `rustc --explain E0382`.
error: could not compile `copiar` (bin "copiar") due to 1 previous error
[codigo de salida: 101]
```

## parte2/05_funciones

`bash parte2/05_funciones/demo.sh` · código de salida `0`

```
--- a_error.rs
fn imprimir(texto: String) {
    println!("recibí: {texto}");
}

fn main() {
    let saludo = String::from("hola");
    imprimir(saludo);
    imprimir(saludo);
}
---
$ cargo run 
error[E0382]: use of moved value: `saludo`
 --> src/main.rs:8:14
  |
6 |     let saludo = String::from("hola");
  |         ------ move occurs because `saludo` has type `String`, which does not implement the `Copy` trait
7 |     imprimir(saludo);
  |              ------ value moved here
8 |     imprimir(saludo);
  |              ^^^^^^ value used here after move
  |
note: consider changing this parameter type in function `imprimir` to borrow instead if owning the value isn't necessary
 --> src/main.rs:1:20
  |
1 | fn imprimir(texto: String) {
  |    --------        ^^^^^^ this parameter takes ownership of the value
  |    |
  |    in this function
help: consider cloning the value if the performance cost is acceptable
  |
7 |     imprimir(saludo.clone());
  |                    ++++++++

For more information about this error, try `rustc --explain E0382`.
error: could not compile `funciones` (bin "funciones") due to 1 previous error
[codigo de salida: 101]
--- b_devolver.rs
fn agregar_signo(mut texto: String) -> String {
    texto.push('!');
    texto
}

fn main() {
    let saludo = String::from("hola");
    let saludo = agregar_signo(saludo);
    let saludo = agregar_signo(saludo);
    println!("{saludo}");
}
---
$ cargo run 
hola!!
[codigo de salida: 0]
```

## parte2/06_drop

`bash parte2/06_drop/demo.sh` · código de salida `0`

```
--- a_drop.rs
struct Recurso(&'static str);

impl Drop for Recurso {
    fn drop(&mut self) {
        println!("  liberado: {}", self.0);
    }
}

fn usar(r: Recurso) {
    println!("  usar() recibe {}", r.0);
}

fn main() {
    println!("empieza main");
    let a = Recurso("a");
    {
        let _b = Recurso("b");
        println!("  fin del bloque interno");
    }
    usar(a);
    println!("después de usar(a)");
    let _c = Recurso("c");
    println!("termina main");
}
---
$ cargo run 
empieza main
  fin del bloque interno
  liberado: b
  usar() recibe a
  liberado: a
después de usar(a)
termina main
  liberado: c
[codigo de salida: 0]
```

## parte3/01_texto

`bash parte3/01_texto/demo.sh` · código de salida `0`

```
--- a_len.rs
fn main() {
    let ave = String::from("ñandú");
    println!("len():           {}", ave.len());
    println!("chars().count(): {}", ave.chars().count());
    println!("bytes: {:?}", ave.as_bytes());
}
---
$ cargo run 
len():           7
chars().count(): 5
bytes: [195, 177, 97, 110, 100, 195, 186]
[codigo de salida: 0]
--- b_indice.rs
fn main() {
    let ave = String::from("ñandú");
    let primera = ave[0];
    println!("{primera}");
}
---
$ cargo run 
error[E0277]: the type `str` cannot be indexed by `{integer}`
 --> src/main.rs:3:23
  |
3 |     let primera = ave[0];
  |                       ^ string indices are ranges of `usize`
  |
  = help: the trait `SliceIndex<str>` is not implemented for `{integer}`
  = note: you can use `.chars().nth()` or `.bytes().nth()`
          for more information, see chapter 8 in The Book: <https://doc.rust-lang.org/book/ch08-02-strings.html#indexing-into-strings>
help: `usize` implements trait `SliceIndex<T>`
 --> /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/core/src/slice/index.rs:179:0
  |
  = note: `SliceIndex<[T]>`
 --> /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/core/src/bstr/traits.rs:197:0
  |
  = note: `SliceIndex<ByteStr>`
  = note: required for `String` to implement `Index<{integer}>`

For more information about this error, try `rustc --explain E0277`.
error: could not compile `texto` (bin "texto") due to 1 previous error
[codigo de salida: 101]
--- c_corte.rs
fn main() {
    let ave = String::from("ñandú");
    println!("los 2 primeros bytes: {}", &ave[0..2]);
    println!("el primer byte: {}", &ave[0..1]);
}
---
$ cargo run 
los 2 primeros bytes: ñ

thread 'main' (7706) panicked at src/main.rs:4:40:
end byte index 1 is not a char boundary; it is inside 'ñ' (bytes 0..2 of string)
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
--- d_chars.rs
fn main() {
    let ave = String::from("ñandú");
    let primera = ave.chars().next().unwrap();
    let ultima = ave.chars().last().unwrap();
    println!("primera: {primera}, última: {ultima}");
    for (i, c) in ave.char_indices() {
        print!("{c}@{i} ");
    }
    println!();
}
---
$ cargo run 
primera: ñ, última: ú
ñ@0 a@2 n@3 d@4 ú@5 
[codigo de salida: 0]
```

## parte3/02_prestar

`bash parte3/02_prestar/demo.sh` · código de salida `0`

```
--- a_prestar.rs
fn contar_letras(texto: &String) -> usize {
    texto.chars().count()
}

fn main() {
    let nombre = String::from("Golandia");
    let n = contar_letras(&nombre);
    println!("{nombre} tiene {n} letras");
    let n2 = contar_letras(&nombre);
    println!("y sigue siendo mío: {nombre} ({n2})");
}
---
$ cargo run 
Golandia tiene 8 letras
y sigue siendo mío: Golandia (8)
[codigo de salida: 0]
```

## parte3/03_mut

`bash parte3/03_mut/demo.sh` · código de salida `0`

```
--- a_error.rs
fn agregar_signo(texto: &String) {
    texto.push('!');
}

fn main() {
    let saludo = String::from("hola");
    agregar_signo(&saludo);
    println!("{saludo}");
}
---
$ cargo run 
error[E0596]: cannot borrow `*texto` as mutable, as it is behind a `&` reference
 --> src/main.rs:2:5
  |
2 |     texto.push('!');
  |     ^^^^^ `texto` is a `&` reference, so it cannot be borrowed as mutable
  |
help: consider changing this to be a mutable reference
  |
1 | fn agregar_signo(texto: &mut String) {
  |                          +++

For more information about this error, try `rustc --explain E0596`.
error: could not compile `mutar` (bin "mutar") due to 1 previous error
[codigo de salida: 101]
--- b_mut.rs
fn agregar_signo(texto: &mut String) {
    texto.push('!');
}

fn main() {
    let mut saludo = String::from("hola");
    agregar_signo(&mut saludo);
    agregar_signo(&mut saludo);
    println!("{saludo}");
}
---
$ cargo run 
hola!!
[codigo de salida: 0]
```

## parte3/04_regla

`bash parte3/04_regla/demo.sh` · código de salida `0`

```
--- a_dos_mut.rs
fn main() {
    let mut s = String::from("hola");
    let r1 = &mut s;
    let r2 = &mut s;
    r1.push('!');
    r2.push('?');
    println!("{s}");
}
---
$ cargo run 
error[E0499]: cannot borrow `s` as mutable more than once at a time
 --> src/main.rs:4:14
  |
3 |     let r1 = &mut s;
  |              ------ first mutable borrow occurs here
4 |     let r2 = &mut s;
  |              ^^^^^^ second mutable borrow occurs here
5 |     r1.push('!');
  |     -- first borrow later used here

For more information about this error, try `rustc --explain E0499`.
error: could not compile `regla` (bin "regla") due to 1 previous error
[codigo de salida: 101]
--- b_vec.rs
fn main() {
    let mut cola = vec![10, 20, 30];
    let primero = &cola[0];
    cola.push(40);
    println!("el primero era {primero}");
}
---
$ cargo run 
error[E0502]: cannot borrow `cola` as mutable because it is also borrowed as immutable
 --> src/main.rs:4:5
  |
3 |     let primero = &cola[0];
  |                    ---- immutable borrow occurs here
4 |     cola.push(40);
  |     ^^^^^^^^^^^^^ mutable borrow occurs here
5 |     println!("el primero era {primero}");
  |                               ------- immutable borrow later used here

For more information about this error, try `rustc --explain E0502`.
error: could not compile `regla` (bin "regla") due to 1 previous error
[codigo de salida: 101]
--- c_vec_ok.rs
fn main() {
    let mut cola = vec![10, 20, 30];
    let _vecino = vec![0u8; 64];
    println!("capacidad antes: {}, datos en {:p}", cola.capacity(), cola.as_ptr());
    let primero = cola[0];
    cola.push(40);
    println!("capacidad después: {}, datos en {:p}", cola.capacity(), cola.as_ptr());
    println!("el primero era {primero}");
}
---
$ cargo run 
capacidad antes: 3, datos en 0x5c0c105f8d60
capacidad después: 6, datos en 0x5c0c105f8d60
el primero era 10
[codigo de salida: 0]
--- d_nll.rs
fn main() {
    let mut s = String::from("hola");
    let r1 = &s;
    let r2 = &s;
    println!("{r1} y {r2}");
    // r1 y r2 ya no se usan más abajo: el préstamo terminó aquí
    let r3 = &mut s;
    r3.push('!');
    println!("{r3}");
}
---
$ cargo run 
hola y hola
hola!
[codigo de salida: 0]
```

## parte3/05_colgante

`bash parte3/05_colgante/demo.sh` · código de salida `0`

```
--- a_error.rs
fn crear_saludo() -> &String {
    let s = String::from("hola");
    &s
}

fn main() {
    let saludo = crear_saludo();
    println!("{saludo}");
}
---
$ cargo run 
error[E0106]: missing lifetime specifier
 --> src/main.rs:1:22
  |
1 | fn crear_saludo() -> &String {
  |                      ^ expected named lifetime parameter
  |
  = help: this function's return type contains a borrowed value, but there is no value for it to be borrowed from
help: consider using the `'static` lifetime, but this is uncommon unless you're returning a borrowed value from a `const` or a `static`
  |
1 | fn crear_saludo() -> &'static String {
  |                       +++++++
help: instead, you are more likely to want to return an owned value
  |
1 - fn crear_saludo() -> &String {
1 + fn crear_saludo() -> String {
  |

For more information about this error, try `rustc --explain E0106`.
error: could not compile `colgante` (bin "colgante") due to 1 previous error
[codigo de salida: 101]
--- b_ok.rs
fn crear_saludo() -> String {
    let s = String::from("hola");
    s
}

fn main() {
    let saludo = crear_saludo();
    println!("{saludo}");
}
---
$ cargo run 
hola
[codigo de salida: 0]
```

## parte3/06_slices

`bash parte3/06_slices/demo.sh` · código de salida `0`

```
--- a_primera.rs
fn primera_palabra(texto: &str) -> &str {
    for (i, b) in texto.bytes().enumerate() {
        if b == b' ' {
            return &texto[..i];
        }
    }
    texto
}

fn main() {
    let frase = String::from("hola mundo");
    let p = primera_palabra(&frase);
    println!("primera palabra: {p}");
    println!("de un literal: {}", primera_palabra("rust es rápido"));

    let numeros = [1, 2, 3, 4, 5];
    let medio: &[i32] = &numeros[1..4];
    println!("medio: {medio:?}");
}
---
$ cargo run 
primera palabra: hola
de un literal: rust
medio: [2, 3, 4]
[codigo de salida: 0]
--- b_error.rs
fn primera_palabra(texto: &str) -> &str {
    texto.split(' ').next().unwrap()
}

fn main() {
    let mut frase = String::from("hola mundo");
    let p = primera_palabra(&frase);
    frase.clear();
    println!("primera palabra: {p}");
}
---
$ cargo run 
error[E0502]: cannot borrow `frase` as mutable because it is also borrowed as immutable
 --> src/main.rs:8:5
  |
7 |     let p = primera_palabra(&frase);
  |                             ------ immutable borrow occurs here
8 |     frase.clear();
  |     ^^^^^^^^^^^^^ mutable borrow occurs here
9 |     println!("primera palabra: {p}");
  |                                 - immutable borrow later used here

For more information about this error, try `rustc --explain E0502`.
error: could not compile `slices` (bin "slices") due to 1 previous error
[codigo de salida: 101]
```

## parte4/01_struct

`bash parte4/01_struct/demo.sh` · código de salida `0`

```
--- a_struct.rs
struct Usuario {
    nombre: String,
    correo: String,
    activo: bool,
    ingresos: u64,
}

fn crear(nombre: String, correo: String) -> Usuario {
    Usuario { nombre, correo, activo: true, ingresos: 1 }
}

fn main() {
    let mut ana = crear(String::from("Ana"), String::from("ana@ejemplo.com"));
    ana.ingresos += 1;
    println!("{} <{}> activo={} ingresos={}", ana.nombre, ana.correo, ana.activo, ana.ingresos);
}
---
$ cargo run 
Ana <ana@ejemplo.com> activo=true ingresos=2
[codigo de salida: 0]
--- b_actualizar.rs
struct Usuario {
    nombre: String,
    correo: String,
    activo: bool,
}

fn main() {
    let ana = Usuario {
        nombre: String::from("Ana"),
        correo: String::from("ana@ejemplo.com"),
        activo: true,
    };
    let beto = Usuario {
        correo: String::from("beto@ejemplo.com"),
        ..ana
    };
    println!("{} {}", beto.nombre, beto.activo);
    println!("{}", ana.nombre);
}
---
$ cargo run 
error[E0382]: borrow of moved value: `ana.nombre`
  --> src/main.rs:18:20
   |
13 |       let beto = Usuario {
   |  ________________-
14 | |         correo: String::from("beto@ejemplo.com"),
15 | |         ..ana
16 | |     };
   | |_____- value moved here
17 |       println!("{} {}", beto.nombre, beto.activo);
18 |       println!("{}", ana.nombre);
   |                      ^^^^^^^^^^ value borrowed here after move
   |
   = note: move occurs because `ana.nombre` has type `String`, which does not implement the `Copy` trait

For more information about this error, try `rustc --explain E0382`.
error: could not compile `usuarios` (bin "usuarios") due to 1 previous error
[codigo de salida: 101]
--- c_parcial.rs
struct Usuario {
    nombre: String,
    correo: String,
    activo: bool,
}

fn main() {
    let ana = Usuario {
        nombre: String::from("Ana"),
        correo: String::from("ana@ejemplo.com"),
        activo: true,
    };
    let beto = Usuario {
        correo: String::from("beto@ejemplo.com"),
        ..ana
    };
    println!("{} {} {}", beto.nombre, beto.correo, beto.activo);
    println!("de ana sigue vivo: {} {}", ana.correo, ana.activo);
}
---
$ cargo run 
Ana beto@ejemplo.com true
de ana sigue vivo: ana@ejemplo.com true
[codigo de salida: 0]
```

## parte4/02_debug

`bash parte4/02_debug/demo.sh` · código de salida `0`

```
--- a_error.rs
struct Punto {
    x: i32,
    y: i32,
}

fn main() {
    let p = Punto { x: 3, y: 4 };
    println!("{p}");
}
---
$ cargo run 
error[E0277]: `Punto` doesn't implement `std::fmt::Display`
 --> src/main.rs:8:15
  |
8 |     println!("{p}");
  |               ^^^ `Punto` cannot be formatted with the default formatter
  |
help: the trait `std::fmt::Display` is not implemented for `Punto`
 --> src/main.rs:1:1
  |
1 | struct Punto {
  | ^^^^^^^^^^^^
  = note: in format strings you may be able to use `{:?}` (or `{:#?}` for pretty-print) instead

For more information about this error, try `rustc --explain E0277`.
error: could not compile `puntos` (bin "puntos") due to 1 previous error
[codigo de salida: 101]
--- b_debug.rs
#[derive(Debug)]
struct Punto {
    x: i32,
    y: i32,
}

fn main() {
    let p = Punto { x: 3, y: 4 };
    println!("{p:?}");
    println!("{p:#?}");
    let area = dbg!(p.x * p.y);
    println!("area = {area}");
}
---
$ cargo run 
Punto { x: 3, y: 4 }
Punto {
    x: 3,
    y: 4,
}
[src/main.rs:11:16] p.x * p.y = 12
area = 12
[codigo de salida: 0]
```

## parte4/03_impl

`bash parte4/03_impl/demo.sh` · código de salida `0`

```
--- a_impl.rs
#[derive(Debug)]
struct Rectangulo {
    ancho: u32,
    alto: u32,
}

impl Rectangulo {
    fn new(ancho: u32, alto: u32) -> Self {
        Self { ancho, alto }
    }

    fn cuadrado(lado: u32) -> Self {
        Self { ancho: lado, alto: lado }
    }

    fn area(&self) -> u32 {
        self.ancho * self.alto
    }

    fn cabe(&self, otro: &Rectangulo) -> bool {
        self.ancho >= otro.ancho && self.alto >= otro.alto
    }

    fn escalar(&mut self, factor: u32) {
        self.ancho *= factor;
        self.alto *= factor;
    }
}

fn main() {
    let mut r = Rectangulo::new(30, 50);
    let c = Rectangulo::cuadrado(40);
    println!("área de r: {}", r.area());
    println!("¿c cabe en r? {}", r.cabe(&c));
    r.escalar(2);
    println!("r escalado: {r:?}, área {}", r.area());
    println!("¿c cabe ahora? {}", r.cabe(&c));
}
---
$ cargo run 
área de r: 1500
¿c cabe en r? false
r escalado: Rectangulo { ancho: 60, alto: 100 }, área 6000
¿c cabe ahora? true
[codigo de salida: 0]
--- b_error.rs
struct Contador {
    n: u32,
}

impl Contador {
    fn sumar(&mut self) {
        self.n += 1;
    }
}

fn main() {
    let c = Contador { n: 0 };
    c.sumar();
    println!("{}", c.n);
}
---
$ cargo run 
error[E0596]: cannot borrow `c` as mutable, as it is not declared as mutable
  --> src/main.rs:13:5
   |
13 |     c.sumar();
   |     ^ cannot borrow as mutable
   |
help: consider changing this to be mutable
   |
12 |     let mut c = Contador { n: 0 };
   |         +++

For more information about this error, try `rustc --explain E0596`.
error: could not compile `rect` (bin "rect") due to 1 previous error
[codigo de salida: 101]
```

## parte4/04_enum

`bash parte4/04_enum/demo.sh` · código de salida `0`

```
--- a_enum.rs
#[derive(Debug)]
enum Comando {
    Salir,
    Mover { x: i32, y: i32 },
    Escribir(String),
    Color(u8, u8, u8),
}

fn ejecutar(c: &Comando) {
    match c {
        Comando::Salir => println!("adiós"),
        Comando::Mover { x, y } => println!("me muevo a ({x}, {y})"),
        Comando::Escribir(texto) => println!("escribo: {texto}"),
        Comando::Color(r, g, b) => println!("color #{r:02x}{g:02x}{b:02x}"),
    }
}

fn main() {
    let comandos = [
        Comando::Mover { x: 3, y: -1 },
        Comando::Escribir(String::from("hola")),
        Comando::Color(222, 165, 132),
        Comando::Salir,
    ];
    for c in &comandos {
        ejecutar(c);
    }
    println!("{:?}", comandos[0]);
}
---
$ cargo run 
me muevo a (3, -1)
escribo: hola
color #dea584
adiós
Mover { x: 3, y: -1 }
[codigo de salida: 0]
```

## parte4/05_option

`bash parte4/05_option/demo.sh` · código de salida `0`

```
--- a_error.rs
fn main() {
    let precio: i32 = 100;
    let descuento: Option<i32> = Some(15);
    let total = precio - descuento;
    println!("{total}");
}
---
$ cargo run 
error[E0277]: cannot subtract `Option<i32>` from `i32`
 --> src/main.rs:4:24
  |
4 |     let total = precio - descuento;
  |                        ^ no implementation for `i32 - Option<i32>`
  |
  = help: the trait `Sub<Option<i32>>` is not implemented for `i32`
help: the following other types implement trait `Sub<Rhs>`
 --> /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/core/src/ops/arith.rs:210:8
  |
  = note: `i32` implements `Sub`
 ::: /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/core/src/ops/arith.rs:225:0
  |
  = note: in this macro invocation
 --> /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/core/src/internal_macros.rs:22:8
  |
  = note: `&i32` implements `Sub<i32>`
 ::: /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/core/src/internal_macros.rs:33:8
  |
  = note: `i32` implements `Sub<&i32>`
 ::: /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/core/src/internal_macros.rs:44:8
  |
  = note: `&i32` implements `Sub`
  = note: this error originates in the macro `sub_impl` (in Nightly builds, run with -Z macro-backtrace for more info)

For more information about this error, try `rustc --explain E0277`.
error: could not compile `menu` (bin "menu") due to 1 previous error
[codigo de salida: 101]
--- b_option.rs
fn buscar_precio(producto: &str) -> Option<u32> {
    match producto {
        "café" => Some(250),
        "arepa" => Some(400),
        _ => None,
    }
}

fn main() {
    for p in ["café", "arepa", "sushi"] {
        match buscar_precio(p) {
            Some(precio) => println!("{p}: {precio}"),
            None => println!("{p}: no está en el menú"),
        }
    }
    if let Some(precio) = buscar_precio("café") {
        println!("con if let: {precio}");
    }
    println!("unwrap_or: {}", buscar_precio("sushi").unwrap_or(0));
    let doble = buscar_precio("arepa").map(|p| p * 2);
    println!("map: {doble:?}");
    println!("unwrap sobre None:");
    let precio = buscar_precio("sushi").unwrap();
    println!("{precio}");
}
---
$ cargo run 
café: 250
arepa: 400
sushi: no está en el menú
con if let: 250
unwrap_or: 0
map: Some(800)
unwrap sobre None:

thread 'main' (12681) panicked at src/main.rs:23:41:
called `Option::unwrap()` on a `None` value
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
```

## parte4/06_match

`bash parte4/06_match/demo.sh` · código de salida `0`

```
--- a_error.rs
enum Estado {
    Pendiente,
    Enviado,
    Entregado,
    Devuelto,
}

fn describir(e: &Estado) -> &'static str {
    match e {
        Estado::Pendiente => "esperando pago",
        Estado::Enviado => "en camino",
        Estado::Entregado => "listo",
    }
}

fn main() {
    println!("{}", describir(&Estado::Devuelto));
}
---
$ cargo run 
error[E0004]: non-exhaustive patterns: `&Estado::Devuelto` not covered
  --> src/main.rs:9:11
   |
 9 |     match e {
   |           ^ pattern `&Estado::Devuelto` not covered
   |
note: `Estado` defined here
  --> src/main.rs:1:6
   |
 1 | enum Estado {
   |      ^^^^^^
...
 5 |     Devuelto,
   |     -------- not covered
   = note: the matched value is of type `&Estado`
help: ensure that all possible cases are being handled by adding a match arm with a wildcard pattern or an explicit pattern as shown
   |
12 ~         Estado::Entregado => "listo",
13 ~         &Estado::Devuelto => todo!(),
   |

For more information about this error, try `rustc --explain E0004`.
error: could not compile `pedidos` (bin "pedidos") due to 1 previous error
[codigo de salida: 101]
--- b_numeros.rs
fn clasificar(n: i32) -> &'static str {
    match n {
        i32::MIN..=-1 => "negativo",
        0 => "cero",
        1 | 2 | 3 => "pequeño",
        x if x % 2 == 0 => "par grande",
        _ => "impar grande",
    }
}

fn main() {
    for n in [-4, 0, 2, 10, 15] {
        println!("{n}: {}", clasificar(n));
    }
}
---
$ cargo run 
-4: negativo
0: cero
2: pequeño
10: par grande
15: impar grande
[codigo de salida: 0]
--- c_sin_comodin.rs
fn main() {
    let n: u8 = 7;
    let texto = match n {
        0 => "cero",
        1..=9 => "un dígito",
    };
    println!("{texto}");
}
---
$ cargo run 
error[E0004]: non-exhaustive patterns: `10_u8..=u8::MAX` not covered
 --> src/main.rs:3:23
  |
3 |     let texto = match n {
  |                       ^ pattern `10_u8..=u8::MAX` not covered
  |
  = note: the matched value is of type `u8`
help: ensure that all possible cases are being handled by adding a match arm with a wildcard pattern or an explicit pattern as shown
  |
5 ~         1..=9 => "un dígito",
6 ~         10_u8..=u8::MAX => todo!(),
  |

For more information about this error, try `rustc --explain E0004`.
error: could not compile `pedidos` (bin "pedidos") due to 1 previous error
[codigo de salida: 101]
```

## parte5/01_vec

`bash parte5/01_vec/demo.sh` · código de salida `0`

```
--- a_vec.rs
fn main() {
    let mut notas: Vec<u32> = Vec::new();
    notas.push(70);
    notas.push(95);
    notas.push(88);
    let mas = vec![60, 100];
    notas.extend(mas);
    println!("notas: {notas:?}, hay {}", notas.len());

    println!("notas[1] = {}", notas[1]);
    println!("get(1)  = {:?}", notas.get(1));
    println!("get(10) = {:?}", notas.get(10));

    for n in &mut notas {
        *n += 5;
    }
    println!("con 5 puntos extra: {notas:?}");

    let suma: u32 = notas.iter().sum();
    let max = notas.iter().max().unwrap();
    println!("suma {suma}, máxima {max}, promedio {}", suma as f64 / notas.len() as f64);

    let aprobadas: Vec<u32> = notas.iter().filter(|&&n| n >= 80).copied().collect();
    println!("aprobadas: {aprobadas:?}");
    println!("pop: {:?}, quedan {:?}", notas.pop(), notas);
}
---
$ cargo run 
notas: [70, 95, 88, 60, 100], hay 5
notas[1] = 95
get(1)  = Some(95)
get(10) = None
con 5 puntos extra: [75, 100, 93, 65, 105]
suma 438, máxima 105, promedio 87.6
aprobadas: [100, 93, 105]
pop: Some(105), quedan [75, 100, 93, 65]
[codigo de salida: 0]
--- b_indice.rs
fn main() {
    let notas = vec![70, 95, 88];
    let i: usize = std::env::args().nth(1).unwrap().parse().unwrap();
    match notas.get(i) {
        Some(n) => println!("con get: {n}"),
        None => println!("con get: no hay nota {i}"),
    }
    println!("con []: {}", notas[i]);
}
---
$ cargo run 1
con get: 95
con []: 95
[codigo de salida: 0]
$ cargo run 5
con get: no hay nota 5

thread 'main' (13895) panicked at src/main.rs:8:33:
index out of bounds: the len is 3 but the index is 5
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
```

## parte5/02_hashmap

`bash parte5/02_hashmap/demo.sh` · código de salida `0`

```
--- a_mapa.rs
use std::collections::HashMap;

fn main() {
    let mut stock: HashMap<String, u32> = HashMap::new();
    stock.insert(String::from("café"), 10);
    stock.insert(String::from("arepa"), 4);
    stock.insert(String::from("café"), 12);
    println!("café: {:?}", stock.get("café"));
    println!("té:   {:?}", stock.get("té"));

    *stock.entry(String::from("arepa")).or_insert(0) -= 1;
    *stock.entry(String::from("té")).or_insert(0) += 7;

    let mut claves: Vec<_> = stock.iter().collect();
    claves.sort();
    println!("{claves:?}");

    let texto = "el gato y el perro y el loro";
    let mut cuenta: HashMap<&str, u32> = HashMap::new();
    for palabra in texto.split_whitespace() {
        *cuenta.entry(palabra).or_insert(0) += 1;
    }
    let mut lista: Vec<_> = cuenta.into_iter().collect();
    lista.sort_by(|a, b| b.1.cmp(&a.1).then(a.0.cmp(b.0)));
    println!("{lista:?}");
}
---
$ cargo run 
café: Some(12)
té:   None
[("arepa", 3), ("café", 12), ("té", 7)]
[("el", 3), ("y", 2), ("gato", 1), ("loro", 1), ("perro", 1)]
[codigo de salida: 0]
--- b_orden.rs
use std::collections::HashMap;

fn main() {
    let mut m = HashMap::new();
    for (i, k) in ["uno", "dos", "tres", "cuatro", "cinco"].iter().enumerate() {
        m.insert(*k, i);
    }
    let claves: Vec<_> = m.keys().collect();
    println!("{claves:?}");
}
---
$ cargo run 
["uno", "cinco", "cuatro", "tres", "dos"]
[codigo de salida: 0]
$ cargo run 
["uno", "dos", "cuatro", "cinco", "tres"]
[codigo de salida: 0]
$ cargo run 
["cinco", "tres", "dos", "uno", "cuatro"]
[codigo de salida: 0]
```

## parte5/03_panic

`bash parte5/03_panic/demo.sh` · código de salida `0`

```
--- a_unwrap.rs
use std::fs;

fn main() {
    let texto = fs::read_to_string("config.txt").unwrap();
    println!("{texto}");
}
---
$ cargo run 

thread 'main' (14696) panicked at src/main.rs:4:50:
called `Result::unwrap()` on an `Err` value: Os { code: 2, kind: NotFound, message: "No such file or directory" }
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
--- b_expect.rs
use std::fs;

fn main() {
    let texto = fs::read_to_string("config.txt")
        .expect("no pude leer config.txt");
    println!("{texto}");
}
---
$ cargo run 

thread 'main' (14933) panicked at src/main.rs:5:10:
no pude leer config.txt: Os { code: 2, kind: NotFound, message: "No such file or directory" }
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
--- c_match.rs
use std::fs;
use std::process;

fn main() {
    let texto = match fs::read_to_string("config.txt") {
        Ok(t) => t,
        Err(e) => {
            eprintln!("error: config.txt: {e}");
            process::exit(1);
        }
    };
    println!("{texto}");
}
---
$ cargo run 
error: config.txt: No such file or directory (os error 2)
[codigo de salida: 1]
$ echo puerto=8080 > config.txt
$ cargo run 
puerto=8080

[codigo de salida: 0]
```

## parte5/04_result

`bash parte5/04_result/demo.sh` · código de salida `0`

```
--- a_result.rs
fn dividir(a: f64, b: f64) -> Result<f64, String> {
    if b == 0.0 {
        return Err(String::from("división entre cero"));
    }
    Ok(a / b)
}

fn main() {
    for (a, b) in [(10.0, 4.0), (1.0, 0.0)] {
        match dividir(a, b) {
            Ok(r) => println!("{a} / {b} = {r}"),
            Err(e) => println!("{a} / {b}: error, {e}"),
        }
    }
    println!("1.0 / 0.0 en f64 sin comprobar = {}", 1.0_f64 / 0.0);
}
---
$ cargo run 
10 / 4 = 2.5
1 / 0: error, división entre cero
1.0 / 0.0 en f64 sin comprobar = inf
[codigo de salida: 0]
--- b_ignorar.rs
fn guardar(dato: &str) -> Result<(), String> {
    if dato.is_empty() {
        return Err(String::from("dato vacío"));
    }
    println!("guardado: {dato}");
    Ok(())
}

fn main() {
    guardar("");
    println!("todo bien (¿seguro?)");
}
---
$ cargo run 
warning: unused `Result` that must be used
  --> src/main.rs:10:5
   |
10 |     guardar("");
   |     ^^^^^^^^^^^
   |
   = note: this `Result` may be an `Err` variant, which should be handled
   = note: `#[warn(unused_must_use)]` (part of `#[warn(unused)]`) on by default
help: use `let _ = ...` to ignore the resulting value
   |
10 |     let _ = guardar("");
   |     +++++++

todo bien (¿seguro?)
[codigo de salida: 0]
```

## parte5/05_interrogacion

`bash parte5/05_interrogacion/demo.sh` · código de salida `0`

```
--- a_q.rs
use std::error::Error;
use std::fs;

fn leer_puerto(ruta: &str) -> Result<u16, Box<dyn Error>> {
    let texto = fs::read_to_string(ruta)?;
    let valor = texto.trim().trim_start_matches("puerto=");
    let puerto: u16 = valor.parse()?;
    Ok(puerto)
}

fn main() -> Result<(), Box<dyn Error>> {
    let ruta = std::env::args().nth(1).unwrap_or(String::from("config.txt"));
    let puerto = leer_puerto(&ruta)?;
    println!("escucharé en el puerto {puerto}");
    Ok(())
}
---
$ cat bueno.txt grande.txt letras.txt
puerto=8080
puerto=80800
puerto=ochenta
$ cargo run bueno.txt
escucharé en el puerto 8080
[codigo de salida: 0]
$ cargo run grande.txt
Error: ParseIntError { kind: PosOverflow }
[codigo de salida: 1]
$ cargo run letras.txt
Error: ParseIntError { kind: InvalidDigit }
[codigo de salida: 1]
$ cargo run no-existe.txt
Error: Os { code: 2, kind: NotFound, message: "No such file or directory" }
[codigo de salida: 1]
```

## parte5/06_proyecto

`bash parte5/06_proyecto/demo.sh` · código de salida `0`

```
--- main.rs
use std::collections::HashMap;
use std::error::Error;
use std::fs;
use std::process;

struct Config {
    ruta: String,
    top: usize,
}

fn leer_args(args: &[String]) -> Result<Config, String> {
    let ruta = match args.get(1) {
        Some(r) => r.clone(),
        None => return Err(String::from("uso: palabras <archivo> [top]")),
    };
    let top = match args.get(2) {
        Some(t) => t.parse().map_err(|_| format!("top inválido: {t}"))?,
        None => 5,
    };
    Ok(Config { ruta, top })
}

fn contar(texto: &str) -> HashMap<String, usize> {
    let mut cuenta = HashMap::new();
    for palabra in texto.split_whitespace() {
        let limpia: String = palabra
            .chars()
            .filter(|c| c.is_alphabetic())
            .flat_map(|c| c.to_lowercase())
            .collect();
        if !limpia.is_empty() {
            *cuenta.entry(limpia).or_insert(0) += 1;
        }
    }
    cuenta
}

fn ejecutar(cfg: &Config) -> Result<(), Box<dyn Error>> {
    let texto = fs::read_to_string(&cfg.ruta)?;
    let cuenta = contar(&texto);
    let mut lista: Vec<(&String, &usize)> = cuenta.iter().collect();
    lista.sort_by(|a, b| b.1.cmp(a.1).then(a.0.cmp(b.0)));
    let total: usize = cuenta.values().sum();
    println!("{} palabras, {} distintas", total, cuenta.len());
    for (palabra, n) in lista.iter().take(cfg.top) {
        println!("{n:>8}  {palabra}");
    }
    Ok(())
}

fn main() {
    let args: Vec<String> = std::env::args().collect();
    let cfg = leer_args(&args).unwrap_or_else(|e| {
        eprintln!("{e}");
        process::exit(2);
    });
    if let Err(e) = ejecutar(&cfg) {
        eprintln!("error: {}: {e}", cfg.ruta);
        process::exit(1);
    }
}
---
$ cat corto.txt
El gato y el perro.
El LORO, el gato... ¡y la Ñandú!
$ cargo run corto.txt 3
12 palabras, 7 distintas
       4  el
       2  gato
       2  y
[codigo de salida: 0]
$ cargo run 
uso: palabras <archivo> [top]
[codigo de salida: 2]
$ cargo run corto.txt tres
top inválido: tres
[codigo de salida: 2]
$ cargo run no-existe.txt
error: no-existe.txt: No such file or directory (os error 2)
[codigo de salida: 1]
$ ls -l grande.txt
22722995 grande.txt
$ time ./target/debug/palabras grande.txt
5000000 palabras, 98 distintas
  505270  el
  208959  la
  159902  de
  135522  que
  118854  y
[tardó 2.123 s]
$ time ./target/release/palabras grande.txt
5000000 palabras, 98 distintas
  505270  el
  208959  la
  159902  de
  135522  que
  118854  y
[tardó 0.247 s]
$ ls -l target/debug/palabras target/release/palabras
5424080 target/debug/palabras
503304 target/release/palabras
```
