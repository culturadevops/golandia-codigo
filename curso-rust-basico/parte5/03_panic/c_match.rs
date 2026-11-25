use std::fs;
use std::process;

fn main() {
    let texto = match fs::read_to_string("config.txt") {
        Ok(t) => t,
        Err(e) => {
            eprintln!("error: config.txt: {e}");
            process::exit(1);
        }
    };
    println!("{texto}");
}
