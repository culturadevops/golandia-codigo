# La imagen de despliegue de Go: el binario estático y NADA más.
FROM scratch
COPY _bin/servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
