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
