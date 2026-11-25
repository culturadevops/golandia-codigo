struct Recurso(&'static str);

impl Drop for Recurso {
    fn drop(&mut self) {
        println!("  liberado: {}", self.0);
    }
}

fn usar(r: Recurso) {
    println!("  usar() recibe {}", r.0);
}

fn main() {
    println!("empieza main");
    let a = Recurso("a");
    {
        let _b = Recurso("b");
        println!("  fin del bloque interno");
    }
    usar(a);
    println!("después de usar(a)");
    let _c = Recurso("c");
    println!("termina main");
}
