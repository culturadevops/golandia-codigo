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
