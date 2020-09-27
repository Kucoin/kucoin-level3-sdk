package kucoin_v2

type Config struct {
	URL  string `mapstructure:"url" validate:"required"`
	Type string `mapstructure:"type" validate:"required"`
	//Verify           bool   `mapstructure:"verify"`
	//VerifyDir        string `mapstructure:"verify_dir" validate:"required_with=Verify"`
}

var defaultConfig = Config{}
