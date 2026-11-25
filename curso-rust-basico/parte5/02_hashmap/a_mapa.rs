use std::collections::HashMap;

fn main() {
    let mut stock: HashMap<String, u32> = HashMap::new();
    stock.insert(String::from("café"), 10);
    stock.insert(String::from("arepa"), 4);
    stock.insert(String::from("café"), 12);
    println!("café: {:?}", stock.get("café"));
    println!("té:   {:?}", stock.get("té"));

    *stock.entry(String::from("arepa")).or_insert(0) -= 1;
    *stock.entry(String::from("té")).or_insert(0) += 7;

    let mut claves: Vec<_> = stock.iter().collect();
    claves.sort();
    println!("{claves:?}");

    let texto = "el gato y el perro y el loro";
    let mut cuenta: HashMap<&str, u32> = HashMap::new();
    for palabra in texto.split_whitespace() {
        *cuenta.entry(palabra).or_insert(0) += 1;
    }
    let mut lista: Vec<_> = cuenta.into_iter().collect();
    lista.sort_by(|a, b| b.1.cmp(&a.1).then(a.0.cmp(b.0)));
    println!("{lista:?}");
}
