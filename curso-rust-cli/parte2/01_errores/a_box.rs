use std::error::Error;
use std::fs;

fn main() -> Result<(), Box<dyn Error>> {
    for a in std::env::args().skip(1) {
        let texto = fs::read_to_string(&a)?;
        println!("{a}: {} líneas", texto.lines().count());
    }
    Ok(())
}
