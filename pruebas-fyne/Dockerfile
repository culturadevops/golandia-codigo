FROM golang:1.26-bookworm
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1-mesa-dev xorg-dev xvfb imagemagick x11-apps \
    libwayland-dev wayland-protocols libxkbcommon-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
