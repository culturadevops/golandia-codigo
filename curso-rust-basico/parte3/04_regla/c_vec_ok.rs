fn main() {
    let mut cola = vec![10, 20, 30];
    let _vecino = vec![0u8; 64];
    println!("capacidad antes: {}, datos en {:p}", cola.capacity(), cola.as_ptr());
    let primero = cola[0];
    cola.push(40);
    println!("capacidad después: {}, datos en {:p}", cola.capacity(), cola.as_ptr());
    println!("el primero era {primero}");
}
