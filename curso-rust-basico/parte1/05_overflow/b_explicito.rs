fn main() {
    let nivel: u8 = 255;
    println!("wrapping_add:   {}", nivel.wrapping_add(1));
    println!("checked_add:    {:?}", nivel.checked_add(1));
    println!("saturating_add: {}", nivel.saturating_add(1));
    println!("checked_add(0): {:?}", nivel.checked_add(0));
}
