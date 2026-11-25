fn clasificar(n: i32) -> &'static str {
    match n {
        i32::MIN..=-1 => "negativo",
        0 => "cero",
        1 | 2 | 3 => "pequeño",
        x if x % 2 == 0 => "par grande",
        _ => "impar grande",
    }
}

fn main() {
    for n in [-4, 0, 2, 10, 15] {
        println!("{n}: {}", clasificar(n));
    }
}
