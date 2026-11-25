fn main() {
    let x = 5;
    {
        let x = x * 2;
        println!("dentro del bloque: {x}");
    }
    println!("fuera del bloque: {x}");
}
