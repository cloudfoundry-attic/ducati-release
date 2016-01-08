FROM gliderlabs/alpine

RUN apk --update add \
  ca-certificates \
  jq \
  iproute2 \
  go \
  alpine-sdk \
  && rm -rf /var/cache/apk/*

ENV GOROOT /usr/lib/go
ENV GOPATH /gopath
ENV GOBIN /gopath/bin
ENV PATH $PATH:$GOROOT/bin:$GOPATH/bin

RUN go get github.com/tools/godep
RUN wget -O consul.zip https://releases.hashicorp.com/consul/0.6.1/consul_0.6.1_linux_amd64.zip && unzip -d /usr/local/bin consul.zip && rm consul.zip

ADD src/github.com/docker/libnetwork /gopath/src/github.com/docker/libnetwork
RUN cd /gopath/src/github.com/docker/libnetwork/cmd/dnet && godep go install .

ADD src/github.com/onsi /gopath/src/github.com/onsi
RUN go install github.com/onsi/ginkgo/ginkgo
