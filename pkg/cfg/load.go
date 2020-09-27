package cfg

import (
	"os"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

var AppConfig = &appConf{}

var defaultMapStructureDecoderConfig = []viper.DecoderConfigOption{
	func(config *mapstructure.DecoderConfig) {
		config.TagName = "mapstructure"
		//config.TagName = "yaml"
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
				if f != reflect.String || t != reflect.String {
					return data, nil
				}
				return os.ExpandEnv(data.(string)), nil
			},
		)
	},
}

func mapStructureParse(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
	}
	for _, opt := range defaultMapStructureDecoderConfig {
		opt(config)
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(input); err != nil {
		return err
	}

	return nil
}

func unmarshalConfig(rawConfig map[string]interface{}, tmpInitCfg interface{}) error {
	_ = gotenv.OverLoad(".env")

	if err := mapStructureParse(rawConfig, tmpInitCfg); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(tmpInitCfg); err != nil {
		return err
	}

	return nil
}

func LoadConfig(cfgFile string, flagSet *pflag.FlagSet, keys map[string]string) (string, error) {
	v := viper.New()
	v.SetConfigFile(cfgFile)
	v.SetConfigType("yaml")
	for key, name := range keys {
		if err := v.BindPFlag(key, flagSet.Lookup(name)); err != nil {
			return "", err
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return "", err
	}

	err := unmarshalConfig(v.AllSettings(), AppConfig)
	if err != nil {
		panic(err)
	}
	return v.ConfigFileUsed(), nil
}
