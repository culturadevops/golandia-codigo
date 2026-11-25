fn main() {
    let ave = String::from("ñandú");
    println!("len():           {}", ave.len());
    println!("chars().count(): {}", ave.chars().count());
    println!("bytes: {:?}", ave.as_bytes());
}
