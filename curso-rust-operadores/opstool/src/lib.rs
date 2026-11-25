//! Utilidades compartidas por los comandos de opstool.
use std::ffi::CString;
use std::fs;

/// 1536 -> "1.5 KiB". Potencias de 1024, como `du -h` y `free -h`.
pub fn humano(bytes: u64) -> String {
    let unidades = ["B", "KiB", "MiB", "GiB", "TiB"];
    let mut v = bytes as f64;
    let mut i = 0;
    while v >= 1024.0 && i < unidades.len() - 1 {
        v /= 1024.0;
        i += 1;
    }
    if i == 0 { format!("{bytes} B") } else { format!("{v:.1} {}", unidades[i]) }
}

/// Lo que devuelve statvfs(2) de un punto de montaje, ya en bytes.
pub struct Espacio {
    pub total: u64,
    pub libre: u64,       // f_bfree: libre de verdad, contando lo reservado para root
    pub disponible: u64,  // f_bavail: lo que puede usar un usuario normal
}

pub fn statvfs(ruta: &str) -> std::io::Result<Espacio> {
    let c = CString::new(ruta).unwrap();
    let mut s: libc::statvfs = unsafe { std::mem::zeroed() };
    if unsafe { libc::statvfs(c.as_ptr(), &mut s) } != 0 {
        return Err(std::io::Error::last_os_error());
    }
    let f = s.f_frsize as u64;
    Ok(Espacio { total: s.f_blocks as u64 * f, libre: s.f_bfree as u64 * f, disponible: s.f_bavail as u64 * f })
}

impl Espacio {
    pub fn usado(&self) -> u64 { self.total - self.libre }
    /// El porcentaje de `df`: usado / (usado + disponible), redondeado hacia arriba.
    pub fn uso_pct(&self) -> u64 {
        let base = self.usado() + self.disponible;
        if base == 0 { 0 } else { (self.usado() * 100).div_ceil(base) }
    }
}

/// Un campo de /proc/meminfo, en bytes.
pub fn meminfo(clave: &str) -> Option<u64> {
    let texto = fs::read_to_string("/proc/meminfo").ok()?;
    texto.lines().find_map(|l| {
        let (k, v) = l.split_once(':')?;
        if k != clave { return None; }
        let kb: u64 = v.trim().trim_end_matches(" kB").parse().ok()?;
        Some(kb * 1024)
    })
}

/// Un archivo de cgroup v2 con un número (o "max").
pub fn cgroup(nombre: &str) -> Option<String> {
    fs::read_to_string(format!("/sys/fs/cgroup/{nombre}")).ok().map(|s| s.trim().to_string())
}

pub fn tam_pagina() -> u64 {
    unsafe { libc::sysconf(libc::_SC_PAGESIZE) as u64 }
}
