fn main() {
    let precio: i32 = 100;
    let descuento: Option<i32> = Some(15);
    let total = precio - descuento;
    println!("{total}");
}
