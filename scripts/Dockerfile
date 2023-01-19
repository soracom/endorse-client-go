FROM golang:1.19.5

RUN apt update -y && \
    apt install -y pcscd libpcsclite1 libpcsclite-dev gosu
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
