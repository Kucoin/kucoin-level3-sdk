package cmd

import (
	"github.com/Kucoin/kucoin-level3-sdk/pkg/bootstrap"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start app",
	Long:  "Start the market application",
	Run: func(cmd *cobra.Command, args []string) {
		cfgFile, err := cmd.Flags().GetString("config")
		if err != nil {
			panic(err)
		}

		bootstrap.Run(cfgFile, cmd.Flags())
	},
}

func init() {
	startCmd.Flags().StringP("config", "c", "config.yaml", "app config file")
	startCmd.Flags().StringP("symbol", "s", "", "symbol")
}
