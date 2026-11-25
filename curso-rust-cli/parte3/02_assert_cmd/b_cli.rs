use assert_cmd::Command;
use predicates::str::contains;

fn fuga() -> Command {
    let mut c = Command::cargo_bin("fuga").unwrap();
    c.env_remove("FUGA_MAX").env_remove("NO_COLOR"); // el test no hereda tu entorno
    c
}

#[test]
fn con_secretos_sale_con_1() {
    fuga().args(["escanear", "../../datos/config.env"])
        .assert().code(1).stdout(contains("aws-access-key"));
}

#[test]
fn limpio_sale_con_0() {
    fuga().args(["escanear", "../../datos/README.md"]).assert().success();
}

#[test]
fn archivo_que_no_existe_sale_con_3() {
    fuga().args(["escanear", "no-existe.env"])
        .assert().code(3).stderr(contains("no pude leer no-existe.env"));
}

#[test]
fn lee_la_entrada_estandar() {
    fuga().args(["escanear", "-"]).write_stdin("password: hunter2\n")
        .assert().code(1).stdout(contains("contrasena"));
}

#[test]
fn muestra_los_dos_hallazgos_de_config_env() {
    fuga().args(["escanear", "../../datos/config.env"])
        .assert().code(1).stdout(contains("github-token"));
}
