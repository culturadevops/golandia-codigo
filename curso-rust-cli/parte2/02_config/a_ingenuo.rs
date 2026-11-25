use clap::Parser;
use serde::Deserialize;
use std::fs;

#[derive(Parser)]
struct Cli {
    /// Máximo de hallazgos por archivo
    #[arg(short, long, default_value_t = 10)]
    max: usize,
}

#[derive(Deserialize)]
struct Config {
    max: Option<usize>,
}

fn main() {
    let cli = Cli::parse();
    let mut max = cli.max;
    if let Ok(t) = fs::read_to_string("fuga.toml") {
        let c: Config = toml::from_str(&t).unwrap();
        if max == 10 {
            max = c.max.unwrap_or(max); // «si el flag es el de por defecto, manda el archivo»
        }
    }
    println!("max efectivo: {max}");
}
