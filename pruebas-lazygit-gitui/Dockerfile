# gl-grb-git:1 — lazygit (Go) y gitui (Rust), binarios OFICIALES de cada release, más git y pyte (un emulador de
# terminal en Python) para ver la pantalla que pinta cada uno.
# lazygit se verifica con su checksums.txt; gitui no publica sumas: se anota la del tar descargado.
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl git python3 python3-pyte time procps \
    ncurses-base && rm -rf /var/lib/apt/lists/*
ARG LAZYGIT=0.66.3
ARG GITUI=0.28.1
WORKDIR /tmp/d
RUN curl -fsSLO https://github.com/jesseduffield/lazygit/releases/download/v${LAZYGIT}/lazygit_${LAZYGIT}_linux_x86_64.tar.gz \
 && curl -fsSLO https://github.com/jesseduffield/lazygit/releases/download/v${LAZYGIT}/checksums.txt \
 && grep " lazygit_${LAZYGIT}_linux_x86_64.tar.gz$" checksums.txt | sha256sum -c - \
 && tar xzf lazygit_${LAZYGIT}_linux_x86_64.tar.gz -C /usr/local/bin lazygit \
 && curl -fsSLO https://github.com/gitui-org/gitui/releases/download/v${GITUI}/gitui-linux-x86_64.tar.gz \
 && sha256sum gitui-linux-x86_64.tar.gz | tee /gitui-tar.sha256 \
 && tar xzf gitui-linux-x86_64.tar.gz -C /usr/local/bin \
 && cd / && rm -rf /tmp/d && lazygit --version && gitui --version && git --version
RUN mkdir -p /root/.config/lazygit && printf 'disableStartupPopups: true\nupdate:\n  method: never\n' > /root/.config/lazygit/config.yml \
 && git config --global user.name banco && git config --global user.email banco@example.com \
 && git config --global init.defaultBranch main && git config --global --add safe.directory '*'
WORKDIR /w
