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
