// El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en Rust 1.99.1 con axum 0.8.9 sobre tokio 1.54.1.
// El mismo contrato: /suma (CPU), /trabajo (objetos, mapa, orden y JSON armado a mano), /espera.
// Para «¿Go o Rust? Lo medí» (golandia 698, 2026-10-29). Código directo, como lo escribiría cualquiera: HashMap con su
// hasher por defecto, String para los nombres, sort_unstable_by (el mismo pdqsort que sort.Slice de Go).
use axum::{extract::Query, http::header, response::IntoResponse, routing::get, Router};
use std::collections::HashMap;
use std::fmt::Write;
use std::time::{Duration, Instant};

type Q = Query<HashMap<String, String>>;

struct Item {
    id: usize,
    nombre: String,
    valor: usize,
}

fn num(q: &Q, k: &str) -> usize {
    q.get(k).and_then(|v| v.parse().ok()).unwrap_or(0)
}

fn json(s: String) -> impl IntoResponse {
    ([(header::CONTENT_TYPE, "application/json")], s)
}

async fn suma(q: Q) -> impl IntoResponse {
    let n = num(&q, "n") as u64;
    let t0 = Instant::now();
    let mut s: u64 = 0;
    for i in 1..=n {
        s += i * i % 1000003;
    }
    json(format!(r#"{{"lang":"rust","n":{n},"suma":{s},"ns":{}}}"#, t0.elapsed().as_nanos()))
}

async fn trabajo(q: Q) -> impl IntoResponse {
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

async fn espera(q: Q) -> impl IntoResponse {
    let ms = num(&q, "ms");
    tokio::time::sleep(Duration::from_millis(ms as u64)).await;
    json(format!(r#"{{"lang":"rust","n":{ms},"suma":0,"ns":0}}"#))
}

#[tokio::main]
async fn main() {
    let arranque = Instant::now();
    let app = Router::new()
        .route("/suma", get(suma))
        .route("/trabajo", get(trabajo))
        .route("/espera", get(espera));
    let escucha = tokio::net::TcpListener::bind("0.0.0.0:8080").await.unwrap();
    let hilos = std::thread::available_parallelism().map(|n| n.get()).unwrap_or(0);
    eprintln!("LISTO_EN_MS {:.3} rust hilos={hilos}", arranque.elapsed().as_secs_f64() * 1000.0);
    axum::serve(escucha, app).await.unwrap();
}
