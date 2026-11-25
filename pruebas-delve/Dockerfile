FROM golang:1.27.1-trixie
WORKDIR /lab
COPY proyecto/ /lab/proyecto/
COPY dentro.sh /lab/dentro.sh
RUN chmod +x /lab/dentro.sh
CMD ["/lab/dentro.sh"]
