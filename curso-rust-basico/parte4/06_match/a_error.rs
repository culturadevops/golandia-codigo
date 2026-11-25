enum Estado {
    Pendiente,
    Enviado,
    Entregado,
    Devuelto,
}

fn describir(e: &Estado) -> &'static str {
    match e {
        Estado::Pendiente => "esperando pago",
        Estado::Enviado => "en camino",
        Estado::Entregado => "listo",
    }
}

fn main() {
    println!("{}", describir(&Estado::Devuelto));
}
