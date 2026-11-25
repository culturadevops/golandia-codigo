use serde::Deserialize;

#[derive(Deserialize)]
struct Config {
    max: Option<usize>,
}

fn main() {
    let c: Config = toml::from_str("maxx = 3").unwrap();
    println!("sin deny_unknown_fields: max = {:?}", c.max);
}
