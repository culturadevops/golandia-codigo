fn main() {
    let mut s = String::from("hola");
    let r1 = &s;
    let r2 = &s;
    println!("{r1} y {r2}");
    // r1 y r2 ya no se usan más abajo: el préstamo terminó aquí
    let r3 = &mut s;
    r3.push('!');
    println!("{r3}");
}
