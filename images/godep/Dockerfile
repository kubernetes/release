FROM golang:1.12

LABEL "source.repo"="github/kubernetes/release/build/godep-builder"

ENV GOPATH /gopath/
ENV PATH $GOPATH/bin:$PATH

RUN go version && \
    go get github.com/tools/godep && \
    godep version && \
    apt-get update && \
    apt-get install -y build-essential

ENTRYPOINT ["/gopath/bin/godep"]
