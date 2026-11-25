fn main() {
    let n: u8 = 7;
    let texto = match n {
        0 => "cero",
        1..=9 => "un dígito",
    };
    println!("{texto}");
}
