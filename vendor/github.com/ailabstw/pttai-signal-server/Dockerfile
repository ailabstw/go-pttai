# Build pttai-signal-server in a stock Go builder container
FROM golang:1.12.4-alpine3.9 as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

RUN mkdir -p /src
ADD . /src/pttai-signal-server
RUN cd /src/pttai-signal-server/cmd/pttai-signal-server && go build .

# Pull pttai-signal-server into a second stage deploy alpine container
FROM alpine:3.9

COPY --from=builder /src/pttai-signal-server/cmd/pttai-signal-server/pttai-signal-server /usr/local/bin/

EXPOSE 8080
