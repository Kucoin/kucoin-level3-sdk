# Kucoin Level3 Market

Kucoin Level3 Market 支持现货和合约的level3消息。

## 文档
  [English Document](README.md)

## 安装

1. 编译

```
CGO_ENABLED=0 go build -ldflags '-s -w' -o kucoin_market cmd/main/market.go
```

或者直接下载已经编译完成的[二进制文件](https://github.com/Kucoin/kucoin-level3-sdk/releases)

## 用法

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

1. 运行命令：

    ```
    ./kucoin_market start -c config.yaml
    ```

## Docker 用法

1. 编译镜像

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
   
1. 运行

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

> python的demo包含了一个本地orderbook的展示

see:[python use_level3 demo](./demo/python-demo/order_book_demo.py)
- Run order_book.py
    ```
    command: python3 order_book_demo.py
    describe: display orderbook
    ```
