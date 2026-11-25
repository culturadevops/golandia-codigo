fn guardar(dato: &str) -> Result<(), String> {
    if dato.is_empty() {
        return Err(String::from("dato vacío"));
    }
    println!("guardado: {dato}");
    Ok(())
}

fn main() {
    guardar("");
    println!("todo bien (¿seguro?)");
}
