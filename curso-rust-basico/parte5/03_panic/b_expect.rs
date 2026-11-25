use std::fs;

fn main() {
    let texto = fs::read_to_string("config.txt")
        .expect("no pude leer config.txt");
    println!("{texto}");
}
