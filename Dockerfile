FROM alpine:3.4

COPY ./documents/config.json /etc/ingestor/
COPY ingestor /usr/local/bin/ingestor

EXPOSE 7780

ENTRYPOINT ["ingestor"]
