# Build GPtt in a stock Go builder container
FROM golang:1.11.1-alpine3.8 as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

RUN mkdir -p /go/src/github.com/ailabstw
ADD . /go/src/github.com/ailabstw/go-pttai
RUN cd /go/src/github.com/ailabstw/go-pttai && make gptt

# Pull GPtt into a second stage deploy alpine container
FROM alpine:3.8

RUN apk add --no-cache ca-certificates
RUN mkdir -p /root/.pttai
COPY --from=builder /go/src/github.com/ailabstw/go-pttai/build/bin/gptt /usr/local/bin/
COPY --from=builder /go/src/github.com/ailabstw/go-pttai/static /static

EXPOSE 9487 9487/udp
# ENTRYPOINT ["gptt", "--testp2p", "--httpdir", "/static", "--httpaddr", "0.0.0.0:9774", "--rpcaddr", "0.0.0.0"]
