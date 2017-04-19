FROM alpine:3.4

COPY ./documents/template.config /etc/ingestor/
COPY ingestor /usr/local/bin/ingestor

EXPOSE 7780

ENTRYPOINT ["./ingestor"]
CMD ["--config", "/etc/ingestor/template.config", "-logtostderr=true", "-v=2"]
