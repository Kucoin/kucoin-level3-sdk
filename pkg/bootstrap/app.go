package bootstrap

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/api"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/app"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/cfg"
	_ "github.com/Kucoin/kucoin-level3-sdk/pkg/includes"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/redis"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/utils/helper"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func Run(cfgFile string, flagSet *pflag.FlagSet) {
	//load cfg
	configFile, err := cfg.LoadConfig(cfgFile, flagSet, map[string]string{
		"symbol": "symbol",
	})
	if err != nil {
		panic(err)
	}

	//init logger
	log.New(cfg.AppConfig.AppDebug)
	defer log.Sync()

	log.Info("using cfg file: " + configFile)
	log.Debug("cfg data: " + helper.ToJsonString(cfg.AppConfig))

	log.Info("init redis connections")
	redis.InitConnections()

	decimal.MarshalJSONWithoutQuotes = true

	// websocket.DefaultDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	websocket.DefaultDialer.ReadBufferSize = 2048000 //2000 kb

	// run market
	marketApp := app.NewApp()

	// rpc server
	go api.InitRpcServer(marketApp)

	fmt.Println("market finished bootstrap")
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	//register for interupt (Ctrl+C) and SIGTERM (docker)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-quit
	log.Warn("app showdown!!!", zap.String("signal", sig.String()))
}
