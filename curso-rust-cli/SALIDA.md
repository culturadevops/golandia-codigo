# SALIDA — Curso de Rust para crear CLI 2027

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Los errores son a propósito: cada lección
> enseña la trampa y lo que pasa de verdad. Cada orden va con su «codigo de salida».

- fecha: `2026-11-13 14:13:27Z`
- `rustc 1.99.1 (58cf8f9c3 2026-10-13)`
- `cargo 1.99.1 (5f94df478 2026-08-27)`
- `jq: jq-1.7`
- `anyhow 1.0.104`
- `assert_cmd 2.2.2`
- `clap 4.6.7`
- `predicates 3.1.4`
- `regex 1.13.1`
- `serde 1.0.229`
- `serde_json 1.0.154`
- `toml 1.2.1+spec-1.1.0`
- `Linux 6.18.33.2-microsoft-standard-WSL2 x86_64`

## parte1/01_a_mano

`bash parte1/01_a_mano/demo.sh` · código de salida `0`

```
--- main.rs
use std::env;
use std::fs;

fn main() {
    let args: Vec<String> = env::args().collect();
    let ruta = &args[1];
    let texto = fs::read_to_string(ruta).unwrap();
    let n = texto.lines().filter(|l| l.contains("AKIA")).count();
    println!("{ruta}: {n} posibles claves de AWS");
}
---
$ fuga datos/config.env
datos/config.env: 1 posibles claves de AWS
[codigo de salida: 0]
$ fuga --help

thread 'main' (145) panicked at src/main.rs:7:42:
called `Result::unwrap()` on an `Err` value: Os { code: 2, kind: NotFound, message: "No such file or directory" }
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
$ fuga

thread 'main' (147) panicked at src/main.rs:6:21:
index out of bounds: the len is 1 but the index is 1
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
[codigo de salida: 101]
```

## parte1/02_clap

`bash parte1/02_clap/demo.sh` · código de salida `0`

```
--- main.rs
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
---
$ fuga --help
fuga: busca secretos filtrados en archivos de texto

Usage: fuga [OPTIONS] <ARCHIVOS>...

Arguments:
  <ARCHIVOS>...  Archivos a revisar

Options:
  -m, --max <MAX>  Máximo de hallazgos por archivo [default: 10]
      --json       Salida en JSON
  -h, --help       Print help
  -V, --version    Print version
[codigo de salida: 0]
$ fuga --version
fuga 1.0.0
[codigo de salida: 0]
$ fuga datos/config.env -m 3
revisaría datos/config.env  (max 3, json false)
[codigo de salida: 0]
$ fuga
error: the following required arguments were not provided:
  <ARCHIVOS>...

Usage: fuga <ARCHIVOS>...

For more information, try '--help'.
[codigo de salida: 2]
$ fuga --maxx 3 datos/config.env
error: unexpected argument '--maxx' found

  tip: a similar argument exists: '--max'

Usage: fuga --max <MAX> <ARCHIVOS>...

For more information, try '--help'.
[codigo de salida: 2]
$ fuga --max tres datos/config.env
error: invalid value 'tres' for '--max <MAX>': invalid digit found in string

For more information, try '--help'.
[codigo de salida: 2]
$ fuga --json false datos/config.env
revisaría false  (max 10, json true)
revisaría datos/config.env  (max 10, json true)
[codigo de salida: 0]
```

## parte1/03_subcomandos

`bash parte1/03_subcomandos/demo.sh` · código de salida `0`

```
--- main.rs
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
---
$ fuga --help
fuga: busca secretos filtrados en archivos de texto

Usage: fuga <COMMAND>

Commands:
  escanear  Revisa archivos en busca de secretos
  patrones  Lista los patrones que busca
  help      Print this message or the help of the given subcommand(s)

Options:
  -h, --help     Print help
  -V, --version  Print version
[codigo de salida: 0]
$ fuga patrones
aws-access-key   \bAKIA[0-9A-Z]{16}\b
github-token     \bghp_[A-Za-z0-9]{36}\b
clave-privada    -----BEGIN [A-Z ]*PRIVATE KEY-----
contrasena       (?i)\b(password|passwd|contraseña)\s*[:=]\s*\S+
[codigo de salida: 0]
$ fuga escanear datos/config.env datos/app.yaml datos/deploy_key.pem datos/README.md
datos/config.env:2  aws-access-key  AKIA****************
datos/config.env:4  github-token  ghp_************************************
datos/app.yaml:4  contrasena  pass*******************
datos/deploy_key.pem:1  clave-privada  ----*******************************
[codigo de salida: 0]
$ fuga --json escanear datos/config.env
error: unexpected argument '--json' found

  tip: 'escanear --json' exists

Usage: fuga <COMMAND>

For more information, try '--help'.
[codigo de salida: 2]
```

## parte1/04_salida

`bash parte1/04_salida/demo.sh` · código de salida `0`

```
--- a_progreso.rs
use regex::Regex;
use std::fs;

fn main() {
    let re = Regex::new(r"\bAKIA[0-9A-Z]{16}\b").unwrap();
    for a in std::env::args().skip(1) {
        println!("revisando {a}...");
        for (i, l) in fs::read_to_string(&a).unwrap().lines().enumerate() {
            if re.is_match(l) {
                println!("{a}:{}  aws-access-key", i + 1);
            }
        }
    }
}
---
$ fuga datos/config.env datos/app.yaml datos/README.md
revisando datos/config.env...
datos/config.env:2  aws-access-key
revisando datos/app.yaml...
revisando datos/README.md...
[codigo de salida: 0]
$ fuga datos/config.env datos/app.yaml datos/README.md | wc -l
4
[codigo de salida: 0]
--- b_stderr.rs
use regex::Regex;
use std::fs;

fn main() {
    let re = Regex::new(r"\bAKIA[0-9A-Z]{16}\b").unwrap();
    for a in std::env::args().skip(1) {
        eprintln!("revisando {a}...");
        for (i, l) in fs::read_to_string(&a).unwrap().lines().enumerate() {
            if re.is_match(l) {
                println!("{a}:{}  aws-access-key", i + 1);
            }
        }
    }
}
---
$ fuga datos/config.env datos/app.yaml datos/README.md | wc -l
revisando datos/config.env...
revisando datos/app.yaml...
revisando datos/README.md...
1
[codigo de salida: 0]
$ fuga datos/config.env datos/app.yaml datos/README.md 2>/dev/null
datos/config.env:2  aws-access-key
[codigo de salida: 0]
--- c_pipe.rs
use regex::Regex;
use std::fs;

fn main() {
    let re = Regex::new(r"\bAKIA[0-9A-Z]{16}\b").unwrap();
    for a in std::env::args().skip(1) {
        for (i, l) in fs::read_to_string(&a).unwrap().lines().enumerate() {
            if re.is_match(l) {
                println!("{a}:{}  aws-access-key", i + 1);
            }
        }
    }
}
---
$ wc -l /tmp/grande.env
200000 /tmp/grande.env
[codigo de salida: 0]
$ fuga /tmp/grande.env | head -2; echo "PIPESTATUS: ${PIPESTATUS[*]}"
/tmp/grande.env:1  aws-access-key
/tmp/grande.env:2  aws-access-key

thread 'main' (704) panicked at /rustc/58cf8f9c35b355199727627d07ae34a4e2846d67/library/std/src/io/stdio.rs:1168:9:
failed printing to stdout: Broken pipe (os error 32)
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
PIPESTATUS: 101 0
[codigo de salida: 0]
--- d_pipe_ok.rs
use regex::Regex;
use std::fs;
use std::io::{self, Write};

fn main() {
    let re = Regex::new(r"\bAKIA[0-9A-Z]{16}\b").unwrap();
    let mut out = io::stdout().lock();
    for a in std::env::args().skip(1) {
        for (i, l) in fs::read_to_string(&a).unwrap().lines().enumerate() {
            if re.is_match(l) {
                if let Err(e) = writeln!(out, "{a}:{}  aws-access-key", i + 1) {
                    if e.kind() == io::ErrorKind::BrokenPipe {
                        return; // quien leía ya cerró: salimos sin ruido
                    }
                    panic!("{e}");
                }
            }
        }
    }
}
---
$ fuga /tmp/grande.env | head -2; echo "PIPESTATUS: ${PIPESTATUS[*]}"
/tmp/grande.env:1  aws-access-key
/tmp/grande.env:2  aws-access-key
PIPESTATUS: 0 0
[codigo de salida: 0]
```

## parte1/05_codigos

`bash parte1/05_codigos/demo.sh` · código de salida `0`

```
--- a_exit.rs
use regex::Regex;
use std::fs;
use std::io::{self, BufWriter, Write};

fn main() {
    let re = Regex::new(r"\bAKIA[0-9A-Z]{16}\b").unwrap();
    let mut out = BufWriter::new(io::stdout().lock());
    let mut hallazgos = 0;
    for a in std::env::args().skip(1) {
        for (i, l) in fs::read_to_string(&a).unwrap().lines().enumerate() {
            if re.is_match(l) {
                writeln!(out, "{a}:{}  aws-access-key", i + 1).unwrap();
                hallazgos += 1;
            }
        }
    }
    if hallazgos > 0 {
        std::process::exit(1); // sale ya: el BufWriter no llega a vaciarse
    }
}
---
$ fuga datos/config.env
[codigo de salida: 1]
$ fuga datos/README.md
[codigo de salida: 0]
--- b_exitcode.rs
use regex::Regex;
use std::fs;
use std::io::{self, BufWriter, Write};
use std::process::ExitCode;

fn main() -> ExitCode {
    let re = Regex::new(r"\bAKIA[0-9A-Z]{16}\b").unwrap();
    let mut out = BufWriter::new(io::stdout().lock());
    let mut hallazgos = 0;
    for a in std::env::args().skip(1) {
        for (i, l) in fs::read_to_string(&a).unwrap().lines().enumerate() {
            if re.is_match(l) {
                writeln!(out, "{a}:{}  aws-access-key", i + 1).unwrap();
                hallazgos += 1;
            }
        }
    }
    if hallazgos > 0 { ExitCode::from(1) } else { ExitCode::SUCCESS }
}
---
$ fuga datos/config.env
datos/config.env:2  aws-access-key
[codigo de salida: 1]
$ fuga datos/README.md && echo "limpio: sigo con el despliegue"
limpio: sigo con el despliegue
[codigo de salida: 0]
$ fuga datos/config.env && echo "limpio: sigo con el despliegue" || echo "FRENO: hay secretos"
datos/config.env:2  aws-access-key
FRENO: hay secretos
[codigo de salida: 0]
```

## parte2/01_errores

`bash parte2/01_errores/demo.sh` · código de salida `0`

```
--- a_box.rs
use std::error::Error;
use std::fs;

fn main() -> Result<(), Box<dyn Error>> {
    for a in std::env::args().skip(1) {
        let texto = fs::read_to_string(&a)?;
        println!("{a}: {} líneas", texto.lines().count());
    }
    Ok(())
}
---
$ fuga datos/config.env datos/no-existe.env datos/app.yaml
datos/config.env: 4 líneas
Error: Os { code: 2, kind: NotFound, message: "No such file or directory" }
[codigo de salida: 1]
--- b_anyhow.rs
use anyhow::{Context, Result};
use std::fs;

fn main() -> Result<()> {
    for a in std::env::args().skip(1) {
        let texto = fs::read_to_string(&a).with_context(|| format!("no pude leer {a}"))?;
        println!("{a}: {} líneas", texto.lines().count());
    }
    Ok(())
}
---
$ fuga datos/config.env datos/no-existe.env datos/app.yaml
datos/config.env: 4 líneas
Error: no pude leer datos/no-existe.env

Caused by:
    No such file or directory (os error 2)
[codigo de salida: 1]
# la herramienta completa: 0 limpio, 1 hay secretos, 3 no pude trabajar, 2 lo pone clap
$ fuga escanear datos/config.env
datos/config.env:2  aws-access-key  AKIA****************
datos/config.env:4  github-token  ghp_************************************
2 hallazgo(s) en 1 archivo(s)
[codigo de salida: 1]
$ fuga escanear datos/no-existe.env
error: no pude leer datos/no-existe.env: No such file or directory (os error 2)
[codigo de salida: 3]
$ fuga escanear datos/README.md
0 hallazgo(s) en 1 archivo(s)
[codigo de salida: 0]
```

## parte2/02_config

`bash parte2/02_config/demo.sh` · código de salida `0`

```
--- a_ingenuo.rs
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
---
$ cat fuga.toml
max = 3
$ fuga
max efectivo: 3
[codigo de salida: 0]
$ fuga --max 10
max efectivo: 3
[codigo de salida: 0]
# la herramienta completa: Option<usize> + env = FUGA_MAX + fuga.toml + 10
$ fuga escanear config.env | wc -l
10 hallazgo(s) en 1 archivo(s)
10
[codigo de salida: 0]
$ echo "max = 3" > fuga.toml
$ fuga escanear config.env | wc -l
3 hallazgo(s) en 1 archivo(s)
3
[codigo de salida: 0]
$ FUGA_MAX=5 fuga escanear config.env | wc -l
5 hallazgo(s) en 1 archivo(s)
5
[codigo de salida: 0]
$ FUGA_MAX=5 fuga escanear config.env --max 7 | wc -l
7 hallazgo(s) en 1 archivo(s)
7
[codigo de salida: 0]
$ fuga escanear config.env --max 10 | wc -l
10 hallazgo(s) en 1 archivo(s)
10
[codigo de salida: 0]
$ FUGA_MAX=cinco fuga escanear config.env
error: invalid value 'cinco' for '--max <MAX>': invalid digit found in string

For more information, try '--help'.
[codigo de salida: 2]
$ echo "maxx = 3" > fuga.toml
$ fuga escanear config.env
error: fuga.toml no es válido: TOML parse error at line 1, column 1
  |
1 | maxx = 3
  | ^^^^
unknown field `maxx`, expected `max`

[codigo de salida: 3]
--- b_sin_deny.rs
use serde::Deserialize;

#[derive(Deserialize)]
struct Config {
    max: Option<usize>,
}

fn main() {
    let c: Config = toml::from_str("maxx = 3").unwrap();
    println!("sin deny_unknown_fields: max = {:?}", c.max);
}
---
$ fuga
sin deny_unknown_fields: max = None
[codigo de salida: 0]
```

## parte2/03_json

`bash parte2/03_json/demo.sh` · código de salida `0`

```
--- a_hashmap.rs
use serde::Serialize;
use std::collections::HashMap;

#[derive(Serialize)]
struct Resumen {
    por_tipo: HashMap<&'static str, usize>,
}

fn main() {
    let mut por_tipo = HashMap::new();
    for t in ["aws-access-key", "github-token", "contrasena", "clave-privada"] {
        por_tipo.insert(t, 1);
    }
    println!("{}", serde_json::to_string(&Resumen { por_tipo }).unwrap());
}
---
$ fuga
{"por_tipo":{"contrasena":1,"github-token":1,"clave-privada":1,"aws-access-key":1}}
[codigo de salida: 0]
$ fuga
{"por_tipo":{"clave-privada":1,"aws-access-key":1,"github-token":1,"contrasena":1}}
[codigo de salida: 0]
$ fuga
{"por_tipo":{"github-token":1,"clave-privada":1,"aws-access-key":1,"contrasena":1}}
[codigo de salida: 0]
$ fuga > a.json; fuga > b.json; diff -q a.json b.json
Files a.json and b.json differ
[codigo de salida: 1]
# la herramienta completa: --json con BTreeMap
$ fuga escanear --json datos/config.env datos/app.yaml
{
  "hallazgos": [
    {
      "archivo": "datos/config.env",
      "linea": 2,
      "tipo": "aws-access-key",
      "muestra": "AKIA****************"
    },
    {
      "archivo": "datos/config.env",
      "linea": 4,
      "tipo": "github-token",
      "muestra": "ghp_************************************"
    },
    {
      "archivo": "datos/app.yaml",
      "linea": 4,
      "tipo": "contrasena",
      "muestra": "pass*******************"
    }
  ],
  "por_tipo": {
    "aws-access-key": 1,
    "contrasena": 1,
    "github-token": 1
  }
}
[codigo de salida: 1]
$ fuga escanear --json datos/*.env datos/*.yaml datos/*.pem > a.json; fuga escanear --json datos/*.env datos/*.yaml datos/*.pem > b.json; diff -q a.json b.json && echo idénticos
idénticos
[codigo de salida: 0]
$ fuga escanear --json datos/* | jq -r ".hallazgos[] | \"\(.archivo):\(.linea) \(.tipo)\""
datos/app.yaml:4 contrasena
datos/config.env:2 aws-access-key
datos/config.env:4 github-token
datos/deploy_key.pem:1 clave-privada
[codigo de salida: 0]
$ fuga escanear --json datos/* | jq ".por_tipo"
{
  "aws-access-key": 1,
  "clave-privada": 1,
  "contrasena": 1,
  "github-token": 1
}
[codigo de salida: 0]
$ fuga escanear --json datos/README.md | jq ".hallazgos | length"
0
[codigo de salida: 0]
```

## parte2/04_color

`bash parte2/04_color/demo.sh` · código de salida `0`

```
--- a_siempre.rs
fn main() {
    for (archivo, tipo) in [("datos/config.env:2", "aws-access-key"), ("datos/app.yaml:4", "contrasena")] {
        println!("{archivo}  \x1b[31m{tipo}\x1b[0m");
    }
}
---
$ fuga > /tmp/informe.txt; cat -v /tmp/informe.txt
datos/config.env:2  ^[[31maws-access-key^[[0m
datos/app.yaml:4  ^[[31mcontrasena^[[0m
[codigo de salida: 0]
$ fuga | grep -c "aws-access-key$"
0
[codigo de salida: 1]
# la herramienta completa: color solo si stdout es una terminal y no hay NO_COLOR
$ fuga escanear datos/config.env 2>/dev/null | cat -v
datos/config.env:2  aws-access-key  AKIA****************
datos/config.env:4  github-token  ghp_************************************
[codigo de salida: 0]
$ script -qec "fuga escanear datos/config.env" /dev/null | cat -v
datos/config.env:2  ^[[31maws-access-key^[[0m  AKIA****************^M
datos/config.env:4  ^[[31mgithub-token^[[0m  ghp_************************************^M
2 hallazgo(s) en 1 archivo(s)^M
[codigo de salida: 0]
$ NO_COLOR=1 script -qec "fuga escanear datos/config.env" /dev/null | cat -v
datos/config.env:2  aws-access-key  AKIA****************^M
datos/config.env:4  github-token  ghp_************************************^M
2 hallazgo(s) en 1 archivo(s)^M
[codigo de salida: 0]
```

## parte3/01_tests

`bash parte3/01_tests/demo.sh` · código de salida `0`

```
$ grep -n "aws-access-key\", Regex" src/main.rs   # versión sin \b
63:        ("aws-access-key", Regex::new(r"AKIA[0-9A-Z]{16}").unwrap()),
$ cargo test
running 4 tests
... 3/4
tests::no_detecta_dentro_de_otra_palabra --- FAILED
failures:
---- tests::no_detecta_dentro_de_otra_palabra stdout ----
thread 'tests::no_detecta_dentro_de_otra_palabra' (2467) panicked at src/main.rs:180:9:
assertion `left == right` failed
  left: 1
 right: 0
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
failures:
    tests::no_detecta_dentro_de_otra_palabra
test result: FAILED. 3 passed; 1 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s
error: test failed, to rerun pass `--bin fuga`
[codigo de salida: 101]
$ fuga escanear datos/notas.txt   # con el patrón sin \b
datos/notas.txt:1  aws-access-key  AKIA****************
datos/notas.txt:2  aws-access-key  AKIA****************
2 hallazgo(s) en 1 archivo(s)
[codigo de salida: 1]
# con \b, la herramienta completa
$ cargo test
running 4 tests
....
test result: ok. 4 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s
[codigo de salida: 0]
$ fuga escanear datos/notas.txt
0 hallazgo(s) en 1 archivo(s)
[codigo de salida: 0]
```

## parte3/02_assert_cmd

`bash parte3/02_assert_cmd/demo.sh` · código de salida `0`

```
--- a_cli.rs
use assert_cmd::Command;
use predicates::str::contains;

fn fuga() -> Command {
    Command::cargo_bin("fuga").unwrap()
}

#[test]
fn con_secretos_sale_con_1() {
    fuga().args(["escanear", "../../datos/config.env"])
        .assert().code(1).stdout(contains("aws-access-key"));
}

#[test]
fn limpio_sale_con_0() {
    fuga().args(["escanear", "../../datos/README.md"]).assert().success();
}

#[test]
fn archivo_que_no_existe_sale_con_3() {
    fuga().args(["escanear", "no-existe.env"])
        .assert().code(3).stderr(contains("no pude leer no-existe.env"));
}

#[test]
fn lee_la_entrada_estandar() {
    fuga().args(["escanear", "-"]).write_stdin("password: hunter2\n")
        .assert().code(1).stdout(contains("contrasena"));
}

#[test]
fn muestra_los_dos_hallazgos_de_config_env() {
    fuga().args(["escanear", "../../datos/config.env"])
        .assert().code(1).stdout(contains("github-token"));
}
---
$ cargo test -q --locked --test cli
test result: ok. 5 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s
[codigo de salida: 0]
$ FUGA_MAX=1 cargo test -q --locked --test cli
thread 'muestra_los_dos_hallazgos_de_config_env' (3215) panicked at tests/cli.rs:34:27:
Unexpected stdout, failed var.contains(github-token)
code=1
stdout="../../datos/config.env:2  aws-access-key  AKIA****************\n"
test result: FAILED. 4 passed; 1 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s
[codigo de salida: 101]
--- b_cli.rs
use assert_cmd::Command;
use predicates::str::contains;

fn fuga() -> Command {
    let mut c = Command::cargo_bin("fuga").unwrap();
    c.env_remove("FUGA_MAX").env_remove("NO_COLOR"); // el test no hereda tu entorno
    c
}

#[test]
fn con_secretos_sale_con_1() {
    fuga().args(["escanear", "../../datos/config.env"])
        .assert().code(1).stdout(contains("aws-access-key"));
}

#[test]
fn limpio_sale_con_0() {
    fuga().args(["escanear", "../../datos/README.md"]).assert().success();
}

#[test]
fn archivo_que_no_existe_sale_con_3() {
    fuga().args(["escanear", "no-existe.env"])
        .assert().code(3).stderr(contains("no pude leer no-existe.env"));
}

#[test]
fn lee_la_entrada_estandar() {
    fuga().args(["escanear", "-"]).write_stdin("password: hunter2\n")
        .assert().code(1).stdout(contains("contrasena"));
}

#[test]
fn muestra_los_dos_hallazgos_de_config_env() {
    fuga().args(["escanear", "../../datos/config.env"])
        .assert().code(1).stdout(contains("github-token"));
}
---
$ FUGA_MAX=1 cargo test -q --locked --test cli
test result: ok. 5 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s
[codigo de salida: 0]
```

## parte3/03_release

`bash parte3/03_release/demo.sh` · código de salida `0`

```
perfil                                            bytes
debug (cargo build)                            38646576
release                                         3841256
+ strip = true                                  2885552
+ lto = fat, codegen-units = 1                  2360512
+ opt-level = z                                 1926280
+ panic = abort                                 1879656
$ wc -l grande.env
400000
debug        0.566 s · 0.576 s · 0.562 s
release      0.059 s · 0.057 s · 0.057 s
lto          0.055 s · 0.054 s · 0.054 s
opt-level z  0.069 s · 0.067 s · 0.070 s
$ FUGA_MAX=1000000 fuga escanear grande.env | wc -l
100000
```

## parte3/04_musl

`bash parte3/04_musl/demo.sh` · código de salida `0`

```
$ file fuga-gnu
ELF 64-bit LSB pie executable, x86-64, dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2
$ ldd fuga-gnu
	linux-vdso.so.1
	libgcc_s.so.1 => /lib/x86_64-linux-gnu/libgcc_s.so.1
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
	/lib64/ld-linux-x86-64.so.2
$ file fuga-musl
ELF 64-bit LSB pie executable, x86-64, static-pie linked, BuildID[sha1]=7404a64bb002229779128510d5825ca2ceb8e9c7
$ ldd fuga-musl
	statically linked
[codigo de salida: 0]
$ ls -l
2360512 fuga-gnu
2446136 fuga-musl
$ chroot vacio /fuga-gnu --version
chroot: failed to run command '/fuga-gnu': No such file or directory
[codigo de salida: 127]
$ chroot vacio /fuga-musl --version
fuga 1.0.0
[codigo de salida: 0]
$ chroot vacio /fuga-musl escanear /config.env
/config.env:2  aws-access-key  AKIA****************
/config.env:4  github-token  ghp_************************************
2 hallazgo(s) en 1 archivo(s)
[codigo de salida: 1]
```
