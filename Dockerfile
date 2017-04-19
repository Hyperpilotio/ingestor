FROM alpine:3.4

COPY ./documents/dev.config /etc/ingestor/
COPY ingestor /usr/local/bin/ingestor

EXPOSE 7780

ENTRYPOINT ["./ingestor"]
CMD ["--config", "/etc/ingestor/dev.config", "-logtostderr=true", "-v=2"]
