FROM golang:1.27.1-trixie
RUN apt-get update -qq && apt-get install -y -qq git && rm -rf /var/lib/apt/lists/*
WORKDIR /lab
COPY proyecto/ /lab/proyecto/
COPY dentro.sh /lab/dentro.sh
RUN chmod +x /lab/dentro.sh
CMD ["/lab/dentro.sh"]
