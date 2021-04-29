package kucoin_v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/cfg"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/events"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/orderbook"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/sdk"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/verify"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"go.uber.org/zap"
)

type Exchange struct {
	exchanges.BasicExchange

	apiService *sdk.Kucoin
	ob         *orderbook.Builder
	ow         *events.OrderWatcher
	verify     *verify.Verify
}

func newExchange() *Exchange {
	apiService := sdk.NewKucoin(
		defaultConfig.URL,
		defaultConfig.Type,
		defaultConfig.Key,
		defaultConfig.Secret,
		defaultConfig.Passphrase,
		false,
		30*time.Second,
	)

	build := orderbook.NewBuilder(apiService, cfg.AppConfig.Symbol)
	var verifyObj *verify.Verify
	//if defaultConfig.Verify {
	//	verifyObj = verify.NewVerify(build, 20, defaultConfig.VerifyDir, cfg.AppConfig.Symbol)
	//}
	ex := &Exchange{
		apiService: apiService,
		ob:         build,
		ow:         events.NewOrderWatcher(),
		verify:     verifyObj,
	}

	//init ob
	go ex.ob.ReloadOrderBook()

	go ex.ow.Run()

	//if defaultConfig.Verify {
	//	go ex.verify.Run()
	//}

	go ex.websocket()

	return ex
}

func (ex *Exchange) websocket() {
	tk, err := ex.apiService.WebSocketPublicToken()
	if err != nil {
		log.Panic("WebSocketPublicToken err", zap.Error(err))
	}

	c := ex.apiService.NewWebSocketClient(tk)

	mc, ec, err := c.Connect()
	if err != nil {
		log.Panic("Connect panic: " + err.Error())
	}
	topic := sdk.L3TopicPrefix(defaultConfig.Type) + cfg.AppConfig.Symbol
	ch := sdk.NewSubscribeMessage(topic, false)
	log.Info("subscribe: " + topic)
	if err := c.Subscribe(ch); err != nil {
		log.Panic("Subscribe panic: "+err.Error(), zap.Error(err))
	}
	log.Info("Subscribe finish", zap.String("topic", topic))
	for {
		select {
		case err := <-ec:
			c.Stop() // Stop subscribing the WebSocket feed
			log.Panic("Connect panic: " + err.Error())

		case msg := <-mc:
			//log.Debug("receive message", zap.Any("data", msg))
			ex.dispatch(msg)
		}
	}
}

func (ex *Exchange) dispatch(msgRawData *sdk.WebSocketDownstreamMessage) {
	//log.Debug("raw message : " + base.ToJsonString(msgRawData))
	ex.ob.Messages <- msgRawData
	ex.ow.Messages <- msgRawData
	//if defaultConfig.Verify {
	//	ex.verify.Messages <- msgRawData
	//}
}

func (ex *Exchange) monitorChanLen() {
	for {
		const msgLenLimit = 50
		if len(ex.ob.Messages) > msgLenLimit {
			log.Info(fmt.Sprintf(
				"msgLenLimit: ex.ob.Messages: %d",
				len(ex.ob.Messages),
			))
		}
		time.Sleep(time.Second)
	}
}

func (ex Exchange) GetPartOrderBook(number int) *exchanges.OrderBook {
	return ex.ob.GetPartOrderBook(number)
}

func (ex Exchange) AddEventClientOidsToChannels(data map[string][]string) error {
	return ex.ow.AddEventClientOidsToChannels(data)
}

func (ex Exchange) AnyCall(method string, args json.RawMessage) (ret interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("AnyCall panic", zap.Any("r", r))

			ret = nil
			err = errors.New("AnyCall panic")
		}
	}()

	switch method {
	case "GetL3PartOrderBook":
		type AnyCallArgs struct {
			Number int `json:"number"`
		}
		var anyCallArgs AnyCallArgs
		if err := json.Unmarshal(args, &anyCallArgs); err != nil {
			return nil, errors.New("unmarshal AnyCallArgs error: " + string(args))
		}

		return ex.ob.GetL3PartOrderBook(anyCallArgs.Number), nil
	default:
		return nil, errors.New("unsupported rpc method: " + method)
	}
}
