FROM scratch
COPY _bin/servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
