fn sumar(a: i32, b: i32) -> i32 {
    a + b
}

fn saludar(nombre: &str) {
    println!("hola, {nombre}");
}

fn main() {
    saludar("Golandia");
    let total = sumar(2, 3);
    let doble = {
        let t = total * 2;
        t + 0
    };
    println!("total {total}, doble {doble}");
}
