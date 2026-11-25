fn main() {
    let notas = vec![70, 95, 88];
    let i: usize = std::env::args().nth(1).unwrap().parse().unwrap();
    match notas.get(i) {
        Some(n) => println!("con get: {n}"),
        None => println!("con get: no hay nota {i}"),
    }
    println!("con []: {}", notas[i]);
}
