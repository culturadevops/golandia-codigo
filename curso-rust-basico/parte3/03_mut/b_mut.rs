fn agregar_signo(texto: &mut String) {
    texto.push('!');
}

fn main() {
    let mut saludo = String::from("hola");
    agregar_signo(&mut saludo);
    agregar_signo(&mut saludo);
    println!("{saludo}");
}
