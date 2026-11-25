source /work/repo/comun.sh
ver a_q.rs; proyecto puerto a_q.rs
cd /tmp/p/puerto
echo 'puerto=8080' > bueno.txt
echo 'puerto=80800' > grande.txt
echo 'puerto=ochenta' > letras.txt
echo '$ cat bueno.txt grande.txt letras.txt'; cat bueno.txt grande.txt letras.txt
correr puerto bueno.txt
correr puerto grande.txt
correr puerto letras.txt
correr puerto no-existe.txt
