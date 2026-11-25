// «¿Go o Rust? La revancha» (golandia 958, 2026-11-24): el MISMO contrato y la MISMA lógica del main.rs de axum (698), con
// actix-web 4 en vez de axum. /suma (CPU), /trabajo (objetos, mapa, orden y JSON armado a mano), /espera.
// Código directo: HashMap con su hasher por defecto, String para los nombres, sort_unstable_by.
use actix_web::{get, web, App, HttpResponse, HttpServer};
use std::collections::HashMap;
use std::fmt::Write;
use std::time::{Duration, Instant};

type Q = web::Query<HashMap<String, String>>;

struct Item {
    id: usize,
    nombre: String,
    valor: usize,
}

fn num(q: &Q, k: &str) -> usize {
    q.get(k).and_then(|v| v.parse().ok()).unwrap_or(0)
}

fn json(s: String) -> HttpResponse {
    HttpResponse::Ok().content_type("application/json").body(s)
}

#[get("/suma")]
async fn suma(q: Q) -> HttpResponse {
    let n = num(&q, "n") as u64;
    let t0 = Instant::now();
    let mut s: u64 = 0;
    for i in 1..=n {
        s += i * i % 1000003;
    }
    json(format!(r#"{{"lang":"rust","n":{n},"suma":{s},"ns":{}}}"#, t0.elapsed().as_nanos()))
}

#[get("/trabajo")]
async fn trabajo(q: Q) -> HttpResponse {
    let n = num(&q, "n");
    let t0 = Instant::now();
    let mut m: HashMap<String, Item> = HashMap::with_capacity(n);
    for i in 0..n {
        let nombre = format!("item-{i}");
        m.insert(nombre.clone(), Item { id: i, nombre, valor: (i * 7919) % 1000 });
    }
    let mut lista: Vec<&Item> = Vec::with_capacity(m.len());
    let mut total = 0;
    for it in m.values() {
        lista.push(it);
        total += it.valor;
    }
    lista.sort_unstable_by(|a, b| a.valor.cmp(&b.valor).then(a.id.cmp(&b.id)));
    let mut b = format!(r#"{{"lang":"rust","n":{n},"total":{total},"primeros":["#);
    for (i, it) in lista.iter().take(50).enumerate() {
        if i > 0 {
            b.push(',');
        }
        write!(b, r#"{{"id":{},"nombre":"{}","valor":{}}}"#, it.id, it.nombre, it.valor).unwrap();
    }
    write!(b, r#"],"ns":{}}}"#, t0.elapsed().as_nanos()).unwrap();
    json(b)
}

#[get("/espera")]
async fn espera(q: Q) -> HttpResponse {
    let ms = num(&q, "ms");
    actix_web::rt::time::sleep(Duration::from_millis(ms as u64)).await;
    json(format!(r#"{{"lang":"rust","n":{ms},"suma":0,"ns":0}}"#))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let arranque = Instant::now();
    let hilos = std::thread::available_parallelism().map(|n| n.get()).unwrap_or(0);
    let srv = HttpServer::new(|| App::new().service(suma).service(trabajo).service(espera))
        .bind(("0.0.0.0", 8080))?
        .run();
    eprintln!("LISTO_EN_MS {:.3} rust actix hilos={hilos}", arranque.elapsed().as_secs_f64() * 1000.0);
    srv.await
}
