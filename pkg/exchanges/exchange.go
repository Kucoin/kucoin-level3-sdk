package exchanges

import (
	"encoding/json"
	"errors"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"go.uber.org/zap"
)

type Exchange interface {
	GetPartOrderBook(number int) *OrderBook
	AddEventClientOidsToChannels(data map[string][]string) error
	AnyCall(method string, args json.RawMessage) (interface{}, error)
}

type OrderBook struct {
	Asks interface{} `json:"asks"`
	Bids interface{} `json:"bids"`
	Time string      `json:"time"`
	Info interface{} `json:"info,omitempty"`
}

type Level3OrderBook struct {
	Asks [][3]string `json:"asks"`
	Bids [][3]string `json:"bids"`
	Info interface{} `json:"info,omitempty"`
}

type BasicExchange struct {
}

func (be *BasicExchange) GetPartOrderBook(number int) *OrderBook {
	return nil
}

func (be *BasicExchange) AddEventClientOidsToChannels(data map[string][]string) error {
	return errors.New("unsupported rpc method: AddEventClientOidsToChannels")
}

func (be *BasicExchange) AnyCall(method string, args json.RawMessage) (ret interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("AnyCall panic", zap.Any("r", r))

			ret = nil
			err = errors.New("AnyCall panic")
		}
	}()

	switch method {
	default:
		return nil, errors.New("unsupported rpc method: " + method)
	}
}
