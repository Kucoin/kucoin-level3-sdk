package kucoin_v2

type Config struct {
	URL  string `mapstructure:"url" validate:"required"`
	Type string `mapstructure:"type" validate:"required"`
	//Verify           bool   `mapstructure:"verify"`
	//VerifyDir        string `mapstructure:"verify_dir" validate:"required_with=Verify"`
	Key        string `mapstructure:"key" validate:"required"`
	Secret     string `mapstructure:"secret" validate:"required"`
	Passphrase string `mapstructure:"passphrase" validate:"required"`
}

var defaultConfig = Config{}
