fn primera_palabra(texto: &str) -> &str {
    texto.split(' ').next().unwrap()
}

fn main() {
    let mut frase = String::from("hola mundo");
    let p = primera_palabra(&frase);
    frase.clear();
    println!("primera palabra: {p}");
}
