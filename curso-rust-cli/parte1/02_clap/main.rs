use clap::Parser;

/// fuga: busca secretos filtrados en archivos de texto
#[derive(Parser)]
#[command(name = "fuga", version, about)]
struct Cli {
    /// Archivos a revisar
    #[arg(required = true)]
    archivos: Vec<String>,

    /// Máximo de hallazgos por archivo
    #[arg(short, long, default_value_t = 10)]
    max: usize,

    /// Salida en JSON
    #[arg(long)]
    json: bool,
}

fn main() {
    let cli = Cli::parse();
    for a in &cli.archivos {
        println!("revisaría {a}  (max {}, json {})", cli.max, cli.json);
    }
}
