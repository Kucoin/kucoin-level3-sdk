package app

import (
	"encoding/json"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/cfg"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"go.uber.org/zap"
)

type App struct {
	exchange exchanges.Exchange
}

func NewApp() *App {
	exchange, err := exchanges.Load(cfg.AppConfig.MarketName())
	if err != nil {
		log.Panic(err.Error(), zap.Error(err))
	}

	app := &App{
		exchange: exchange,
	}

	return app
}

func (app *App) PartOrderBook(number int) *exchanges.OrderBook {
	return app.exchange.GetPartOrderBook(number)
}

func (app *App) AddEventClientOidsToChannels(data map[string][]string) error {
	return app.exchange.AddEventClientOidsToChannels(data)
}

func (app *App) AnyCall(method string, args json.RawMessage) (interface{}, error) {
	return app.exchange.AnyCall(method, args)
}
