fn main() {
    let s = String::from("hola");
    println!("String en la pila: {} bytes", std::mem::size_of::<String>());
    println!("texto en el montón: {} bytes, capacidad {}", s.len(), s.capacity());
    let literal: &str = "hola";
    println!("un &str ocupa {} bytes", std::mem::size_of::<&str>());
    println!("{literal}");
}
