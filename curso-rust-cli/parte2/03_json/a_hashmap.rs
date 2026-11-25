use serde::Serialize;
use std::collections::HashMap;

#[derive(Serialize)]
struct Resumen {
    por_tipo: HashMap<&'static str, usize>,
}

fn main() {
    let mut por_tipo = HashMap::new();
    for t in ["aws-access-key", "github-token", "contrasena", "clave-privada"] {
        por_tipo.insert(t, 1);
    }
    println!("{}", serde_json::to_string(&Resumen { por_tipo }).unwrap());
}
