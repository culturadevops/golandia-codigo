fn agregar_signo(texto: &String) {
    texto.push('!');
}

fn main() {
    let saludo = String::from("hola");
    agregar_signo(&saludo);
    println!("{saludo}");
}
