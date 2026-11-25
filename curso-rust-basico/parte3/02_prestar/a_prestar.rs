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
