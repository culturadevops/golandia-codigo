#[derive(Debug)]
struct Rectangulo {
    ancho: u32,
    alto: u32,
}

impl Rectangulo {
    fn new(ancho: u32, alto: u32) -> Self {
        Self { ancho, alto }
    }

    fn cuadrado(lado: u32) -> Self {
        Self { ancho: lado, alto: lado }
    }

    fn area(&self) -> u32 {
        self.ancho * self.alto
    }

    fn cabe(&self, otro: &Rectangulo) -> bool {
        self.ancho >= otro.ancho && self.alto >= otro.alto
    }

    fn escalar(&mut self, factor: u32) {
        self.ancho *= factor;
        self.alto *= factor;
    }
}

fn main() {
    let mut r = Rectangulo::new(30, 50);
    let c = Rectangulo::cuadrado(40);
    println!("área de r: {}", r.area());
    println!("¿c cabe en r? {}", r.cabe(&c));
    r.escalar(2);
    println!("r escalado: {r:?}, área {}", r.area());
    println!("¿c cabe ahora? {}", r.cabe(&c));
}
