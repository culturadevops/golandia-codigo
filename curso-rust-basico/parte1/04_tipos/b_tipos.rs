fn main() {
    let edad: u32 = "42".parse().unwrap();
    let temperatura: i32 = -7;
    let precio: f64 = 19.99;
    let activo: bool = true;
    let cangrejo: char = '🦀';
    println!("{edad} {temperatura} {precio} {activo} {cangrejo}");
    println!("u8:  {} a {}", u8::MIN, u8::MAX);
    println!("i32: {} a {}", i32::MIN, i32::MAX);
    println!("u64: {} a {}", u64::MIN, u64::MAX);
    println!("0.1 + 0.2 = {}", 0.1 + 0.2);
}
