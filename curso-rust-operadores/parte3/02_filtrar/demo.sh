# Errores por minuto en un log de 200.000 líneas, y lo que cuesta compilar la regex en el bucle.
source /work/repo/comun.sh
tmp
# Un log siempre igual: semilla fija, 30 minutos, 2 % de errores y un pico en las 03:17.
awk 'BEGIN { srand(7); split("INFO INFO INFO INFO WARN", niv, " ");
  for (i = 0; i < 200000; i++) {
    s = int(i * 1800 / 200000); m = int(s / 60);
    p = (m == 17) ? 0.30 : 0.02;
    lvl = (rand() < p) ? "ERROR" : niv[int(rand() * 5) + 1];
    printf "2026-11-13T03:%02d:%02d %s pedido=%d servicio=pagos\n", m, s % 60, lvl, i
  } }' > app.log
c 'wc -l app.log'
c 'head -3 app.log'
c 'ops-filtrar app.log'
c 'ops-filtrar app.log --lento'
c 'grep -c " ERROR " app.log'
