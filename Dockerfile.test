# Build GPtt in a stock Go builder container
FROM golang:1.12.4-alpine3.9

RUN apk add --no-cache make gcc musl-dev linux-headers git openssh bash

RUN mkdir -p /go/src/gitlab.corp.ailabs.tw/ptt.ai
ADD . /go/src/gitlab.corp.ailabs.tw/ptt.ai/go-pttai
RUN mkdir -p /go/src/gitlab.corp.ailabs.tw/ptt.ai/go-pttai/build/_workspace
ENV GOPATH=/go/src/gitlab.corp.ailabs.tw/ptt.ai/go-pttai/build/_workspace
RUN cd /go/src/gitlab.corp.ailabs.tw/ptt.ai/go-pttai && go get github.com/rakyll/gotest && go install github.com/rakyll/gotest && make test
