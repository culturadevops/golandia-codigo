fn main() {
    let ave = String::from("ñandú");
    let primera = ave.chars().next().unwrap();
    let ultima = ave.chars().last().unwrap();
    println!("primera: {primera}, última: {ultima}");
    for (i, c) in ave.char_indices() {
        print!("{c}@{i} ");
    }
    println!();
}
