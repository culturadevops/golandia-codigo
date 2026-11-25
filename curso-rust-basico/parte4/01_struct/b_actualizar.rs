struct Usuario {
    nombre: String,
    correo: String,
    activo: bool,
}

fn main() {
    let ana = Usuario {
        nombre: String::from("Ana"),
        correo: String::from("ana@ejemplo.com"),
        activo: true,
    };
    let beto = Usuario {
        correo: String::from("beto@ejemplo.com"),
        ..ana
    };
    println!("{} {}", beto.nombre, beto.activo);
    println!("{}", ana.nombre);
}
