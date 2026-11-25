# Compilar y capturar Wails en Linux: Go 1.26 (Wails 2.16 pide >= 1.25), el WebKit del sistema y
# GTK (lo que Wails usa como ventana en Linux), Xvfb (una pantalla sin pantalla) e ImageMagick
# para fotografiarla. Las capturas del curso son ventanas de verdad, hechas aquí.
FROM golang:1.26-bookworm
RUN apt-get update -qq && apt-get install -y -qq --no-install-recommends \
      libgtk-3-dev libwebkit2gtk-4.1-dev xvfb imagemagick x11-utils procps xdotool \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
