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
