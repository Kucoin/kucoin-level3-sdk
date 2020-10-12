package kucoin_v2

import (
	"github.com/Kucoin/kucoin-level3-sdk/pkg/cfg"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/utils/helper"
	"go.uber.org/zap"
)

const (
	KucoinName       = "kucoin_v2"
	KucoinFutureName = "kumex_v2"
	PoloniexSwapName = "poloniex_swap_v2"
)

func init() {
	exchanges.RegisterType(KucoinName, setup(KucoinName))
	exchanges.RegisterType(KucoinFutureName, setup(KucoinFutureName))
	exchanges.RegisterType(PoloniexSwapName, setup(PoloniexSwapName))
}

func setup(name string) exchanges.Factory {
	return func() (exchanges.Exchange, error) {
		// parse config
		if err := cfg.AppConfig.Unpack(&defaultConfig); err != nil {
			log.Panic("Unpack panic", zap.Error(err))
		}
		log.Debug("market config for " + name + ":" + helper.ToJsonString(defaultConfig))

		ex := newExchange()

		return ex, nil
	}
}
