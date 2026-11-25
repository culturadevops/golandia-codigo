#!/bin/sh
# Claves de host propias de este contenedor, la clave pública del cliente y el «rol» del servidor.
set -e
ssh-keygen -A >/dev/null
printf '%s\n' "$CLAVE_PUB" > /home/admin/.ssh/authorized_keys
chmod 600 /home/admin/.ssh/authorized_keys
chown admin:admin /home/admin/.ssh/authorized_keys
printf '%s\n' "${ROL:-normal}" > /etc/rol
exec /usr/sbin/sshd -D -e
