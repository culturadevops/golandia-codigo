# Servidor de laboratorio del vídeo 682 (golandia): sshd en alpine, versiones fijas.
# Cada contenedor genera SUS claves de host al arrancar (entrada.sh): así cada servidor tiene una huella distinta,
# como en la vida real, y known_hosts tiene algo que comprobar.
FROM alpine:3.22
RUN apk add --no-cache openssh-server=10.0_p1-r10 \
 && adduser -D -s /bin/sh admin \
 && passwd -u admin \
 && mkdir -p /home/admin/.ssh && chmod 700 /home/admin/.ssh && chown admin:admin /home/admin/.ssh \
 && printf 'PasswordAuthentication no\nKbdInteractiveAuthentication no\nPermitRootLogin no\nPubkeyAuthentication yes\n' > /etc/ssh/sshd_config.d/lab.conf
COPY revisar /usr/local/bin/revisar
COPY entrada.sh /entrada.sh
RUN chmod 755 /usr/local/bin/revisar /entrada.sh
EXPOSE 22
ENTRYPOINT ["/entrada.sh"]
