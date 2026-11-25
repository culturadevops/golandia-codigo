fn dividir(a: f64, b: f64) -> Result<f64, String> {
    if b == 0.0 {
        return Err(String::from("división entre cero"));
    }
    Ok(a / b)
}

fn main() {
    for (a, b) in [(10.0, 4.0), (1.0, 0.0)] {
        match dividir(a, b) {
            Ok(r) => println!("{a} / {b} = {r}"),
            Err(e) => println!("{a} / {b}: error, {e}"),
        }
    }
    println!("1.0 / 0.0 en f64 sin comprobar = {}", 1.0_f64 / 0.0);
}
