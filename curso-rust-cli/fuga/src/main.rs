use std::collections::BTreeMap;
use std::fs;
use std::io::{self, IsTerminal, Read, Write};
use std::path::PathBuf;
use std::process::ExitCode;

use anyhow::{Context, Result};
use clap::{Parser, Subcommand};
use regex::Regex;
use serde::{Deserialize, Serialize};

/// fuga: busca secretos filtrados (claves, tokens, contraseñas) en archivos de texto
#[derive(Parser)]
#[command(name = "fuga", version, about)]
struct Cli {
    #[command(subcommand)]
    orden: Orden,
}

#[derive(Subcommand)]
enum Orden {
    /// Revisa archivos (o la entrada estándar con -)
    Escanear {
        /// Archivos a revisar; - lee la entrada estándar
        #[arg(required = true)]
        archivos: Vec<String>,
        /// Máximo de hallazgos por archivo (si no: FUGA_MAX, fuga.toml, 10)
        #[arg(short, long, env = "FUGA_MAX")]
        max: Option<usize>,
        /// Archivo de configuración
        #[arg(long, default_value = "fuga.toml")]
        config: PathBuf,
        /// Salida en JSON para jq
        #[arg(long)]
        json: bool,
    },
    /// Lista los patrones que busca
    Patrones,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct Config {
    max: Option<usize>,
}

#[derive(Serialize)]
struct Hallazgo {
    archivo: String,
    linea: usize,
    tipo: &'static str,
    muestra: String,
}

#[derive(Serialize)]
struct Informe {
    hallazgos: Vec<Hallazgo>,
    por_tipo: BTreeMap<&'static str, usize>,
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

fn detectar(archivo: &str, texto: &str, max: usize, pats: &[(&'static str, Regex)]) -> Vec<Hallazgo> {
    let mut v = Vec::new();
    for (i, linea) in texto.lines().enumerate() {
        for (tipo, re) in pats {
            if let Some(m) = re.find(linea) {
                if v.len() == max {
                    return v;
                }
                v.push(Hallazgo { archivo: archivo.to_string(), linea: i + 1, tipo, muestra: tapar(m.as_str()) });
            }
        }
    }
    v
}

fn leer(ruta: &str) -> Result<String> {
    if ruta == "-" {
        let mut s = String::new();
        io::stdin().read_to_string(&mut s).context("no pude leer la entrada estándar")?;
        return Ok(s);
    }
    fs::read_to_string(ruta).with_context(|| format!("no pude leer {ruta}"))
}

/// flag > FUGA_MAX (clap los junta en `flag`) > fuga.toml > 10
fn max_efectivo(flag: Option<usize>, config: &PathBuf) -> Result<usize> {
    if let Some(m) = flag {
        return Ok(m);
    }
    match fs::read_to_string(config) {
        Ok(t) => {
            let c: Config = toml::from_str(&t).with_context(|| format!("{} no es válido", config.display()))?;
            Ok(c.max.unwrap_or(10))
        }
        Err(e) if e.kind() == io::ErrorKind::NotFound => Ok(10),
        Err(e) => Err(e).with_context(|| format!("no pude leer {}", config.display())),
    }
}

fn escanear(archivos: &[String], max: Option<usize>, config: &PathBuf, json: bool) -> Result<bool> {
    let max = max_efectivo(max, config)?;
    let pats = patrones();
    let mut hallazgos = Vec::new();
    for a in archivos {
        hallazgos.extend(detectar(a, &leer(a)?, max, &pats));
    }
    let hay = !hallazgos.is_empty();
    let mut out = io::stdout().lock();
    if json {
        let mut por_tipo = BTreeMap::new();
        for h in &hallazgos {
            *por_tipo.entry(h.tipo).or_insert(0) += 1;
        }
        serde_json::to_writer_pretty(&mut out, &Informe { hallazgos, por_tipo })?;
        writeln!(out)?;
        return Ok(hay);
    }
    let color = io::stdout().is_terminal() && std::env::var_os("NO_COLOR").is_none();
    for h in &hallazgos {
        let tipo = if color { format!("\x1b[31m{}\x1b[0m", h.tipo) } else { h.tipo.to_string() };
        writeln!(out, "{}:{}  {}  {}", h.archivo, h.linea, tipo, h.muestra)?;
    }
    eprintln!("{} hallazgo(s) en {} archivo(s)", hallazgos.len(), archivos.len());
    Ok(hay)
}

fn main() -> ExitCode {
    let cli = Cli::parse();
    let r = match cli.orden {
        Orden::Escanear { archivos, max, config, json } => escanear(&archivos, max, &config, json),
        Orden::Patrones => {
            for (nombre, re) in patrones() {
                println!("{nombre:16} {re}");
            }
            Ok(false)
        }
    };
    match r {
        Ok(false) => ExitCode::SUCCESS,              // 0: limpio
        Ok(true) => ExitCode::from(1),               // 1: hay secretos
        Err(e) if e.downcast_ref::<io::Error>().is_some_and(|e| e.kind() == io::ErrorKind::BrokenPipe) => {
            ExitCode::SUCCESS                        // quien leía cerró la tubería
        }
        Err(e) => {
            eprintln!("error: {e:#}");
            ExitCode::from(3)                        // 3: no pude trabajar (2 es de clap)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn detecta_la_clave_de_aws() {
        let h = detectar("x", "KEY=AKIAIOSFODNN7EXAMPLE", 10, &patrones());
        assert_eq!(h.len(), 1);
        assert_eq!(h[0].tipo, "aws-access-key");
    }

    #[test]
    fn no_detecta_dentro_de_otra_palabra() {
        let h = detectar("x", "pedido XAKIAIOSFODNN7EXAMPLEX", 10, &patrones());
        assert_eq!(h.len(), 0);
    }

    #[test]
    fn tapa_todo_menos_cuatro() {
        assert_eq!(tapar("AKIAIOSFODNN7EXAMPLE"), "AKIA****************");
    }

    #[test]
    fn respeta_el_maximo() {
        let texto = "AKIAIOSFODNN7EXAMPLE\n".repeat(5);
        assert_eq!(detectar("x", &texto, 2, &patrones()).len(), 2);
    }
}
