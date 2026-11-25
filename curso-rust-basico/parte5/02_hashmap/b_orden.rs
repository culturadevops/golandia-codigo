use std::collections::HashMap;

fn main() {
    let mut m = HashMap::new();
    for (i, k) in ["uno", "dos", "tres", "cuatro", "cinco"].iter().enumerate() {
        m.insert(*k, i);
    }
    let claves: Vec<_> = m.keys().collect();
    println!("{claves:?}");
}
