package sdk

import (
	"net/http"
	"time"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/sdk/http_client"
)

const (
	urlSpotSnapshot = "/api/v3/market/orderbook/level3"

	urlFuturesSnapshot = "/api/v2/level3/snapshot"

	topicSpotL3Prefix = "/spotMarket/level3:"

	topicFutureL3Prefix = "/contractMarket/level3v2:"
)

type Kucoin struct {
	httpClient *http_client.Client

	typ string
}

func NewKucoin(baseUrl, typ, apiKey, apiSecret, apiPassphrase string, skipVerifyTls bool, timeout time.Duration) *Kucoin {
	client := http_client.NewClient(baseUrl, apiKey, apiSecret, apiPassphrase, skipVerifyTls, timeout)
	kucoin := &Kucoin{
		httpClient: client,

		typ: typ,
	}
	return kucoin
}

func L3TopicPrefix(typ string) string {
	var ret string
	switch typ {
	case "spot":
		ret = topicSpotL3Prefix
	case "future":
		ret = topicFutureL3Prefix
	default:
		log.Panic("market type error, must be spot or future")
	}

	return ret
}

func (kucoin *Kucoin) AtomicFullOrderBook(symbol string) (*http_client.Response, error) {
	var url string
	switch kucoin.typ {
	case "spot":
		url = urlSpotSnapshot
	case "future":
		url = urlFuturesSnapshot
	default:
		log.Panic("market type error, must be spot or future")
	}

	return kucoin.httpClient.Request(http.MethodGet, url, map[string]string{"symbol": symbol})
}
