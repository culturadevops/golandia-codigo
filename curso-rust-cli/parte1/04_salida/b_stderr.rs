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
