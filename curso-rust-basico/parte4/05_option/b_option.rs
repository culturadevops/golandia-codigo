fn buscar_precio(producto: &str) -> Option<u32> {
    match producto {
        "café" => Some(250),
        "arepa" => Some(400),
        _ => None,
    }
}

fn main() {
    for p in ["café", "arepa", "sushi"] {
        match buscar_precio(p) {
            Some(precio) => println!("{p}: {precio}"),
            None => println!("{p}: no está en el menú"),
        }
    }
    if let Some(precio) = buscar_precio("café") {
        println!("con if let: {precio}");
    }
    println!("unwrap_or: {}", buscar_precio("sushi").unwrap_or(0));
    let doble = buscar_precio("arepa").map(|p| p * 2);
    println!("map: {doble:?}");
    println!("unwrap sobre None:");
    let precio = buscar_precio("sushi").unwrap();
    println!("{precio}");
}
