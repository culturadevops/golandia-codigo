$ go version
go1.26.5
$ ollama --version
ollama version is 0.40.5
$ nvidia-smi --query-gpu=name,memory.total --format=csv,noheader
NVIDIA GeForce RTX 5070 Ti, 16303 MiB
$ ls -l medir-llm | awk "{print \$5}"   # el binario, en bytes
9012224

$ ./medir-llm -rondas 5 qwen2.5-coder:1.5b qwen2.5-coder:3b qwen2.5-coder:7b
prompt: "Explica en tres frases qué es un contenedor de Docker." · máx. 128 tokens · temperatura 0 · GPU
modelo                 frío: carga frío: TTFT TTFT caliente  tokens/s
qwen2.5-coder:1.5b          1.72 s     1.79 s          9 ms     416.6
qwen2.5-coder:3b            2.34 s     2.40 s         13 ms     283.7
qwen2.5-coder:7b            3.87 s     3.97 s         16 ms     154.4

$ ./medir-llm -rondas 5 -cpu qwen2.5-coder:1.5b qwen2.5-coder:3b qwen2.5-coder:7b
prompt: "Explica en tres frases qué es un contenedor de Docker." · máx. 128 tokens · temperatura 0 · SOLO CPU
modelo                 frío: carga frío: TTFT TTFT caliente  tokens/s
qwen2.5-coder:1.5b          1.80 s     1.89 s         26 ms      54.4
qwen2.5-coder:3b            2.57 s     2.73 s         45 ms      28.0
qwen2.5-coder:7b            5.34 s     5.69 s         87 ms      12.8

$ # el mismo modelo con el cliente de Python (vídeo 472), para comparar clientes
$ python medir_llm.py qwen2.5-coder:7b --rondas 5
prompt: 'Explica en tres frases qué es un contenedor de Docker.' · máx. 128 tokens · temperatura 0 · GPU
modelo                 frío: carga frío: TTFT TTFT caliente  tokens/s
qwen2.5-coder:7b            3.72 s     3.80 s         33 ms     154.5
$ ./medir-llm -rondas 5 qwen2.5-coder:7b
prompt: "Explica en tres frases qué es un contenedor de Docker." · máx. 128 tokens · temperatura 0 · GPU
modelo                 frío: carga frío: TTFT TTFT caliente  tokens/s
qwen2.5-coder:7b            2.10 s     2.17 s         18 ms     154.4
