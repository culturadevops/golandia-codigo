# Una «Raspberry Pi» de mentira: Debian 13 ARM64 bajo QEMU, con systemd como proceso 1 y el servicio del sensor instalado.
# docker build --platform linux/arm64 -t gl-prac-rpi-systemd:1 -f systemd/Dockerfile .
FROM debian:trixie-slim
RUN apt-get update -qq && apt-get install -y -qq --no-install-recommends systemd systemd-sysv curl ca-certificates \
 && rm -rf /var/lib/apt/lists/*
COPY _bin/sensor-arm64 /usr/local/bin/sensor
COPY systemd/sensor.service systemd/sensor-simple.service /etc/systemd/system/
RUN systemctl enable sensor.service
STOPSIGNAL SIGRTMIN+3
CMD ["/lib/systemd/systemd"]
