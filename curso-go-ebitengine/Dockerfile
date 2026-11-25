# Ebitengine en Linux, sin pantalla: Go 1.26, lo que pide GLFW (X11 + OpenGL), Mesa para dibujar
# por software, Xvfb como pantalla virtual e ImageMagick para fotografiarla.
FROM golang:1.26-bookworm
RUN apt-get update -qq && apt-get install -y -qq --no-install-recommends \
      libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev \
      libxxf86vm-dev libasound2-dev pkg-config libgl1-mesa-dri xvfb imagemagick procps xdotool \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
