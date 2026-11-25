source /work/repo/comun.sh
ver main.rs; proyecto palabras main.rs
cd /tmp/p/palabras
printf 'El gato y el perro.\nEl LORO, el gato... ¡y la Ñandú!\n' > corto.txt
echo '$ cat corto.txt'; cat corto.txt
correr palabras corto.txt 3
correr palabras
correr palabras corto.txt tres
correr palabras no-existe.txt
# un texto grande y siempre igual: 5 millones de palabras con semilla fija
awk 'BEGIN{srand(42); split("el la de que y en un ser se no haber por con su para como estar tener le lo todo pero más hacer o poder decir este ir otro ese si me ya ver porque dar cuando él muy sin vez mucho saber qué sobre mi alguno mismo yo también hasta año dos querer entre así primero desde grande eso ni nos llegar pasar tiempo ella sí día uno bien poco deber entonces poner cosa tanto hombre parecer nuestro tan donde ahora parte después vida quedar siempre creer hablar llevar dejar nada cada seguir menos nuevo encontrar", w, " "); n=length(w); for(i=0;i<5000000;i++){ printf "%s%s", w[int(rand()^2*n)+1], (i%12==11?"\n":" ") } }' > grande.txt
echo '$ ls -l grande.txt'; ls -l grande.txt | awk '{print $5, $9}'
cargo build -q 2>&1; cargo build -q --release 2>&1
TIMEFORMAT='[tardó %R s]'
echo '$ time ./target/debug/palabras grande.txt'
time ./target/debug/palabras grande.txt
echo '$ time ./target/release/palabras grande.txt'
time ./target/release/palabras grande.txt
echo '$ ls -l target/debug/palabras target/release/palabras'
ls -l target/debug/palabras target/release/palabras | awk '{print $5, $9}'
