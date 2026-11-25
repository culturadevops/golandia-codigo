fn main() {
    let puntos = [10, 20, 30, 40, 50];
    let i: usize = std::env::args().nth(1).unwrap().parse().unwrap();
    println!("pido la posición {i}");
    println!("valor: {}", puntos[i]);
    println!("fin del programa");
}
