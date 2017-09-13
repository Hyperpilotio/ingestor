# Build Stage
FROM lacion/docker-alpine:gobuildimage AS build-stage

LABEL app="build-ingestor"
LABEL REPO="https://github.com/hyperpilotio/ingestor"

ENV GOROOT=/usr/lib/go \
    GOPATH=/gopath \
    GOBIN=/gopath/bin \
    PROJPATH=/gopath/src/github.com/hyperpilotio/ingestor

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /gopath/src/github.com/hyperpilotio/ingestor
WORKDIR /gopath/src/github.com/hyperpilotio/ingestor

RUN make build-alpine

# Final Stage
FROM lacion/docker-alpine:latest

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/hyperpilotio/ingestor"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/ingestor/bin

WORKDIR /opt/ingestor/bin

COPY ./documents/template.config /etc/ingestor/

EXPOSE 7780

COPY --from=build-stage /gopath/src/github.com/hyperpilotio/ingestor/bin/ingestor /opt/ingestor/bin/
RUN chmod +x /opt/ingestor/bin/ingestor

CMD ["/opt/ingestor/bin/ingestor", "--config", "/etc/ingestor/template.config", "-logtostderr=true", "-v=2"]