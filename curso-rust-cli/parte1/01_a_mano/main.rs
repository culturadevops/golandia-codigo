use std::env;
use std::fs;

fn main() {
    let args: Vec<String> = env::args().collect();
    let ruta = &args[1];
    let texto = fs::read_to_string(ruta).unwrap();
    let n = texto.lines().filter(|l| l.contains("AKIA")).count();
    println!("{ruta}: {n} posibles claves de AWS");
}
