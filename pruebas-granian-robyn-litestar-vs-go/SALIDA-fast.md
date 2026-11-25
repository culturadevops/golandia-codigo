cpu.max: 200000 100000
os.cpu_count(): 24
robyn --fast -> processes = 49 · workers = 2
os.process_cpu_count(): 24

## robyn --fast en el contenedor de 2 CPU y 1 GiB (2026-11-09, a mano, contenedor «rfast», nada más corriendo)
1ª respuesta: 175 ms
procesos srv_robyn: 50 · LISTO: 47 (a los 10 s) · memoria en reposo: 310.1MiB / 1GiB
/suma?n=200000, hey -c 16, 10 s: 190.9 pet/s · p50 95.2 ms · p99 205.4 ms · [200] 1927 respuestas
memoria tras carga: 312.1MiB · OOMKilled: false
(Con --processes 2, SALIDA.md sección 5: 238 pet/s, 37 MiB bajo carga; sección 3: 34.8 MiB en reposo.)
