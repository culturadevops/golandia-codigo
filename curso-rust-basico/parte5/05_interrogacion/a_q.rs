use std::error::Error;
use std::fs;

fn leer_puerto(ruta: &str) -> Result<u16, Box<dyn Error>> {
    let texto = fs::read_to_string(ruta)?;
    let valor = texto.trim().trim_start_matches("puerto=");
    let puerto: u16 = valor.parse()?;
    Ok(puerto)
}

fn main() -> Result<(), Box<dyn Error>> {
    let ruta = std::env::args().nth(1).unwrap_or(String::from("config.txt"));
    let puerto = leer_puerto(&ruta)?;
    println!("escucharé en el puerto {puerto}");
    Ok(())
}
