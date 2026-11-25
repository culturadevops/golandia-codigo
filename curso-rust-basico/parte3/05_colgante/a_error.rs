fn crear_saludo() -> &String {
    let s = String::from("hola");
    &s
}

fn main() {
    let saludo = crear_saludo();
    println!("{saludo}");
}
