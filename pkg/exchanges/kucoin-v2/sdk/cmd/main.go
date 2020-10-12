package main

import (
	"time"

	kucoin "github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/sdk"
)

func main() {
	apiService := kucoin.NewKucoin("http://openapi-v2.kucoin.net", "spot", "", "", "", true, 30*time.Second)

	tk, err := apiService.WebSocketPublicToken()
	tk, err = apiService.WebSocketPrivateToken()
	if err != nil {
		panic(err)
	}

	c := apiService.NewWebSocketClient(tk)

	mc, ec, err := c.Connect()
	if err != nil {
		panic(err)
	}

	symbol := "KCS-BTC"
	ch := kucoin.NewSubscribeMessage("/spotMarket/level2Depth5:"+symbol, false)
	ch = kucoin.NewSubscribeMessage("/spotMarket/level2Depth50:"+symbol, false)
	//ch = kucoin.NewSubscribeMessage("/spotMarket/level3:"+symbol, false)
	ch = kucoin.NewSubscribeMessage("/spotMarket/tradeOrders", true)

	//ch = kucoin.NewSubscribeMessage("/spotMarket/advancedOrders", true)

	//ch = kucoin.NewSubscribeMessage("/contractMarket/level2depth5:XBTUSDM", false)
	//ch = kucoin.NewSubscribeMessage("/contractMarket/level2depth50:XBTUSDM", false)
	//ch = kucoin.NewSubscribeMessage("/contractMarket/level3v2:XBTUSDM", false)
	//ch = kucoin.NewSubscribeMessage("/contractMarket/tradeOrders", true)

	if err := c.Subscribe(ch); err != nil {
		panic(err)
	}

	for {
		select {
		case err := <-ec:
			c.Stop() // Stop subscribing the WebSocket feed
			panic(err)

		case <-mc:
			//log.Println(base.ToJsonString(msg))
		}
	}
}
