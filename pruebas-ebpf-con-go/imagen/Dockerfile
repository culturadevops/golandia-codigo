# Go 1.27.2 + lo que necesita bpf2go para compilar el programa eBPF en C.
FROM golang:1.27.2
RUN apt-get update && apt-get install -y --no-install-recommends clang llvm libbpf-dev curl iproute2 && rm -rf /var/lib/apt/lists/*
