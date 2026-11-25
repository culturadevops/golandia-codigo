use clap::{Parser, Subcommand};
use regex::Regex;
use std::fs;

/// fuga: busca secretos filtrados en archivos de texto
#[derive(Parser)]
#[command(name = "fuga", version, about)]
struct Cli {
    #[command(subcommand)]
    orden: Orden,
}

#[derive(Subcommand)]
enum Orden {
    /// Revisa archivos en busca de secretos
    Escanear {
        /// Archivos a revisar
        #[arg(required = true)]
        archivos: Vec<String>,
        /// Salida en JSON
        #[arg(long)]
        json: bool,
    },
    /// Lista los patrones que busca
    Patrones,
}

fn patrones() -> Vec<(&'static str, Regex)> {
    vec![
        ("aws-access-key", Regex::new(r"\bAKIA[0-9A-Z]{16}\b").unwrap()),
        ("github-token", Regex::new(r"\bghp_[A-Za-z0-9]{36}\b").unwrap()),
        ("clave-privada", Regex::new(r"-----BEGIN [A-Z ]*PRIVATE KEY-----").unwrap()),
        ("contrasena", Regex::new(r"(?i)\b(password|passwd|contraseña)\s*[:=]\s*\S+").unwrap()),
    ]
}

fn tapar(s: &str) -> String {
    let c: Vec<char> = s.chars().collect();
    let n = c.len().min(4);
    c[..n].iter().collect::<String>() + &"*".repeat(c.len() - n)
}

fn main() {
    let cli = Cli::parse();
    match cli.orden {
        Orden::Patrones => {
            for (nombre, re) in patrones() {
                println!("{nombre:16} {re}");
            }
        }
        Orden::Escanear { archivos, json: _ } => {
            let pats = patrones();
            for a in &archivos {
                let texto = fs::read_to_string(a).unwrap();
                for (i, linea) in texto.lines().enumerate() {
                    for (tipo, re) in &pats {
                        if let Some(m) = re.find(linea) {
                            println!("{a}:{}  {tipo}  {}", i + 1, tapar(m.as_str()));
                        }
                    }
                }
            }
        }
    }
}
