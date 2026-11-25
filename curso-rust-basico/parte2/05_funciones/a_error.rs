fn imprimir(texto: String) {
    println!("recibí: {texto}");
}

fn main() {
    let saludo = String::from("hola");
    imprimir(saludo);
    imprimir(saludo);
}
