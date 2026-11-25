// Handler de prueba: EXACTAMENTE el mismo trabajo que en Go, Python y Node (vídeo 121).
// Rust con lambda_runtime: un binario `bootstrap` estático (musl) para provided.al2023.
use lambda_runtime::{service_fn, Error, LambdaEvent};
use serde::{Deserialize, Serialize};
use std::time::Duration;

#[derive(Deserialize)]
struct Peticion {
    #[serde(default)]
    n: u64,
    #[serde(default)]
    io_ms: u64,
}

#[derive(Serialize)]
struct Respuesta {
    lang: &'static str,
    n: u64,
    suma: u64,
    io_ms: u64,
}

async fn manejar(ev: LambdaEvent<Peticion>) -> Result<Respuesta, Error> {
    let p = ev.payload;
    let mut suma: u64 = 0;
    for i in 1..=p.n {
        suma += i * i % 1000003;
    }
    if p.io_ms > 0 {
        tokio::time::sleep(Duration::from_millis(p.io_ms)).await;
    }
    Ok(Respuesta { lang: "rust", n: p.n, suma, io_ms: p.io_ms })
}

#[tokio::main]
async fn main() -> Result<(), Error> {
    lambda_runtime::run(service_fn(manejar)).await
}
