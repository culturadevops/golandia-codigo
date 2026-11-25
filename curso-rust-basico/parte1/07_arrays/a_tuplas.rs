fn main() {
    let jugador: (&str, u32, f64) = ("Ana", 3, 1.75);
    let (nombre, vidas, altura) = jugador;
    println!("{nombre} tiene {vidas} vidas y mide {altura}");
    println!("campo 1 de la tupla: {}", jugador.1);

    let puntos = [10, 20, 30, 40, 50];
    let ceros = [0; 4];
    println!("primero {}, largo {}, ceros {:?}", puntos[0], puntos.len(), ceros);
}
