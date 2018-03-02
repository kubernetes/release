FROM golang

LABEL "source.repo"="github/kubernetes/release/build/godep-builder"

ENV GOPATH /gopath/
ENV PATH $GOPATH/bin:$PATH

RUN go version
RUN go get github.com/tools/godep
RUN godep version
RUN apt-get update && \
    apt-get install -y build-essential

ENTRYPOINT ["/gopath/bin/godep"]
