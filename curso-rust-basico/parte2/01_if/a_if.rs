fn main() {
    let temperatura = 31;
    let estado = if temperatura > 30 { "calor" } else { "normal" };
    println!("hoy: {estado}");

    let nota = 72;
    if nota >= 90 {
        println!("excelente");
    } else if nota >= 70 {
        println!("aprobado");
    } else {
        println!("a repasar");
    }
}
