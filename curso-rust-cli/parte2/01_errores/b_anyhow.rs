use anyhow::{Context, Result};
use std::fs;

fn main() -> Result<()> {
    for a in std::env::args().skip(1) {
        let texto = fs::read_to_string(&a).with_context(|| format!("no pude leer {a}"))?;
        println!("{a}: {} líneas", texto.lines().count());
    }
    Ok(())
}
