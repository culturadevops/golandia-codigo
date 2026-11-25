FROM golang:1.26-bookworm
RUN apt-get update -qq && apt-get install -y -qq gcc binutils >/dev/null && rm -rf /var/lib/apt/lists/*
WORKDIR /lab
