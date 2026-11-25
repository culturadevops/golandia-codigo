fn main() {
    // loop devuelve un valor con break
    let mut intentos = 0;
    let clave = loop {
        intentos += 1;
        if intentos == 3 {
            break intentos * 100;
        }
    };
    println!("clave {clave} tras {intentos} intentos");

    // while
    let mut cuenta = 3;
    while cuenta > 0 {
        println!("{cuenta}...");
        cuenta -= 1;
    }
    println!("¡despegue!");

    // for sobre rangos
    for i in 1..4 {
        print!("{i} ");
    }
    println!("<- 1..4");
    for i in 1..=4 {
        print!("{i} ");
    }
    println!("<- 1..=4");
    for i in (1..=3).rev() {
        print!("{i} ");
    }
    println!("<- rev");

    // for sobre un array
    let frutas = ["mango", "guayaba", "parchita"];
    for fruta in frutas {
        println!("fruta: {fruta}");
    }
}
