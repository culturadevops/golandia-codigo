use std::collections::HashMap;
use std::error::Error;
use std::fs;
use std::process;

struct Config {
    ruta: String,
    top: usize,
}

fn leer_args(args: &[String]) -> Result<Config, String> {
    let ruta = match args.get(1) {
        Some(r) => r.clone(),
        None => return Err(String::from("uso: palabras <archivo> [top]")),
    };
    let top = match args.get(2) {
        Some(t) => t.parse().map_err(|_| format!("top inválido: {t}"))?,
        None => 5,
    };
    Ok(Config { ruta, top })
}

fn contar(texto: &str) -> HashMap<String, usize> {
    let mut cuenta = HashMap::new();
    for palabra in texto.split_whitespace() {
        let limpia: String = palabra
            .chars()
            .filter(|c| c.is_alphabetic())
            .flat_map(|c| c.to_lowercase())
            .collect();
        if !limpia.is_empty() {
            *cuenta.entry(limpia).or_insert(0) += 1;
        }
    }
    cuenta
}

fn ejecutar(cfg: &Config) -> Result<(), Box<dyn Error>> {
    let texto = fs::read_to_string(&cfg.ruta)?;
    let cuenta = contar(&texto);
    let mut lista: Vec<(&String, &usize)> = cuenta.iter().collect();
    lista.sort_by(|a, b| b.1.cmp(a.1).then(a.0.cmp(b.0)));
    let total: usize = cuenta.values().sum();
    println!("{} palabras, {} distintas", total, cuenta.len());
    for (palabra, n) in lista.iter().take(cfg.top) {
        println!("{n:>8}  {palabra}");
    }
    Ok(())
}

fn main() {
    let args: Vec<String> = std::env::args().collect();
    let cfg = leer_args(&args).unwrap_or_else(|e| {
        eprintln!("{e}");
        process::exit(2);
    });
    if let Err(e) = ejecutar(&cfg) {
        eprintln!("error: {}: {e}", cfg.ruta);
        process::exit(1);
    }
}
