FROM golang:1.27.1-trixie
WORKDIR /lab
COPY etapas/ /lab/etapas/
COPY dentro.sh /lab/dentro.sh
RUN chmod +x /lab/dentro.sh
CMD ["/lab/dentro.sh"]
