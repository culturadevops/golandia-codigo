source /work/repo/comun.sh
ver a_nivel.rs; proyecto nivel a_nivel.rs; correr nivel; correr_release nivel
ver b_explicito.rs; proyecto nivel b_explicito.rs; correr nivel
proyecto nivel a_nivel.rs
printf '\n[profile.release]\noverflow-checks = true\n' >> /tmp/p/nivel/Cargo.toml
echo '$ tail -2 Cargo.toml'; tail -2 /tmp/p/nivel/Cargo.toml
correr_release nivel
