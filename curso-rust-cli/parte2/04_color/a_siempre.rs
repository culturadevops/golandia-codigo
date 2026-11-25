fn main() {
    for (archivo, tipo) in [("datos/config.env:2", "aws-access-key"), ("datos/app.yaml:4", "contrasena")] {
        println!("{archivo}  \x1b[31m{tipo}\x1b[0m");
    }
}
