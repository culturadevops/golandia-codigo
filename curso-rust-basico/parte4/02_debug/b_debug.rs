#[derive(Debug)]
struct Punto {
    x: i32,
    y: i32,
}

fn main() {
    let p = Punto { x: 3, y: 4 };
    println!("{p:?}");
    println!("{p:#?}");
    let area = dbg!(p.x * p.y);
    println!("area = {area}");
}
