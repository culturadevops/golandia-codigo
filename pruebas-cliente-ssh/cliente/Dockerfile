# Máquina de control del vídeo 682: aquí corren el binario de Go, el bucle de Bash con ssh y Ansible. Versiones fijas.
FROM python:3.13.16-slim-trixie
RUN apt-get update \
 && apt-get install -y --no-install-recommends openssh-client=1:10.0p1-7+deb13u4 \
 && rm -rf /var/lib/apt/lists/* \
 && pip install --no-cache-dir ansible-core==2.21.5
CMD ["sleep", "infinity"]
