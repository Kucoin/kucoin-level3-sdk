FROM golang:1.13-stretch as builder

RUN export GO111MODULE=on \
    && export GOPROXY=https://goproxy.io \
    && mkdir -p /go/src/github.com/Kucoin/kucoin-level3-sdk

COPY . /go/src/github.com/Kucoin/kucoin-level3-sdk

RUN cd /go/src/github.com/Kucoin/kucoin-level3-sdk \
    && CGO_ENABLED=0 go build -ldflags '-s -w' -o /go/bin/kucoin_market cmd/main/market.go

FROM debian:stretch

RUN apt-get update \
    && apt-get install ca-certificates -y

COPY --from=builder /go/bin/kucoin_market /usr/local/bin/

# .env => /app/.env
WORKDIR /app
VOLUME /app

EXPOSE 9090

CMD ["kucoin_market", "start", "-c", "config.yaml"]
