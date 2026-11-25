cd /tmp && rm -rf hola
echo '$ cargo new hola'
cargo new hola 2>&1
cd hola
echo '$ find . -type f -not -path "./.git/*"'
find . -type f -not -path './.git/*' | sort
echo '$ cat Cargo.toml'
cat Cargo.toml
echo '$ cat src/main.rs'
cat src/main.rs
echo '$ cargo run'
cargo run 2>&1 | sed -E 's/in [0-9.]+s/in X.XXs/'
echo '$ cargo build --release'
cargo build --release 2>&1 | sed -E 's/in [0-9.]+s/in X.XXs/'
echo '$ ls -l target/debug/hola target/release/hola'
ls -l target/debug/hola target/release/hola | awk '{print $5, $9}'
echo '$ ./target/release/hola'
./target/release/hola
