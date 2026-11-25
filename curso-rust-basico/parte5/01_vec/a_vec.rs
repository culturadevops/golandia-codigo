fn main() {
    let mut notas: Vec<u32> = Vec::new();
    notas.push(70);
    notas.push(95);
    notas.push(88);
    let mas = vec![60, 100];
    notas.extend(mas);
    println!("notas: {notas:?}, hay {}", notas.len());

    println!("notas[1] = {}", notas[1]);
    println!("get(1)  = {:?}", notas.get(1));
    println!("get(10) = {:?}", notas.get(10));

    for n in &mut notas {
        *n += 5;
    }
    println!("con 5 puntos extra: {notas:?}");

    let suma: u32 = notas.iter().sum();
    let max = notas.iter().max().unwrap();
    println!("suma {suma}, máxima {max}, promedio {}", suma as f64 / notas.len() as f64);

    let aprobadas: Vec<u32> = notas.iter().filter(|&&n| n >= 80).copied().collect();
    println!("aprobadas: {aprobadas:?}");
    println!("pop: {:?}, quedan {:?}", notas.pop(), notas);
}
