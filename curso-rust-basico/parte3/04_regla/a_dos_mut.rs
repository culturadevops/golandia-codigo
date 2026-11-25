fn main() {
    let mut s = String::from("hola");
    let r1 = &mut s;
    let r2 = &mut s;
    r1.push('!');
    r2.push('?');
    println!("{s}");
}
