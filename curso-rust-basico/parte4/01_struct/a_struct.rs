struct Usuario {
    nombre: String,
    correo: String,
    activo: bool,
    ingresos: u64,
}

fn crear(nombre: String, correo: String) -> Usuario {
    Usuario { nombre, correo, activo: true, ingresos: 1 }
}

fn main() {
    let mut ana = crear(String::from("Ana"), String::from("ana@ejemplo.com"));
    ana.ingresos += 1;
    println!("{} <{}> activo={} ingresos={}", ana.nombre, ana.correo, ana.activo, ana.ingresos);
}
