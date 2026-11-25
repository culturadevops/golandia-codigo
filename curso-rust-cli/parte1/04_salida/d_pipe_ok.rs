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
