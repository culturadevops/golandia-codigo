fn main() {
    let s1 = String::from("hola");
    let s2 = s1.clone();
    println!("s1 = {s1}, s2 = {s2}");
    println!("s1 en {:p}", s1.as_ptr());
    println!("s2 en {:p}", s2.as_ptr());
}
