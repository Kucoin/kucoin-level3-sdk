package cfg

import (
	"github.com/go-playground/validator/v10"
)

type appConf struct {
	AppDebug bool `mapstructure:"app_debug"`

	Symbol string `mapstructure:"symbol" validate:"required"`

	App App `mapstructure:"app" validate:"required,dive"`

	ApiServer ApiServer `mapstructure:"api_server" validate:"required,dive"`

	Redis Redis `mapstructure:"redis"`

	Market map[string]interface{} `mapstructure:"market" validate:"required,len=1"`
}

type App struct {
	Name    string `mapstructure:"name" validate:"required"`
	LogFile string `mapstructure:"log_file"  validate:"required"`
}

type ApiServer struct {
	Network string `mapstructure:"network" validate:"required"`
	Address string `mapstructure:"address" validate:"required"`
	Token   string `mapstructure:"token" validate:"required"`
}

type Redis struct {
	Addr     string `mapstructure:"addr" validate:"required"`
	Password string `mapstructure:"password"`
	Db       int    `mapstructure:"db"`
}

func (cfg appConf) MarketName() string {
	for name := range cfg.Market {
		return name
	}

	panic("undefined market name")
}

func (cfg *appConf) Unpack(output interface{}) error {
	marketName := cfg.MarketName()
	marketCfg := cfg.Market[marketName]
	if err := mapStructureParse(marketCfg, output); err != nil {
		return err
	}

	validate := validator.New()
	return validate.Struct(output)
}
