# Kucoin Level3 Market

## Guide
  [中文文档](README_CN.md)

## Installation

1. build

```
CGO_ENABLED=0 go build -ldflags '-s -w' -o kucoin_market cmd/main/market.go
```

or you can download the latest available [release](https://github.com/Kucoin/kucoin-level3-sdk/releases)

## Usage

1. [vim config.yaml](config.example.yaml):
   ```
    app_debug: true
    
    symbol: KCS-USDT
    #symbol: XBTUSDM
    
    app:
      name: market
      log_file: "./runtime/log/market.log"
    
    api_server:
      network: tcp
      address: 0.0.0.0:9090
      token: your-rpc-token
    
    market.kucoin_v2:
      url: "https://api.kucoin.com"
      type: "spot"
      # url: "https://api-futures.kucoin.com"
      # type: "future"
   
    redis:
      addr: 127.0.0.1:6379
      password: ""
      db: 0
    ```

1. Run Command：

    ```
    ./kucoin_market start -c config.yaml
    ```

## Docker Usage

1. Build docker image

   ```
   docker build -t kucoin_market .
   ```

1. [vim config.yaml](config.example.yaml):
    ```
    app_debug: true
    
    symbol: KCS-USDT
    #symbol: XBTUSDM
    
    app:
      name: market
      log_file: "./runtime/log/market.log"
    
    api_server:
      network: tcp
      address: 0.0.0.0:9090
      token: your-rpc-token
    
    market.kucoin_v2:
      url: "https://api.kucoin.com"
      type: "spot"
      # url: "https://api-futures.kucoin.com"
      # type: "future"

    redis:
      addr: 127.0.0.1:6379
      password: ""
      db: 0
    ```

1. Run

  ```
  docker run --rm -it -v $(pwd)/config.yaml:/app/config.yaml --net=host kucoin_market
  ```

## RPC Method

> default endpoint : 127.0.0.1:9090
> the sdk rpc is based on golang jsonrpc 1.0 over tcp.

see:[python jsonrpc client demo](./demo/python-demo/level3/rpc.py)

* Get Part Order Book
    ```
    {"method": "Server.GetOrderBook", "params": [{"token": "your-rpc-token", "number": 1}], "id": 0}
    ```

* Add Event ClientOids To Channels
    ```
    {"method": "Server.AddEventClientOidsToChannels", "params": [{"token": "your-rpc-token", "data": {"clientOid": ["channel-1", "channel-2"]}}], "id": 0}
    ```

## Python-Demo

> the demo including orderbook display

see:[python use_level3 demo](./demo/python-demo/order_book_demo.py)
- Run order_book.py
    ```
    command: python3 order_book_demo.py
    describe: display orderbook
    ```
