fn main() {
    'filas: for fila in 1..=3 {
        for col in 1..=3 {
            if fila * col == 4 {
                println!("encontré el 4 en ({fila}, {col})");
                break 'filas;
            }
        }
    }
    println!("fin");
}
