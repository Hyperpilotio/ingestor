FROM golang

RUN go get github.com/Masterminds/glide
RUN go build github.com/Masterminds/glide

COPY . /go/src/github.com/hyperpilotio/ingestor
WORKDIR /go/src/github.com/hyperpilotio/ingestor
RUN glide update
RUN go build

EXPOSE 7780
ENTRYPOINT ["./ingestor"]
CMD ["--config", "./documents/dev.config", "-logtostderr=true", "-v=2"]
