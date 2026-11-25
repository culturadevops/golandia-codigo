use std::fs;

fn main() {
    let texto = fs::read_to_string("config.txt").unwrap();
    println!("{texto}");
}
